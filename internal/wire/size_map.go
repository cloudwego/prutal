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

// SizeMapFunc computes the total wire size of a scalar-valued map field,
// including field tags and length prefixes for each entry.
type SizeMapFunc func(p unsafe.Pointer, wireTagSize int) int

var mapSizeFuncs = map[mapEncoderFuncKey]SizeMapFunc{}

func GetMapSizeFunc(k, v CoderType) SizeMapFunc {
	return mapSizeFuncs[mapEncoderFuncKey{K: k, V: v}]
}

// Value-level size helpers (no unsafe.Pointer, no heap escape)

func szVarintU32(v uint32) int { return protowire.SizeVarint(uint64(v)) }
func szVarintU64(v uint64) int { return protowire.SizeVarint(v) }
func szVarintI32(v int32) int  { return protowire.SizeVarint(uint64(int64(v))) }
func szZigZag32(v int32) int {
	return protowire.SizeVarint(uint64(uint32(v<<1) ^ uint32(v>>31)))
}
func szZigZag64(v int64) int {
	return protowire.SizeVarint(uint64(v<<1) ^ uint64(v>>63))
}
func szFixed32(_ uint32) int   { return 4 }
func szFixed64(_ uint64) int   { return 8 }
func szBool(_ bool) int        { return 1 }
func szString(v string) int    { return protowire.SizeBytes(len(v)) }
func szByteSlice(v []byte) int { return protowire.SizeBytes(len(v)) }

// makeSizeMap creates a typed map size function using Go generics.
// Uses typed range iteration (no reflection overhead) and value-level
// key/value size functions (no unsafe.Pointer, no heap escape).
func makeSizeMap[K comparable, V any](kf func(K) int, vf func(V) int) SizeMapFunc {
	return func(p unsafe.Pointer, wireTagSize int) int {
		n := 0
		for k, v := range *(*map[K]V)(p) {
			entrySz := 1 + kf(k) + 1 + vf(v)
			n += wireTagSize + protowire.SizeBytes(entrySz)
		}
		return n
	}
}

