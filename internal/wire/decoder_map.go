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

package wire

import (
	"unsafe"

	"github.com/cloudwego/prutal/internal/protowire"
)

func ensureMapNotNil[K comparable, V any](mp unsafe.Pointer) *map[K]V {
	m := (*map[K]V)(mp)
	if *m == nil {
		*m = make(map[K]V)
	}
	return m
}

func init() {
	register := func(k, v CoderType, f DecodeFunc) {
		mapDecoderFuncs[mapDecoderFuncKey{K: k, V: v}] = f
	}
	register(CoderVarintU32, CoderVarintU32, DecodeMap_VarintU32_VarintU32)
	register(CoderVarintU32, CoderVarint64, DecodeMap_VarintU32_VarintU64)
	register(CoderVarintU32, CoderZigZag32, DecodeMap_VarintU32_ZigZag32)
	register(CoderVarintU32, CoderZigZag64, DecodeMap_VarintU32_ZigZag64)
	register(CoderVarintU32, CoderFixed32, DecodeMap_VarintU32_Fixed32)
	register(CoderVarintU32, CoderFixed64, DecodeMap_VarintU32_Fixed64)
	register(CoderVarintU32, CoderBool, DecodeMap_VarintU32_Bool)

	// CoderVarintI32 uses the same decoders as CoderVarintU32: varint decode truncates to int32
	register(CoderVarintI32, CoderVarintI32, DecodeMap_VarintU32_VarintU32)
	register(CoderVarintI32, CoderVarintU32, DecodeMap_VarintU32_VarintU32)
	register(CoderVarintI32, CoderVarint64, DecodeMap_VarintU32_VarintU64)
	register(CoderVarintI32, CoderZigZag32, DecodeMap_VarintU32_ZigZag32)
	register(CoderVarintI32, CoderZigZag64, DecodeMap_VarintU32_ZigZag64)
	register(CoderVarintI32, CoderFixed32, DecodeMap_VarintU32_Fixed32)
	register(CoderVarintI32, CoderFixed64, DecodeMap_VarintU32_Fixed64)
	register(CoderVarintI32, CoderBool, DecodeMap_VarintU32_Bool)

	register(CoderVarint64, CoderVarintI32, DecodeMap_VarintU64_VarintU32)
	register(CoderVarint64, CoderVarintU32, DecodeMap_VarintU64_VarintU32)
	register(CoderVarint64, CoderVarint64, DecodeMap_VarintU64_VarintU64)
	register(CoderVarint64, CoderZigZag32, DecodeMap_VarintU64_ZigZag32)
	register(CoderVarint64, CoderZigZag64, DecodeMap_VarintU64_ZigZag64)
	register(CoderVarint64, CoderFixed32, DecodeMap_VarintU64_Fixed32)
	register(CoderVarint64, CoderFixed64, DecodeMap_VarintU64_Fixed64)
	register(CoderVarint64, CoderBool, DecodeMap_VarintU64_Bool)

	register(CoderZigZag32, CoderVarintI32, DecodeMap_ZigZag32_VarintU32)
	register(CoderZigZag32, CoderVarintU32, DecodeMap_ZigZag32_VarintU32)
	register(CoderZigZag32, CoderVarint64, DecodeMap_ZigZag32_VarintU64)
	register(CoderZigZag32, CoderZigZag32, DecodeMap_ZigZag32_ZigZag32)
	register(CoderZigZag32, CoderZigZag64, DecodeMap_ZigZag32_ZigZag64)
	register(CoderZigZag32, CoderFixed32, DecodeMap_ZigZag32_Fixed32)
	register(CoderZigZag32, CoderFixed64, DecodeMap_ZigZag32_Fixed64)
	register(CoderZigZag32, CoderBool, DecodeMap_ZigZag32_Bool)

	register(CoderZigZag64, CoderVarintI32, DecodeMap_ZigZag64_VarintU32)
	register(CoderZigZag64, CoderVarintU32, DecodeMap_ZigZag64_VarintU32)
	register(CoderZigZag64, CoderVarint64, DecodeMap_ZigZag64_VarintU64)
	register(CoderZigZag64, CoderZigZag32, DecodeMap_ZigZag64_ZigZag32)
	register(CoderZigZag64, CoderZigZag64, DecodeMap_ZigZag64_ZigZag64)
	register(CoderZigZag64, CoderFixed32, DecodeMap_ZigZag64_Fixed32)
	register(CoderZigZag64, CoderFixed64, DecodeMap_ZigZag64_Fixed64)
	register(CoderZigZag64, CoderBool, DecodeMap_ZigZag64_Bool)

	register(CoderFixed32, CoderVarintI32, DecodeMap_Fixed32_VarintU32)
	register(CoderFixed32, CoderVarintU32, DecodeMap_Fixed32_VarintU32)
	register(CoderFixed32, CoderVarint64, DecodeMap_Fixed32_VarintU64)
	register(CoderFixed32, CoderZigZag32, DecodeMap_Fixed32_ZigZag32)
	register(CoderFixed32, CoderZigZag64, DecodeMap_Fixed32_ZigZag64)
	register(CoderFixed32, CoderFixed32, DecodeMap_Fixed32_Fixed32)
	register(CoderFixed32, CoderFixed64, DecodeMap_Fixed32_Fixed64)
	register(CoderFixed32, CoderBool, DecodeMap_Fixed32_Bool)

	register(CoderFixed64, CoderVarintI32, DecodeMap_Fixed64_VarintU32)
	register(CoderFixed64, CoderVarintU32, DecodeMap_Fixed64_VarintU32)
	register(CoderFixed64, CoderVarint64, DecodeMap_Fixed64_VarintU64)
	register(CoderFixed64, CoderZigZag32, DecodeMap_Fixed64_ZigZag32)
	register(CoderFixed64, CoderZigZag64, DecodeMap_Fixed64_ZigZag64)
	register(CoderFixed64, CoderFixed32, DecodeMap_Fixed64_Fixed32)
	register(CoderFixed64, CoderFixed64, DecodeMap_Fixed64_Fixed64)
	register(CoderFixed64, CoderBool, DecodeMap_Fixed64_Bool)

	register(CoderBool, CoderVarintU32, DecodeMap_Bool_VarintU32)
	register(CoderBool, CoderVarint64, DecodeMap_Bool_VarintU64)
	register(CoderBool, CoderZigZag32, DecodeMap_Bool_ZigZag32)
	register(CoderBool, CoderZigZag64, DecodeMap_Bool_ZigZag64)
	register(CoderBool, CoderFixed32, DecodeMap_Bool_Fixed32)
	register(CoderBool, CoderFixed64, DecodeMap_Bool_Fixed64)
	register(CoderBool, CoderBool, DecodeMap_Bool_Bool)
}

