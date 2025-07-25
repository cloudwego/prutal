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
	"io"
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
	register(CoderVarint32, CoderVarint32, DecodeMap_VarintU32_VarintU32)
	register(CoderVarint32, CoderVarint64, DecodeMap_VarintU32_VarintU64)
	register(CoderVarint32, CoderZigZag32, DecodeMap_VarintU32_ZigZag32)
	register(CoderVarint32, CoderZigZag64, DecodeMap_VarintU32_ZigZag64)
	register(CoderVarint32, CoderFixed32, DecodeMap_VarintU32_Fixed32)
	register(CoderVarint32, CoderFixed64, DecodeMap_VarintU32_Fixed64)
	register(CoderVarint32, CoderBool, DecodeMap_VarintU32_Bool)

	register(CoderVarint64, CoderVarint32, DecodeMap_VarintU64_VarintU32)
	register(CoderVarint64, CoderVarint64, DecodeMap_VarintU64_VarintU64)
	register(CoderVarint64, CoderZigZag32, DecodeMap_VarintU64_ZigZag32)
	register(CoderVarint64, CoderZigZag64, DecodeMap_VarintU64_ZigZag64)
	register(CoderVarint64, CoderFixed32, DecodeMap_VarintU64_Fixed32)
	register(CoderVarint64, CoderFixed64, DecodeMap_VarintU64_Fixed64)
	register(CoderVarint64, CoderBool, DecodeMap_VarintU64_Bool)

	register(CoderZigZag32, CoderVarint32, DecodeMap_ZigZag32_VarintU32)
	register(CoderZigZag32, CoderVarint64, DecodeMap_ZigZag32_VarintU64)
	register(CoderZigZag32, CoderZigZag32, DecodeMap_ZigZag32_ZigZag32)
	register(CoderZigZag32, CoderZigZag64, DecodeMap_ZigZag32_ZigZag64)
	register(CoderZigZag32, CoderFixed32, DecodeMap_ZigZag32_Fixed32)
	register(CoderZigZag32, CoderFixed64, DecodeMap_ZigZag32_Fixed64)
	register(CoderZigZag32, CoderBool, DecodeMap_ZigZag32_Bool)

	register(CoderZigZag64, CoderVarint32, DecodeMap_ZigZag64_VarintU32)
	register(CoderZigZag64, CoderVarint64, DecodeMap_ZigZag64_VarintU64)
	register(CoderZigZag64, CoderZigZag32, DecodeMap_ZigZag64_ZigZag32)
	register(CoderZigZag64, CoderZigZag64, DecodeMap_ZigZag64_ZigZag64)
	register(CoderZigZag64, CoderFixed32, DecodeMap_ZigZag64_Fixed32)
	register(CoderZigZag64, CoderFixed64, DecodeMap_ZigZag64_Fixed64)
	register(CoderZigZag64, CoderBool, DecodeMap_ZigZag64_Bool)

	register(CoderFixed32, CoderVarint32, DecodeMap_Fixed32_VarintU32)
	register(CoderFixed32, CoderVarint64, DecodeMap_Fixed32_VarintU64)
	register(CoderFixed32, CoderZigZag32, DecodeMap_Fixed32_ZigZag32)
	register(CoderFixed32, CoderZigZag64, DecodeMap_Fixed32_ZigZag64)
	register(CoderFixed32, CoderFixed32, DecodeMap_Fixed32_Fixed32)
	register(CoderFixed32, CoderFixed64, DecodeMap_Fixed32_Fixed64)
	register(CoderFixed32, CoderBool, DecodeMap_Fixed32_Bool)

	register(CoderFixed64, CoderVarint32, DecodeMap_Fixed64_VarintU32)
	register(CoderFixed64, CoderVarint64, DecodeMap_Fixed64_VarintU64)
	register(CoderFixed64, CoderZigZag32, DecodeMap_Fixed64_ZigZag32)
	register(CoderFixed64, CoderZigZag64, DecodeMap_Fixed64_ZigZag64)
	register(CoderFixed64, CoderFixed32, DecodeMap_Fixed64_Fixed32)
	register(CoderFixed64, CoderFixed64, DecodeMap_Fixed64_Fixed64)
	register(CoderFixed64, CoderBool, DecodeMap_Fixed64_Bool)

	register(CoderBool, CoderVarint32, DecodeMap_Bool_VarintU32)
	register(CoderBool, CoderVarint64, DecodeMap_Bool_VarintU64)
	register(CoderBool, CoderZigZag32, DecodeMap_Bool_ZigZag32)
	register(CoderBool, CoderZigZag64, DecodeMap_Bool_ZigZag64)
	register(CoderBool, CoderFixed32, DecodeMap_Bool_Fixed32)
	register(CoderBool, CoderFixed64, DecodeMap_Bool_Fixed64)
	register(CoderBool, CoderBool, DecodeMap_Bool_Bool)
}

