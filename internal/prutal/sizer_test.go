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
)

func assertSizeMatchesMarshal(t *testing.T, v interface{}) {
	t.Helper()
	sz, err := Size(v)
	assert.NoError(t, err)
	b, err := MarshalAppend(nil, v)
	assert.NoError(t, err)
	assert.Equal(t, len(b), sz)
}

func TestSize_Nil(t *testing.T) {
	sz, err := Size(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, sz)
	sz, err = Size((*TestOneofMessage)(nil))
	assert.NoError(t, err)
	assert.Equal(t, 0, sz)
}

func TestSize_Empty(t *testing.T) {
	assertSizeMatchesMarshal(t, &TestOneofMessage{})
}

func TestSize_Oneof(t *testing.T) {
	p := &TestOneofMessage{}
	p.OneOfFieldA = &TestOneofMessage_Field2{Field2: 123}
	p.OneOfFieldB = &TestOneofMessage_Field4{Field4: "helloworld"}
	p.OneOfFieldC = &TestOneofMessage_Field5{&TestofNestedMessage{true}}
	assertSizeMatchesMarshal(t, p)
}

func TestSize_Scalars(t *testing.T) {
	type S struct {
		A int32   `protobuf:"varint,1,opt"`
		B int64   `protobuf:"varint,2,opt"`
		C uint32  `protobuf:"varint,3,opt"`
		D uint64  `protobuf:"varint,4,opt"`
		E int32   `protobuf:"zigzag32,5,opt"`
		F int64   `protobuf:"zigzag64,6,opt"`
		G uint32  `protobuf:"fixed32,7,opt"`
		H uint64  `protobuf:"fixed64,8,opt"`
		I float32 `protobuf:"fixed32,9,opt"`
		J float64 `protobuf:"fixed64,10,opt"`
		K bool    `protobuf:"varint,11,opt"`
		L string  `protobuf:"bytes,12,opt"`
		M []byte  `protobuf:"bytes,13,opt"`
	}
	p := &S{
		A: 42, B: 1 << 40, C: 300, D: 1 << 50,
		E: -100, F: -1 << 30, G: 0xDEADBEEF, H: 0xCAFEBABEDEADBEEF,
		I: 3.14, J: 2.718, K: true,
		L: "hello protobuf", M: []byte("bytes data"),
	}
	assertSizeMatchesMarshal(t, p)
}

func TestSize_ListPacked(t *testing.T) {
	p := &TestStruct_Benchmark_Encode_List_Packed{
		PackedInt32s:   testutils.Repeat(50, rand.Int31()),
		PackedInt64s:   testutils.Repeat(50, rand.Int63()),
		PackedUint32s:  testutils.Repeat(50, rand.Uint32()),
		PackedUint64s:  testutils.Repeat(50, rand.Uint64()),
		PackedBools:    testutils.RandomBoolSlice(50),
		PackedFixed32:  testutils.Repeat(50, rand.Uint32()),
		PackedFixed64:  testutils.Repeat(50, rand.Uint64()),
		PackedFloat:    testutils.Repeat(50, rand.Float32()),
		PackedDouble:   testutils.Repeat(50, rand.Float64()),
		PackedZigZag32: testutils.Repeat(50, rand.Int31()),
		PackedZigZag64: testutils.Repeat(50, rand.Int63()),
	}
	assertSizeMatchesMarshal(t, p)
}

func TestSize_ListNoPack(t *testing.T) {
	p := &TestStruct_Benchmark_Encode_List_NoPack{
		PackedInt32s:   testutils.Repeat(50, rand.Int31()),
		PackedInt64s:   testutils.Repeat(50, rand.Int63()),
		PackedUint32s:  testutils.Repeat(50, rand.Uint32()),
		PackedUint64s:  testutils.Repeat(50, rand.Uint64()),
		PackedBools:    testutils.RandomBoolSlice(50),
		PackedFixed32:  testutils.Repeat(50, rand.Uint32()),
		PackedFixed64:  testutils.Repeat(50, rand.Uint64()),
		PackedFloat:    testutils.Repeat(50, rand.Float32()),
		PackedDouble:   testutils.Repeat(50, rand.Float64()),
		PackedZigZag32: testutils.Repeat(50, rand.Int31()),
		PackedZigZag64: testutils.Repeat(50, rand.Int63()),
	}
	assertSizeMatchesMarshal(t, p)
}

func TestSize_String(t *testing.T) {
	p := &TestStruct_Benchmark_Encode_String{
		S1: testutils.RandomStr(10),
		S2: testutils.RandomStr(50),
		S3: testutils.RandomStr(100),
		S4: testutils.RandomStr(200),
		S5: testutils.RandomStr(400),
		SS: testutils.Repeat(100, "helloworld"),
	}
	assertSizeMatchesMarshal(t, p)
}

func TestSize_MapScalar(t *testing.T) {
	p := &TestStruct_Benchmark_Map_Scalar{}
	oo := testutils.DefaultFillOptions()
	oo.Seed = 12345
	oo.MapMinSize = 50
	oo.MapMaxSize = 50
	testutils.RandFill(p, oo)
	assertSizeMatchesMarshal(t, p)
}

func TestSize_MapString(t *testing.T) {
	p := &TestStruct_Benchmark_Map_String{}
	oo := testutils.DefaultFillOptions()
	oo.Seed = 12345
	oo.MapMinSize = 50
	oo.MapMaxSize = 50
	oo.StringMaxLen = 20
	testutils.RandFill(p, oo)
	assertSizeMatchesMarshal(t, p)
}

