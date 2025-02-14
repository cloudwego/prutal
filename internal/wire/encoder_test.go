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
	"strings"
	"testing"
	"unsafe"

	"github.com/cloudwego/prutal/internal/protowire"
	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestAppendVarint(t *testing.T) {
	v := uint64(1)
	for i := 0; i < 30; i++ {
		b0 := protowire.AppendVarint([]byte{}, v)
		b1 := UnsafeAppendVarintU64([]byte{}, unsafe.Pointer(&v))
		assert.BytesEqual(t, b0, b1)

		v32 := uint32(v)
		b0 = protowire.AppendVarint([]byte{}, uint64(v32))
		b1 = UnsafeAppendVarintU32([]byte{}, unsafe.Pointer(&v32))
		assert.BytesEqual(t, b0, b1)

		v *= 17
	}
}

func TestAppendZigZag(t *testing.T) {
	v := int64(-1)
	for i := 0; i < 30; i++ {
		b0 := protowire.AppendVarint([]byte{}, protowire.EncodeZigZag(v))
		b1 := UnsafeAppendZigZag64([]byte{}, unsafe.Pointer(&v))
		assert.BytesEqual(t, b0, b1)

		v32 := int32(v)
		b0 = protowire.AppendVarint([]byte{}, protowire.EncodeZigZag(int64(v32)))
		b1 = UnsafeAppendZigZag32([]byte{}, unsafe.Pointer(&v32))
		assert.BytesEqual(t, b0, b1)

		v *= 17
	}
}

func TestAppendFixed(t *testing.T) {
	v := uint64(1)
	for i := 0; i < 30; i++ {
		b0 := protowire.AppendFixed64([]byte{}, v)
		b1 := UnsafeAppendFixed64([]byte{}, unsafe.Pointer(&v))
		assert.BytesEqual(t, b0, b1)

		v32 := uint32(v)
		b0 = protowire.AppendFixed32([]byte{}, v32)
		b1 = UnsafeAppendFixed32([]byte{}, unsafe.Pointer(&v32))
		assert.BytesEqual(t, b0, b1)

		v *= 17
	}
}

func TestAppendBool(t *testing.T) {
	v := true
	b0 := protowire.AppendVarint([]byte{}, protowire.EncodeBool(v))
	b1 := UnsafeAppendBool([]byte{}, unsafe.Pointer(&v))
	assert.BytesEqual(t, b0, b1)

	v = false
	b0 = protowire.AppendVarint([]byte{}, protowire.EncodeBool(v))
	b1 = UnsafeAppendBool([]byte{}, unsafe.Pointer(&v))
	assert.BytesEqual(t, b0, b1)
}

func TestAppendString(t *testing.T) {
	v := "hello"
	b0 := protowire.AppendString([]byte{}, v)
	b1 := UnsafeAppendString([]byte{}, unsafe.Pointer(&v))
	assert.BytesEqual(t, b0, b1)

	v = strings.Repeat("hello", 1000)
	b0 = protowire.AppendString([]byte{}, v)
	b1 = UnsafeAppendString([]byte{}, unsafe.Pointer(&v))
	assert.BytesEqual(t, b0, b1)
}

func TestReserveLen(t *testing.T) {
	n := 300
	b := make([]byte, n)
	for i := 0; i < 300; i++ {
		b[i] = 'a' + byte(i)%26
	}
	for i := 0; i < 300; i++ {
		teststr := string(b[:i])
		b0 := LenReserve([]byte{})
		x := append(b0, teststr...)
		b0 = LenPack(x, len(x)-len(b0))
		b1 := protowire.AppendString([]byte{}, teststr)
		assert.BytesEqual(t, b1, b0)
	}
}
