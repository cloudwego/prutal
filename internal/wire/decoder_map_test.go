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

// Map entry: key (field 1) = 1 (canonical), value (field 2) = the over-long
// varint 0x80 0x00 (= 0, i.e. false). The pre-fix code read only the first
// byte (0x80 != 0) and wrongly returned true. Bool map keys/values are
// varints on the wire, not single bytes.
func TestDecodeMap_Bool_Bool(t *testing.T) {
	b := []byte{0x08, 0x01, 0x10, 0x80, 0x00}

	m := map[bool]bool{}
	assert.NoError(t, DecodeMap_Bool_Bool(b, unsafe.Pointer(&m)))
	assert.MapEqual(t, map[bool]bool{true: false}, m)
}

// A map entry is a message with two optional fields, so the decoders must
// accept anything the wire format allows and not only the canonical key-then-
// value form written by encoder_map.go.
func TestDecodeMapEntryShapes(t *testing.T) {
	t.Run("no value", func(t *testing.T) {
		m := map[int32]int32{}
		b := (&Builder{}).AppendVarintField(1, 7).Bytes()
		assert.NoError(t, DecodeMap_VarintU32_VarintU32(b, unsafe.Pointer(&m)))
		assert.MapEqual(t, map[int32]int32{7: 0}, m)
	})

	t.Run("no key", func(t *testing.T) {
		m := map[int32]int32{}
		b := (&Builder{}).AppendVarintField(2, 9).Bytes()
		assert.NoError(t, DecodeMap_VarintU32_VarintU32(b, unsafe.Pointer(&m)))
		assert.MapEqual(t, map[int32]int32{0: 9}, m)
	})

	t.Run("empty entry", func(t *testing.T) {
		m := map[int32]int32{}
		assert.NoError(t, DecodeMap_VarintU32_VarintU32(nil, unsafe.Pointer(&m)))
		assert.MapEqual(t, map[int32]int32{0: 0}, m)
	})

	t.Run("value before key", func(t *testing.T) {
		m := map[int32]int32{}
		b := (&Builder{}).AppendVarintField(2, 9).AppendVarintField(1, 7).Bytes()
		assert.NoError(t, DecodeMap_VarintU32_VarintU32(b, unsafe.Pointer(&m)))
		assert.MapEqual(t, map[int32]int32{7: 9}, m)
	})

	t.Run("unknown entry fields", func(t *testing.T) {
		m := map[int32]int32{}
		b := (&Builder{}).AppendStringField(5, "x").
			AppendVarintField(1, 7).
			AppendFixed64Field(3, 1).
			AppendVarintField(2, 9).Bytes()
		assert.NoError(t, DecodeMap_VarintU32_VarintU32(b, unsafe.Pointer(&m)))
		assert.MapEqual(t, map[int32]int32{7: 9}, m)
	})

	// A key or value tagged with an unexpected wire type is skipped like any
	// other unknown field, same as the reference implementation.
	t.Run("wrong wire types", func(t *testing.T) {
		m := map[int32]int32{}
		b := (&Builder{}).AppendFixed32Field(1, 7).AppendStringField(2, "x").Bytes()
		assert.NoError(t, DecodeMap_VarintU32_VarintU32(b, unsafe.Pointer(&m)))
		assert.MapEqual(t, map[int32]int32{0: 0}, m)
	})

	t.Run("repeated fields keep the last", func(t *testing.T) {
		m := map[int32]int32{}
		b := (&Builder{}).AppendVarintField(1, 1).AppendVarintField(2, 2).
			AppendVarintField(1, 3).AppendVarintField(2, 4).Bytes()
		assert.NoError(t, DecodeMap_VarintU32_VarintU32(b, unsafe.Pointer(&m)))
		assert.MapEqual(t, map[int32]int32{3: 4}, m)
	})

	t.Run("fixed key and value", func(t *testing.T) {
		m := map[uint32]uint64{}
		b := (&Builder{}).AppendFixed64Field(2, 9).AppendFixed32Field(1, 7).Bytes()
		assert.NoError(t, DecodeMap_Fixed32_Fixed64(b, unsafe.Pointer(&m)))
		assert.MapEqual(t, map[uint32]uint64{7: 9}, m)
	})

	t.Run("bool key and value", func(t *testing.T) {
		m := map[bool]bool{}
		b := (&Builder{}).AppendVarintField(2, 1).Bytes()
		assert.NoError(t, DecodeMap_Bool_Bool(b, unsafe.Pointer(&m)))
		assert.MapEqual(t, map[bool]bool{false: true}, m)
	})

	// Entry tags are canonically one byte, but an over-long tag varint is
	// legal and must still be recognised as the key or the value.
	t.Run("over-long entry tags", func(t *testing.T) {
		m := map[int32]int32{}
		// 0x88 0x00 = field 1 varint, 0x90 0x00 = field 2 varint
		b := []byte{0x88, 0x00, 7, 0x90, 0x00, 9}
		assert.NoError(t, DecodeMap_VarintU32_VarintU32(b, unsafe.Pointer(&m)))
		assert.MapEqual(t, map[int32]int32{7: 9}, m)
	})

	// An unknown field with a two-byte tag must be skipped by its real length.
	t.Run("unknown field with multi-byte tag", func(t *testing.T) {
		m := map[int32]int32{}
		b := (&Builder{}).AppendVarintField(1, 7).
			AppendVarintField(2000, 1).
			AppendVarintField(2, 9).Bytes()
		assert.NoError(t, DecodeMap_VarintU32_VarintU32(b, unsafe.Pointer(&m)))
		assert.MapEqual(t, map[int32]int32{7: 9}, m)
	})
}

// The reference implementation range-checks the entry field number before
// using it; a number past the maximum makes the message invalid rather than
// contributing an unknown field.
func TestDecodeMapEntryFieldNumberTooLarge(t *testing.T) {
	m := map[int32]int32{}
	b := []byte{0x80, 0x80, 0x80, 0x80, 0x10, 0x01} // field 1<<29, varint
	assert.ErrorContains(t, DecodeMap_VarintU32_VarintU32(b, unsafe.Pointer(&m)), "invalid field number")
}

func TestDecodeMapEntryTruncated(t *testing.T) {
	m := map[int32]int32{}
	for _, b := range [][]byte{
		{0x08},             // key tag without payload
		{0x08, 0x80},       // truncated key varint
		{0x08, 0x01, 0x15}, // truncated fixed32 of an unknown field
		{0x0d, 0x01},       // truncated fixed32 for a mismatched key
		{0x08, 0x01, 0x80}, // truncated tag after a valid key
	} {
		if err := DecodeMap_VarintU32_VarintU32(b, unsafe.Pointer(&m)); err == nil {
			t.Errorf("entry %x: no error", b)
		}
	}
}
