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

// mapEntry wraps an entry body as a LEN record of the given map field.
func mapEntry(num int, body []byte) []byte {
	return (&wire.Builder{}).AppendBytesField(num, body).Bytes()
}

// A map entry may omit its key or value field; the missing one then takes the
// zero value instead of making the whole message unparseable.
//
// This covers the reflection decode path (string keys, bytes and message
// values); see internal/wire for the scalar fast path.
func TestDecodeMapEntryOptionalFields(t *testing.T) {
	const (
		fMapStringString = 901
		fMapStringBytes  = 902
		fMapStringStruct = 903
	)
	key := func() []byte { return (&wire.Builder{}).AppendStringField(1, "k").Bytes() }
	val := func() []byte { return (&wire.Builder{}).AppendStringField(2, "v").Bytes() }

	t.Run("no value", func(t *testing.T) {
		var v TestStruct
		assert.NoError(t, Unmarshal(mapEntry(fMapStringString, key()), &v))
		assert.MapEqual(t, map[string]string{"k": ""}, v.MapStringString)
	})

	t.Run("no key", func(t *testing.T) {
		var v TestStruct
		assert.NoError(t, Unmarshal(mapEntry(fMapStringString, val()), &v))
		assert.MapEqual(t, map[string]string{"": "v"}, v.MapStringString)
	})

	t.Run("empty entry", func(t *testing.T) {
		var v TestStruct
		assert.NoError(t, Unmarshal(mapEntry(fMapStringString, nil), &v))
		assert.MapEqual(t, map[string]string{"": ""}, v.MapStringString)
	})

	t.Run("value before key", func(t *testing.T) {
		var v TestStruct
		assert.NoError(t, Unmarshal(mapEntry(fMapStringString, append(val(), key()...)), &v))
		assert.MapEqual(t, map[string]string{"k": "v"}, v.MapStringString)
	})

	t.Run("unknown entry fields", func(t *testing.T) {
		b := (&wire.Builder{}).AppendVarintField(3, 7).Bytes()
		b = append(b, key()...)
		b = append(b, (&wire.Builder{}).AppendStringField(4, "ignored").Bytes()...)
		b = append(b, val()...)
		var v TestStruct
		assert.NoError(t, Unmarshal(mapEntry(fMapStringString, b), &v))
		assert.MapEqual(t, map[string]string{"k": "v"}, v.MapStringString)
	})

	// A key or value carrying an unexpected wire type is skipped like any
	// other unknown field, same as the reference implementation.
	t.Run("key with wrong wire type", func(t *testing.T) {
		b := (&wire.Builder{}).AppendVarintField(1, 7).Bytes()
		var v TestStruct
		assert.NoError(t, Unmarshal(mapEntry(fMapStringString, append(b, val()...)), &v))
		assert.MapEqual(t, map[string]string{"": "v"}, v.MapStringString)
	})

	t.Run("bytes value omitted", func(t *testing.T) {
		var v TestStruct
		assert.NoError(t, Unmarshal(mapEntry(fMapStringBytes, key()), &v))
		assert.Equal(t, 1, len(v.MapStringBytes))
		assert.Equal(t, 0, len(v.MapStringBytes["k"]))
	})

	// An omitted message value still yields an empty message, never a nil
	// pointer that callers would have to guard against.
	t.Run("message value omitted", func(t *testing.T) {
		var v TestStruct
		assert.NoError(t, Unmarshal(mapEntry(fMapStringStruct, key()), &v))
		assert.Equal(t, 1, len(v.MapStringStructA))
		assert.NotNil(t, v.MapStringStructA["k"])
		assert.Equal(t, uint64(0), v.MapStringStructA["k"].V)
	})
}

// Over-long entry tags are legal and must still be recognised as the key or
// the value; an unknown field with a multi-byte tag must be skipped by its
// real length.
func TestDecodeMapEntryLongTags(t *testing.T) {
	const fMapStringString = 901

	// 0x8a 0x00 = field 1 LEN, 0x92 0x00 = field 2 LEN
	b := []byte{0x8a, 0x00, 1, 'k', 0x92, 0x00, 1, 'v'}
	var v TestStruct
	assert.NoError(t, Unmarshal(mapEntry(fMapStringString, b), &v))
	assert.MapEqual(t, map[string]string{"k": "v"}, v.MapStringString)

	b = (&wire.Builder{}).AppendStringField(1, "k").
		AppendVarintField(2000, 1).
		AppendStringField(2, "v").Bytes()
	var v2 TestStruct
	assert.NoError(t, Unmarshal(mapEntry(fMapStringString, b), &v2))
	assert.MapEqual(t, map[string]string{"k": "v"}, v2.MapStringString)
}

// The reference implementation range-checks the entry field number before
// using it; a number past the maximum makes the message invalid rather than
// contributing an unknown field.
func TestDecodeMapEntryFieldNumberTooLarge(t *testing.T) {
	const fMapStringString = 901
	body := []byte{0x80, 0x80, 0x80, 0x80, 0x10, 0x01} // field 1<<29, varint
	var v TestStruct
	assert.ErrorContains(t, Unmarshal(mapEntry(fMapStringString, body), &v), "invalid field number")
}

