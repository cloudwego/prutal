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

import "unsafe"

type mapEncoderFuncKey struct {
	K, V CoderType
}

var (
	mapEncoderFuncs = map[mapEncoderFuncKey]AppendRepeatedFunc{}
)

func GetMapEncoderFunc(k, v CoderType) AppendRepeatedFunc {
	return mapEncoderFuncs[mapEncoderFuncKey{K: k, V: v}]
}

func init() {
	register := func(k, v CoderType, f AppendRepeatedFunc) {
		mapEncoderFuncs[mapEncoderFuncKey{K: k, V: v}] = f
	}

	register(CoderVarint32, CoderVarint32, AppendMap_VarintU32_VarintU32)
	register(CoderVarint32, CoderVarint64, AppendMap_VarintU32_VarintU64)
	register(CoderVarint32, CoderZigZag32, AppendMap_VarintU32_ZigZag32)
	register(CoderVarint32, CoderZigZag64, AppendMap_VarintU32_ZigZag64)
	register(CoderVarint32, CoderFixed32, AppendMap_VarintU32_Fixed32)
	register(CoderVarint32, CoderFixed64, AppendMap_VarintU32_Fixed64)
	register(CoderVarint32, CoderBool, AppendMap_VarintU32_Bool)
	register(CoderVarint32, CoderBytes, AppendMap_VarintU32_Bytes)
	register(CoderVarint32, CoderString, AppendMap_VarintU32_String)

	register(CoderVarint64, CoderVarint32, AppendMap_VarintU64_VarintU32)
	register(CoderVarint64, CoderVarint64, AppendMap_VarintU64_VarintU64)
	register(CoderVarint64, CoderZigZag32, AppendMap_VarintU64_ZigZag32)
	register(CoderVarint64, CoderZigZag64, AppendMap_VarintU64_ZigZag64)
	register(CoderVarint64, CoderFixed32, AppendMap_VarintU64_Fixed32)
	register(CoderVarint64, CoderFixed64, AppendMap_VarintU64_Fixed64)
	register(CoderVarint64, CoderBool, AppendMap_VarintU64_Bool)
	register(CoderVarint64, CoderBytes, AppendMap_VarintU64_Bytes)
	register(CoderVarint64, CoderString, AppendMap_VarintU64_String)

	register(CoderZigZag32, CoderVarint32, AppendMap_ZigZag32_VarintU32)
	register(CoderZigZag32, CoderVarint64, AppendMap_ZigZag32_VarintU64)
	register(CoderZigZag32, CoderZigZag32, AppendMap_ZigZag32_ZigZag32)
	register(CoderZigZag32, CoderZigZag64, AppendMap_ZigZag32_ZigZag64)
	register(CoderZigZag32, CoderFixed32, AppendMap_ZigZag32_Fixed32)
	register(CoderZigZag32, CoderFixed64, AppendMap_ZigZag32_Fixed64)
	register(CoderZigZag32, CoderBool, AppendMap_ZigZag32_Bool)
	register(CoderZigZag32, CoderBytes, AppendMap_ZigZag32_Bytes)
	register(CoderZigZag32, CoderString, AppendMap_ZigZag32_String)

	register(CoderZigZag64, CoderVarint32, AppendMap_ZigZag64_VarintU32)
	register(CoderZigZag64, CoderVarint64, AppendMap_ZigZag64_VarintU64)
	register(CoderZigZag64, CoderZigZag32, AppendMap_ZigZag64_ZigZag32)
	register(CoderZigZag64, CoderZigZag64, AppendMap_ZigZag64_ZigZag64)
	register(CoderZigZag64, CoderFixed32, AppendMap_ZigZag64_Fixed32)
	register(CoderZigZag64, CoderFixed64, AppendMap_ZigZag64_Fixed64)
	register(CoderZigZag64, CoderBool, AppendMap_ZigZag64_Bool)
	register(CoderZigZag64, CoderBytes, AppendMap_ZigZag64_Bytes)
	register(CoderZigZag64, CoderString, AppendMap_ZigZag64_String)

	register(CoderFixed32, CoderVarint32, AppendMap_Fixed32_VarintU32)
	register(CoderFixed32, CoderVarint64, AppendMap_Fixed32_VarintU64)
	register(CoderFixed32, CoderZigZag32, AppendMap_Fixed32_ZigZag32)
	register(CoderFixed32, CoderZigZag64, AppendMap_Fixed32_ZigZag64)
	register(CoderFixed32, CoderFixed32, AppendMap_Fixed32_Fixed32)
	register(CoderFixed32, CoderFixed64, AppendMap_Fixed32_Fixed64)
	register(CoderFixed32, CoderBool, AppendMap_Fixed32_Bool)
	register(CoderFixed32, CoderBytes, AppendMap_Fixed32_Bytes)
	register(CoderFixed32, CoderString, AppendMap_Fixed32_String)

	register(CoderFixed64, CoderVarint32, AppendMap_Fixed64_VarintU32)
	register(CoderFixed64, CoderVarint64, AppendMap_Fixed64_VarintU64)
	register(CoderFixed64, CoderZigZag32, AppendMap_Fixed64_ZigZag32)
	register(CoderFixed64, CoderZigZag64, AppendMap_Fixed64_ZigZag64)
	register(CoderFixed64, CoderFixed32, AppendMap_Fixed64_Fixed32)
	register(CoderFixed64, CoderFixed64, AppendMap_Fixed64_Fixed64)
	register(CoderFixed64, CoderBool, AppendMap_Fixed64_Bool)
	register(CoderFixed64, CoderBytes, AppendMap_Fixed64_Bytes)
	register(CoderFixed64, CoderString, AppendMap_Fixed64_String)

	register(CoderBool, CoderVarint32, AppendMap_Bool_VarintU32)
	register(CoderBool, CoderVarint64, AppendMap_Bool_VarintU64)
	register(CoderBool, CoderZigZag32, AppendMap_Bool_ZigZag32)
	register(CoderBool, CoderZigZag64, AppendMap_Bool_ZigZag64)
	register(CoderBool, CoderFixed32, AppendMap_Bool_Fixed32)
	register(CoderBool, CoderFixed64, AppendMap_Bool_Fixed64)
	register(CoderBool, CoderBool, AppendMap_Bool_Bool)
	register(CoderBool, CoderBytes, AppendMap_Bool_Bytes)
	register(CoderBool, CoderString, AppendMap_Bool_String)

	register(CoderString, CoderVarint32, AppendMap_String_VarintU32)
	register(CoderString, CoderVarint64, AppendMap_String_VarintU64)
	register(CoderString, CoderZigZag32, AppendMap_String_ZigZag32)
	register(CoderString, CoderZigZag64, AppendMap_String_ZigZag64)
	register(CoderString, CoderFixed32, AppendMap_String_Fixed32)
	register(CoderString, CoderFixed64, AppendMap_String_Fixed64)
	register(CoderString, CoderBool, AppendMap_String_Bool)
	register(CoderString, CoderBytes, AppendMap_String_Bytes)
	register(CoderString, CoderString, AppendMap_String_String)
}