func init() {
	reg := func(k, v CoderType, f SizeMapFunc) {
		mapSizeFuncs[mapEncoderFuncKey{K: k, V: v}] = f
	}

	// CoderVarintU32 keys (Go: uint32)
	reg(CoderVarintU32, CoderVarintU32, makeSizeMap(szVarintU32, szVarintU32))
	reg(CoderVarintU32, CoderVarintI32, makeSizeMap(szVarintU32, szVarintI32))
	reg(CoderVarintU32, CoderVarint64, makeSizeMap(szVarintU32, szVarintU64))
	reg(CoderVarintU32, CoderZigZag32, makeSizeMap(szVarintU32, szZigZag32))
	reg(CoderVarintU32, CoderZigZag64, makeSizeMap(szVarintU32, szZigZag64))
	reg(CoderVarintU32, CoderFixed32, makeSizeMap(szVarintU32, szFixed32))
	reg(CoderVarintU32, CoderFixed64, makeSizeMap(szVarintU32, szFixed64))
	reg(CoderVarintU32, CoderBool, makeSizeMap(szVarintU32, szBool))
	reg(CoderVarintU32, CoderBytes, makeSizeMap(szVarintU32, szByteSlice))
	reg(CoderVarintU32, CoderString, makeSizeMap(szVarintU32, szString))

	// CoderVarintI32 keys (Go: int32, signed varint)
	reg(CoderVarintI32, CoderVarintU32, makeSizeMap(szVarintI32, szVarintU32))
	reg(CoderVarintI32, CoderVarintI32, makeSizeMap(szVarintI32, szVarintI32))
	reg(CoderVarintI32, CoderVarint64, makeSizeMap(szVarintI32, szVarintU64))
	reg(CoderVarintI32, CoderZigZag32, makeSizeMap(szVarintI32, szZigZag32))
	reg(CoderVarintI32, CoderZigZag64, makeSizeMap(szVarintI32, szZigZag64))
	reg(CoderVarintI32, CoderFixed32, makeSizeMap(szVarintI32, szFixed32))
	reg(CoderVarintI32, CoderFixed64, makeSizeMap(szVarintI32, szFixed64))
	reg(CoderVarintI32, CoderBool, makeSizeMap(szVarintI32, szBool))
	reg(CoderVarintI32, CoderBytes, makeSizeMap(szVarintI32, szByteSlice))
	reg(CoderVarintI32, CoderString, makeSizeMap(szVarintI32, szString))

	// CoderVarint64 keys (Go: uint64)
	reg(CoderVarint64, CoderVarintU32, makeSizeMap(szVarintU64, szVarintU32))
	reg(CoderVarint64, CoderVarintI32, makeSizeMap(szVarintU64, szVarintI32))
	reg(CoderVarint64, CoderVarint64, makeSizeMap(szVarintU64, szVarintU64))
	reg(CoderVarint64, CoderZigZag32, makeSizeMap(szVarintU64, szZigZag32))
	reg(CoderVarint64, CoderZigZag64, makeSizeMap(szVarintU64, szZigZag64))
	reg(CoderVarint64, CoderFixed32, makeSizeMap(szVarintU64, szFixed32))
	reg(CoderVarint64, CoderFixed64, makeSizeMap(szVarintU64, szFixed64))
	reg(CoderVarint64, CoderBool, makeSizeMap(szVarintU64, szBool))
	reg(CoderVarint64, CoderBytes, makeSizeMap(szVarintU64, szByteSlice))
	reg(CoderVarint64, CoderString, makeSizeMap(szVarintU64, szString))

	// CoderZigZag32 keys (Go: int32)
	reg(CoderZigZag32, CoderVarintU32, makeSizeMap(szZigZag32, szVarintU32))
	reg(CoderZigZag32, CoderVarintI32, makeSizeMap(szZigZag32, szVarintI32))
	reg(CoderZigZag32, CoderVarint64, makeSizeMap(szZigZag32, szVarintU64))
	reg(CoderZigZag32, CoderZigZag32, makeSizeMap(szZigZag32, szZigZag32))
	reg(CoderZigZag32, CoderZigZag64, makeSizeMap(szZigZag32, szZigZag64))
	reg(CoderZigZag32, CoderFixed32, makeSizeMap(szZigZag32, szFixed32))
	reg(CoderZigZag32, CoderFixed64, makeSizeMap(szZigZag32, szFixed64))
	reg(CoderZigZag32, CoderBool, makeSizeMap(szZigZag32, szBool))
	reg(CoderZigZag32, CoderBytes, makeSizeMap(szZigZag32, szByteSlice))
	reg(CoderZigZag32, CoderString, makeSizeMap(szZigZag32, szString))

	// CoderZigZag64 keys (Go: int64)
	reg(CoderZigZag64, CoderVarintU32, makeSizeMap(szZigZag64, szVarintU32))
	reg(CoderZigZag64, CoderVarintI32, makeSizeMap(szZigZag64, szVarintI32))
	reg(CoderZigZag64, CoderVarint64, makeSizeMap(szZigZag64, szVarintU64))
	reg(CoderZigZag64, CoderZigZag32, makeSizeMap(szZigZag64, szZigZag32))
	reg(CoderZigZag64, CoderZigZag64, makeSizeMap(szZigZag64, szZigZag64))
	reg(CoderZigZag64, CoderFixed32, makeSizeMap(szZigZag64, szFixed32))
	reg(CoderZigZag64, CoderFixed64, makeSizeMap(szZigZag64, szFixed64))
	reg(CoderZigZag64, CoderBool, makeSizeMap(szZigZag64, szBool))
	reg(CoderZigZag64, CoderBytes, makeSizeMap(szZigZag64, szByteSlice))
	reg(CoderZigZag64, CoderString, makeSizeMap(szZigZag64, szString))

	// CoderFixed32 keys (Go: uint32)
	reg(CoderFixed32, CoderVarintU32, makeSizeMap(szFixed32, szVarintU32))
	reg(CoderFixed32, CoderVarintI32, makeSizeMap(szFixed32, szVarintI32))
	reg(CoderFixed32, CoderVarint64, makeSizeMap(szFixed32, szVarintU64))
	reg(CoderFixed32, CoderZigZag32, makeSizeMap(szFixed32, szZigZag32))
	reg(CoderFixed32, CoderZigZag64, makeSizeMap(szFixed32, szZigZag64))
	reg(CoderFixed32, CoderFixed32, makeSizeMap(szFixed32, szFixed32))
	reg(CoderFixed32, CoderFixed64, makeSizeMap(szFixed32, szFixed64))
	reg(CoderFixed32, CoderBool, makeSizeMap(szFixed32, szBool))
	reg(CoderFixed32, CoderBytes, makeSizeMap(szFixed32, szByteSlice))
	reg(CoderFixed32, CoderString, makeSizeMap(szFixed32, szString))

	// CoderFixed64 keys (Go: uint64)
	reg(CoderFixed64, CoderVarintU32, makeSizeMap(szFixed64, szVarintU32))
	reg(CoderFixed64, CoderVarintI32, makeSizeMap(szFixed64, szVarintI32))
	reg(CoderFixed64, CoderVarint64, makeSizeMap(szFixed64, szVarintU64))
	reg(CoderFixed64, CoderZigZag32, makeSizeMap(szFixed64, szZigZag32))
	reg(CoderFixed64, CoderZigZag64, makeSizeMap(szFixed64, szZigZag64))
	reg(CoderFixed64, CoderFixed32, makeSizeMap(szFixed64, szFixed32))
	reg(CoderFixed64, CoderFixed64, makeSizeMap(szFixed64, szFixed64))
	reg(CoderFixed64, CoderBool, makeSizeMap(szFixed64, szBool))
	reg(CoderFixed64, CoderBytes, makeSizeMap(szFixed64, szByteSlice))
	reg(CoderFixed64, CoderString, makeSizeMap(szFixed64, szString))

	// CoderBool keys (Go: bool)
	reg(CoderBool, CoderVarintU32, makeSizeMap(szBool, szVarintU32))
	reg(CoderBool, CoderVarintI32, makeSizeMap(szBool, szVarintI32))
	reg(CoderBool, CoderVarint64, makeSizeMap(szBool, szVarintU64))
	reg(CoderBool, CoderZigZag32, makeSizeMap(szBool, szZigZag32))
	reg(CoderBool, CoderZigZag64, makeSizeMap(szBool, szZigZag64))
	reg(CoderBool, CoderFixed32, makeSizeMap(szBool, szFixed32))
	reg(CoderBool, CoderFixed64, makeSizeMap(szBool, szFixed64))
	reg(CoderBool, CoderBool, makeSizeMap(szBool, szBool))
	reg(CoderBool, CoderBytes, makeSizeMap(szBool, szByteSlice))
	reg(CoderBool, CoderString, makeSizeMap(szBool, szString))

	// CoderString keys (Go: string)
	reg(CoderString, CoderVarintU32, makeSizeMap(szString, szVarintU32))
	reg(CoderString, CoderVarintI32, makeSizeMap(szString, szVarintI32))
	reg(CoderString, CoderVarint64, makeSizeMap(szString, szVarintU64))
	reg(CoderString, CoderZigZag32, makeSizeMap(szString, szZigZag32))
	reg(CoderString, CoderZigZag64, makeSizeMap(szString, szZigZag64))
	reg(CoderString, CoderFixed32, makeSizeMap(szString, szFixed32))
	reg(CoderString, CoderFixed64, makeSizeMap(szString, szFixed64))
	reg(CoderString, CoderBool, makeSizeMap(szString, szBool))
	reg(CoderString, CoderBytes, makeSizeMap(szString, szByteSlice))
	reg(CoderString, CoderString, makeSizeMap(szString, szString))
}
