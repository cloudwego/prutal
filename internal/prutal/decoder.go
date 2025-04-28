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
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"github.com/cloudwego/prutal/internal/desc"
	"github.com/cloudwego/prutal/internal/hack"
	"github.com/cloudwego/prutal/internal/protowire"
	"github.com/cloudwego/prutal/internal/wire"
)

var decoderPool = sync.Pool{
	New: func() interface{} {
		d := &Decoder{}
		d.s.init()
		return d
	},
}

const defaultMinMakeSliceCap = 8

type Decoder struct {
	s span
}

func (d *Decoder) Malloc(n, align int, abiType uintptr) unsafe.Pointer {
	if n > defaultDecoderSpanSize/4 || abiType != 0 {
		// too large, or it needs GC to scan (MallocAbiType != 0 of tType)
		return mallocgc(uintptr(n), unsafe.Pointer(abiType), abiType != 0)
	}
	return d.s.Malloc(n, align) // only for noscan objects like string.Data, []int etc...
}

func (d *Decoder) mallocIfPointer(p unsafe.Pointer, t *desc.Type) (ret unsafe.Pointer) {
	if !t.IsPointer {
		return p
	}

	// we need to malloc the type first before assigning a value to it
	// *p = new(type)
	t = t.V
	ret = d.Malloc(int(t.Size), t.Align, t.MallocAbiType)
	*(*unsafe.Pointer)(p) = ret
	return
}

func resetUnknownFields(s *desc.StructDesc, base unsafe.Pointer) {
	p := unsafe.Add(base, s.UnknownFieldsOffset)
	if s.UnknownFieldsPointer {
		p = *(*unsafe.Pointer)(p)
	}
	if p != nil {
		(*hack.SliceHeader)(p).Len = 0
	}
}

func appendToUnknownFields(s *desc.StructDesc, base unsafe.Pointer, b []byte) {
	p := unsafe.Add(base, s.UnknownFieldsOffset)
	var x *[]byte
	if s.UnknownFieldsPointer {
		if *(*unsafe.Pointer)(p) == nil {
			*(*unsafe.Pointer)(p) = unsafe.Pointer(&[]byte{})
		}
		x = (*[]byte)(*(*unsafe.Pointer)(p))
	} else {
		x = (*[]byte)(p)
	}
	*x = append(*x, b...)
}

