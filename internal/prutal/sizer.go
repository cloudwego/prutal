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

package prutal

import (
	"reflect"
	"unsafe"

	"github.com/cloudwego/prutal/internal/desc"
	"github.com/cloudwego/prutal/internal/hack"
	"github.com/cloudwego/prutal/internal/protowire"
)

func SizeStruct(base unsafe.Pointer, s *desc.StructDesc, maxdepth int) (int, error) {
	if maxdepth == 0 {
		return 0, errMaxDepthExceeded
	}
	if base == nil {
		base = s.Empty
	}
	n := 0
	for _, f := range s.Fields {
		p := unsafe.Add(base, f.Offset)
		t := f.T

		if f.IsOneof() {
			data := hack.IfaceData(p)
			if data == nil {
				continue
			}
			if hack.ReflectTypePtr(f.OneofType) != hack.IfaceTypePtr(p) {
				continue
			}
			p = data
		}

		// skip zero values, same logic and order as the encoder
		if k := f.ZeroKind; k == desc.ZeroKindU64 {
			if *(*uint64)(p) == 0 {
				continue
			}
		} else if k == desc.ZeroKindU32 {
			if *(*uint32)(p) == 0 {
				continue
			}
		} else if k == desc.ZeroKindLen { // slice or string, both can use StringHeader
			if ((*hack.StringHeader)(p)).Len == 0 {
				continue
			}
		} else if k == desc.ZeroKindU8 {
			if *(*uint8)(p) == 0 {
				continue
			}
		}

		if t.IsPointer {
			t = t.V
			p = *(*unsafe.Pointer)(p)
		}

		// Non-repeated scalar and struct fields
		if !f.Repeated {
			if f.SizeFunc != nil {
				n += f.WireTagSize + f.SizeFunc(p)
				continue
			}
			if t.Kind != reflect.Struct {
				panic("[BUG] not struct")
			}
			sz, err := SizeStruct(p, t.S, maxdepth-1)
			if err != nil {
				return 0, err
			}
			n += f.WireTagSize + protowire.SizeBytes(sz)
			continue
		}

		// Repeated scalar list or map fast path
		if f.SizeRepeated != nil {
			n += f.SizeRepeated(p, f.WireTagSize)
			continue
		}

		// List fields (struct lists)
		if f.IsList {
			sz, err := sizeStructListField(f, p, t, maxdepth)
			if err != nil {
				return 0, err
			}
			n += sz
			continue
		}

		// Map fields (struct-valued maps)
		if f.IsMap {
			sz, err := sizeMapField(f, p, maxdepth)
			if err != nil {
				return 0, err
			}
			n += sz
			continue
		}
	}

	if s.HasUnknownFields {
		n += sizeUnknownFields(s, base)
	}
	return n, nil
}

func sizeStructListField(f *desc.FieldDesc, p unsafe.Pointer, t *desc.Type, maxdepth int) (int, error) {
	vt := t.V
	if vt.IsPointer {
		s := vt.V.S
		vv := *(*[]unsafe.Pointer)(p)
		n := len(vv) * f.WireTagSize
		for _, base := range vv {
			sz, err := SizeStruct(base, s, maxdepth-1)
			if err != nil {
				return 0, err
			}
			n += protowire.SizeBytes(sz)
		}
		return n, nil
	}

	s := vt.S
	h := (*hack.SliceHeader)(p)
	if h.Len == 0 {
		return 0, nil
	}
	n := h.Len * f.WireTagSize
	ep := h.Data
	for i := 0; i < h.Len; i++ {
		if i != 0 {
			ep = unsafe.Add(ep, vt.Size)
		}
		sz, err := SizeStruct(ep, s, maxdepth-1)
		if err != nil {
			return 0, err
		}
		n += protowire.SizeBytes(sz)
	}
	return n, nil
}

