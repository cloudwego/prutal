/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package compat

import (
	"math/rand"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// filler populates any message through protoreflect, so well-known types and
// generated messages of every API level are covered by the same code.
//
// Every singular field is set, lists and maps get 1 to 4 entries, and
// strings mix ASCII with multi-byte runes. Messages nest until depth runs
// out, which with the depth of 3 used by the tests takes a self-referencing
// message through a few levels without blowing up; the seeds 1 to 5 are
// arbitrary and merely make a failure reproducible.
type filler struct {
	r *rand.Rand
}

func newFiller(seed int64) *filler { return &filler{r: rand.New(rand.NewSource(seed))} }

func (f *filler) fill(m protoreflect.Message, depth int) {
	fds := m.Descriptor().Fields()
	seen := map[protoreflect.FullName]bool{}
	for i := 0; i < fds.Len(); i++ {
		fd := fds.Get(i)
		if od := fd.ContainingOneof(); od != nil && !od.IsSynthetic() {
			if seen[od.FullName()] {
				continue
			}
			seen[od.FullName()] = true
			members := od.Fields()
			fd = members.Get(f.r.Intn(members.Len()))
		}
		required := fd.Cardinality() == protoreflect.Required
		switch {
		case fd.IsList():
			l := m.Mutable(fd).List()
			for j, n := 0, 1+f.r.Intn(4); j < n; j++ {
				if fd.Message() != nil && depth <= 0 {
					break
				}
				l.Append(f.value(fd, l.NewElement(), depth-1))
			}
		case fd.IsMap():
			mp := m.Mutable(fd).Map()
			for j, n := 0, 1+f.r.Intn(4); j < n; j++ {
				if fd.MapValue().Message() != nil && depth <= 0 {
					break
				}
				k := f.value(fd.MapKey(), protoreflect.Value{}, depth).MapKey()
				mp.Set(k, f.value(fd.MapValue(), mp.NewValue(), depth-1))
			}
		case fd.Message() != nil:
			if depth <= 0 && !required {
				continue
			}
			f.fill(m.Mutable(fd).Message(), depth-1)
		default:
			m.Set(fd, f.value(fd, protoreflect.Value{}, depth))
		}
	}
}

func (f *filler) value(fd protoreflect.FieldDescriptor, v protoreflect.Value, depth int) protoreflect.Value {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return protoreflect.ValueOfBool(f.r.Intn(2) == 1)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return protoreflect.ValueOfInt32(int32(f.r.Uint32()))
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return protoreflect.ValueOfInt64(int64(f.r.Uint64()))
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return protoreflect.ValueOfUint32(f.r.Uint32())
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return protoreflect.ValueOfUint64(f.r.Uint64())
	case protoreflect.FloatKind:
		return protoreflect.ValueOfFloat32(f.r.Float32()*2e6 - 1e6)
	case protoreflect.DoubleKind:
		return protoreflect.ValueOfFloat64(f.r.Float64()*2e12 - 1e12)
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(f.str())
	case protoreflect.BytesKind:
		n := f.r.Intn(12)
		if n == 0 && !fd.HasPresence() && !fd.IsList() && !fd.ContainingMessage().IsMapEntry() &&
			fd.ParentFile().Syntax() == protoreflect.Editions {
			// an empty value of an editions implicit bytes field is the known
			// divergence, see TestKnownDivergenceEditionImplicitBytes
			n = 1
		}
		b := make([]byte, n)
		f.r.Read(b)
		return protoreflect.ValueOfBytes(b)
	case protoreflect.EnumKind:
		vals := fd.Enum().Values()
		return protoreflect.ValueOfEnum(vals.Get(f.r.Intn(vals.Len())).Number())
	case protoreflect.MessageKind, protoreflect.GroupKind:
		if depth >= 0 {
			f.fill(v.Message(), depth)
		}
		return v
	}
	panic("unhandled kind " + fd.Kind().String())
}

func (f *filler) str() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789 ,.-_/中文"
	rs := []rune(alphabet)
	n := f.r.Intn(16)
	out := make([]rune, n)
	for i := range out {
		out[i] = rs[f.r.Intn(len(rs))]
	}
	return string(out)
}