// decodeMapEntry decodes one map entry whose key and value are both scalar
// (VARINT, I32 or I64) types and returns their raw wire values.
//
// It parses whatever the wire format allows instead of only the canonical form
// produced by encoder_map.go: entry fields may come in any order, either of
// them may be missing (its zero value applies), the last occurrence of a
// repeated field wins, and unknown fields are skipped. A key or value tagged
// with an unexpected wire type is skipped as unknown as well, matching the
// reference implementation.
//
// The reflection decoder in internal/prutal scans entries the same way for the
// key and value types that have no specialized decoder here; keep the two in
// sync.
func decodeMapEntry(b []byte, ktyp, vtyp Type) (k, v uint64, err error) {
	// A canonical entry tag fits in one byte: field #1 or #2 with a scalar
	// wire type encodes to 0x15 at most, which can never start a longer
	// varint. Over-long tag varints are legal too, so a byte matching neither
	// falls back to a full tag decode instead of being written off as another
	// field.
	ktag := byte(MapKeyFieldNum)<<3 | byte(ktyp)
	vtag := byte(MapValFieldNum)<<3 | byte(vtyp)

	// Fast path: the canonical entry every known encoder writes, key then
	// value, both present, nothing else. This is the hot path of all the
	// specialized map decoders, so it runs straight through, calling out only
	// for a varint wider than two bytes; whatever does not fit is left to the
	// general scan below, which starts over from the first byte.
	//
	// The scalar is dispatched on its wire type right here rather than in a
	// helper: a helper covering all three types is too big to be inlined, and
	// a call per entry field costs more than the decoding itself.
	if len(b) > 0 && b[0] == ktag {
		var kn, vn int
		switch ktyp {
		case TypeFixed32:
			var u32 uint32
			u32, kn = protowire.ConsumeFixed32(b[1:])
			k = uint64(u32)
		case TypeFixed64:
			k, kn = protowire.ConsumeFixed64(b[1:])
		default: // TypeVarint
			if k, kn = consumeVarintFast(b[1:]); kn == 0 {
				k, kn = protowire.ConsumeVarint(b[1:])
			}
		}
		if i := 1 + kn; kn > 0 && i < len(b) && b[i] == vtag {
			switch vtyp {
			case TypeFixed32:
				var u32 uint32
				u32, vn = protowire.ConsumeFixed32(b[i+1:])
				v = uint64(u32)
			case TypeFixed64:
				v, vn = protowire.ConsumeFixed64(b[i+1:])
			default: // TypeVarint
				if v, vn = consumeVarintFast(b[i+1:]); vn == 0 {
					v, vn = protowire.ConsumeVarint(b[i+1:])
				}
			}
			if vn > 0 && i+1+vn == len(b) {
				return k, v, nil
			}
		}
		k, v = 0, 0
	}

	for i := 0; i < len(b); {
		var n int
		switch b[i] {
		case ktag:
			k, n, err = consumeMapScalar(b[i+1:], ktyp)
			n++ // entry tag byte

		case vtag:
			v, n, err = consumeMapScalar(b[i+1:], vtyp)
			n++ // entry tag byte

		default:
			num, typ, m, terr := ConsumeMapEntryTag(b[i:])
			if terr != nil {
				return 0, 0, terr
			}
			switch {
			case num == MapKeyFieldNum && typ == ktyp:
				k, n, err = consumeMapScalar(b[i+m:], ktyp)
			case num == MapValFieldNum && typ == vtyp:
				v, n, err = consumeMapScalar(b[i+m:], vtyp)
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
			return 0, 0, err
		}
		i += n
	}
	return k, v, nil
}

// consumeVarintFast decodes a varint of at most two bytes, which is what map
// keys and values usually are. It reports n == 0 for anything else, truncated
// input included, leaving the caller to take the general path. Small enough to
// be inlined; keep it so.
func consumeVarintFast(b []byte) (v uint64, n int) {
	if len(b) > 0 && b[0] < 0x80 {
		return uint64(b[0]), 1
	}
	if len(b) > 1 && b[1] < 0x80 {
		return uint64(b[0]&0x7f) | uint64(b[1])<<7, 2
	}
	return 0, 0
}

// consumeMapScalar decodes one scalar map key or value. typ has already been
// matched against the expected key or value type by the caller, so it is
// always one of the three scalar wire types.
func consumeMapScalar(b []byte, typ Type) (v uint64, n int, err error) {
	switch typ {
	case TypeFixed32:
		var u32 uint32
		u32, n = protowire.ConsumeFixed32(b)
		v = uint64(u32)
	case TypeFixed64:
		v, n = protowire.ConsumeFixed64(b)
	default: // TypeVarint
		v, n = protowire.ConsumeVarint(b)
	}
	if n < 0 {
		return 0, 0, protowire.ParseError(n)
	}
	return v, n, nil
}

func DecodeMap_VarintU32_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint32](mp)
	(*m)[uint32(k)] = uint32(v)
	return nil
}

