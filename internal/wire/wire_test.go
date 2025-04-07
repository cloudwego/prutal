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
	"encoding/binary"
	"testing"

	"github.com/cloudwego/prutal/internal/protowire"
	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestEncodeTag(t *testing.T) {
	v := EncodeTag(1233, TypeBytes)
	n, wt := DecodeTag(v)
	assert.Equal(t, int32(1233), n)
	assert.Equal(t, TypeBytes, wt)
}

func TestConsumeKVTag(t *testing.T) {
	v := EncodeTag(2, TypeBytes)
	n, wt := ConsumeKVTag([]byte{byte(v)})
	assert.Equal(t, int32(2), n)
	assert.Equal(t, TypeBytes, wt)

	n, wt = ConsumeKVTag([]byte{})
	assert.Equal(t, int32(-1), n)
	assert.Equal(t, Type(-1), wt)
}

var (
	_b0 = binary.AppendUvarint([]byte{}, 1)     // nolint: unused
	_b1 = binary.AppendUvarint([]byte{}, 1<<7)  // nolint: unused
	_b2 = binary.AppendUvarint([]byte{}, 1<<14) // nolint: unused
	_b3 = binary.AppendUvarint([]byte{}, 1<<21) // nolint: unused
	_b4 = binary.AppendUvarint([]byte{}, 1<<28) // nolint: unused
	_b5 = binary.AppendUvarint([]byte{}, 1<<35) // nolint: unused
	_b6 = binary.AppendUvarint([]byte{}, 1<<42) // nolint: unused
	_b7 = binary.AppendUvarint([]byte{}, 1<<49) // nolint: unused
	_b8 = binary.AppendUvarint([]byte{}, 1<<56) // nolint: unused
	_b9 = binary.AppendUvarint([]byte{}, 1<<63) // nolint: unused
)

func BenchmarkUvarint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = binary.Uvarint(_b0)
		_, _ = binary.Uvarint(_b2)
		_, _ = binary.Uvarint(_b4)
		_, _ = binary.Uvarint(_b6)
	}
}

func BenchmarkConsumeVarint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = protowire.ConsumeVarint(_b0)
		_, _ = protowire.ConsumeVarint(_b2)
		_, _ = protowire.ConsumeVarint(_b4)
		_, _ = protowire.ConsumeVarint(_b6)
	}
}
