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
	"unsafe"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestDecodePackedVarintU64(t *testing.T) {
	vv0 := []uint64{1, 1 << 12, 1 << 22}
	p := NewBuilder()
	defer p.Free()
	b := p.AppendVarintU64(vv0...).Bytes()

	vv1 := []uint64{}
	assert.NoError(t, DecodePackedVarintU64(b, unsafe.Pointer(&vv1)))
	assert.SliceEqual(t, vv0, vv1)
}

func TestDecodePackedVarintU32(t *testing.T) {
	vv0 := []uint32{1, 1 << 12, 1 << 22}
	p := NewBuilder()
	defer p.Free()
	b := p.AppendVarintU32(vv0...).Bytes()

	vv1 := []uint32{}
	assert.NoError(t, DecodePackedVarintU32(b, unsafe.Pointer(&vv1)))
	assert.SliceEqual(t, vv0, vv1)
}

func TestDecodePackedZigZag64(t *testing.T) {
	vv0 := []int64{-1, 0, 1 << 12, 1 << 22}
	p := NewBuilder()
	defer p.Free()

	b := p.AppendZigZag64(vv0...).Bytes()

	vv1 := []int64{}
	assert.NoError(t, DecodePackedZigZag64(b, unsafe.Pointer(&vv1)))
	assert.SliceEqual(t, vv0, vv1)
}

func TestDecodePackedZigZag32(t *testing.T) {
	vv0 := []int32{-1, 0, 1 << 12, 1 << 22}
	p := NewBuilder()
	defer p.Free()

	b := p.AppendZigZag32(vv0...).Bytes()

	vv1 := []int32{}
	assert.NoError(t, DecodePackedZigZag32(b, unsafe.Pointer(&vv1)))
	assert.SliceEqual(t, vv0, vv1)
}

func TestDecodePackedFixed64(t *testing.T) {
	vv0 := []uint64{1, 1 << 12, 1 << 22}
	p := NewBuilder()
	defer p.Free()
	b := p.AppendFixed64(vv0...).Bytes()

	vv1 := []uint64{}
	assert.NoError(t, DecodePackedFixed64(b, unsafe.Pointer(&vv1)))
	assert.SliceEqual(t, vv0, vv1)
}

func TestDecodePackedFixed32(t *testing.T) {
	vv0 := []uint32{1, 1 << 12, 1 << 22}
	p := NewBuilder()
	defer p.Free()
	b := p.AppendFixed32(vv0...).Bytes()

	vv1 := []uint32{}
	assert.NoError(t, DecodePackedFixed32(b, unsafe.Pointer(&vv1)))
	assert.SliceEqual(t, vv0, vv1)
}

func TestDecodePackedBool(t *testing.T) {
	vv0 := []bool{true, false, true, false}
	b := []byte{1, 0, 1, 0}

	vv1 := []bool{}
	assert.NoError(t, DecodePackedBool(b, unsafe.Pointer(&vv1)))
	assert.SliceEqual(t, vv0, vv1)

}
