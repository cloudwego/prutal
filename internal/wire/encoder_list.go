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

type AppendListFunc func(b []byte, id int32, p unsafe.Pointer) []byte

var appendListFuncs = map[CoderType]AppendListFunc{
	CoderVarint32: UnsafeAppendVarintU32List,
	CoderVarint64: UnsafeAppendVarintU64List,
	CoderZigZag32: UnsafeAppendZigZag32List,
	CoderZigZag64: UnsafeAppendZigZag64List,
	CoderFixed32:  UnsafeAppendFixed32List,
	CoderFixed64:  UnsafeAppendFixed64List,
	CoderBool:     UnsafeAppendBoolList,
}

func GetAppendListFunc(t CoderType) AppendListFunc {
	return appendListFuncs[t]
}

func UnsafeAppendVarintU64List(b []byte, id int32, p unsafe.Pointer) []byte {
	for _, v := range *(*[]uint64)(p) {
		b = AppendVarint(b, EncodeTag(id, TypeVarint))
		for v >= 0x80 {
			b = append(b, byte(v)|0x80)
			v >>= 7
		}
		b = append(b, byte(v))
	}
	return b
}

func UnsafeAppendVarintU32List(b []byte, id int32, p unsafe.Pointer) []byte {
	for _, v := range *(*[]uint32)(p) {
		b = AppendVarint(b, EncodeTag(id, TypeVarint))
		for v >= 0x80 {
			b = append(b, byte(v)|0x80)
			v >>= 7
		}
		b = append(b, byte(v))
	}
	return b
}

func UnsafeAppendZigZag64List(b []byte, id int32, p unsafe.Pointer) []byte {
	for _, x := range *(*[]int64)(p) {
		b = AppendVarint(b, EncodeTag(id, TypeVarint))
		v := uint64(x<<1) ^ uint64(x>>63)
		for v >= 0x80 {
			b = append(b, byte(v)|0x80)
			v >>= 7
		}
		b = append(b, byte(v))
	}
	return b
}

func UnsafeAppendZigZag32List(b []byte, id int32, p unsafe.Pointer) []byte {
	for _, x := range *(*[]int32)(p) {
		b = AppendVarint(b, EncodeTag(id, TypeVarint))
		v := uint32(x<<1) ^ uint32(x>>31)
		for v >= 0x80 {
			b = append(b, byte(v)|0x80)
			v >>= 7
		}
		b = append(b, byte(v))
	}
	return b
}

func UnsafeAppendFixed64List(b []byte, id int32, p unsafe.Pointer) []byte {
	for _, v := range *(*[]uint64)(p) {
		b = AppendVarint(b, EncodeTag(id, TypeFixed64))
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

func UnsafeAppendFixed32List(b []byte, id int32, p unsafe.Pointer) []byte {
	for _, v := range *(*[]uint32)(p) {
		b = AppendVarint(b, EncodeTag(id, TypeFixed32))
		b = append(b, byte(v>>0),
			byte(v>>8),
			byte(v>>16),
			byte(v>>24))
	}
	return b
}

func UnsafeAppendBoolList(b []byte, id int32, p unsafe.Pointer) []byte {
	for _, v := range *(*[]byte)(p) {
		b = AppendVarint(b, EncodeTag(id, TypeVarint))
		b = append(b, v)
	}
	return b
}

func UnsafeAppendStringList(b []byte, id int32, p unsafe.Pointer) []byte {
	for _, v := range *(*[]string)(p) {
		b = AppendVarint(b, EncodeTag(id, TypeBytes))
		b = AppendVarint(b, uint64(len(v)))
		b = append(b, v...)
	}
	return b
}

func UnsafeAppendBytesList(b []byte, id int32, p unsafe.Pointer) []byte {
	for _, v := range *(*[][]byte)(p) {
		b = AppendVarint(b, EncodeTag(id, TypeBytes))
		b = AppendVarint(b, uint64(len(v)))
		b = append(b, v...)
	}
	return b
}
