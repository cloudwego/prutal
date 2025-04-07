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

func init() {
	appendPackedFuncs[CoderVarint32] = UnsafeAppendPackedVarintU32
	appendPackedFuncs[CoderVarint64] = UnsafeAppendPackedVarintU64
	appendPackedFuncs[CoderZigZag32] = UnsafeAppendPackedZigZag32
	appendPackedFuncs[CoderZigZag64] = UnsafeAppendPackedZigZag64
	appendPackedFuncs[CoderFixed32] = UnsafeAppendPackedFixed32
	appendPackedFuncs[CoderFixed64] = UnsafeAppendPackedFixed64
	appendPackedFuncs[CoderBool] = UnsafeAppendPackedBool
}

func UnsafeAppendPackedVarintU64(b []byte, p unsafe.Pointer) []byte {
	b = LenReserve(b)
	sz0 := len(b)
	for _, v := range *(*[]uint64)(p) {
		for v >= 0x80 {
			b = append(b, byte(v)|0x80)
			v >>= 7
		}
		b = append(b, byte(v))
	}
	b = LenPack(b, len(b)-sz0)
	return b
}

func UnsafeAppendPackedVarintU32(b []byte, p unsafe.Pointer) []byte {
	b = LenReserve(b)
	sz0 := len(b)
	for _, v := range *(*[]uint32)(p) {
		for v >= 0x80 {
			b = append(b, byte(v)|0x80)
			v >>= 7
		}
		b = append(b, byte(v))
	}
	b = LenPack(b, len(b)-sz0)
	return b
}

func UnsafeAppendPackedZigZag64(b []byte, p unsafe.Pointer) []byte {
	b = LenReserve(b)
	sz0 := len(b)
	for _, x := range *(*[]int64)(p) {
		v := uint64(x<<1) ^ uint64(x>>63)
		for v >= 0x80 {
			b = append(b, byte(v)|0x80)
			v >>= 7
		}
		b = append(b, byte(v))
	}
	b = LenPack(b, len(b)-sz0)
	return b
}

func UnsafeAppendPackedZigZag32(b []byte, p unsafe.Pointer) []byte {
	b = LenReserve(b)
	sz0 := len(b)
	for _, x := range *(*[]int32)(p) {
		v := uint32(x<<1) ^ uint32(x>>31)
		for v >= 0x80 {
			b = append(b, byte(v)|0x80)
			v >>= 7
		}
		b = append(b, byte(v))
	}
	b = LenPack(b, len(b)-sz0)
	return b
}

func UnsafeAppendPackedFixed64(b []byte, p unsafe.Pointer) []byte {
	vv := *(*[]uint64)(p)
	b = AppendVarint(b, uint64(8*len(vv)))
	for _, v := range vv {
		b = append(b, byte(v>>0),
			byte(v>>8),
			byte(v>>16),
			byte(v>>24),
			byte(v>>32),
			byte(v>>40),
			byte(v>>48),
			byte(v>>56))
	}
	return b
}

func UnsafeAppendPackedFixed32(b []byte, p unsafe.Pointer) []byte {
	vv := *(*[]uint32)(p)
	b = AppendVarint(b, uint64(4*len(vv)))
	for _, v := range vv {
		b = append(b, byte(v>>0),
			byte(v>>8),
			byte(v>>16),
			byte(v>>24))
	}
	return b
}

func UnsafeAppendPackedBool(b []byte, p unsafe.Pointer) []byte {
	vv := *(*[]byte)(p)
	b = AppendVarint(b, uint64(len(vv)))
	return append(b, vv...)
}