func decodeMap_Varint(b []byte, num int32) (v uint64, n int, err error) {
	fn, typ := ConsumeKVTag(b)
	if fn != num {
		return 0, 0, newFieldNumErr(fn, num)
	}
	if typ != TypeVarint {
		return 0, 0, newTypeNotMatchErr(typ, TypeVarint)
	}
	b = b[1:]
	if len(b) >= 1 && b[0] < 0x80 {
		v = uint64(b[0])
		n = 1
	} else if len(b) >= 2 && b[1] < 0x80 {
		v = uint64(b[0]&0x7f) + uint64(b[1])<<7
		n = 2
	} else {
		v, n = protowire.ConsumeVarint(b)
	}
	if n < 0 {
		return 0, 0, protowire.ParseError(n)
	}
	n++ // +1 for the tag byte
	return
}

func decodeMap_Fixed32(b []byte, num int32) (v uint32, n int, err error) {
	fn, typ := ConsumeKVTag(b)
	if fn != num {
		return 0, 0, newFieldNumErr(fn, num)
	}
	if typ != TypeFixed32 {
		return 0, 0, newTypeNotMatchErr(typ, TypeFixed32)
	}
	b = b[1:]
	v, n = protowire.ConsumeFixed32(b)
	if n < 0 {
		return 0, 0, protowire.ParseError(n)
	}
	n++ // +1 for the tag byte
	return
}

func decodeMap_Fixed64(b []byte, num int32) (v uint64, n int, err error) {
	fn, typ := ConsumeKVTag(b)
	if fn != num {
		return 0, 0, newFieldNumErr(fn, num)
	}
	if typ != TypeFixed64 {
		return 0, 0, newTypeNotMatchErr(typ, TypeFixed64)
	}
	b = b[1:]
	v, n = protowire.ConsumeFixed64(b)
	if n < 0 {
		return 0, 0, protowire.ParseError(n)
	}
	n++ // +1 for the tag byte
	return
}

func decodeMap_Bool(b []byte, num int32) (bool, int, error) {
	if len(b) < 2 {
		return false, 0, io.ErrUnexpectedEOF
	}
	fn, typ := ConsumeKVTag(b)
	if fn != num {
		return false, 0, newFieldNumErr(fn, num)
	}
	if typ != TypeVarint {
		return false, 0, newTypeNotMatchErr(typ, TypeVarint)
	}
	return b[1] != 0, 2, nil
}

func DecodeMap_VarintU32_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint32](mp)
	(*m)[uint32(k)] = uint32(v)
	return nil
}

func DecodeMap_VarintU32_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, int64](mp)
	(*m)[uint32(k)] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_VarintU32_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint64](mp)
	(*m)[uint32(k)] = v
	return nil
}

func DecodeMap_VarintU32_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, int32](mp)
	(*m)[uint32(k)] = int32(protowire.DecodeZigZag(v))
	return nil
}

func DecodeMap_VarintU32_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed64(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint64](mp)
	(*m)[uint32(k)] = v
	return nil
}

func DecodeMap_VarintU32_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed32(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint32](mp)
	(*m)[uint32(k)] = v
	return nil
}

func DecodeMap_VarintU32_Bool(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Bool(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, bool](mp)
	(*m)[uint32(k)] = v
	return nil
}

func DecodeMap_VarintU64_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint64](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_VarintU64_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint32](mp)
	(*m)[k] = uint32(v)
	return nil
}

func DecodeMap_VarintU64_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, int64](mp)
	(*m)[k] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_VarintU64_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, int32](mp)
	(*m)[k] = int32(protowire.DecodeZigZag(v))
	return nil
}

func DecodeMap_VarintU64_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed64(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint64](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_VarintU64_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed32(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint32](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_VarintU64_Bool(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Bool(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, bool](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_ZigZag64_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, uint64](mp)
	(*m)[protowire.DecodeZigZag(k)] = v
	return nil
}

func DecodeMap_ZigZag64_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, uint32](mp)
	(*m)[protowire.DecodeZigZag(k)] = uint32(v)
	return nil
}

