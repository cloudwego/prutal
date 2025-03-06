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
	"testing"
	"unsafe"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestUnsafeAppendVarintU64List(t *testing.T) {
	vv := []uint64{rand.Uint64(), rand.Uint64(), rand.Uint64()}
	p := &Builder{}
	b0 := p.AppendVarintField(1, vv[0]).
		AppendVarintField(1, vv[1]).
		AppendVarintField(1, vv[2]).Bytes()
	b1 := UnsafeAppendVarintU64List(nil, 1, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendVarintU32List(t *testing.T) {
	vv := []uint32{rand.Uint32(), rand.Uint32(), rand.Uint32()}
	p := &Builder{}
	b0 := p.AppendVarintField(1, uint64(vv[0])).
		AppendVarintField(1, uint64(vv[1])).
		AppendVarintField(1, uint64(vv[2])).Bytes()
	b1 := UnsafeAppendVarintU32List(nil, 1, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendZigZag64List(t *testing.T) {
	vv := []int64{-1, 0, 1}
	p := &Builder{}
	b0 := p.AppendZigZagField(1, vv[0]).
		AppendZigZagField(1, vv[1]).
		AppendZigZagField(1, vv[2]).Bytes()
	b1 := UnsafeAppendZigZag64List(nil, 1, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendZigZag32List(t *testing.T) {
	vv := []int32{-1, 0, 1}
	p := &Builder{}
	b0 := p.AppendZigZagField(1, int64(vv[0])).
		AppendZigZagField(1, int64(vv[1])).
		AppendZigZagField(1, int64(vv[2])).Bytes()
	b1 := UnsafeAppendZigZag32List(nil, 1, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendFixed64List(t *testing.T) {
	vv := []uint64{rand.Uint64(), rand.Uint64(), rand.Uint64()}
	p := &Builder{}
	b0 := p.AppendFixed64Field(1, vv[0]).
		AppendFixed64Field(1, vv[1]).
		AppendFixed64Field(1, vv[2]).Bytes()
	b1 := UnsafeAppendFixed64List(nil, 1, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendFixed32List(t *testing.T) {
	vv := []uint32{rand.Uint32(), rand.Uint32(), rand.Uint32()}
	p := &Builder{}
	b0 := p.AppendFixed32Field(1, vv[0]).
		AppendFixed32Field(1, vv[1]).
		AppendFixed32Field(1, vv[2]).Bytes()
	b1 := UnsafeAppendFixed32List(nil, 1, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendBool(t *testing.T) {
	vv := []bool{true, false, true}
	b0 := []byte{
		1 << 3, 1,
		1 << 3, 0,
		1 << 3, 1}
	b1 := UnsafeAppendBoolList(nil, 1, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendStringList(t *testing.T) {
	vv := []string{"s1", "s2", "s3"}
	p := &Builder{}
	b0 := p.AppendStringField(1, vv[0]).
		AppendStringField(1, vv[1]).
		AppendStringField(1, vv[2]).Bytes()
	b1 := UnsafeAppendStringList(nil, 1, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}

func TestUnsafeAppendBytesList(t *testing.T) {
	vv := [][]byte{[]byte("s1"), []byte("s2"), []byte("s3")}
	p := &Builder{}
	b0 := p.AppendBytesField(1, vv[0]).
		AppendBytesField(1, vv[1]).
		AppendBytesField(1, vv[2]).Bytes()
	b1 := UnsafeAppendBytesList(nil, 1, unsafe.Pointer(&vv))
	assert.BytesEqual(t, b0, b1)
}
