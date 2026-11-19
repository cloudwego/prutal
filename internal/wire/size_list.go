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
	"unsafe"

	"github.com/cloudwego/prutal/internal/protowire"
)

var sizePackedFuncs = map[CoderType]SizeRepeatedFunc{
	CoderVarintU32: UnsafeSizePackedVarintU32,
	CoderVarintI32: UnsafeSizePackedVarintI32,
	CoderVarint64:  UnsafeSizePackedVarintU64,
	CoderZigZag32:  UnsafeSizePackedZigZag32,
	CoderZigZag64:  UnsafeSizePackedZigZag64,
	CoderFixed32:   UnsafeSizePackedFixed32,
	CoderFixed64:   UnsafeSizePackedFixed64,
	CoderBool:      UnsafeSizePackedBool,
}

var sizeListFuncs = map[CoderType]SizeRepeatedFunc{
	CoderVarintU32: UnsafeSizeVarintU32List,
	CoderVarintI32: UnsafeSizeVarintI32List,
	CoderVarint64:  UnsafeSizeVarintU64List,
	CoderZigZag32:  UnsafeSizeZigZag32List,
	CoderZigZag64:  UnsafeSizeZigZag64List,
	CoderFixed32:   UnsafeSizeFixed32List,
	CoderFixed64:   UnsafeSizeFixed64List,
	CoderBool:      UnsafeSizeBoolList,
	CoderString:    UnsafeSizeStringList,
	CoderBytes:     UnsafeSizeBytesList,
}

// GetSizeListFunc returns the specialized size function for a scalar list field.
// All returned functions return 0 for empty slices (SizeStruct skips empty
// slices earlier, while direct callers still get wire-accurate sizes).
func GetSizeListFunc(t CoderType, packed bool) SizeRepeatedFunc {
	if packed {
		return sizePackedFuncs[t]
	}
	return sizeListFuncs[t]
}

// UnsafeSizePackedVarintU64 returns the wire size of a packed uint64/int64 varint slice.
func UnsafeSizePackedVarintU64(p unsafe.Pointer, wireTagSize int) int {
	vv := *(*[]uint64)(p)
	if len(vv) == 0 {
		return 0
	}
	content := 0
	for _, v := range vv {
		content += protowire.SizeVarint(v)
	}
	return wireTagSize + protowire.SizeBytes(content)
}

// UnsafeSizePackedVarintI32 returns the wire size of a packed signed int32 varint slice.
func UnsafeSizePackedVarintI32(p unsafe.Pointer, wireTagSize int) int {
	vv := *(*[]int32)(p)
	if len(vv) == 0 {
		return 0
	}
	content := 0
	for _, v := range vv {
		content += protowire.SizeVarint(uint64(int64(v)))
	}
	return wireTagSize + protowire.SizeBytes(content)
}

// UnsafeSizePackedVarintU32 returns the wire size of a packed uint32 varint slice.
func UnsafeSizePackedVarintU32(p unsafe.Pointer, wireTagSize int) int {
	vv := *(*[]uint32)(p)
	if len(vv) == 0 {
		return 0
	}
	content := 0
	for _, v := range vv {
		content += protowire.SizeVarint(uint64(v))
	}
	return wireTagSize + protowire.SizeBytes(content)
}

// UnsafeSizePackedZigZag64 returns the wire size of a packed sint64 slice.
func UnsafeSizePackedZigZag64(p unsafe.Pointer, wireTagSize int) int {
	vv := *(*[]int64)(p)
	if len(vv) == 0 {
		return 0
	}
	content := 0
	for _, v := range vv {
		content += protowire.SizeVarint(uint64(v<<1) ^ uint64(v>>63))
	}
	return wireTagSize + protowire.SizeBytes(content)
}

// UnsafeSizePackedZigZag32 returns the wire size of a packed sint32 slice.
func UnsafeSizePackedZigZag32(p unsafe.Pointer, wireTagSize int) int {
	vv := *(*[]int32)(p)
	if len(vv) == 0 {
		return 0
	}
	content := 0
	for _, v := range vv {
		content += protowire.SizeVarint(uint64(uint32(v<<1) ^ uint32(v>>31)))
	}
	return wireTagSize + protowire.SizeBytes(content)
}

