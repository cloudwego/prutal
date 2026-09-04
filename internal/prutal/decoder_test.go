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

	"github.com/cloudwego/prutal/internal/testutils"
	"github.com/cloudwego/prutal/internal/testutils/assert"
	"github.com/cloudwego/prutal/internal/wire"
)

func TestDecodeOneof(t *testing.T) {
	n := 0x0000ffff
	s := "helloworld"
	tmp := wire.Builder{}
	buf := wire.Builder{}
	buf.AppendVarintField(2, uint64(n)).
		AppendStringField(4, s).
		AppendBytesField(5, tmp.AppendVarintField(1, 1).Bytes())
	b := buf.Bytes()
	p := &TestOneofMessage{}
	err := Unmarshal(b, p)
	assert.NoError(t, err)

	f2, ok := p.OneOfFieldA.(*TestOneofMessage_Field2)
	assert.True(t, ok)
	assert.Equal(t, int64(n), f2.Field2)
	f4, ok := p.OneOfFieldB.(*TestOneofMessage_Field4)
	assert.True(t, ok)
	assert.Equal(t, s, f4.Field4)

	f5, ok := p.OneOfFieldC.(*TestOneofMessage_Field5)
	assert.True(t, ok)
	assert.Equal(t, true, f5.Field5.Field1)
}

// Map decoding is the hot path most sensitive to how map entries are scanned;
// these mirror the Benchmark_Encode_Map_* set so a change to either side of
// the wire shows up in `go test -bench=.`.
func Benchmark_Decode_Map_Scalar(b *testing.B) {
	p := &TestStruct_Benchmark_Map_Scalar{}
	oo := testutils.DefaultFillOptions()
	oo.Seed = 12345
	oo.MapMinSize = 50
	oo.MapMaxSize = 50
	testutils.RandFill(p, oo)
	data, err := MarshalAppend(nil, p)
	assert.NoError(b, err)

	v := &TestStruct_Benchmark_Map_Scalar{}
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*v = TestStruct_Benchmark_Map_Scalar{}
		if err := Unmarshal(data, v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_Map_String(b *testing.B) {
	p := &TestStruct_Benchmark_Map_String{}
	oo := testutils.DefaultFillOptions()
	oo.Seed = 12345
	oo.MapMinSize = 100
	oo.MapMaxSize = 100
	oo.StringMaxLen = 20
	testutils.RandFill(p, oo)
	data, err := MarshalAppend(nil, p)
	assert.NoError(b, err)

	v := &TestStruct_Benchmark_Map_String{}
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*v = TestStruct_Benchmark_Map_String{}
		if err := Unmarshal(data, v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_Map_Struct(b *testing.B) {
	p := &TestStruct_Benchmark_Map_Struct{}
	oo := testutils.DefaultFillOptions()
	oo.Seed = 12345
	oo.MapMinSize = 100
	oo.MapMaxSize = 100
	testutils.RandFill(p, oo)
	data, err := MarshalAppend(nil, p)
	assert.NoError(b, err)

	v := &TestStruct_Benchmark_Map_Struct{}
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*v = TestStruct_Benchmark_Map_Struct{}
		if err := Unmarshal(data, v); err != nil {
			b.Fatal(err)
		}
	}
}