func (d *Decoder) DecodeStruct(b []byte, base unsafe.Pointer, s *desc.StructDesc, maxdepth int) (int, error) {
	if maxdepth == 0 {
		return 0, errMaxDepthExceeded
	}
	i := 0

	var (
		f   *desc.FieldDesc  // cache last field, optmize for repeated
		tmv *desc.TmpMapVars // cache for map, optmize for repeated map
	)

	// reset unknownfields = unknownfields[:0]
	if s.HasUnknownFields {
		resetUnknownFields(s, base)
	}

	for i < len(b) {
		// next field tag
		num, typ, n := protowire.ConsumeTag(b[i:])
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		i += n

		// get field or skip
		if f == nil || f.ID != int32(num) {
			if tmv != nil {
				f.T.MapTmpVarsPool.Put(tmv)
				tmv = nil
			}
			f = s.GetField(int32(num))
			if f == nil {
				// field not found, skip bytes
				m := protowire.ConsumeFieldValue(num, typ, b[i:])
				if m < 0 {
					return i, protowire.ParseError(m)
				}
				i += m
				if s.HasUnknownFields {
					appendToUnknownFields(s, base, b[i-m-n:i])
				}
				continue
			}
		}

		p := unsafe.Add(base, f.Offset)
		t := f.T
		tag := f.TagType
		if f.IsOneof() {
			data := d.Malloc(int(t.Size), t.Align, t.MallocAbiType)
			hack.IfaceUpdate(p, f.IfaceTab, data)
			p = data
		}

		if t.IsPointer {
			p = d.mallocIfPointer(p, t)
			t = t.V
		}

		if f.Repeated && t.IsSlice {
			t = t.V
			if typ == wire.TypeBytes && f.Packed {
				// packed repeated fields, only scalar types except string or bytes
				if f.DecodeFunc == nil {
					panic(fmt.Sprintf("BUG? unknown packed field %q (#%d)", f.Name, f.ID))
				}
				packed, n := protowire.ConsumeBytes(b[i:])
				if n < 0 {
					return i, protowire.ParseError(n)
				}
				i += n
				if err := f.DecodeFunc(packed, p); err != nil {
					return i, err
				}
				continue
			}
			h := (*hack.SliceHeader)(p)
			if h.Cap == 0 {
				d.ReallocSlice(h, t, defaultMinMakeSliceCap)
			} else if h.Len == h.Cap {
				d.ReallocSlice(h, t, 2*h.Cap)
			}
			h.Len++
			p = unsafe.Add(h.Data, uintptr(h.Len-1)*t.Size) // p = &d[len(d-1)]
			if t.IsPointer {
				p = d.mallocIfPointer(p, t)
				t = t.V
			}
		}

		switch typ {
		case wire.TypeVarint: // case: VARINT
			if tag != desc.TypeVarint && tag != desc.TypeZigZag32 && tag != desc.TypeZigZag64 {
				return i, newWireTypeNotMatch(typ, tag)
			}

			u64, n := protowire.ConsumeVarint(b[i:])
			if n < 0 {
				return 0, protowire.ParseError(n)
			}
			if tag == desc.TypeZigZag32 || tag == desc.TypeZigZag64 {
				u64 = uint64(protowire.DecodeZigZag(u64))
			}
			switch t.Kind {
			case reflect.Int32:
				*(*int32)(p) = int32(u64)
			case reflect.Uint32:
				*(*uint32)(p) = uint32(u64)
			case reflect.Int64:
				*(*int64)(p) = int64(u64)
			case reflect.Uint64:
				*(*uint64)(p) = u64
			case reflect.Bool: // 1 for true, 0 for false
				*(*byte)(p) = byte(u64 & 0x1)
			}
			i += n

		case wire.TypeFixed32: // case: I32
			if tag != desc.TypeFixed32 {
				return i, newWireTypeNotMatch(typ, tag)
			}

			u32, n := protowire.ConsumeFixed32(b[i:])
			if n < 0 {
				return 0, protowire.ParseError(n)
			}
			switch t.Kind {
			case reflect.Int32:
				*(*int32)(p) = int32(u32)
			case reflect.Uint32, reflect.Float32:
				*(*uint32)(p) = u32
			}
			i += n

		case wire.TypeFixed64: // case: I64
			if tag != desc.TypeFixed64 {
				return i, newWireTypeNotMatch(typ, tag)
			}

			u64, n := protowire.ConsumeFixed64(b[i:])
			if n < 0 {
				return 0, protowire.ParseError(n)
			}
			switch t.Kind {
			case reflect.Int64:
				*(*int64)(p) = int64(u64)
			case reflect.Uint64, reflect.Float64:
				*(*uint64)(p) = u64
			}
			i += n

		case wire.TypeBytes: // case: LEN
			// string, bytes, embedded messages (struct or map), packed repeated fields
			fb, n := protowire.ConsumeBytes(b[i:])
			if n < 0 {
				return 0, protowire.ParseError(n)
			}
			i += n
			switch t.Kind {
			case desc.KindBytes:
				if len(fb) > 0 {
					data := d.Malloc(len(fb), 1, 0)
					*(*[]byte)(p) = unsafe.Slice((*byte)(data), len(fb))
					copy(*(*[]byte)(p), fb)
				} else {
					*(*[]byte)(p) = []byte{}
				}
			case reflect.String:
				if len(fb) > 0 {
					data := d.Malloc(len(fb), 1, 0)
					copy(unsafe.Slice((*byte)(data), len(fb)), fb)
					h := (*hack.StringHeader)(p)
					h.Data = data
					h.Len = len(fb)
				} else {
					*(*string)(p) = ""
				}
			case reflect.Map:
				if f.DecodeFunc != nil { // fast path for using docoders in wire pkg
					if err := f.DecodeFunc(fb, p); err != nil {
						return i, err
					}
				} else {
					if tmv == nil {
						tmv = f.T.MapTmpVarsPool.Get().(*desc.TmpMapVars)
					}
					if _, err := d.DecodeMapPair(fb, p, f, tmv, maxdepth-1); err != nil {
						return i, err
					}
				}
			case reflect.Struct:
				if _, err := d.DecodeStruct(fb, p, t.S, maxdepth-1); err != nil {
					return i, err
				}
			default:
				return i, newWireTypeNotMatch(typ, tag)
			}

		default:
			// unknown wiretype
			return i, newWireTypeNotMatch(typ, tag)
		}

	} // end of decoding field loop

	if tmv != nil { // use defer? if no performance issue
		f.T.MapTmpVarsPool.Put(tmv)
	}
	return i, nil
}

