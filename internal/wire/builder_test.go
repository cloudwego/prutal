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
	"testing"

	"github.com/cloudwego/prutal/internal/protowire"
	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestBuilder(t *testing.T) {
	p := &Builder{}
	b := make([]byte, 0, 100)

	resetfn := func() {
		p.Reset()
		b = b[:0]
	}

	// AppendVarintField
	resetfn()
	b = protowire.AppendTag(b, 1, protowire.VarintType)
	b = protowire.AppendVarint(b, 129)
	assert.BytesEqual(t, b, p.AppendVarintField(1, 129).Bytes())

	// AppendZigZagField
	resetfn()
	b = protowire.AppendTag(b, 1, protowire.VarintType)
	b = protowire.AppendVarint(b, protowire.EncodeZigZag(-200))
	assert.BytesEqual(t, b, p.AppendZigZagField(1, -200).Bytes())

	// AppendFixed32Field
	resetfn()
	b = protowire.AppendTag(b, 1, protowire.Fixed32Type)
	b = protowire.AppendFixed32(b, 300)
	assert.BytesEqual(t, b, p.AppendFixed32Field(1, 300).Bytes())

	// AppendFixed64Field
	resetfn()
	b = protowire.AppendTag(b, 1, protowire.Fixed64Type)
	b = protowire.AppendFixed64(b, 300)
	assert.BytesEqual(t, b, p.AppendFixed64Field(1, 300).Bytes())

	// AppendStringField
	resetfn()
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendString(b, "hello")
	assert.BytesEqual(t, b, p.AppendStringField(1, "hello").Bytes())

	// AppendBytesField
	resetfn()
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendString(b, "hello")
	assert.BytesEqual(t, b, p.AppendBytesField(1, []byte("hello")).Bytes())

	// AppendPackedVarintField
	resetfn()
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendVarint(b, uint64(protowire.SizeVarint(100))+uint64(protowire.SizeVarint(1000)))
	b = protowire.AppendVarint(b, 100)
	b = protowire.AppendVarint(b, 1000)
	assert.BytesEqual(t, b, p.AppendPackedVarintField(1, 100, 1000).Bytes())

	// AppendPackedZigZagField
	resetfn()
	z1, z2 := protowire.EncodeZigZag(-100), protowire.EncodeZigZag(-1000)
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendVarint(b, uint64(protowire.SizeVarint(z1))+uint64(protowire.SizeVarint(z2)))
	b = protowire.AppendVarint(b, z1)
	b = protowire.AppendVarint(b, z2)
	assert.BytesEqual(t, b, p.AppendPackedZigZagField(1, -100, -1000).Bytes())

	// AppendPackedFixed32Field
	resetfn()
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendVarint(b, 2*4)
	b = protowire.AppendFixed32(b, 100)
	b = protowire.AppendFixed32(b, 1000)
	assert.BytesEqual(t, b, p.AppendPackedFixed32Field(1, 100, 1000).Bytes())

	// AppendPackedFixed64Field
	resetfn()
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendVarint(b, 2*8)
	b = protowire.AppendFixed64(b, 100)
	b = protowire.AppendFixed64(b, 1000)
	assert.BytesEqual(t, b, p.AppendPackedFixed64Field(1, 100, 1000).Bytes())

}