// UnsafeSizePackedFixed64 returns the wire size of a packed 64-bit fixed slice.
func UnsafeSizePackedFixed64(p unsafe.Pointer, wireTagSize int) int {
	n := len(*(*[]uint64)(p))
	if n == 0 {
		return 0
	}
	return wireTagSize + protowire.SizeBytes(n*8)
}

// UnsafeSizePackedFixed32 returns the wire size of a packed 32-bit fixed slice.
func UnsafeSizePackedFixed32(p unsafe.Pointer, wireTagSize int) int {
	n := len(*(*[]uint32)(p))
	if n == 0 {
		return 0
	}
	return wireTagSize + protowire.SizeBytes(n*4)
}

// UnsafeSizePackedBool returns the wire size of a packed bool slice.
func UnsafeSizePackedBool(p unsafe.Pointer, wireTagSize int) int {
	n := len(*(*[]bool)(p))
	if n == 0 {
		return 0
	}
	return wireTagSize + protowire.SizeBytes(n)
}

// UnsafeSizeVarintU64List returns the wire size of an unpacked uint64/int64 varint slice.
func UnsafeSizeVarintU64List(p unsafe.Pointer, wireTagSize int) int {
	vv := *(*[]uint64)(p)
	n := len(vv) * wireTagSize
	for _, v := range vv {
		n += protowire.SizeVarint(v)
	}
	return n
}

// UnsafeSizeVarintI32List returns the wire size of an unpacked signed int32 varint slice.
func UnsafeSizeVarintI32List(p unsafe.Pointer, wireTagSize int) int {
	vv := *(*[]int32)(p)
	n := len(vv) * wireTagSize
	for _, v := range vv {
		n += protowire.SizeVarint(uint64(int64(v)))
	}
	return n
}

// UnsafeSizeVarintU32List returns the wire size of an unpacked uint32 varint slice.
func UnsafeSizeVarintU32List(p unsafe.Pointer, wireTagSize int) int {
	vv := *(*[]uint32)(p)
	n := len(vv) * wireTagSize
	for _, v := range vv {
		n += protowire.SizeVarint(uint64(v))
	}
	return n
}

// UnsafeSizeZigZag64List returns the wire size of an unpacked sint64 slice.
func UnsafeSizeZigZag64List(p unsafe.Pointer, wireTagSize int) int {
	vv := *(*[]int64)(p)
	n := len(vv) * wireTagSize
	for _, v := range vv {
		n += protowire.SizeVarint(uint64(v<<1) ^ uint64(v>>63))
	}
	return n
}

// UnsafeSizeZigZag32List returns the wire size of an unpacked sint32 slice.
func UnsafeSizeZigZag32List(p unsafe.Pointer, wireTagSize int) int {
	vv := *(*[]int32)(p)
	n := len(vv) * wireTagSize
	for _, v := range vv {
		n += protowire.SizeVarint(uint64(uint32(v<<1) ^ uint32(v>>31)))
	}
	return n
}

// UnsafeSizeFixed64List returns the wire size of an unpacked 64-bit fixed slice.
func UnsafeSizeFixed64List(p unsafe.Pointer, wireTagSize int) int {
	return len(*(*[]uint64)(p)) * (wireTagSize + 8)
}

// UnsafeSizeFixed32List returns the wire size of an unpacked 32-bit fixed slice.
func UnsafeSizeFixed32List(p unsafe.Pointer, wireTagSize int) int {
	return len(*(*[]uint32)(p)) * (wireTagSize + 4)
}

// UnsafeSizeBoolList returns the wire size of an unpacked bool slice.
func UnsafeSizeBoolList(p unsafe.Pointer, wireTagSize int) int {
	return len(*(*[]bool)(p)) * (wireTagSize + 1)
}

// UnsafeSizeStringList returns the wire size of a string slice.
func UnsafeSizeStringList(p unsafe.Pointer, wireTagSize int) int {
	vv := *(*[]string)(p)
	n := len(vv) * wireTagSize
	for _, v := range vv {
		n += protowire.SizeBytes(len(v))
	}
	return n
}

// UnsafeSizeBytesList returns the wire size of a [][]byte slice.
func UnsafeSizeBytesList(p unsafe.Pointer, wireTagSize int) int {
	vv := *(*[][]byte)(p)
	n := len(vv) * wireTagSize
	for _, v := range vv {
		n += protowire.SizeBytes(len(v))
	}
	return n
}