// DecodeMapKey ...
// keyType = "int32" | "int64" | "uint32" | "uint64" | "sint32" | "sint64" |
// "fixed32" | "fixed64" | "sfixed32" | "sfixed64" | "bool" | "string"
func (d *Decoder) DecodeMapKey(b []byte, p unsafe.Pointer, f *desc.FieldDesc) (int, error) {
	num, typ := wire.ConsumeKVTag(b)
	if num != 1 { // key num
		return 0, fmt.Errorf("key field num not match, got %d expect 1", num)
	}
	i := 1 // one byte for map key tag

	t := f.T.K
	tag := f.KeyType

	switch typ {
	case wire.TypeVarint: // case: VARINT
		if tag != desc.TypeVarint && tag != desc.TypeZigZag32 && tag != desc.TypeZigZag64 {
			return i, newWireTypeNotMatch(typ, tag)
		}

		u64, n := protowire.ConsumeVarint(b[i:])
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		if tag == desc.TypeZigZag32 || tag == desc.TypeZigZag64 {
			u64 = uint64(protowire.DecodeZigZag(u64))
		}
		switch t.Kind {
		case reflect.Int32:
			*(*int32)(p) = int32(u64)
		case reflect.Uint32:
			*(*uint32)(p) = uint32(u64)
		case reflect.Int64:
			*(*int64)(p) = int64(u64)
		case reflect.Uint64:
			*(*uint64)(p) = u64
		case reflect.Bool: // 1 for true, 0 for false
			*(*byte)(p) = byte(u64 & 0x1)
		}
		i += n

	case wire.TypeFixed32: // case: I32
		if tag != desc.TypeFixed32 {
			return i, newWireTypeNotMatch(typ, tag)
		}

		u32, n := protowire.ConsumeFixed32(b[i:])
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		switch t.Kind {
		case reflect.Int32:
			*(*int32)(p) = int32(u32)
		case reflect.Uint32, reflect.Float32:
			*(*uint32)(p) = u32
		}
		i += n

	case wire.TypeFixed64: // case: I64
		if tag != desc.TypeFixed64 {
			return i, newWireTypeNotMatch(typ, tag)
		}

		u64, n := protowire.ConsumeFixed64(b[i:])
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		switch t.Kind {
		case reflect.Int64:
			*(*int64)(p) = int64(u64)
		case reflect.Uint64, reflect.Float64:
			*(*uint64)(p) = u64
		}
		i += n

	case wire.TypeBytes: // case: LEN
		// for map, can only be string
		if t.Kind != reflect.String {
			return i, newWireTypeNotMatch(typ, tag)
		}
		fb, n := protowire.ConsumeBytes(b[i:])
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		i += n
		if len(fb) > 0 {
			data := d.Malloc(len(fb), 1, 0)
			copy(unsafe.Slice((*byte)(data), len(fb)), fb)
			h := (*hack.StringHeader)(p)
			h.Data = data
			h.Len = len(fb)
		} else {
			*(*string)(p) = ""
		}

	default:
		// unknown wiretype
		return i, newWireTypeNotMatch(typ, tag)
	}

	return i, nil
}

