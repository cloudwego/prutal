/*
 * Copyright 2026 CloudWeGo Authors
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
	"math"
	"testing"
	"unsafe"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestGetSizeListFunc(t *testing.T) {
	packedTypes := []CoderType{
		CoderVarintU32, CoderVarintI32, CoderVarint64,
		CoderZigZag32, CoderZigZag64,
		CoderFixed32, CoderFixed64, CoderBool,
	}
	for _, ct := range packedTypes {
		assert.True(t, GetSizeListFunc(ct, true) != nil)
		assert.True(t, GetSizeListFunc(ct, false) != nil)
	}
	assert.True(t, GetSizeListFunc(CoderString, true) == nil)
	assert.True(t, GetSizeListFunc(CoderString, false) != nil)
	assert.True(t, GetSizeListFunc(CoderBytes, true) == nil)
	assert.True(t, GetSizeListFunc(CoderBytes, false) != nil)
	assert.True(t, GetSizeListFunc(CoderStruct, false) == nil)
	assert.True(t, GetSizeListFunc(CoderStruct, true) == nil)
}

func TestSizePackedMatchesAppend(t *testing.T) {
	tagNums := []int32{1, 16, 2048, 536870911}

	check := func(name string, ct CoderType, p unsafe.Pointer, empty bool) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			sizeFn := GetSizeListFunc(ct, true)
			appendFn := GetAppendFunc(ct, true)
			for _, num := range tagNums {
				tag := EncodeTag(num, TypeBytes)
				tagSize := SizeVarint(tag)
				got := sizeFn(p, tagSize)
				if empty {
					assert.Equal(t, 0, got)
					continue
				}
				encoded := AppendVarint(nil, tag)
				encoded = appendFn(encoded, p)
				assert.Equal(t, len(encoded), got)
			}
		})
	}

	var nilU64 []uint64
	emptyU64 := []uint64{}
	u64s := []uint64{0, 1, 127, 128, 1 << 20, math.MaxUint64}
	u64s2ByteLen := make([]uint64, 20) // 200 bytes -> 2-byte packed length prefix
	for i := range u64s2ByteLen {
		u64s2ByteLen[i] = math.MaxUint64
	}
	u64s3ByteLen := make([]uint64, 1700) // 17000 bytes -> 3-byte packed length prefix
	for i := range u64s3ByteLen {
		u64s3ByteLen[i] = math.MaxUint64
	}
	check("u64_nil", CoderVarint64, unsafe.Pointer(&nilU64), true)
	check("u64_empty", CoderVarint64, unsafe.Pointer(&emptyU64), true)
	check("u64", CoderVarint64, unsafe.Pointer(&u64s), false)
	check("u64_2byte_len", CoderVarint64, unsafe.Pointer(&u64s2ByteLen), false)
	check("u64_3byte_len", CoderVarint64, unsafe.Pointer(&u64s3ByteLen), false)

	i64s := []int64{math.MinInt64, -1, 0, 1, math.MaxInt64}
	check("i64_varint", CoderVarint64, unsafe.Pointer(&i64s), false)

	var nilI32 []int32
	i32s := []int32{math.MinInt32, -128, -1, 0, 1, 127, 128, math.MaxInt32}
	negI32Large := make([]int32, 20) // 200 bytes -> 2-byte packed length prefix
	for i := range negI32Large {
		negI32Large[i] = -1
	}
	check("i32_nil", CoderVarintI32, unsafe.Pointer(&nilI32), true)
	check("i32", CoderVarintI32, unsafe.Pointer(&i32s), false)
	check("i32_neg_2byte_len", CoderVarintI32, unsafe.Pointer(&negI32Large), false)

	var nilU32 []uint32
	u32s := []uint32{0, 1, 127, 128, 1 << 21, math.MaxUint32}
	check("u32_nil", CoderVarintU32, unsafe.Pointer(&nilU32), true)
	check("u32", CoderVarintU32, unsafe.Pointer(&u32s), false)

	var nilZ32 []int32
	zz32Large := make([]int32, 40) // 200 bytes -> 2-byte packed length prefix
	for i := range zz32Large {
		zz32Large[i] = math.MinInt32
	}
	check("zz32_nil", CoderZigZag32, unsafe.Pointer(&nilZ32), true)
	check("zz32", CoderZigZag32, unsafe.Pointer(&i32s), false)
	check("zz32_2byte_len", CoderZigZag32, unsafe.Pointer(&zz32Large), false)

	var nilZ64 []int64
	zz64Large := make([]int64, 20) // 200 bytes -> 2-byte packed length prefix
	for i := range zz64Large {
		zz64Large[i] = math.MinInt64
	}
	check("zz64_nil", CoderZigZag64, unsafe.Pointer(&nilZ64), true)
	check("zz64", CoderZigZag64, unsafe.Pointer(&i64s), false)
	check("zz64_2byte_len", CoderZigZag64, unsafe.Pointer(&zz64Large), false)

	var nilF32 []uint32
	f32s := make([]uint32, 40) // 160 bytes -> 2-byte packed length prefix
	for i := range f32s {
		f32s[i] = uint32(i)
	}
	check("fixed32_nil", CoderFixed32, unsafe.Pointer(&nilF32), true)
	check("fixed32_large", CoderFixed32, unsafe.Pointer(&f32s), false)

	var nilF64 []uint64
	f64s := make([]uint64, 25) // 200 bytes -> 2-byte packed length prefix
	for i := range f64s {
		f64s[i] = uint64(i)
	}
	check("fixed64_nil", CoderFixed64, unsafe.Pointer(&nilF64), true)
	check("fixed64_large", CoderFixed64, unsafe.Pointer(&f64s), false)

	var nilBool []bool
	bools := make([]bool, 150) // 150 bytes -> 2-byte packed length prefix
	for i := range bools {
		bools[i] = i%2 == 0
	}
	check("bool_nil", CoderBool, unsafe.Pointer(&nilBool), true)
	check("bool_large", CoderBool, unsafe.Pointer(&bools), false)
}

func TestSizeListMatchesAppend(t *testing.T) {
	tagNums := []int32{1, 16, 2048, 536870911}

	check := func(name string, ct CoderType, wt Type, p unsafe.Pointer) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			sizeFn := GetSizeListFunc(ct, false)
			appendFn := GetAppendListFunc(ct)
			for _, num := range tagNums {
				tag := EncodeTag(num, wt)
				tagSize := SizeVarint(tag)
				encoded := appendFn(nil, tag, p)
				assert.Equal(t, len(encoded), sizeFn(p, tagSize))
			}
		})
	}

	var nilU64 []uint64
	u64s := []uint64{0, 1, 127, 128, math.MaxUint64}
	check("u64_nil", CoderVarint64, TypeVarint, unsafe.Pointer(&nilU64))
	check("u64", CoderVarint64, TypeVarint, unsafe.Pointer(&u64s))

	i64s := []int64{math.MinInt64, -1, 0, 1, math.MaxInt64}
	check("i64_varint", CoderVarint64, TypeVarint, unsafe.Pointer(&i64s))
	check("zz64", CoderZigZag64, TypeVarint, unsafe.Pointer(&i64s))

	i32s := []int32{math.MinInt32, -1, 0, 127, 128, math.MaxInt32}
	check("i32", CoderVarintI32, TypeVarint, unsafe.Pointer(&i32s))
	check("zz32", CoderZigZag32, TypeVarint, unsafe.Pointer(&i32s))

	u32s := []uint32{0, 1, 127, 128, math.MaxUint32}
	check("u32", CoderVarintU32, TypeVarint, unsafe.Pointer(&u32s))

	f32s := []uint32{0, 1, math.MaxUint32}
	check("fixed32", CoderFixed32, TypeFixed32, unsafe.Pointer(&f32s))

	f64s := []uint64{0, 1, math.MaxUint64}
	check("fixed64", CoderFixed64, TypeFixed64, unsafe.Pointer(&f64s))

	bools := []bool{false, true, false}
	check("bool", CoderBool, TypeVarint, unsafe.Pointer(&bools))

	var nilStr []string
	strs := []string{"", "a", "hello world", string(make([]byte, 140))}
	check("string_nil", CoderString, TypeBytes, unsafe.Pointer(&nilStr))
	check("string", CoderString, TypeBytes, unsafe.Pointer(&strs))

	var nilBytes [][]byte
	bs := [][]byte{nil, {}, []byte("abc"), make([]byte, 140)}
	check("bytes_nil", CoderBytes, TypeBytes, unsafe.Pointer(&nilBytes))
	check("bytes", CoderBytes, TypeBytes, unsafe.Pointer(&bs))
}