func appendMapKey_VarintU32(b []byte, k uint32) []byte {
	b = append(b, byte(1)<<3|byte(TypeVarint))
	for k >= 0x80 {
		b = append(b, byte(k)|0x80)
		k >>= 7
	}
	return append(b, byte(k))
}

func appendMapKey_VarintU64(b []byte, k uint64) []byte {
	b = append(b, byte(1)<<3|byte(TypeVarint))
	for k >= 0x80 {
		b = append(b, byte(k)|0x80)
		k >>= 7
	}
	return append(b, byte(k))
}

func appendMapKey_ZigZag32(b []byte, k int32) []byte {
	b = append(b, byte(1)<<3|byte(TypeVarint))
	v := uint32(k<<1) ^ uint32(k>>31)
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}
	return append(b, byte(v))
}

func appendMapKey_ZigZag64(b []byte, k int64) []byte {
	b = append(b, byte(1)<<3|byte(TypeVarint))
	v := uint64(k<<1) ^ uint64(k>>63)
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}
	return append(b, byte(v))
}

func appendMapKey_Fixed32(b []byte, k uint32) []byte {
	b = append(b, byte(1)<<3|byte(TypeFixed32))
	return append(b,
		byte(k>>0),
		byte(k>>8),
		byte(k>>16),
		byte(k>>24))
}

