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
	"math/rand"
	"unsafe"

	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestUnsafeAppendPackedVarintU64(t *testing.T) {
	vv := []uint64{rand.Uint64(), rand.Uint64(), rand.Uint64()}
	p := &Builder{}
	b0 := p.AppendPackedVarintField(1, vv...).Bytes()[1:] // skip tag
	b1 := UnsafeAppendPackedVarintU64(nil, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendPackedVarintU32(t *testing.T) {
	vv := []uint32{rand.Uint32(), rand.Uint32(), rand.Uint32()}
	p := &Builder{}
	b0 := p.AppendPackedVarintField(1, toUint64s(vv)...).Bytes()[1:] // skip tag
	b1 := UnsafeAppendPackedVarintU32(nil, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendPackedZigZag64(t *testing.T) {
	vv := []int64{rand.Int63(), rand.Int63(), rand.Int63()}
	p := &Builder{}
	b0 := p.AppendPackedZigZagField(1, vv...).Bytes()[1:] // skip tag
	b1 := UnsafeAppendPackedZigZag64(nil, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendPackedZigZag32(t *testing.T) {
	vv := []int32{rand.Int31(), rand.Int31(), rand.Int31()}
	p := &Builder{}
	b0 := p.AppendPackedZigZagField(1, toInt64s(vv)...).Bytes()[1:] // skip tag
	b1 := UnsafeAppendPackedZigZag32(nil, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendPackedFixed64(t *testing.T) {
	vv := []uint64{rand.Uint64(), rand.Uint64(), rand.Uint64()}
	p := &Builder{}
	b0 := p.AppendPackedFixed64Field(1, vv...).Bytes()[1:] // skip tag
	b1 := UnsafeAppendPackedFixed64(nil, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendPackedFixed32(t *testing.T) {
	vv := []uint32{rand.Uint32(), rand.Uint32(), rand.Uint32()}
	p := &Builder{}
	b0 := p.AppendPackedFixed32Field(1, vv...).Bytes()[1:] // skip tag
	b1 := UnsafeAppendPackedFixed32(nil, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendPackedBool(t *testing.T) {
	vv := []bool{true, false, true}
	b0 := []byte{3, 1, 0, 1}
	b1 := UnsafeAppendPackedBool(nil, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}