func sizeMapField(f *desc.FieldDesc, p unsafe.Pointer, maxdepth int) (int, error) {
	t := f.T
	vt := t.V

	// fast path: typed iteration for map[primitive]*struct patterns
	if f.ValSizeFunc == nil && vt.IsPointer {
		switch t.K.Kind {
		case reflect.Int64, reflect.Uint64:
			return sizeMapField_u64(f, p, maxdepth)
		case reflect.Int32, reflect.Uint32:
			return sizeMapField_u32(f, p, maxdepth)
		case reflect.String:
			return sizeMapField_str(f, p, maxdepth)
		}
	}

	tmp := t.MapTmpVarsPool.Get().(*desc.TmpMapVars)
	m := tmp.MapWithPtr(p)
	it := hack.NewMapIter(m)
	n := 0
	var retErr error
	for {
		kp, vp := it.Next()
		if kp == nil {
			break
		}
		entrySz := 1 + f.KeySizeFunc(kp) // key tag (1 byte) + key value
		if f.ValSizeFunc != nil {
			entrySz += 1 + f.ValSizeFunc(vp) // val tag (1 byte) + val value
		} else {
			s := vt.S
			if vt.IsPointer {
				s = vt.V.S
				vp = *(*unsafe.Pointer)(vp)
			}
			sz, err := SizeStruct(vp, s, maxdepth-1)
			if err != nil {
				retErr = err
				break
			}
			entrySz += 1 + protowire.SizeBytes(sz) // val tag (1 byte) + len-prefixed struct
		}
		n += f.WireTagSize + protowire.SizeBytes(entrySz)
	}
	t.MapTmpVarsPool.Put(tmp)
	return n, retErr
}

func sizeMapField_u64(f *desc.FieldDesc, p unsafe.Pointer, maxdepth int) (int, error) {
	s := f.T.V.V.S
	keyType := f.KeyType
	n := 0
	for k, vp := range *(*map[uint64]unsafe.Pointer)(p) {
		keySz := 1 + sizeKeyU64(keyType, k) // key tag (1 byte) + key value
		valSz, err := SizeStruct(vp, s, maxdepth-1)
		if err != nil {
			return 0, err
		}
		entrySz := keySz + 1 + protowire.SizeBytes(valSz)
		n += f.WireTagSize + protowire.SizeBytes(entrySz)
	}
	return n, nil
}

func sizeMapField_u32(f *desc.FieldDesc, p unsafe.Pointer, maxdepth int) (int, error) {
	s := f.T.V.V.S
	keyType := f.KeyType
	signed := f.T.K.Kind == reflect.Int32
	n := 0
	for k, vp := range *(*map[uint32]unsafe.Pointer)(p) {
		keySz := 1 + sizeKeyU32(keyType, k, signed) // key tag (1 byte) + key value
		valSz, err := SizeStruct(vp, s, maxdepth-1)
		if err != nil {
			return 0, err
		}
		entrySz := keySz + 1 + protowire.SizeBytes(valSz)
		n += f.WireTagSize + protowire.SizeBytes(entrySz)
	}
	return n, nil
}

func sizeMapField_str(f *desc.FieldDesc, p unsafe.Pointer, maxdepth int) (int, error) {
	s := f.T.V.V.S
	n := 0
	for k, vp := range *(*map[string]unsafe.Pointer)(p) {
		keySz := 1 + protowire.SizeBytes(len(k)) // key tag (1 byte) + key value
		valSz, err := SizeStruct(vp, s, maxdepth-1)
		if err != nil {
			return 0, err
		}
		entrySz := keySz + 1 + protowire.SizeBytes(valSz)
		n += f.WireTagSize + protowire.SizeBytes(entrySz)
	}
	return n, nil
}

func sizeKeyU64(t desc.TagType, k uint64) int {
	switch t {
	case desc.TypeFixed64:
		return 8
	case desc.TypeZigZag64:
		return protowire.SizeVarint(k<<1 ^ uint64(int64(k)>>63))
	default: // TypeVarint
		return protowire.SizeVarint(k)
	}
}

func sizeKeyU32(t desc.TagType, k uint32, signed bool) int {
	switch t {
	case desc.TypeFixed32:
		return 4
	case desc.TypeZigZag32:
		return protowire.SizeVarint(uint64(k<<1 ^ uint32(int32(k)>>31)))
	default: // TypeVarint
		if signed {
			// int32 varint: sign-extend to 64-bit per protobuf spec
			return protowire.SizeVarint(uint64(int64(int32(k))))
		}
		return protowire.SizeVarint(uint64(k))
	}
}

func sizeUnknownFields(s *desc.StructDesc, base unsafe.Pointer) int {
	p := unsafe.Add(base, s.UnknownFieldsOffset)
	var x *[]byte
	if s.UnknownFieldsPointer {
		pp := *(*unsafe.Pointer)(p)
		if pp == nil {
			return 0
		}
		x = (*[]byte)(pp)
	} else {
		x = (*[]byte)(p)
	}
	return len(*x)
}
