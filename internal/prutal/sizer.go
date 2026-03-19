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

		// skip zero values — same logic as encoder
		skipzero := false
		switch {
		case t.Size == 8:
			skipzero = *(*uint64)(p) == 0
		case t.Size == 4:
			skipzero = *(*uint32)(p) == 0
		case t.Size == 1:
			skipzero = *(*uint8)(p) == 0
		case t.SliceLike:
			skipzero = ((*hack.StringHeader)(p)).Len == 0
		}
		if skipzero {
			continue
		}

		if t.IsPointer {
			t = t.V
			p = *(*unsafe.Pointer)(p)
		}

		// Scalar fields
		if !f.Repeated && f.SizeFunc != nil {
			n += f.WireTagSize + f.SizeFunc(p)
			continue
		}

		// List fields
		if f.IsList {
			sz, err := sizeListField(f, p, t, maxdepth)
			if err != nil {
				return 0, err
			}
			n += sz
			continue
		}

		// Map fields
		if f.IsMap {
			sz, err := sizeMapField(f, p, maxdepth)
			if err != nil {
				return 0, err
			}
			n += sz
			continue
		}

		// Struct fields
		if t.Kind != reflect.Struct {
			panic("[BUG] not struct")
		}
		sz, err := SizeStruct(p, t.S, maxdepth-1)
		if err != nil {
			return 0, err
		}
		n += f.WireTagSize + protowire.SizeBytes(sz)
	}

	if s.HasUnknownFields {
		n += sizeUnknownFields(s, base)
	}
	return n, nil
}

func sizeListField(f *desc.FieldDesc, p unsafe.Pointer, t *desc.Type, maxdepth int) (int, error) {
	vt := t.V
	h := (*hack.SliceHeader)(p)
	if h.Len == 0 {
		return 0, nil
	}

	if f.Packed {
		// packed: tag + SizeBytes(content)
		content := 0
		ep := h.Data
		for i := 0; i < h.Len; i++ {
			if i != 0 {
				ep = unsafe.Add(ep, vt.Size)
			}
			content += f.SizeFunc(ep)
		}
		return f.WireTagSize + protowire.SizeBytes(content), nil
	}

	if f.SizeFunc != nil {
		// scalar list: each element has tag + value
		n := 0
		ep := h.Data
		for i := 0; i < h.Len; i++ {
			if i != 0 {
				ep = unsafe.Add(ep, vt.Size)
			}
			n += f.WireTagSize + f.SizeFunc(ep)
		}
		return n, nil
	}

	// struct list
	s := vt.S
	if vt.IsPointer {
		s = vt.V.S
	}
	n := 0
	ep := h.Data
	for i := 0; i < h.Len; i++ {
		if i != 0 {
			ep = unsafe.Add(ep, vt.Size)
		}
		base := ep
		if vt.IsPointer {
			base = *(*unsafe.Pointer)(ep)
		}
		sz, err := SizeStruct(base, s, maxdepth-1)
		if err != nil {
			return 0, err
		}
		n += f.WireTagSize + protowire.SizeBytes(sz)
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