func DecodeMap_VarintU32_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, int64](mp)
	(*m)[uint32(k)] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_VarintU32_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint64](mp)
	(*m)[uint32(k)] = v
	return nil
}

func DecodeMap_VarintU32_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, int32](mp)
	(*m)[uint32(k)] = DecodeZigZag32(v)
	return nil
}

func DecodeMap_VarintU32_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeFixed64)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint64](mp)
	(*m)[uint32(k)] = v
	return nil
}

func DecodeMap_VarintU32_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeFixed32)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint32](mp)
	(*m)[uint32(k)] = uint32(v)
	return nil
}

func DecodeMap_VarintU32_Bool(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, bool](mp)
	(*m)[uint32(k)] = v != 0
	return nil
}

func DecodeMap_VarintU64_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint64](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_VarintU64_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint32](mp)
	(*m)[k] = uint32(v)
	return nil
}

func DecodeMap_VarintU64_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, int64](mp)
	(*m)[k] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_VarintU64_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, int32](mp)
	(*m)[k] = DecodeZigZag32(v)
	return nil
}

func DecodeMap_VarintU64_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeFixed64)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint64](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_VarintU64_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeFixed32)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint32](mp)
	(*m)[k] = uint32(v)
	return nil
}

func DecodeMap_VarintU64_Bool(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, bool](mp)
	(*m)[k] = v != 0
	return nil
}

func DecodeMap_ZigZag64_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, uint64](mp)
	(*m)[protowire.DecodeZigZag(k)] = v
	return nil
}

func DecodeMap_ZigZag64_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, uint32](mp)
	(*m)[protowire.DecodeZigZag(k)] = uint32(v)
	return nil
}

