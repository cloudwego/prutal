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
package compat

import (
	"testing"

	"github.com/cloudwego/prutal"
	"github.com/cloudwego/prutal/internal/testutils/assert"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

// TestPackedEitherWay sends fields declared unpacked in packed form and a
// packed one unpacked: the wire format allows either, protobuf-go takes
// both, and so must prutal.
func TestPackedEitherWay(t *testing.T) {
	b := protowire.AppendTag(nil, 18, protowire.BytesType) // unpacked_i32, packed
	b = protowire.AppendBytes(b, []byte{1, 2})
	b = protowire.AppendTag(b, 19, protowire.BytesType) // unpacked_color, packed
	b = protowire.AppendBytes(b, []byte{1})
	b = protowire.AppendTag(b, 20, protowire.BytesType) // unpacked_db, packed
	b = protowire.AppendBytes(b, []byte{0, 0, 0, 0, 0, 0, 0xf8, 0x3f})
	b = protowire.AppendTag(b, 21, protowire.BytesType) // unpacked_b, packed
	b = protowire.AppendBytes(b, []byte{1, 0})
	b = protowire.AppendTag(b, 1, protowire.VarintType) // i32, unpacked
	b = protowire.AppendVarint(b, 7)
	b = protowire.AppendTag(b, 1, protowire.VarintType)
	b = protowire.AppendVarint(b, 8)

	ref := &Repeateds{}
	assert.NoError(t, proto.Unmarshal(b, ref))
	m := &Repeateds{}
	assert.NoError(t, prutal.Unmarshal(b, m))
	assert.True(t, proto.Equal(ref, m))
	assert.SliceEqual(t, []int32{1, 2}, m.UnpackedI32)
	assert.SliceEqual(t, []Color{Color_RED}, m.UnpackedColor)
	assert.SliceEqual(t, []float64{1.5}, m.UnpackedDb)
	assert.SliceEqual(t, []bool{true, false}, m.UnpackedB)
	assert.SliceEqual(t, []int32{7, 8}, m.I32)
	for _, e := range check(m) { // the canonical output still agrees
		t.Error(e)
	}
}
