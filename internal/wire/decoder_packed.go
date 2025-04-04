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

func DecodePackedVarintU64(b []byte, h unsafe.Pointer) error {
	sz := 0
	for _, v := range b {
		if v < 0x80 {
			sz++
		}
	}
	vv := make([]uint64, 0, sz)
	for len(b) > 0 {
		var v uint64
		var n int
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
			return io.ErrUnexpectedEOF
		}
		vv = append(vv, v)
		b = b[n:]
	}
	*(*[]uint64)(h) = vv
	return nil
}

func DecodePackedVarintU32(b []byte, h unsafe.Pointer) error {
	sz := 0
	for _, v := range b {
		if v < 0x80 {
			sz++
		}
	}
	vv := make([]uint32, 0, sz)
	for len(b) > 0 {
		var v uint64
		var n int
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
			return io.ErrUnexpectedEOF
		}
		vv = append(vv, uint32(v))
		b = b[n:]
	}
	*(*[]uint32)(h) = vv
	return nil
}

func DecodePackedZigZag64(b []byte, h unsafe.Pointer) error {
	sz := 0
	for _, v := range b {
		if v < 0x80 {
			sz++
		}
	}
	vv := make([]int64, 0, sz)
	for len(b) > 0 {
		var v uint64
		var n int
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
			return io.ErrUnexpectedEOF
		}
		vv = append(vv, protowire.DecodeZigZag(v))
		b = b[n:]
	}
	*(*[]int64)(h) = vv
	return nil
}

func DecodePackedZigZag32(b []byte, h unsafe.Pointer) error {
	sz := 0
	for _, v := range b {
		if v < 0x80 {
			sz++
		}
	}
	vv := make([]int32, 0, sz)
	for len(b) > 0 {
		var v uint64
		var n int
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
			return io.ErrUnexpectedEOF
		}
		vv = append(vv, int32(protowire.DecodeZigZag(v)))
		b = b[n:]
	}
	*(*[]int32)(h) = vv
	return nil
}

func DecodePackedFixed64(b []byte, h unsafe.Pointer) error {
	if len(b)&7 != 0 {
		return io.ErrUnexpectedEOF
	}
	vv := make([]uint64, 0, len(b)/8)
	for len(b) > 0 {
		_ = b[7]
		vv = append(vv,
			uint64(b[0])<<0|uint64(b[1])<<8|uint64(b[2])<<16|uint64(b[3])<<24|
				uint64(b[4])<<32|uint64(b[5])<<40|uint64(b[6])<<48|uint64(b[7])<<56)
		b = b[8:]
	}
	*(*[]uint64)(h) = vv
	return nil
}

func DecodePackedFixed32(b []byte, h unsafe.Pointer) error {
	if len(b)&3 != 0 {
		return io.ErrUnexpectedEOF
	}
	vv := make([]uint32, 0, len(b)/4)
	for len(b) > 0 {
		_ = b[3]
		vv = append(vv, uint32(b[0])<<0|uint32(b[1])<<8|uint32(b[2])<<16|uint32(b[3])<<24)
		b = b[4:]
	}
	*(*[]uint32)(h) = vv
	return nil
}

func DecodePackedBool(b []byte, h unsafe.Pointer) error {
	vv := make([]bool, len(b))
	for i, c := range b {
		vv[i] = c > 0
	}
	*(*[]bool)(h) = vv
	return nil
}