func DecodeMap_ZigZag64_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, int64](mp)
	(*m)[protowire.DecodeZigZag(k)] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_ZigZag64_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, int32](mp)
	(*m)[protowire.DecodeZigZag(k)] = DecodeZigZag32(v)
	return nil
}

func DecodeMap_ZigZag64_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeFixed64)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, uint64](mp)
	(*m)[protowire.DecodeZigZag(k)] = v
	return nil
}

func DecodeMap_ZigZag64_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeFixed32)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, uint32](mp)
	(*m)[protowire.DecodeZigZag(k)] = uint32(v)
	return nil
}

func DecodeMap_ZigZag64_Bool(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, bool](mp)
	(*m)[protowire.DecodeZigZag(k)] = v != 0
	return nil
}

func DecodeMap_ZigZag32_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, uint64](mp)
	(*m)[DecodeZigZag32(k)] = v
	return nil
}

func DecodeMap_ZigZag32_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, uint32](mp)
	(*m)[DecodeZigZag32(k)] = uint32(v)
	return nil
}

func DecodeMap_ZigZag32_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, int64](mp)
	(*m)[DecodeZigZag32(k)] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_ZigZag32_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, int32](mp)
	(*m)[DecodeZigZag32(k)] = DecodeZigZag32(v)
	return nil
}

func DecodeMap_ZigZag32_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeFixed64)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, uint64](mp)
	(*m)[DecodeZigZag32(k)] = v
	return nil
}

func DecodeMap_ZigZag32_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeFixed32)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, uint32](mp)
	(*m)[DecodeZigZag32(k)] = uint32(v)
	return nil
}

func DecodeMap_ZigZag32_Bool(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, bool](mp)
	(*m)[DecodeZigZag32(k)] = v != 0
	return nil
}

func DecodeMap_Fixed64_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed64, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint64](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_Fixed64_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed64, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint32](mp)
	(*m)[k] = uint32(v)
	return nil
}

func DecodeMap_Fixed64_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed64, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, int64](mp)
	(*m)[k] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_Fixed64_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed64, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, int32](mp)
	(*m)[k] = DecodeZigZag32(v)
	return nil
}

func DecodeMap_Fixed64_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed64, TypeFixed64)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint64](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_Fixed64_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed64, TypeFixed32)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint32](mp)
	(*m)[k] = uint32(v)
	return nil
}

func DecodeMap_Fixed64_Bool(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed64, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, bool](mp)
	(*m)[k] = v != 0
	return nil
}

// Fixed32 key decoders
func DecodeMap_Fixed32_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed32, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint64](mp)
	(*m)[uint32(k)] = v
	return nil
}

func DecodeMap_Fixed32_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed32, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint32](mp)
	(*m)[uint32(k)] = uint32(v)
	return nil
}

func DecodeMap_Fixed32_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed32, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, int64](mp)
	(*m)[uint32(k)] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_Fixed32_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed32, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, int32](mp)
	(*m)[uint32(k)] = DecodeZigZag32(v)
	return nil
}

func DecodeMap_Fixed32_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed32, TypeFixed64)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint64](mp)
	(*m)[uint32(k)] = v
	return nil
}

func DecodeMap_Fixed32_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed32, TypeFixed32)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint32](mp)
	(*m)[uint32(k)] = uint32(v)
	return nil
}

func DecodeMap_Fixed32_Bool(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeFixed32, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, bool](mp)
	(*m)[uint32(k)] = v != 0
	return nil
}

// Bool key decoders
func DecodeMap_Bool_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, uint64](mp)
	(*m)[k != 0] = v
	return nil
}

func DecodeMap_Bool_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, uint32](mp)
	(*m)[k != 0] = uint32(v)
	return nil
}

func DecodeMap_Bool_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, int64](mp)
	(*m)[k != 0] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_Bool_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, int32](mp)
	(*m)[k != 0] = DecodeZigZag32(v)
	return nil
}

func DecodeMap_Bool_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeFixed64)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, uint64](mp)
	(*m)[k != 0] = v
	return nil
}

func DecodeMap_Bool_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeFixed32)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, uint32](mp)
	(*m)[k != 0] = uint32(v)
	return nil
}

func DecodeMap_Bool_Bool(b []byte, mp unsafe.Pointer) error {
	k, v, err := decodeMapEntry(b, TypeVarint, TypeVarint)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, bool](mp)
	(*m)[k != 0] = v != 0
	return nil
}