func appendMapKey_Fixed64(b []byte, k uint64) []byte {
	b = append(b, byte(1)<<3|byte(TypeFixed64))
	return append(b,
		byte(k>>0),
		byte(k>>8),
		byte(k>>16),
		byte(k>>24),
		byte(k>>32),
		byte(k>>40),
		byte(k>>48),
		byte(k>>56))
}

func appendMapKey_Bool(b []byte, k bool) []byte {
	b = append(b, byte(1)<<3|byte(TypeVarint))
	if k {
		return append(b, 1)
	}
	return append(b, 0)
}

func appendMapKey_String(b []byte, k string) []byte {
	b = append(b, byte(1)<<3|byte(TypeBytes))
	b = AppendVarint(b, uint64(len(k)))
	return append(b, k...)
}

func appendMapValue_VarintU32(b []byte, v uint32) []byte {
	b = append(b, byte(2)<<3|byte(TypeVarint))
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}
	return append(b, byte(v))
}

func appendMapValue_VarintU64(b []byte, v uint64) []byte {
	b = append(b, byte(2)<<3|byte(TypeVarint))
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}
	return append(b, byte(v))
}

func appendMapValue_ZigZag32(b []byte, v int32) []byte {
	b = append(b, byte(2)<<3|byte(TypeVarint))
	z := uint32(v<<1) ^ uint32(v>>31)
	for z >= 0x80 {
		b = append(b, byte(z)|0x80)
		z >>= 7
	}
	return append(b, byte(z))
}

func appendMapValue_ZigZag64(b []byte, v int64) []byte {
	b = append(b, byte(2)<<3|byte(TypeVarint))
	z := uint64(v<<1) ^ uint64(v>>63)
	for z >= 0x80 {
		b = append(b, byte(z)|0x80)
		z >>= 7
	}
	return append(b, byte(z))
}

func appendMapValue_Fixed32(b []byte, v uint32) []byte {
	b = append(b, byte(2)<<3|byte(TypeFixed32))
	return append(b,
		byte(v>>0),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24))
}

func appendMapValue_Fixed64(b []byte, v uint64) []byte {
	b = append(b, byte(2)<<3|byte(TypeFixed64))
	return append(b,
		byte(v>>0),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
		byte(v>>32),
		byte(v>>40),
		byte(v>>48),
		byte(v>>56))
}

func appendMapValue_Bool(b []byte, v bool) []byte {
	b = append(b, byte(2)<<3|byte(TypeVarint))
	if v {
		return append(b, 1)
	}
	return append(b, 0)
}

func appendMapValue_Bytes(b []byte, v []byte) []byte {
	b = append(b, byte(2)<<3|byte(TypeBytes))
	b = AppendVarint(b, uint64(len(v)))
	return append(b, v...)
}

func appendMapValue_String(b []byte, v string) []byte {
	b = append(b, byte(2)<<3|byte(TypeBytes))
	b = AppendVarint(b, uint64(len(v)))
	return append(b, v...)
}

