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

type mergeSub struct {
	V    int64   `protobuf:"varint,1,opt"`
	List []int64 `protobuf:"varint,2,rep"`

	unknownFields []byte
}

type mergeMsg struct {
	Sub  *mergeSub   `protobuf:"bytes,1,opt"`
	Subs []*mergeSub `protobuf:"bytes,2,rep"`
}

// protobuf defines concatenating two payloads as merging them, so a singular
// message field occurring twice must merge rather than be replaced, keeping
// the unknown fields of both occurrences.
func TestUnmarshalMergesSingularMessageField(t *testing.T) {
	b := (&wire.Builder{}).
		AppendBytesField(1, (&wire.Builder{}).AppendVarintField(1, 7).AppendVarintField(8, 1).Bytes()).
		AppendBytesField(1, (&wire.Builder{}).AppendVarintField(2, 9).AppendVarintField(9, 2).Bytes()).
		Bytes()

	v := &mergeMsg{}
	assert.NoError(t, Unmarshal(b, v))
	assert.Equal(t, int64(7), v.Sub.V)
	assert.SliceEqual(t, []int64{9}, v.Sub.List)
	// the unknown fields of both occurrences must survive the merge
	want := (&wire.Builder{}).AppendVarintField(8, 1).AppendVarintField(9, 2).Bytes()
	assert.BytesEqual(t, want, v.Sub.unknownFields)
}

// A repeated message field still appends: each occurrence is its own element.
func TestUnmarshalRepeatedMessageFieldAppends(t *testing.T) {
	b := (&wire.Builder{}).
		AppendBytesField(2, (&wire.Builder{}).AppendVarintField(1, 7).Bytes()).
		AppendBytesField(2, (&wire.Builder{}).AppendVarintField(1, 9).Bytes()).
		Bytes()

	v := &mergeMsg{}
	assert.NoError(t, Unmarshal(b, v))
	assert.Equal(t, 2, len(v.Subs))
	assert.Equal(t, int64(7), v.Subs[0].V)
	assert.Equal(t, int64(9), v.Subs[1].V)
}
