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
	"math/rand"
	"testing"

	"github.com/cloudwego/prutal/internal/testutils"
	"github.com/cloudwego/prutal/internal/testutils/assert"
	"github.com/cloudwego/prutal/internal/wire"
)

func TestEncodeOneof(t *testing.T) {
	a := &TestOneofMessage_Field2{Field2: 123}
	b := &TestOneofMessage_Field4{Field4: "helloworld"}
	c := &TestOneofMessage_Field5{&TestofNestedMessage{true}}
	p := &TestOneofMessage{}
	p.OneOfFieldA = a
	p.OneOfFieldB = b
	p.OneOfFieldC = c

	tmp := wire.Builder{}
	buf := wire.Builder{}
	buf.AppendVarintField(2, uint64(a.Field2)).
		AppendStringField(4, b.Field4).
		AppendBytesField(5, tmp.AppendVarintField(1, 1).Bytes())
	expect := buf.Bytes()
	data, err := MarshalAppend(nil, p)
	assert.NoError(t, err)
	assert.BytesEqual(t, expect, data)
}

type TestStruct_Benchmark_Encode_List_Packed struct {
	PackedInt32s  []int32  `protobuf:"varint,501,rep,packed"`
	PackedInt64s  []int64  `protobuf:"varint,502,rep,packed"`
	PackedUint32s []uint32 `protobuf:"varint,503,rep,packed"`
	PackedUint64s []uint64 `protobuf:"varint,504,rep,packed"`
	PackedBools   []bool   `protobuf:"varint,505,rep,packed"`

	PackedFixed32 []uint32  `protobuf:"fixed32,511,rep,packed"`
	PackedFixed64 []uint64  `protobuf:"fixed64,512,rep,packed"`
	PackedFloat   []float32 `protobuf:"fixed32,513,rep,packed"`
	PackedDouble  []float64 `protobuf:"fixed64,514,rep,packed"`

	PackedZigZag32 []int32 `protobuf:"zigzag32,521,rep,packed"`
	PackedZigZag64 []int64 `protobuf:"zigzag64,522,rep,packed"`
}

func Benchmark_Encode_List_Packed(b *testing.B) {
	p := &TestStruct_Benchmark_Encode_List_Packed{
		PackedInt32s:  testutils.Repeat(50, rand.Int31()),
		PackedInt64s:  testutils.Repeat(50, rand.Int63()),
		PackedUint32s: testutils.Repeat(50, rand.Uint32()),
		PackedUint64s: testutils.Repeat(50, rand.Uint64()),
		PackedBools:   testutils.RandomBoolSlice(50),

		PackedFixed32: testutils.Repeat(50, rand.Uint32()),
		PackedFixed64: testutils.Repeat(50, rand.Uint64()),
		PackedFloat:   testutils.Repeat(50, rand.Float32()),
		PackedDouble:  testutils.Repeat(50, rand.Float64()),

		PackedZigZag32: testutils.Repeat(50, rand.Int31()),
		PackedZigZag64: testutils.Repeat(50, rand.Int63()),
	}

	buf := make([]byte, 0, 16<<10)
	for i := 0; i < b.N; i++ {
		_, _ = MarshalAppend(buf[:0], p)
	}
}

type TestStruct_Benchmark_Encode_List_NoPack struct {
	PackedInt32s  []int32  `protobuf:"varint,501,rep"`
	PackedInt64s  []int64  `protobuf:"varint,502,rep"`
	PackedUint32s []uint32 `protobuf:"varint,503,rep"`
	PackedUint64s []uint64 `protobuf:"varint,504,rep"`
	PackedBools   []bool   `protobuf:"varint,505,rep"`

	PackedFixed32 []uint32  `protobuf:"fixed32,511,rep"`
	PackedFixed64 []uint64  `protobuf:"fixed64,512,rep"`
	PackedFloat   []float32 `protobuf:"fixed32,513,rep"`
	PackedDouble  []float64 `protobuf:"fixed64,514,rep"`

	PackedZigZag32 []int32 `protobuf:"zigzag32,521,rep"`
	PackedZigZag64 []int64 `protobuf:"zigzag64,522,rep"`
}

func Benchmark_Encode_List_NoPack(b *testing.B) {
	p := &TestStruct_Benchmark_Encode_List_NoPack{
		PackedInt32s:  testutils.Repeat(50, rand.Int31()),
		PackedInt64s:  testutils.Repeat(50, rand.Int63()),
		PackedUint32s: testutils.Repeat(50, rand.Uint32()),
		PackedUint64s: testutils.Repeat(50, rand.Uint64()),
		PackedBools:   testutils.RandomBoolSlice(50),

		PackedFixed32: testutils.Repeat(50, rand.Uint32()),
		PackedFixed64: testutils.Repeat(50, rand.Uint64()),
		PackedFloat:   testutils.Repeat(50, rand.Float32()),
		PackedDouble:  testutils.Repeat(50, rand.Float64()),

		PackedZigZag32: testutils.Repeat(50, rand.Int31()),
		PackedZigZag64: testutils.Repeat(50, rand.Int63()),
	}

	buf := make([]byte, 0, 16<<10)
	for i := 0; i < b.N; i++ {
		_, _ = MarshalAppend(buf[:0], p)
	}
}

type TestStruct_Benchmark_Encode_String struct {
	S1 string   `protobuf:"bytes,1"`
	S2 string   `protobuf:"bytes,2"`
	S3 string   `protobuf:"bytes,3"`
	S4 string   `protobuf:"bytes,4"`
	S5 string   `protobuf:"bytes,5"`
	SS []string `protobuf:"bytes,6,rep"`
}

func Benchmark_Encode_String(b *testing.B) {
	p := &TestStruct_Benchmark_Encode_String{
		S1: testutils.RandomStr(10),
		S2: testutils.RandomStr(50),
		S3: testutils.RandomStr(100),
		S4: testutils.RandomStr(200),
		S5: testutils.RandomStr(400),
		SS: testutils.Repeat(100, "helloworld"),
	}

	buf := make([]byte, 0, 16<<10)
	for i := 0; i < b.N; i++ {
		_, _ = MarshalAppend(buf[:0], p)
	}
}
