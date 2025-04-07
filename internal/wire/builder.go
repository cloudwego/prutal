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

import "sync"

// Builder implements a wire format builder for testing
type Builder struct {
	b []byte
}

var poolBuilder = sync.Pool{
	New: func() any {
		return &Builder{}
	},
}

func NewBuilder() *Builder {
	p := poolBuilder.Get().(*Builder)
	p.Reset()
	return p
}

func (p *Builder) Free() {
	poolBuilder.Put(p)
}

func (p *Builder) Reset() *Builder {
	p.b = p.b[:0]
	return p
}

func (p *Builder) Bytes() []byte {
	// copy, in case caller will ref to the []byte
	// it's OK for testing
	return append([]byte{}, p.b...)
}

func xappendVarint(b []byte, v uint64) []byte {
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}
	return append(b, byte(v))
}

func (p *Builder) appendVarint(v uint64) *Builder {
	p.b = xappendVarint(p.b, v)
	return p
}

func (p *Builder) appendFixed32(v uint32) *Builder {
	p.b = append(p.b,
		byte(v>>0),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24))
	return p
}

func (p *Builder) appendFixed64(v uint64) *Builder {
	p.b = append(p.b,
		byte(v>>0),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
		byte(v>>32),
		byte(v>>40),
		byte(v>>48),
		byte(v>>56))
	return p
}

func (p *Builder) appendTag(num int, t Type) *Builder {
	return p.appendVarint(uint64(num)<<3 | uint64(t&7))
}

func (p *Builder) appendString(v string) *Builder {
	p.appendVarint(uint64(len(v)))
	p.b = append(p.b, v...)
	return p
}

func (p *Builder) appendBytes(v []byte) *Builder {
	p.appendVarint(uint64(len(v)))
	p.b = append(p.b, v...)
	return p
}

func (p *Builder) AppendVarintU64(vv ...uint64) *Builder {
	for _, v := range vv {
		p.appendVarint(v)
	}
	return p
}

func (p *Builder) AppendVarintU32(vv ...uint32) *Builder {
	for _, v := range vv {
		p.appendVarint(uint64(v))
	}
	return p
}

func (p *Builder) AppendZigZag32(vv ...int32) *Builder {
	for _, v := range vv {
		p.appendVarint(uint64(uint32(v<<1) ^ uint32(v>>31)))
	}
	return p
}

func (p *Builder) AppendZigZag64(vv ...int64) *Builder {
	for _, v := range vv {
		p.appendVarint(uint64(v<<1) ^ uint64(v>>63))
	}
	return p
}

func (p *Builder) AppendFixed64(vv ...uint64) *Builder {
	for _, v := range vv {
		p.appendFixed64(v)
	}
	return p
}

func (p *Builder) AppendFixed32(vv ...uint32) *Builder {
	for _, v := range vv {
		p.appendFixed32(v)
	}
	return p
}

func (p *Builder) AppendVarintField(num int, v uint64) *Builder {
	return p.appendTag(num, TypeVarint).appendVarint(v)
}

func (p *Builder) AppendZigZagField(num int, v int64) *Builder {
	return p.appendTag(num, TypeVarint).appendVarint(uint64(v<<1) ^ uint64(v>>63))
}

func (p *Builder) AppendFixed32Field(num int, v uint32) *Builder {
	return p.appendTag(num, TypeFixed32).appendFixed32(v)
}

func (p *Builder) AppendFixed64Field(num int, v uint64) *Builder {
	return p.appendTag(num, TypeFixed64).appendFixed64(v)
}

func (p *Builder) AppendStringField(num int, v string) *Builder {
	return p.appendTag(num, TypeBytes).appendString(v)
}

func (p *Builder) AppendBytesField(num int, v []byte) *Builder {
	return p.appendTag(num, TypeBytes).appendBytes(v)
}

func (p *Builder) AppendPackedVarintField(num int, vv ...uint64) *Builder {
	p.appendTag(num, TypeBytes)
	tmp := NewBuilder()
	defer tmp.Free()

	tmp.AppendVarintU64(vv...)
	return p.appendBytes(tmp.b)
}

func (p *Builder) AppendPackedZigZagField(num int, vv ...int64) *Builder {
	p.appendTag(num, TypeBytes)

	tmp := NewBuilder()
	defer tmp.Free()

	tmp.AppendZigZag64(vv...)
	return p.appendBytes(tmp.b)
}

func (p *Builder) AppendPackedFixed32Field(num int, vv ...uint32) *Builder {
	p.appendTag(num, TypeBytes).appendVarint(uint64(len(vv) * 4))
	for _, v := range vv {
		p.appendFixed32(v)
	}
	return p
}

func (p *Builder) AppendPackedFixed64Field(num int, vv ...uint64) *Builder {
	p.appendTag(num, TypeBytes).appendVarint(uint64(len(vv) * 8))
	for _, v := range vv {
		p.appendFixed64(v)
	}
	return p
}
