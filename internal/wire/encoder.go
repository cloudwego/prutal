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

func AppendVarint(b []byte, v uint64) []byte {
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}
	return append(b, byte(v))
}

func UnsafeAppendVarintU64(b []byte, p unsafe.Pointer) []byte {
	v := *(*uint64)(p)
	switch {
	case v < 1<<7:
		return append(b, byte(v))
	case v < 1<<14:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte(v>>7))
	case v < 1<<21:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte(v>>14))
	case v < 1<<28:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte(v>>21))
	case v < 1<<35:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte(v>>28))
	case v < 1<<42:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte(v>>35))
	case v < 1<<49:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte(v>>42))
	case v < 1<<56:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte((v>>42)&0x7f|0x80),
			byte(v>>49))
	case v < 1<<63:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte((v>>42)&0x7f|0x80),
			byte((v>>49)&0x7f|0x80),
			byte(v>>56))
	default:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte((v>>42)&0x7f|0x80),
			byte((v>>49)&0x7f|0x80),
			byte((v>>56)&0x7f|0x80),
			1)
	}
}

func UnsafeAppendVarintU32(b []byte, p unsafe.Pointer) []byte {
	v := *(*uint32)(p)
	switch {
	case v < 1<<7:
		return append(b, byte(v))
	case v < 1<<14:
		return append(b,
			byte((v>>0)|0x80),
			byte(v>>7))
	case v < 1<<21:
		return append(b,
			byte((v>>0)|0x80),
			byte((v>>7)|0x80),
			byte(v>>14))
	case v < 1<<28:
		return append(b,
			byte((v>>0)|0x80),
			byte((v>>7)|0x80),
			byte((v>>14)|0x80),
			byte(v>>21))
	default: // v < 1<<35:
		return append(b,
			byte((v>>0)|0x80),
			byte((v>>7)|0x80),
			byte((v>>14)|0x80),
			byte((v>>21)|0x80),
			byte(v>>28))
	}
}

func UnsafeAppendZigZag64(b []byte, p unsafe.Pointer) []byte {
	x := *(*int64)(p)
	v := uint64(x<<1) ^ uint64(x>>63)
	switch {
	case v < 1<<7:
		return append(b, byte(v))
	case v < 1<<14:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte(v>>7))
	case v < 1<<21:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte(v>>14))
	case v < 1<<28:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte(v>>21))
	case v < 1<<35:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte(v>>28))
	case v < 1<<42:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte(v>>35))
	case v < 1<<49:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte(v>>42))
	case v < 1<<56:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte((v>>42)&0x7f|0x80),
			byte(v>>49))
	case v < 1<<63:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte((v>>42)&0x7f|0x80),
			byte((v>>49)&0x7f|0x80),
			byte(v>>56))
	default:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte((v>>42)&0x7f|0x80),
			byte((v>>49)&0x7f|0x80),
			byte((v>>56)&0x7f|0x80),
			1)
	}
}

func UnsafeAppendZigZag32(b []byte, p unsafe.Pointer) []byte {
	x := *(*int32)(p)
	v := uint32(x<<1) ^ uint32(x>>31)
	switch {
	case v < 1<<7:
		return append(b, byte(v))
	case v < 1<<14:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte(v>>7))
	case v < 1<<21:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte(v>>14))
	case v < 1<<28:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte(v>>21))
	default:
		return append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte(v>>28))
	}
}

func UnsafeAppendBool(b []byte, p unsafe.Pointer) []byte {
	return append(b, *(*byte)(p)&0x1)
}

var UnsafeAppendBytes = UnsafeAppendString // should be the same

func UnsafeAppendString(b []byte, p unsafe.Pointer) []byte {
	s := *(*string)(p)
	b = AppendVarint(b, uint64(len(s)))
	return append(b, s...)
}

func UnsafeAppendFixed32(b []byte, p unsafe.Pointer) []byte {
	v := *(*uint32)(p)
	return append(b,
		byte(v>>0),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24))
}

func UnsafeAppendFixed64(b []byte, p unsafe.Pointer) []byte {
	v := *(*uint64)(p)
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

// LenReserve reserves one byte for LEN encoding.
// use `LenPack` after bytes are appended to the end of the []byte with the size of it
func LenReserve(b []byte) []byte {
	return append(b, 0)
}

func LenPack(b []byte, sz int) []byte {
	if sz < 128 {
		// fast path for most cases that can be inlined with cost 79
		b[len(b)-1-sz] = byte(sz) //  1 byte varint
		return b
	}
	return lenPackSlow(b, sz)
}

func lenPackSlow(b []byte, sz int) []byte {
	m := protowire.SizeVarint(uint64(sz))
	for i := m; i > 1; i-- {
		// reserved one byte, then m-1 bytes needed
		b = append(b, 0)
	}
	pos := len(b) - sz - m // pos varint
	copy(b[pos+m:], b[pos+1:])
	protowire.AppendVarint(b[pos:][:0], uint64(sz))
	return b
}
