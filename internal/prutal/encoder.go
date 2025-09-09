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
	"sync"
	"unsafe"

	"github.com/cloudwego/prutal/internal/desc"
	"github.com/cloudwego/prutal/internal/hack"
	"github.com/cloudwego/prutal/internal/wire"
)

var encoderPool = sync.Pool{
	New: func() interface{} {
		e := &Encoder{}
		e.p = unsafe.Pointer(&e.tmp[0])
		return e
	},
}

type Encoder struct {
	p   unsafe.Pointer // Always points to &tmp[0]
	tmp [32]byte       // Temporary buffer for encoding operations
}

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
			data := hack.IfaceData(p)
			if data == nil {
				continue
			}
			if hack.ReflectTypePtr(f.OneofType) != hack.IfaceTypePtr(p) {
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

		// Scalar fields
		if !f.Repeated && f.AppendFunc != nil {
			// scalar types without `repeated`
			b = wire.AppendVarintSmall(b, f.WireTag)
			b = f.AppendFunc(b, p)
			continue
		}

		// List fields
		if f.IsList {
			if f.Packed {
				// fast path for using funcs in wire package
				b = wire.AppendVarintSmall(b, f.WireTag)
				b = f.AppendFunc(b, p)
			} else if f.AppendRepeated != nil {
				// fast path for using funcs in wire package
				b = f.AppendRepeated(b, f.WireTag, p)
			} else {
				b, err = e.AppendListField(b, f, p, maxdepth)
				if err != nil {
					return b, err
				}
			}
			continue
		}

		// Map fields
		if f.IsMap {
			if f.AppendRepeated != nil {
				// Fast path: use pre-generated encoding functions from wire package
				// This is used when both key and value types have specialized encoders
				// (e.g., map[int32]string, map[string]int64, etc.)
				b = f.AppendRepeated(b, f.WireTag, p)
			} else {
				// Slow path: manual encoding when value type is a struct
				// AppendRepeated is nil only when the map value type is CoderStruct,
				// which happens when ValType is TypeBytes and the underlying type is reflect.Struct
				b, err = e.AppendMapField(b, f, p, maxdepth)
				if err != nil {
					return b, err
				}
			}
			continue
		}

		// Struct fields
		if t.Kind != reflect.Struct {
			panic("[BUG] not struct") // assert reflect.Struct
		}
		b = wire.AppendVarintSmall(b, f.WireTag)
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

func (e *Encoder) AppendListField(b []byte, f *desc.FieldDesc, p unsafe.Pointer, maxdepth int) (_ []byte, err error) {
	// pb doesn't support nested slice or map, can only be struct
	t := f.T
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
		b = wire.AppendVarintSmall(b, f.WireTag)
		base := p
		if vt.IsPointer {
			base = *(*unsafe.Pointer)(p)
		}
		b, err = e.AppendStruct(b, base, s, true, maxdepth-1)
		if err != nil {
			break
		}
	}
	return b, err
}

