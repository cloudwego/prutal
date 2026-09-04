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

package prutal

import (
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
	"github.com/cloudwego/prutal/internal/wire"
)

type varintMsg struct {
	Bool        bool            `protobuf:"varint,1,opt"`
	OptBool     *bool           `protobuf:"varint,2,opt"`
	PackedBools []bool          `protobuf:"varint,3,rep,packed"`
	MapStrBool  map[string]bool `protobuf:"bytes,4,rep" protobuf_key:"bytes,1,opt" protobuf_val:"varint,2,opt"`
	MapBoolStr  map[bool]string `protobuf:"bytes,5,rep" protobuf_key:"varint,1,opt" protobuf_val:"bytes,2,opt"`
	MapBoolBool map[bool]bool   `protobuf:"bytes,6,rep" protobuf_key:"varint,1,opt" protobuf_val:"varint,2,opt"`

	Sint32        int32            `protobuf:"zigzag32,11,opt"`
	PackedSint32s []int32          `protobuf:"zigzag32,12,rep,packed"`
	MapStrSint32  map[string]int32 `protobuf:"bytes,13,rep" protobuf_key:"bytes,1,opt" protobuf_val:"zigzag32,2,opt"`
	MapSint32Str  map[int32]string `protobuf:"bytes,14,rep" protobuf_key:"zigzag32,1,opt" protobuf_val:"bytes,2,opt"`
	MapSint32     map[int32]int32  `protobuf:"bytes,15,rep" protobuf_key:"zigzag32,1,opt" protobuf_val:"zigzag32,2,opt"`
}

// Any non-zero varint is true; encoders are not obliged to write 1. Reading
// only the low bit turns 2 into false, which is how a field that used to be an
// int32 silently changes value after being redeclared as a bool.
func TestDecodeNonCanonicalBool(t *testing.T) {
	const raw = 2

	b := (&wire.Builder{}).
		AppendVarintField(1, raw).
		AppendVarintField(2, raw).
		AppendPackedVarintField(3, raw).
		AppendBytesField(4, (&wire.Builder{}).AppendStringField(1, "k").AppendVarintField(2, raw).Bytes()).
		AppendBytesField(5, (&wire.Builder{}).AppendVarintField(1, raw).AppendStringField(2, "v").Bytes()).
		AppendBytesField(6, (&wire.Builder{}).AppendVarintField(1, raw).AppendVarintField(2, raw).Bytes()).
		Bytes()

	v := &varintMsg{}
	assert.NoError(t, Unmarshal(b, v))
	assert.True(t, v.Bool)
	assert.True(t, *v.OptBool)
	assert.SliceEqual(t, []bool{true}, v.PackedBools)
	assert.MapEqual(t, map[string]bool{"k": true}, v.MapStrBool)
	assert.MapEqual(t, map[bool]string{true: "v"}, v.MapBoolStr)
	assert.MapEqual(t, map[bool]bool{true: true}, v.MapBoolBool)
}

// A sint32 takes the low 32 bits of its varint and zigzag decodes those. Doing
// it the other way round -- decode the full varint, then narrow -- gives a
// different number for any value wider than 32 bits.
func TestDecodeWideSint32(t *testing.T) {
	const (
		raw  = 34093304179
		want = int32(-2014266554)
	)

	b := (&wire.Builder{}).
		AppendVarintField(11, raw).
		AppendPackedVarintField(12, raw).
		AppendBytesField(13, (&wire.Builder{}).AppendStringField(1, "k").AppendVarintField(2, raw).Bytes()).
		AppendBytesField(14, (&wire.Builder{}).AppendVarintField(1, raw).AppendStringField(2, "v").Bytes()).
		AppendBytesField(15, (&wire.Builder{}).AppendVarintField(1, raw).AppendVarintField(2, raw).Bytes()).
		Bytes()

	v := &varintMsg{}
	assert.NoError(t, Unmarshal(b, v))
	assert.Equal(t, want, v.Sint32)
	assert.SliceEqual(t, []int32{want}, v.PackedSint32s)
	assert.MapEqual(t, map[string]int32{"k": want}, v.MapStrSint32)
	assert.MapEqual(t, map[int32]string{want: "v"}, v.MapSint32Str)
	assert.MapEqual(t, map[int32]int32{want: want}, v.MapSint32)
}