func DecodeMap_ZigZag64_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, int64](mp)
	(*m)[protowire.DecodeZigZag(k)] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_ZigZag64_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, int32](mp)
	(*m)[protowire.DecodeZigZag(k)] = int32(protowire.DecodeZigZag(v))
	return nil
}

func DecodeMap_ZigZag64_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed64(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, uint64](mp)
	(*m)[protowire.DecodeZigZag(k)] = v
	return nil
}

func DecodeMap_ZigZag64_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed32(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, uint32](mp)
	(*m)[protowire.DecodeZigZag(k)] = v
	return nil
}

func DecodeMap_ZigZag64_Bool(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Bool(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int64, bool](mp)
	(*m)[protowire.DecodeZigZag(k)] = v
	return nil
}

func DecodeMap_ZigZag32_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, uint64](mp)
	(*m)[int32(protowire.DecodeZigZag(k))] = v
	return nil
}

func DecodeMap_ZigZag32_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, uint32](mp)
	(*m)[int32(protowire.DecodeZigZag(k))] = uint32(v)
	return nil
}

func DecodeMap_ZigZag32_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, int64](mp)
	(*m)[int32(protowire.DecodeZigZag(k))] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_ZigZag32_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, int32](mp)
	(*m)[int32(protowire.DecodeZigZag(k))] = int32(protowire.DecodeZigZag(v))
	return nil
}

func DecodeMap_ZigZag32_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed64(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, uint64](mp)
	(*m)[int32(protowire.DecodeZigZag(k))] = v
	return nil
}

func DecodeMap_ZigZag32_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed32(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, uint32](mp)
	(*m)[int32(protowire.DecodeZigZag(k))] = v
	return nil
}

func DecodeMap_ZigZag32_Bool(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Varint(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Bool(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[int32, bool](mp)
	(*m)[int32(protowire.DecodeZigZag(k))] = v
	return nil
}

func DecodeMap_Fixed64_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed64(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint64](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_Fixed64_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed64(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint32](mp)
	(*m)[k] = uint32(v)
	return nil
}

func DecodeMap_Fixed64_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed64(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, int64](mp)
	(*m)[k] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_Fixed64_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed64(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, int32](mp)
	(*m)[k] = int32(protowire.DecodeZigZag(v))
	return nil
}

func DecodeMap_Fixed64_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed64(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed64(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint64](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_Fixed64_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed64(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed32(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, uint32](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_Fixed64_Bool(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed64(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Bool(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint64, bool](mp)
	(*m)[k] = v
	return nil
}

// Fixed32 key decoders
func DecodeMap_Fixed32_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed32(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint64](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_Fixed32_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed32(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint32](mp)
	(*m)[k] = uint32(v)
	return nil
}

func DecodeMap_Fixed32_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed32(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, int64](mp)
	(*m)[k] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_Fixed32_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed32(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, int32](mp)
	(*m)[k] = int32(protowire.DecodeZigZag(v))
	return nil
}

func DecodeMap_Fixed32_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed32(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed64(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint64](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_Fixed32_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed32(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed32(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, uint32](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_Fixed32_Bool(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Fixed32(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Bool(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[uint32, bool](mp)
	(*m)[k] = v
	return nil
}

// Bool key decoders
func DecodeMap_Bool_VarintU64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Bool(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, uint64](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_Bool_VarintU32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Bool(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, uint32](mp)
	(*m)[k] = uint32(v)
	return nil
}

func DecodeMap_Bool_ZigZag64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Bool(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, int64](mp)
	(*m)[k] = protowire.DecodeZigZag(v)
	return nil
}

func DecodeMap_Bool_ZigZag32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Bool(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Varint(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, int32](mp)
	(*m)[k] = int32(protowire.DecodeZigZag(v))
	return nil
}

func DecodeMap_Bool_Fixed64(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Bool(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed64(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, uint64](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_Bool_Fixed32(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Bool(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Fixed32(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, uint32](mp)
	(*m)[k] = v
	return nil
}

func DecodeMap_Bool_Bool(b []byte, mp unsafe.Pointer) error {
	k, n, err := decodeMap_Bool(b, 1)
	if err != nil {
		return err
	}
	b = b[n:]
	v, _, err := decodeMap_Bool(b, 2)
	if err != nil {
		return err
	}
	m := ensureMapNotNil[bool, bool](mp)
	(*m)[k] = v
	return nil
}