// The key and value vars are pooled and reused across the entries of one
// field, so an entry that omits a field must not inherit the previous entry's
// value. Each value kind is protected by a different mechanism, so cover them
// all: string and bytes by ZeroVal, a message pointer by the per-entry
// allocation, a non-pointer struct by Reset.
func TestDecodeMapEntryNoLeakBetweenEntries(t *testing.T) {
	const (
		fMapStringString  = 901
		fMapStringBytes   = 902
		fMapStringStructA = 903
		fMapStringStructB = 904
	)
	key := func(s string) []byte { return (&wire.Builder{}).AppendStringField(1, s).Bytes() }

	t.Run("string", func(t *testing.T) {
		b := mapEntry(fMapStringString, (&wire.Builder{}).
			AppendStringField(1, "k").AppendStringField(2, "v").Bytes())
		b = append(b, mapEntry(fMapStringString, nil)...)
		var v TestStruct
		assert.NoError(t, Unmarshal(b, &v))
		assert.MapEqual(t, map[string]string{"k": "v", "": ""}, v.MapStringString)
	})

	t.Run("bytes", func(t *testing.T) {
		b := mapEntry(fMapStringBytes, (&wire.Builder{}).
			AppendStringField(1, "a").AppendStringField(2, "xy").Bytes())
		b = append(b, mapEntry(fMapStringBytes, key("b"))...)
		var v TestStruct
		assert.NoError(t, Unmarshal(b, &v))
		assert.Equal(t, 2, len(v.MapStringBytes))
		assert.Equal(t, 0, len(v.MapStringBytes["b"]))
	})

	// Two entries must never end up sharing one message.
	t.Run("message pointer", func(t *testing.T) {
		b := mapEntry(fMapStringStructA, (&wire.Builder{}).AppendStringField(1, "a").
			AppendBytesField(2, (&wire.Builder{}).AppendVarintField(1, 42).Bytes()).Bytes())
		b = append(b, mapEntry(fMapStringStructA, key("b"))...)
		b = append(b, mapEntry(fMapStringStructA, key("c"))...)
		var v TestStruct
		assert.NoError(t, Unmarshal(b, &v))
		assert.Equal(t, 3, len(v.MapStringStructA))
		assert.Equal(t, uint64(42), v.MapStringStructA["a"].V)
		assert.Equal(t, uint64(0), v.MapStringStructA["b"].V)
		assert.True(t, v.MapStringStructA["b"] != v.MapStringStructA["c"],
			"entries must not share one message")
	})

	// A present but empty value record sets hasVal, so only Reset stands
	// between the two entries here.
	t.Run("non-pointer struct", func(t *testing.T) {
		b := mapEntry(fMapStringStructB, (&wire.Builder{}).AppendStringField(1, "a").
			AppendBytesField(2, (&wire.Builder{}).AppendVarintField(1, 42).Bytes()).Bytes())
		b = append(b, mapEntry(fMapStringStructB, append(key("b"),
			(&wire.Builder{}).AppendBytesField(2, nil).Bytes()...))...)
		var v TestStruct
		assert.NoError(t, Unmarshal(b, &v))
		assert.Equal(t, uint64(42), v.MapStringStructB["a"].V)
		assert.Equal(t, uint64(0), v.MapStringStructB["b"].V)
	})
}

// Of repeated key or value fields inside one entry the last one wins, and
// repeated message values merge into a single message.
func TestDecodeMapEntryRepeatedFields(t *testing.T) {
	const (
		fMapStringString = 901
		fMapStringStruct = 903
	)

	b := (&wire.Builder{}).
		AppendStringField(1, "a").AppendStringField(1, "b").
		AppendStringField(2, "x").AppendStringField(2, "y").Bytes()
	var v TestStruct
	assert.NoError(t, Unmarshal(mapEntry(fMapStringString, b), &v))
	assert.MapEqual(t, map[string]string{"b": "y"}, v.MapStringString)

	// Two value records merge into one message: the second must not discard
	// what the first decoded. TestStructS has a single field, so record #1
	// sets it and record #2 carries only an unknown field.
	b = (&wire.Builder{}).AppendStringField(1, "k").
		AppendBytesField(2, (&wire.Builder{}).AppendVarintField(1, 42).Bytes()).
		AppendBytesField(2, (&wire.Builder{}).AppendVarintField(9, 7).Bytes()).Bytes()
	var v2 TestStruct
	assert.NoError(t, Unmarshal(mapEntry(fMapStringStruct, b), &v2))
	assert.NotNil(t, v2.MapStringStructA["k"])
	assert.Equal(t, uint64(42), v2.MapStringStructA["k"].V)
}

// A truncated entry must still be reported instead of silently accepted.
func TestDecodeMapEntryTruncated(t *testing.T) {
	const fMapStringString = 901
	for _, body := range [][]byte{
		{0x0a},             // key tag without payload
		{0x0a, 0x05, 'a'},  // key string shorter than its length
		{0x12, 0x02, 'a'},  // value string shorter than its length
		{0x08, 0x80},       // truncated varint of an unknown field
		{0x0a, 0x00, 0x80}, // truncated tag after a valid key
	} {
		var v TestStruct
		if err := Unmarshal(mapEntry(fMapStringString, body), &v); err == nil {
			t.Errorf("body %x: no error", body)
		}
	}

	// a malformed message value must fail too, not decode partially
	const fMapStringStruct = 903
	body := (&wire.Builder{}).AppendStringField(1, "k").
		AppendBytesField(2, []byte{0x08, 0x80}).Bytes()
	var v TestStruct
	if err := Unmarshal(mapEntry(fMapStringStruct, body), &v); err == nil {
		t.Errorf("malformed message value: no error")
	}
}
