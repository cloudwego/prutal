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
)

// unknown field #9, used to drive the unknownFields handling paths below.
func unknownFieldBytes(t *testing.T) []byte {
	t.Helper()
	b, err := MarshalAppend(nil, &struct {
		X int64 `protobuf:"varint,9,opt"`
	}{X: 42})
	assert.NoError(t, err)
	return b
}

// A non-[]byte field named unknownFields must NOT be treated as unknown-fields
// storage, otherwise the decoder writes a slice header over it.
func TestUnknownFieldsWrongType(t *testing.T) {
	type M struct {
		A             int64 `protobuf:"varint,1,opt"`
		unknownFields string
	}
	var v M
	assert.NoError(t, Unmarshal(unknownFieldBytes(t), &v))
	assert.Equal(t, "", v.unknownFields)
}

// A promoted unknownFields field from an embedded struct has an offset relative
// to the embedded struct; it must not be applied to the outer base.
func TestUnknownFieldsPromoted(t *testing.T) {
	type base struct {
		Pad              [3]uint64
		XXX_unrecognized []byte
	}
	type M struct {
		A int64 `protobuf:"varint,1,opt"`
		base
	}
	var v M
	assert.NoError(t, Unmarshal(unknownFieldBytes(t), &v))
	assert.Equal(t, uint64(0), v.Pad[0])
	assert.Equal(t, uint64(0), v.Pad[1])
	assert.Equal(t, uint64(0), v.Pad[2])
}

// A directly-declared []byte unknownFields field still works.
func TestUnknownFieldsBytes(t *testing.T) {
	type M struct {
		A                int64 `protobuf:"varint,1,opt"`
		XXX_unrecognized []byte
	}
	var v M
	assert.NoError(t, Unmarshal(unknownFieldBytes(t), &v))
	assert.True(t, len(v.XXX_unrecognized) > 0)
}

// A nested repeated slice (e.g. [][]int32) must be rejected at decode time
// instead of causing type confusion.
func TestNestedRepeatedRejected(t *testing.T) {
	src, err := MarshalAppend(nil, &struct {
		F []int32 `protobuf:"varint,1,rep,packed"`
	}{F: []int32{1, 2, 3}})
	assert.NoError(t, err)

	var v struct {
		F [][]int32 `protobuf:"varint,1,rep,packed"`
	}
	err = Unmarshal(src, &v)
	assert.True(t, err != nil)
	assert.Equal(t, 0, len(v.F))
}

// Field number exactly at the direct-map boundary (1000) must round-trip.
func TestFieldIDBoundary(t *testing.T) {
	type M struct {
		F int64 `protobuf:"varint,1000,opt"`
	}
	b, err := MarshalAppend(nil, &M{F: 42})
	assert.NoError(t, err)
	var v M
	assert.NoError(t, Unmarshal(b, &v))
	assert.Equal(t, int64(42), v.F)
}

// Unmarshal into a typed-nil message pointer must return an error, not crash.
func TestUnmarshalNilMessage(t *testing.T) {
	err := Unmarshal([]byte{0x08, 0x01}, (*TestOneofMessage)(nil))
	assert.True(t, err != nil)
}

// A set oneof member must be serialized even when its value is the zero
// value, otherwise the chosen case is lost on round-trip.
func TestOneofZeroValueRoundTrip(t *testing.T) {
	src := &TestOneofMessage{OneOfFieldA: &TestOneofMessage_Field1{}}

	b, err := MarshalAppend(nil, src)
	assert.NoError(t, err)
	sz, err := Size(src)
	assert.NoError(t, err)
	assert.Equal(t, sz, len(b)) // tag + len-0 payload

	var dst TestOneofMessage
	assert.NoError(t, Unmarshal(b, &dst))
	_, ok := dst.OneOfFieldA.(*TestOneofMessage_Field1)
	assert.True(t, ok)
}

// Unpacked and packed occurrences of the same repeated field must be
// concatenated on decode, per protobuf merge semantics.
func TestPackedAndUnpackedMixDecode(t *testing.T) {
	// field 1 (uint64): unpacked element 5, then packed run [1, 2]
	b := []byte{0x08, 0x05, 0x0a, 0x02, 0x01, 0x02}
	var v struct {
		F []uint64 `protobuf:"varint,1,rep,packed"`
	}
	assert.NoError(t, Unmarshal(b, &v))
	assert.SliceEqual(t, []uint64{5, 1, 2}, v.F)
}

// Map nesting must cost one recursion level on both the encode and decode
// sides. The pre-fix decoder consumed two levels per map hop, so messages
// deeper than half the budget passed Marshal but failed Unmarshal.
func TestMapDepthBudgetRoundTrip(t *testing.T) {
	const depth = 600 // > defaultRecursionMaxDepth/2, < the budget itself
	bottom := &mapDepthMsg{Leaf: 1}
	v := bottom
	for i := 0; i < depth; i++ {
		m := map[int64]*mapDepthMsg{1: v}
		v = &mapDepthMsg{Next: m}
	}

	b, err := MarshalAppend(nil, v)
	assert.NoError(t, err)
	var dst mapDepthMsg
	assert.NoError(t, Unmarshal(b, &dst))
}

type mapDepthMsg struct {
	Leaf int64                  `protobuf:"varint,1,opt"`
	Next map[int64]*mapDepthMsg `protobuf:"bytes,2,rep" protobuf_key:"varint,1,opt" protobuf_val:"bytes,2,opt"`
	_    [0]func()              // no-copy pointer: keep this type unpointable
}
