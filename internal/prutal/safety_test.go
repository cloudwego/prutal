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