// VarintU32 key encoders
func AppendMap_VarintU32_VarintU32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU32(b, k)
		b = appendMapValue_VarintU32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU32_VarintU64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU32(b, k)
		b = appendMapValue_VarintU64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU32_ZigZag32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]int32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU32(b, k)
		b = appendMapValue_ZigZag32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU32_ZigZag64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]int64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU32(b, k)
		b = appendMapValue_ZigZag64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU32_Fixed32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU32(b, k)
		b = appendMapValue_Fixed32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU32_Fixed64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU32(b, k)
		b = appendMapValue_Fixed64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU32_Bool(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]bool)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU32(b, k)
		b = appendMapValue_Bool(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU32_Bytes(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32][]byte)(p)
	if len(m) == 0 {
		return b
	}
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU32(b, k)
		b = appendMapValue_Bytes(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU32_String(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]string)(p)
	if len(m) == 0 {
		return b
	}
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU32(b, k)
		b = appendMapValue_String(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

// VarintU64 key encoders
func AppendMap_VarintU64_VarintU32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU64(b, k)
		b = appendMapValue_VarintU32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU64_VarintU64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU64(b, k)
		b = appendMapValue_VarintU64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU64_ZigZag32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]int32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU64(b, k)
		b = appendMapValue_ZigZag32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU64_ZigZag64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]int64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU64(b, k)
		b = appendMapValue_ZigZag64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU64_Fixed32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU64(b, k)
		b = appendMapValue_Fixed32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU64_Fixed64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU64(b, k)
		b = appendMapValue_Fixed64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU64_Bool(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]bool)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU64(b, k)
		b = appendMapValue_Bool(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU64_Bytes(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64][]byte)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU64(b, k)
		b = appendMapValue_Bytes(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_VarintU64_String(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]string)(p)
	if len(m) == 0 {
		return b
	}
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_VarintU64(b, k)
		b = appendMapValue_String(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

// ZigZag32 key encoders
func AppendMap_ZigZag32_VarintU32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int32]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag32(b, k)
		b = appendMapValue_VarintU32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag32_VarintU64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int32]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag32(b, k)
		b = appendMapValue_VarintU64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag32_ZigZag32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int32]int32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag32(b, k)
		b = appendMapValue_ZigZag32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag32_ZigZag64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int32]int64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag32(b, k)
		b = appendMapValue_ZigZag64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag32_Fixed32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int32]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag32(b, k)
		b = appendMapValue_Fixed32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag32_Fixed64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int32]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag32(b, k)
		b = appendMapValue_Fixed64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag32_Bool(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int32]bool)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag32(b, k)
		b = appendMapValue_Bool(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag32_Bytes(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int32][]byte)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag32(b, k)
		b = appendMapValue_Bytes(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag32_String(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int32]string)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag32(b, k)
		b = appendMapValue_String(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

// ZigZag64 key encoders
func AppendMap_ZigZag64_VarintU32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int64]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag64(b, k)
		b = appendMapValue_VarintU32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag64_VarintU64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int64]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag64(b, k)
		b = appendMapValue_VarintU64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag64_ZigZag32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int64]int32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag64(b, k)
		b = appendMapValue_ZigZag32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag64_ZigZag64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int64]int64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag64(b, k)
		b = appendMapValue_ZigZag64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag64_Fixed32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int64]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag64(b, k)
		b = appendMapValue_Fixed32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag64_Fixed64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int64]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag64(b, k)
		b = appendMapValue_Fixed64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag64_Bool(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int64]bool)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag64(b, k)
		b = appendMapValue_Bool(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag64_Bytes(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int64][]byte)(p)
	if len(m) == 0 {
		return b
	}
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag64(b, k)
		b = appendMapValue_Bytes(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_ZigZag64_String(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[int64]string)(p)
	if len(m) == 0 {
		return b
	}
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_ZigZag64(b, k)
		b = appendMapValue_String(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

// Fixed32 key encoders
func AppendMap_Fixed32_VarintU32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed32(b, k)
		b = appendMapValue_VarintU32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed32_VarintU64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed32(b, k)
		b = appendMapValue_VarintU64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed32_ZigZag32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]int32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed32(b, k)
		b = appendMapValue_ZigZag32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed32_ZigZag64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]int64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed32(b, k)
		b = appendMapValue_ZigZag64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed32_Fixed32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed32(b, k)
		b = appendMapValue_Fixed32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed32_Fixed64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed32(b, k)
		b = appendMapValue_Fixed64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed32_Bool(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]bool)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed32(b, k)
		b = appendMapValue_Bool(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed32_Bytes(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32][]byte)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed32(b, k)
		b = appendMapValue_Bytes(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed32_String(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint32]string)(p)
	if len(m) == 0 {
		return b
	}
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed32(b, k)
		b = appendMapValue_String(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

// Fixed64 key encoders
func AppendMap_Fixed64_VarintU32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed64(b, k)
		b = appendMapValue_VarintU32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed64_VarintU64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed64(b, k)
		b = appendMapValue_VarintU64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed64_ZigZag32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]int32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed64(b, k)
		b = appendMapValue_ZigZag32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed64_ZigZag64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]int64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed64(b, k)
		b = appendMapValue_ZigZag64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed64_Fixed32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed64(b, k)
		b = appendMapValue_Fixed32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed64_Fixed64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed64(b, k)
		b = appendMapValue_Fixed64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed64_Bool(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]bool)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed64(b, k)
		b = appendMapValue_Bool(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed64_Bytes(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64][]byte)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed64(b, k)
		b = appendMapValue_Bytes(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Fixed64_String(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[uint64]string)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Fixed64(b, k)
		b = appendMapValue_String(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

// Bool key encoders
func AppendMap_Bool_VarintU32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[bool]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Bool(b, k)
		b = appendMapValue_VarintU32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Bool_VarintU64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[bool]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Bool(b, k)
		b = appendMapValue_VarintU64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Bool_ZigZag32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[bool]int32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Bool(b, k)
		b = appendMapValue_ZigZag32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Bool_ZigZag64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[bool]int64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Bool(b, k)
		b = appendMapValue_ZigZag64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Bool_Fixed32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[bool]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Bool(b, k)
		b = appendMapValue_Fixed32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Bool_Fixed64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[bool]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Bool(b, k)
		b = appendMapValue_Fixed64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Bool_Bool(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[bool]bool)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Bool(b, k)
		b = appendMapValue_Bool(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Bool_Bytes(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[bool][]byte)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Bool(b, k)
		b = appendMapValue_Bytes(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_Bool_String(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[bool]string)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_Bool(b, k)
		b = appendMapValue_String(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

// String key encoders
func AppendMap_String_VarintU32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[string]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_String(b, k)
		b = appendMapValue_VarintU32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_String_VarintU64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[string]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_String(b, k)
		b = appendMapValue_VarintU64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_String_ZigZag32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[string]int32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_String(b, k)
		b = appendMapValue_ZigZag32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_String_ZigZag64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[string]int64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_String(b, k)
		b = appendMapValue_ZigZag64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_String_Fixed32(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[string]uint32)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_String(b, k)
		b = appendMapValue_Fixed32(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_String_Fixed64(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[string]uint64)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_String(b, k)
		b = appendMapValue_Fixed64(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_String_Bool(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[string]bool)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_String(b, k)
		b = appendMapValue_Bool(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_String_Bytes(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[string][]byte)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_String(b, k)
		b = appendMapValue_Bytes(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}

func AppendMap_String_String(b []byte, tag uint64, p unsafe.Pointer) []byte {
	m := *(*map[string]string)(p)
	for k, v := range m {
		b = AppendVarintSmall(b, tag)
		b = LenReserve(b)
		sz0 := len(b)
		b = appendMapKey_String(b, k)
		b = appendMapValue_String(b, v)
		b = LenPack(b, len(b)-sz0)
	}
	return b
}
