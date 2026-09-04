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
		//nolint:govet // mallocgc expects the runtime type pointer stored in abiType
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

// DecodeStruct decodes b into the struct at base, merging into whatever is
// already there: repeated fields are appended to, message fields decode into
// the message they already point at, and unknown fields are appended. That is
// what protobuf requires of a message that occurs more than once, and it makes
// concatenated payloads decode the same as a merge. Callers that need a clean
// result must hand over a zero struct.
func (d *Decoder) DecodeStruct(b []byte, base unsafe.Pointer, s *desc.StructDesc, maxdepth int) (int, error) {
	if maxdepth == 0 {
		return 0, errMaxDepthExceeded
	}
	i := 0

	var (
		f   *desc.FieldDesc  // cache last field, optimize for repeated
		tmv *desc.TmpMapVars // cache for map, optimize for repeated map
	)

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
			// A field may occur more than once, and protobuf merges the
			// occurrences, so decode into the message allocated by the first
			// one instead of replacing it. Every pointer the decoder writes
			// into lives in zeroed memory -- a struct allocated by
			// mallocIfPointer, a oneof wrapper, a grown slice element -- so a
			// nil pointer reliably marks the first occurrence.
			if q := *(*unsafe.Pointer)(p); q != nil {
				p = q
			} else {
				p = d.mallocIfPointer(p, t)
			}
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
				// Fast path: use specialized decoders from wire package for primitive key-value pairs
				// These are pre-generated functions optimized for specific map types (e.g., map[uint32]int64)
				// Available for numeric/bool keys with numeric/bool values only
				if f.DecodeFunc != nil {
					if err := f.DecodeFunc(fb, p); err != nil {
						return i, err
					}
				} else {
					// Fallback: use generic reflection-based map decoding
					// Used for:
					// - String keys (most common case: map[string]T)
					// - Complex values: structs/messages (map[K]*MyMessage)
					// - Bytes values (map[K][]byte)
					// - String values when key is string (map[string]string)
					if tmv == nil {
						tmv = f.T.MapTmpVarsPool.Get().(*desc.TmpMapVars)
					}
					if _, err := d.DecodeMapPair(fb, p, f, tmv, maxdepth); err != nil {
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

// decodeMapKey decodes a map key into p from b, which starts right after the
// entry field tag. The caller has already matched the tag wire type against
// f.KeyWireType, so the key type alone decides how to read the value.
//
// keyType = "int32" | "int64" | "uint32" | "uint64" | "sint32" | "sint64" |
// "fixed32" | "fixed64" | "sfixed32" | "sfixed64" | "bool" | "string"
func (d *Decoder) decodeMapKey(b []byte, p unsafe.Pointer, f *desc.FieldDesc) (int, error) {
	t := f.T.K
	tag := f.KeyType

	switch tag {
	case desc.TypeVarint, desc.TypeZigZag32, desc.TypeZigZag64: // case: VARINT
		u64, n := protowire.ConsumeVarint(b)
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
		return n, nil

	case desc.TypeFixed32: // case: I32
		u32, n := protowire.ConsumeFixed32(b)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		switch t.Kind {
		case reflect.Int32:
			*(*int32)(p) = int32(u32)
		case reflect.Uint32:
			*(*uint32)(p) = u32
		}
		return n, nil

	case desc.TypeFixed64: // case: I64
		u64, n := protowire.ConsumeFixed64(b)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		switch t.Kind {
		case reflect.Int64:
			*(*int64)(p) = int64(u64)
		case reflect.Uint64:
			*(*uint64)(p) = u64
		}
		return n, nil

	case desc.TypeBytes: // case: LEN
		// p is sized for the key type, so writing a string header through it
		// is only safe once the key really is a string; desc guarantees that,
		// but the check keeps a descriptor regression from corrupting memory
		if t.Kind != reflect.String {
			return 0, newWireTypeNotMatch(f.KeyWireType, tag)
		}
		fb, n := protowire.ConsumeBytes(b)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		if len(fb) > 0 {
			data := d.Malloc(len(fb), 1, 0)
			copy(unsafe.Slice((*byte)(data), len(fb)), fb)
			h := (*hack.StringHeader)(p)
			h.Data = data
			h.Len = len(fb)
		} else {
			*(*string)(p) = ""
		}
		return n, nil
	}
	return 0, fmt.Errorf("unsupported map key type %s", tag)
}

// decodeMapValue decodes a map value into p from b, which starts right after
// the entry field tag. Like decodeMapKey, plus "bytes" | messageType |
// enumType. The caller has already matched the tag wire type against
// f.ValWireType and, for a message value, dereferenced p to the message it
// allocated for the entry.
func (d *Decoder) decodeMapValue(b []byte, p unsafe.Pointer, f *desc.FieldDesc, maxdepth int) (int, error) {
	t := f.T.V
	tag := f.ValType

	if t.IsPointer {
		t = t.V
	}

	switch tag {
	case desc.TypeVarint, desc.TypeZigZag32, desc.TypeZigZag64: // case: VARINT
		u64, n := protowire.ConsumeVarint(b)
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
		return n, nil

	case desc.TypeFixed32: // case: I32
		u32, n := protowire.ConsumeFixed32(b)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		switch t.Kind {
		case reflect.Int32:
			*(*int32)(p) = int32(u32)
		case reflect.Uint32, reflect.Float32:
			*(*uint32)(p) = u32
		}
		return n, nil

	case desc.TypeFixed64: // case: I64
		u64, n := protowire.ConsumeFixed64(b)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		switch t.Kind {
		case reflect.Int64:
			*(*int64)(p) = int64(u64)
		case reflect.Uint64, reflect.Float64:
			*(*uint64)(p) = u64
		}
		return n, nil

	case desc.TypeBytes: // case: LEN
		// string, bytes or embedded message
		fb, n := protowire.ConsumeBytes(b)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
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
				return n, err
			}
		default:
			return n, newWireTypeNotMatch(f.ValWireType, tag)
		}
		return n, nil
	}
	return 0, fmt.Errorf("unsupported map value type %s", tag)
}

// DecodeMapPair decodes one map entry of f into the map at p.
//
// Map entries are encoded exactly like a message with two fields: #1 for the
// key and #2 for the value. Both are optional on the wire, may come in either
// order and may repeat, and an entry may carry unknown fields, so it is
// scanned like a struct instead of being assumed to be in the canonical form
// that AppendMapField writes. A missing key or value contributes its zero
// value, and of repeated occurrences the last one wins.
func (d *Decoder) DecodeMapPair(b []byte, p unsafe.Pointer, f *desc.FieldDesc, tmp *desc.TmpMapVars, maxdepth int) (int, error) {
	tmp.Reset()

	var m reflect.Value
	if *(*unsafe.Pointer)(p) == nil {
		m = reflect.MakeMap(f.T.T)
		*(*unsafe.Pointer)(p) = m.UnsafePointer()
	}

	// tmp is pooled, so every entry must fully overwrite the key and value
	// vars. A message value is allocated here rather than on the first value
	// field: that both clears the pointer the previous entry left behind --
	// without it two keys would end up sharing one message -- and lets
	// repeated value fields merge into one message instead of replacing it.
	ptrVal := f.T.V.IsPointer
	vp := tmp.ValPointer()
	if ptrVal {
		vp = d.mallocIfPointer(vp, f.T.V)
	}

	// Entry tag bytes, see decodeMapEntry in internal/wire for why the
	// single-byte match is safe and why a miss still decodes the full tag.
	// Keep the two loops in sync.
	ktag := byte(wire.MapKeyFieldNum)<<3 | byte(f.KeyWireType)
	vtag := byte(wire.MapValFieldNum)<<3 | byte(f.ValWireType)

	i := 0
	hasKey, hasVal := false, false

	// Fast path: the canonical entry, key then value. The general scan below
	// carries on from wherever this stops, so anything after or instead of
	// the two fields is still handled.
	if len(b) > 0 && b[0] == ktag {
		n, err := d.decodeMapKey(b[1:], tmp.KeyPointer(), f)
		if err != nil {
			return 1, err
		}
		i = 1 + n
		hasKey = true
		if i < len(b) && b[i] == vtag {
			n, err = d.decodeMapValue(b[i+1:], vp, f, maxdepth)
			if err != nil {
				return i, err
			}
			i += 1 + n
			hasVal = true
		}
	}

	for i < len(b) {
		var n int
		var err error
		switch b[i] {
		case ktag:
			n, err = d.decodeMapKey(b[i+1:], tmp.KeyPointer(), f)
			n++ // entry tag byte
			hasKey = true

		case vtag:
			n, err = d.decodeMapValue(b[i+1:], vp, f, maxdepth)
			n++ // entry tag byte
			hasVal = true

		default:
			num, typ, m, terr := wire.ConsumeMapEntryTag(b[i:])
			if terr != nil {
				return i, terr
			}
			switch {
			case num == wire.MapKeyFieldNum && typ == f.KeyWireType:
				n, err = d.decodeMapKey(b[i+m:], tmp.KeyPointer(), f)
				hasKey = true
			case num == wire.MapValFieldNum && typ == f.ValWireType:
				n, err = d.decodeMapValue(b[i+m:], vp, f, maxdepth)
				hasVal = true
			default:
				// an unknown field, or a key or value carrying an unexpected
				// wire type: the reference implementation skips both the same way
				if n = protowire.ConsumeFieldValue(num, typ, b[i+m:]); n < 0 {
					err = protowire.ParseError(n)
				}
			}
			n += m
		}
		if err != nil {
			return i, err
		}
		i += n
	}

	if !hasKey {
		tmp.ZeroKey()
	}
	if !hasVal && !ptrVal { // a message value was already allocated above
		tmp.ZeroVal()
	}
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