// AppendMapField encodes a map field where the value is a struct type.
// This function is only called when f.AppendRepeated is nil, which occurs when
// the map value cannot be encoded using the pre-generated functions in the wire package.
func (e *Encoder) AppendMapField(b []byte, f *desc.FieldDesc, p unsafe.Pointer, maxdepth int) (_ []byte, err error) {
	t := f.T
	kt := t.K
	vt := t.V

	// Optimization: Use specialized unsafe functions for common map[primitive]*struct patterns
	// These avoid reflection overhead for map iteration when:
	// 1. The value has no specialized append function (f.ValAppendFunc == nil), meaning it's a struct
	// 2. The value is a pointer type (vt.IsPointer)
	// This optimization is safe because:
	// - The field descriptor ensures the types match what protobuf expects
	// - We're only optimizing the map iteration, not the struct encoding itself
	if f.ValAppendFunc == nil && vt.IsPointer { // value is struct pointer
		switch kt.Kind {
		case reflect.Int64, reflect.Uint64:
			return e.appendMapField_u64_unsafe(b, f, p, maxdepth)
		case reflect.Int32, reflect.Uint32:
			return e.appendMapField_u32_unsafe(b, f, p, maxdepth)
		case reflect.String:
			return e.appendMapField_str_unsafe(b, f, p, maxdepth)
		}
	}
	tmp := t.MapTmpVarsPool.Get().(*desc.TmpMapVars)
	m := tmp.MapWithPtr(p)
	it := hack.NewMapIter(m)
	for {
		kp, vp := it.Next()
		if kp == nil {
			break
		}
		// LEN for each map record
		b = wire.AppendVarintSmall(b, f.WireTag)
		b = wire.LenReserve(b)
		beforesz := len(b)

		// Key
		b = wire.AppendKeyTag(b, f.KeyWireType)
		b = f.KeyAppendFunc(b, kp)

		// Val
		b = wire.AppendValTag(b, f.ValWireType)
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
				break
			}
		}
		b = wire.LenPack(b, len(b)-beforesz)
	}
	t.MapTmpVarsPool.Put(tmp)
	return b, err
}

func (e *Encoder) appendMapField_u64_unsafe(b []byte, f *desc.FieldDesc, p unsafe.Pointer, maxdepth int) (_ []byte, err error) {
	s := f.T.V.V.S
	for k, vp := range *(*map[uint64]unsafe.Pointer)(p) {
		*(*uint64)(e.p) = k

		// LEN for each map record
		b = wire.AppendVarintSmall(b, f.WireTag)
		b = wire.LenReserve(b)
		beforesz := len(b)

		// Key
		b = wire.AppendKeyTag(b, f.KeyWireType)
		b = f.KeyAppendFunc(b, e.p)

		// Val
		b = wire.AppendValTag(b, f.ValWireType)
		b, err = e.AppendStruct(b, vp, s, true, maxdepth-1)
		if err != nil {
			break
		}
		b = wire.LenPack(b, len(b)-beforesz)
	}
	return b, err
}

func (e *Encoder) appendMapField_u32_unsafe(b []byte, f *desc.FieldDesc, p unsafe.Pointer, maxdepth int) (_ []byte, err error) {
	s := f.T.V.V.S
	for k, vp := range *(*map[uint32]unsafe.Pointer)(p) {
		*(*uint32)(e.p) = k

		// LEN for each map record
		b = wire.AppendVarintSmall(b, f.WireTag)
		b = wire.LenReserve(b)
		beforesz := len(b)

		// Key
		b = wire.AppendKeyTag(b, f.KeyWireType)
		b = f.KeyAppendFunc(b, e.p)

		// Val
		b = wire.AppendValTag(b, f.ValWireType)
		b, err = e.AppendStruct(b, vp, s, true, maxdepth-1)
		if err != nil {
			break
		}
		b = wire.LenPack(b, len(b)-beforesz)
	}
	return b, err
}

func (e *Encoder) appendMapField_str_unsafe(b []byte, f *desc.FieldDesc, p unsafe.Pointer, maxdepth int) (_ []byte, err error) {
	t := f.T
	s := t.V.V.S
	for k, vp := range *(*map[string]unsafe.Pointer)(p) {
		// LEN for each map record
		b = wire.AppendVarintSmall(b, f.WireTag)
		b = wire.LenReserve(b)
		beforesz := len(b)

		// Key
		b = wire.AppendKeyTag(b, f.KeyWireType)
		b = wire.AppendVarint(b, uint64(len(k)))
		b = append(b, k...)

		// Val
		b = wire.AppendValTag(b, f.ValWireType)
		b, err = e.AppendStruct(b, vp, s, true, maxdepth-1)
		if err != nil {
			break
		}
		b = wire.LenPack(b, len(b)-beforesz)
	}
	return b, err
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
