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

type SizeFunc func(p unsafe.Pointer) int

var sizeFuncs = map[CoderType]SizeFunc{
	CoderVarint32:  UnsafeSizeVarintU32,
	CoderVarintI32: UnsafeSizeVarintI32,
	CoderVarint64:  UnsafeSizeVarintU64,
	CoderZigZag32:  UnsafeSizeZigZag32,
	CoderZigZag64:  UnsafeSizeZigZag64,
	CoderFixed32:   UnsafeSizeFixed32,
	CoderFixed64:   UnsafeSizeFixed64,
	CoderString:    UnsafeSizeString,
	CoderBytes:     UnsafeSizeBytes,
	CoderBool:      UnsafeSizeBool,
}

func GetSizeFunc(t CoderType) SizeFunc {
	return sizeFuncs[t]
}

func UnsafeSizeVarintU64(p unsafe.Pointer) int {
	return protowire.SizeVarint(*(*uint64)(p))
}

func UnsafeSizeVarintI32(p unsafe.Pointer) int {
	return protowire.SizeVarint(uint64(int64(*(*int32)(p))))
}

func UnsafeSizeVarintU32(p unsafe.Pointer) int {
	return protowire.SizeVarint(uint64(*(*uint32)(p)))
}

func UnsafeSizeZigZag64(p unsafe.Pointer) int {
	x := *(*int64)(p)
	return protowire.SizeVarint(uint64(x<<1) ^ uint64(x>>63))
}

func UnsafeSizeZigZag32(p unsafe.Pointer) int {
	x := *(*int32)(p)
	return protowire.SizeVarint(uint64(uint32(x<<1) ^ uint32(x>>31)))
}

func UnsafeSizeFixed32(_ unsafe.Pointer) int { return 4 }

func UnsafeSizeFixed64(_ unsafe.Pointer) int { return 8 }

func UnsafeSizeString(p unsafe.Pointer) int {
	return protowire.SizeBytes(len(*(*string)(p)))
}

var UnsafeSizeBytes = UnsafeSizeString // same layout as string

func UnsafeSizeBool(_ unsafe.Pointer) int { return 1 }
