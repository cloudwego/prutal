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
	"github.com/cloudwego/prutal/internal/wire"
)

type Encoder struct{}

func (e *Encoder) AppendStruct(b []byte, base unsafe.Pointer, s *desc.StructDesc, encodeLen bool, maxdepth int) (_ []byte, err error) {
	if base == nil {
		base = s.Empty // empty struct
	}
	if maxdepth == 0 {
		return b, errMaxDepthExceeded
	}
	var beforeStructSize int
	if encodeLen {
		b = wire.LenReserve(b)
		beforeStructSize = len(b)
	}
	for _, f := range s.Fields {
		p := unsafe.Add(base, f.Offset)
		t := f.T

		if f.IsOneof() {
			typ, data := hack.ExtratIface(p)
			if data == nil {
				continue
			}
			if hack.ReflectTypePtr(f.OneofType) != typ {
				continue
			}
			p = data
		}

		// this also checks pointer types
		skipzero := false
		switch {
		case t.Size == 8: // int64, uint64, float64 or pointer on amd64
			skipzero = *(*uint64)(p) == 0
		case t.Size == 4: // int32, uint32, float32 or pointer on 386
			skipzero = *(*uint32)(p) == 0
		case t.Size == 1: // bool?
			skipzero = *(*uint8)(p) == 0
		case t.SliceLike: // for slice or string, both can use StringHeader
			skipzero = ((*hack.StringHeader)(p)).Len == 0
		}
		if skipzero {
			continue
		}

		if t.IsPointer {
			t = t.V
			p = *(*unsafe.Pointer)(p) // dereference
		}

		if !f.Repeated && f.AppendFunc != nil { // scalar types without `repeated`
			b = wire.AppendVarint(b, wire.EncodeTag(f.ID, f.WireType))
			b = f.AppendFunc(b, p)
			continue
		}

		if f.IsList {
			// fast path for packed list
			if f.Packed {
				b = wire.AppendVarint(b, wire.EncodeTag(f.ID, wire.TypeBytes))
				b = f.AppendFunc(b, p)
				continue
			}
			// fast path for repeated list except struct
			if f.AppendRepeated != nil {
				b = f.AppendRepeated(b, f.ID, p)
				continue
			}

			// pb doesn't support nested slice or map, can only be struct
			vt := t.V
			s := vt.S
			if vt.IsPointer {
				s = vt.V.S
			}
			if s == nil {
				panic("[BUG] not struct")
			}

			h := (*hack.SliceHeader)(p)
			p = h.Data
			for i := 0; i < h.Len; i++ {
				if i != 0 {
					p = unsafe.Add(p, vt.Size)
				}
				b = wire.AppendVarint(b, wire.EncodeTag(f.ID, f.WireType))
				base := p
				if vt.IsPointer {
					base = *(*unsafe.Pointer)(p)
				}
				b, err = e.AppendStruct(b, base, s, true, maxdepth-1)
				if err != nil {
					return b, err
				}
			}
			continue
		} // end of list field

		if f.IsMap {
			tmp := t.MapTmpVarsPool.Get().(*desc.TmpMapVars)
			m := tmp.MapWithPtr(p)
			vt := t.V
			it := hack.NewMapIter(m)
			for {
				kp, vp := it.Next()
				if kp == nil {
					break
				}
				// LEN for each map record
				b = wire.AppendVarint(b, wire.EncodeTag(f.ID, wire.TypeBytes))
				b = wire.LenReserve(b)
				beforesz := len(b)

				// Key
				b = wire.AppendVarint(b, wire.EncodeTag(1, f.KeyWireType))
				b = f.KeyAppendFunc(b, kp)

				// Val
				b = wire.AppendVarint(b, wire.EncodeTag(2, f.ValWireType))
				if f.ValAppendFunc != nil {
					b = f.ValAppendFunc(b, vp)
				} else {
					s := vt.S
					if vt.IsPointer { // likely it's a pointer for struct
						s = vt.V.S
						vp = *(*unsafe.Pointer)(vp)
					}
					b, err = e.AppendStruct(b, vp, s, true, maxdepth-1)
					if err != nil {
						return b, err
					}
				}

				b = wire.LenPack(b, len(b)-beforesz)
			}
			t.MapTmpVarsPool.Put(tmp)
			continue
		} // end of map field

		// case Struct
		if t.Kind != reflect.Struct {
			panic("[BUG] not struct") // assert reflect.Struct
		}

		b = wire.AppendVarint(b, wire.EncodeTag(f.ID, f.WireType))
		b, err = e.AppendStruct(b, p, t.S, true, maxdepth-1)
		if err != nil {
			return b, err
		}

	} // end of encoding field loop

	if s.HasUnknownFields {
		b = appendUnknownFields(b, s, base)
	}

	if encodeLen {
		b = wire.LenPack(b, len(b)-beforeStructSize)
	}
	return b, nil
}

func appendUnknownFields(b []byte, s *desc.StructDesc, base unsafe.Pointer) []byte {
	p := unsafe.Add(base, s.UnknownFieldsOffset)
	var x *[]byte
	if s.UnknownFieldsPointer {
		x = (*[]byte)(*(*unsafe.Pointer)(p))
	} else {
		x = (*[]byte)(p)
	}
	return append(b, (*x)...)
}