func TestSize_MapStruct(t *testing.T) {
	p := &TestStruct_Benchmark_Map_Struct{}
	oo := testutils.DefaultFillOptions()
	oo.Seed = 12345
	oo.MapMinSize = 50
	oo.MapMaxSize = 50
	testutils.RandFill(p, oo)
	assertSizeMatchesMarshal(t, p)
}

func TestSize_MapStructAllKeyTypes(t *testing.T) {
	type V struct {
		X int64 `protobuf:"varint,1,opt"`
	}

	// map[int32]*struct — sizeMapField_u32 with varint key
	type M1 struct {
		M map[int32]*V `protobuf:"bytes,1,rep" protobuf_key:"varint,1,opt" protobuf_val:"bytes,2,opt"`
	}
	assertSizeMatchesMarshal(t, &M1{M: map[int32]*V{-1: {X: 1}, 0: {X: 2}, 100: {X: 3}}})

	// map[uint32]*struct — sizeMapField_u32 with varint key
	type M2 struct {
		M map[uint32]*V `protobuf:"bytes,1,rep" protobuf_key:"varint,1,opt" protobuf_val:"bytes,2,opt"`
	}
	assertSizeMatchesMarshal(t, &M2{M: map[uint32]*V{0: {X: 1}, 100: {X: 2}, 0xFFFFFFFF: {X: 3}}})

	// map[int64]*struct — sizeMapField_u64 with varint key
	type M3 struct {
		M map[int64]*V `protobuf:"bytes,1,rep" protobuf_key:"varint,1,opt" protobuf_val:"bytes,2,opt"`
	}
	assertSizeMatchesMarshal(t, &M3{M: map[int64]*V{-1: {X: 1}, 0: {X: 2}, 1 << 40: {X: 3}}})

	// map[uint64]*struct — sizeMapField_u64 with fixed64 key
	type M4 struct {
		M map[uint64]*V `protobuf:"bytes,1,rep" protobuf_key:"fixed64,1,opt" protobuf_val:"bytes,2,opt"`
	}
	assertSizeMatchesMarshal(t, &M4{M: map[uint64]*V{0: {X: 1}, 1 << 50: {X: 2}}})

	// map[int32]*struct — sizeMapField_u32 with zigzag32 key
	type M5 struct {
		M map[int32]*V `protobuf:"bytes,1,rep" protobuf_key:"zigzag32,1,opt" protobuf_val:"bytes,2,opt"`
	}
	assertSizeMatchesMarshal(t, &M5{M: map[int32]*V{-100: {X: 1}, 0: {X: 2}, 100: {X: 3}}})

	// map[string]*struct — sizeMapField_str
	type M6 struct {
		M map[string]*V `protobuf:"bytes,1,rep" protobuf_key:"bytes,1,opt" protobuf_val:"bytes,2,opt"`
	}
	assertSizeMatchesMarshal(t, &M6{M: map[string]*V{"": {X: 1}, "hello": {X: 2}, "world": {X: 3}}})
}

func TestSize_ListStruct(t *testing.T) {
	type Inner struct {
		A int32  `protobuf:"varint,1,opt"`
		B string `protobuf:"bytes,2,opt"`
	}
	type S struct {
		Items []*Inner `protobuf:"bytes,1,rep"`
	}
	p := &S{
		Items: []*Inner{
			{A: 1, B: "hello"},
			{A: 1000, B: "world"},
			{A: 0, B: ""},
		},
	}
	assertSizeMatchesMarshal(t, p)
}

func TestSize_NestedStruct(t *testing.T) {
	type Inner struct {
		X int64 `protobuf:"varint,1,opt"`
	}
	type Outer struct {
		A *Inner `protobuf:"bytes,1,opt"`
		B int32  `protobuf:"varint,2,opt"`
	}
	p := &Outer{
		A: &Inner{X: 999999},
		B: 42,
	}
	assertSizeMatchesMarshal(t, p)
}

func BenchmarkSize_MapStruct(b *testing.B) {
	p := &TestStruct_Benchmark_Map_Struct{}
	oo := testutils.DefaultFillOptions()
	oo.Seed = 12345
	oo.MapMinSize = 100
	oo.MapMaxSize = 100
	testutils.RandFill(p, oo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Size(p)
	}
}

func BenchmarkSize_MapScalar(b *testing.B) {
	p := &TestStruct_Benchmark_Map_Scalar{}
	oo := testutils.DefaultFillOptions()
	oo.Seed = 12345
	oo.MapMinSize = 50
	oo.MapMaxSize = 50
	testutils.RandFill(p, oo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Size(p)
	}
}

func BenchmarkSize_ListPacked(b *testing.B) {
	p := &TestStruct_Benchmark_Encode_List_Packed{
		PackedInt32s:   testutils.Repeat(50, rand.Int31()),
		PackedInt64s:   testutils.Repeat(50, rand.Int63()),
		PackedUint32s:  testutils.Repeat(50, rand.Uint32()),
		PackedUint64s:  testutils.Repeat(50, rand.Uint64()),
		PackedBools:    testutils.RandomBoolSlice(50),
		PackedFixed32:  testutils.Repeat(50, rand.Uint32()),
		PackedFixed64:  testutils.Repeat(50, rand.Uint64()),
		PackedFloat:    testutils.Repeat(50, rand.Float32()),
		PackedDouble:   testutils.Repeat(50, rand.Float64()),
		PackedZigZag32: testutils.Repeat(50, rand.Int31()),
		PackedZigZag64: testutils.Repeat(50, rand.Int63()),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Size(p)
	}
}