// DecodeMapValue ...
// like DecodeMapKey plus "bytes" | messageType | enumType
func (d *Decoder) DecodeMapValue(b []byte, p unsafe.Pointer, f *desc.FieldDesc, maxdepth int) (int, error) {
	num, typ := wire.ConsumeKVTag(b)
	if num != 2 { // val num
		return 0, fmt.Errorf("val field num not match, got %d expect 2", num)
	}
	i := 1 // one byte for map value tag

	t := f.T.V
	tag := f.ValType

	if t.IsPointer {
		p = d.mallocIfPointer(p, t)
		t = t.V
	}

	switch typ {
	case wire.TypeVarint: // case: VARINT
		if tag != desc.TypeVarint && tag != desc.TypeZigZag32 && tag != desc.TypeZigZag64 {
			return i, newWireTypeNotMatch(typ, tag)
		}

		u64, n := protowire.ConsumeVarint(b[i:])
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		if tag == desc.TypeZigZag32 || tag == desc.TypeZigZag64 {
			u64 = uint64(protowire.DecodeZigZag(u64))
		}
		switch t.Kind {
		case reflect.Int32:
			*(*int32)(p) = int32(u64)
		case reflect.Uint32:
			*(*uint32)(p) = uint32(u64)
		case reflect.Int64:
			*(*int64)(p) = int64(u64)
		case reflect.Uint64:
			*(*uint64)(p) = u64
		case reflect.Bool: // 1 for true, 0 for false
			*(*byte)(p) = byte(u64 & 0x1)
		}
		i += n

	case wire.TypeFixed32: // case: I32
		if tag != desc.TypeFixed32 {
			return i, newWireTypeNotMatch(typ, tag)
		}

		u32, n := protowire.ConsumeFixed32(b[i:])
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		switch t.Kind {
		case reflect.Int32:
			*(*int32)(p) = int32(u32)
		case reflect.Uint32, reflect.Float32:
			*(*uint32)(p) = u32
		}
		i += n

	case wire.TypeFixed64: // case: I64
		if tag != desc.TypeFixed64 {
			return i, newWireTypeNotMatch(typ, tag)
		}

		u64, n := protowire.ConsumeFixed64(b[i:])
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		switch t.Kind {
		case reflect.Int64:
			*(*int64)(p) = int64(u64)
		case reflect.Uint64, reflect.Float64:
			*(*uint64)(p) = u64
		}
		i += n

	case wire.TypeBytes: // case: LEN
		// string, bytes, embedded messages (struct or map), packed repeated fields
		fb, n := protowire.ConsumeBytes(b[i:])
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		i += n
		switch t.Kind {
		case desc.KindBytes:
			if len(fb) > 0 {
				data := d.Malloc(len(fb), 1, 0)
				*(*[]byte)(p) = unsafe.Slice((*byte)(data), len(fb))
				copy(*(*[]byte)(p), fb)
			} else {
				*(*[]byte)(p) = []byte{}
			}
		case reflect.String:
			if len(fb) > 0 {
				data := d.Malloc(len(fb), 1, 0)
				copy(unsafe.Slice((*byte)(data), len(fb)), fb)
				h := (*hack.StringHeader)(p)
				h.Data = data
				h.Len = len(fb)
			} else {
				*(*string)(p) = ""
			}
		case reflect.Struct:
			if _, err := d.DecodeStruct(fb, p, t.S, maxdepth-1); err != nil {
				return i, err
			}
		default:
			return i, newWireTypeNotMatch(typ, tag)
		}

	default:
		// unknown wiretype
		return i, newWireTypeNotMatch(typ, tag)
	}

	return i, nil
}

func (d *Decoder) DecodeMapPair(b []byte, p unsafe.Pointer, f *desc.FieldDesc, tmp *desc.TmpMapVars, maxdepth int) (int, error) {
	tmp.Reset()

	// maps are encoded exactly like a repeated message field:
	// as a sequence of LEN-typed records, with two fields each.
	// #1 for key and #2 for value.
	var m reflect.Value
	if *(*unsafe.Pointer)(p) == nil {
		m = reflect.MakeMap(f.T.T)
		*(*unsafe.Pointer)(p) = m.UnsafePointer()
	}

	i := 0
	n, err := d.DecodeMapKey(b, tmp.KeyPointer(), f)
	if err != nil {
		return 0, err
	}
	i += n

	n, err = d.DecodeMapValue(b[i:], tmp.ValPointer(), f, maxdepth)
	if err != nil {
		return i, err
	}
	i += n

	tmp.Update(tmp, p)
	return i, nil
}

func (d *Decoder) ReallocSlice(h *hack.SliceHeader, t *desc.Type, c int) {
	if h.Cap >= c {
		return
	}
	data := d.Malloc(c*int(t.Size), t.Align, t.MallocAbiType)
	if h.Len > 0 {
		copyn(data, h.Data, h.Len*int(t.Size))
	}
	h.Data = data
	h.Cap = c
}

// copyn copies n bytes from src to dst addr.
func copyn(dst, src unsafe.Pointer, n int) {
	copy(
		unsafe.Slice((*byte)(dst), n),
		unsafe.Slice((*byte)(src), n),
	)
}
