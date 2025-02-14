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
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
	"github.com/cloudwego/prutal/internal/wire"
)

func P[T any](v T) *T { return &v }

type testcase struct {
	Name   string
	Bytes  func() []byte
	Struct func() interface{}
}

func runTest(t *testing.T, tc *testcase) {
	t.Helper()

	// make sure the decoder works
	// and then we can use Unmarshal for verifying Marshal
	// this also fixes inconsist bytes when encoding map
	if runDecoderTest(t, tc) {
		runEncoderTest(t, tc)
	}
}

func runEncoderTest(t *testing.T, tc *testcase) bool {
	t.Helper()
	return t.Run(tc.Name+"#Encoder", func(t *testing.T) {
		t.Helper()

		b, err := MarshalAppend([]byte{}, tc.Struct())
		assert.NoError(t, err)

		p := &TestStruct{}
		err = Unmarshal(b, p)
		assert.NoError(t, err)
		assert.DeepEqual(t, tc.Struct(), p)
	})
}

func runDecoderTest(t *testing.T, tc *testcase) bool {
	t.Helper()
	return t.Run(tc.Name+"#Decoder", func(t *testing.T) {
		t.Helper()
		p := &TestStruct{}
		err := Unmarshal(tc.Bytes(), p)
		assert.NoError(t, err)
		assert.DeepEqual(t, tc.Struct(), p)
	})
}

// fix constant overflows
func u32(v int32) uint32 {
	return uint32(v)
}

// fix constant overflows
func u64(v int64) uint64 {
	return uint64(v)
}

type TestStructS struct {
	V uint64 `protobuf:"varint,1,opt"`
}

type TestStruct struct {
	// for TestEncoderDecoderVarint

	OptionalInt32 *int32 `protobuf:"varint,101,opt"`
	Int32         int32  `protobuf:"varint,102,opt"`

	OptionalInt64 *int64 `protobuf:"varint,103,opt"`
	Int64         int64  `protobuf:"varint,104,opt"`

	OptionalUint32 *uint32 `protobuf:"varint,105,opt"`
	Uint32         uint32  `protobuf:"varint,106,opt"`

	OptionalUint64 *uint64 `protobuf:"varint,107,opt"`
	Uint64         uint64  `protobuf:"varint,108,opt"`

	OptionalBool *bool `protobuf:"varint,109,opt"`
	Bool         bool  `protobuf:"varint,110,opt"`

	// for TestEncodedDecoderFixed

	OptionalFixed32 *uint32 `protobuf:"fixed32,200,opt"`
	Fixed32         uint32  `protobuf:"fixed32,201,opt"`

	OptionalFixed64 *uint64 `protobuf:"fixed64,202,opt"`
	Fixed64         uint64  `protobuf:"fixed64,203,opt"`

	OptionalSfixed32 *int32 `protobuf:"fixed32,204,opt"`
	Sfixed32         int32  `protobuf:"fixed32,205,opt"`

	OptionalSfixed64 *int64 `protobuf:"fixed64,206,opt"`
	Sfixed64         int64  `protobuf:"fixed64,207,opt"`

	OptionalFloat *float32 `protobuf:"fixed32,208,opt"`
	Float         float32  `protobuf:"fixed32,209,opt"`

	OptionalDouble *float64 `protobuf:"fixed64,210,opt"`
	Double         float64  `protobuf:"fixed64,211,opt"`

	// for TestEncoderDecoderZigZag

	OptionalSint32 *int32 `protobuf:"zigzag32,301,opt"`
	Sint32         int32  `protobuf:"zigzag32,302,opt"`

	OptionalSint64 *int64 `protobuf:"zigzag64,303,opt"`
	Sint64         int64  `protobuf:"zigzag64,304,opt"`

	// for TestEncoderDecoderBytes

	OptionalStr *string `protobuf:"bytes,401,opt"`
	Str         string  `protobuf:"bytes,402,opt"`

	OptionalB *[]byte `protobuf:"bytes,411,opt"`
	B         []byte  `protobuf:"bytes,412,opt"`

	StructA *TestStructS `protobuf:"bytes,421,opt"`
	StructB TestStructS  `protobuf:"bytes,422,opt"`

	// for TestEncoderDecoderPacked

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

	// for TestEncoderDecoderRepeated
	Repeated       []uint64 `protobuf:"varint,601,rep"`
	PackedRepeated []uint64 `protobuf:"varint,602,rep,packed"`

	StructsA []*TestStructS `protobuf:"bytes,611,rep"`
	StructsB []TestStructS  `protobuf:"bytes,612,rep"`

	// for TestEncoderDecoderMapVarint
	MapBool     map[bool]bool     `protobuf:"bytes,701,rep" protobuf_key:"varint,1,opt" protobuf_val:"varint,2,opt"`
	MapInt32    map[int32]int32   `protobuf:"bytes,702,rep" protobuf_key:"varint,1,opt" protobuf_val:"varint,2,opt"`
	MapInt64    map[int64]int64   `protobuf:"bytes,703,rep" protobuf_key:"varint,1,opt" protobuf_val:"varint,2,opt"`
	MapUint32   map[uint32]uint32 `protobuf:"bytes,704,rep" protobuf_key:"varint,1,opt" protobuf_val:"varint,2,opt"`
	MapUint64   map[uint64]uint64 `protobuf:"bytes,705,rep" protobuf_key:"varint,1,opt" protobuf_val:"varint,2,opt"`
	MapZigZag32 map[int32]int32   `protobuf:"bytes,706,rep" protobuf_key:"zigzag32,1,opt" protobuf_val:"zigzag32,2,opt"`
	MapZigZag64 map[int64]int64   `protobuf:"bytes,707,rep" protobuf_key:"zigzag64,1,opt" protobuf_val:"zigzag64,2,opt"`

	// for TestEncoderDecoderMapFixed
	MapFixed32  map[uint32]uint32   `protobuf:"bytes,801,rep" protobuf_key:"fixed32,1,opt" protobuf_val:"fixed32,2,opt"`
	MapFixed64  map[uint64]uint64   `protobuf:"bytes,802,rep" protobuf_key:"fixed64,1,opt" protobuf_val:"fixed64,2,opt"`
	MapSfixed32 map[int32]int32     `protobuf:"bytes,803,rep" protobuf_key:"fixed32,1,opt" protobuf_val:"fixed32,2,opt"`
	MapSfixed64 map[int64]int64     `protobuf:"bytes,804,rep" protobuf_key:"fixed64,1,opt" protobuf_val:"fixed64,2,opt"`
	MapFloat    map[float32]float32 `protobuf:"bytes,805,rep" protobuf_key:"fixed32,1,opt" protobuf_val:"fixed32,2,opt"`
	MapDouble   map[float64]float64 `protobuf:"bytes,806,rep" protobuf_key:"fixed64,1,opt" protobuf_val:"fixed64,2,opt"`

	// for TestEncoderDecoderMapBytes
	MapStringString map[string]string `protobuf:"bytes,901,rep" protobuf_key:"bytes,1,opt" protobuf_val:"bytes,2,opt"`
	MapStringBytes  map[string][]byte `protobuf:"bytes,902,rep" protobuf_key:"bytes,1,opt" protobuf_val:"bytes,2,opt"`

	MapStringStructA map[string]*TestStructS `protobuf:"bytes,903,rep" protobuf_key:"bytes,1,opt" protobuf_val:"bytes,2,opt"`
	MapStringStructB map[string]TestStructS  `protobuf:"bytes,904,rep" protobuf_key:"bytes,1,opt" protobuf_val:"bytes,2,opt"`
}

func TestEncoderDecoderVarint(t *testing.T) {
	x := &wire.Builder{}

	runTest(t, &testcase{Name: "Int32",
		Bytes: func() []byte {
			return x.Reset().
				AppendVarintField(101, 100).
				AppendVarintField(102, 200).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{OptionalInt32: P(int32(100)), Int32: 200}
		},
	})
	runTest(t, &testcase{
		Name: "Int64",
		Bytes: func() []byte {
			return x.Reset().
				AppendVarintField(103, 100).
				AppendVarintField(104, 200).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{OptionalInt64: P(int64(100)), Int64: 200}
		},
	})
	runTest(t, &testcase{
		Name: "Uint32",
		Bytes: func() []byte {
			return x.Reset().
				AppendVarintField(105, 100).
				AppendVarintField(106, 200).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{OptionalUint32: P(uint32(100)), Uint32: 200}
		},
	})
	runTest(t, &testcase{
		Name: "Uint64",
		Bytes: func() []byte {
			return x.Reset().
				AppendVarintField(107, 100).
				AppendVarintField(108, 200).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{OptionalUint64: P(uint64(100)), Uint64: 200}
		},
	})
	runTest(t, &testcase{
		Name: "Bool",
		Bytes: func() []byte {
			return x.Reset().
				AppendVarintField(109, 1).
				AppendVarintField(110, 1).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{OptionalBool: P(true), Bool: true}
		},
	})
}

func TestEncodedDecoderFixed(t *testing.T) {
	x := &wire.Builder{}
	runTest(t, &testcase{
		Name: "Fixed32_uint32",
		Bytes: func() []byte {
			return x.Reset().
				AppendFixed32Field(200, 100).
				AppendFixed32Field(201, 1000).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalFixed32: P(uint32(100)),
				Fixed32:         uint32(1000),
			}
		},
	})
	runTest(t, &testcase{
		Name: "Fixed64_uint64",
		Bytes: func() []byte {
			return x.Reset().
				AppendFixed64Field(202, 100).
				AppendFixed64Field(203, 1000).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalFixed64: P(uint64(100)),
				Fixed64:         uint64(1000),
			}
		},
	})

	runTest(t, &testcase{
		Name: "Fixed32_int32",
		Bytes: func() []byte {
			return x.Reset().
				AppendFixed32Field(204, u32(-100)).
				AppendFixed32Field(205, u32(-1000)).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalSfixed32: P(int32(-100)),
				Sfixed32:         int32(-1000),
			}
		},
	})
	runTest(t, &testcase{
		Name: "Fixed64_int64",
		Bytes: func() []byte {
			return x.Reset().
				AppendFixed64Field(206, u64(-100)).
				AppendFixed64Field(207, u64(-1000)).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalSfixed64: P(int64(-100)),
				Sfixed64:         int64(-1000),
			}
		},
	})

	runTest(t, &testcase{
		Name: "Fixed32_float32",
		Bytes: func() []byte {
			return x.Reset().
				AppendFixed32Field(208, math.Float32bits(100)).
				AppendFixed32Field(209, math.Float32bits(1000)).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalFloat: P(float32(100)),
				Float:         float32(1000),
			}
		},
	})

	runTest(t, &testcase{
		Name: "Fixed64_float64",
		Bytes: func() []byte {
			return x.Reset().
				AppendFixed64Field(210, math.Float64bits(100)).
				AppendFixed64Field(211, math.Float64bits(1000)).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalDouble: P(float64(100)),
				Double:         float64(1000),
			}
		},
	})

}

func TestEncoderDecoderZigZag(t *testing.T) {
	x := &wire.Builder{}

	runTest(t, &testcase{
		Name: "ZigZag32",
		Bytes: func() []byte {
			return x.Reset().
				AppendZigZagField(301, -100).
				AppendZigZagField(302, -1000).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalSint32: P(int32(-100)),
				Sint32:         int32(-1000),
			}
		},
	})
	runTest(t, &testcase{
		Name: "ZigZag64",
		Bytes: func() []byte {
			return x.Reset().
				AppendZigZagField(303, -100).
				AppendZigZagField(304, -1000).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalSint64: P(int64(-100)),
				Sint64:         int64(-1000),
			}
		},
	})
}

func TestEncoderDecoderBytes(t *testing.T) {
	x := &wire.Builder{}
	runTest(t, &testcase{
		Name: "String",
		Bytes: func() []byte {
			return x.Reset().
				AppendStringField(401, "hello").
				AppendStringField(402, "world").Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalStr: P("hello"),
				Str:         "world",
			}
		},
	})

	// only for decoder, coz encoder will skip zero value
	runDecoderTest(t, &testcase{
		Name: "String_Empty",
		Bytes: func() []byte {
			return x.Reset().
				AppendStringField(401, "").
				AppendStringField(402, "").Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalStr: P(""),
			}
		},
	})

	// only for encoder, coz the case above covers this case
	runEncoderTest(t, &testcase{
		Name: "String_Empty",
		Bytes: func() []byte {
			return x.Reset().
				AppendStringField(401, "").Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalStr: P(""),
			}
		},
	})

	runTest(t, &testcase{
		Name: "Bytes",
		Bytes: func() []byte {
			return x.Reset().
				AppendStringField(411, "hello").
				AppendStringField(412, "world").Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalB: P([]byte("hello")),
				B:         []byte("world"),
			}
		},
	})

	// only for Decoder, coz encoder will not encode []byte{} which is zero value
	runDecoderTest(t, &testcase{
		Name: "Bytes_Empty",
		Bytes: func() []byte {
			return x.Reset().
				AppendStringField(411, "").
				AppendStringField(412, "").Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalB: P([]byte("")),
				B:         []byte{},
			}
		},
	})

	//  only for Encoder, coz the case above covers this case
	runEncoderTest(t, &testcase{
		Name: "Bytes_Empty",
		Bytes: func() []byte {
			return x.Reset().
				AppendStringField(411, "").Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				OptionalB: P([]byte("")),
			}
		},
	})
	runTest(t, &testcase{
		Name: "Struct",
		Bytes: func() []byte {
			b := x.Reset().AppendVarintField(1, 1).Bytes() // TestStructS{V:1}
			return x.Reset().AppendBytesField(421, b).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{StructA: &TestStructS{V: 1}}
		},
	})
	runTest(t, &testcase{
		Name: "Struct_NoPointer",
		Bytes: func() []byte {
			b := x.Reset().AppendVarintField(1, 1).Bytes() // TestStructS{V:1}
			return x.Reset().AppendBytesField(422, b).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{StructB: TestStructS{V: 1}}
		},
	})
}

func TestEncoderDecoderPacked(t *testing.T) {
	x := &wire.Builder{}
	runTest(t, &testcase{
		Name: "Varint",
		Bytes: func() []byte {
			return x.Reset().
				AppendPackedVarintField(501, uint64(u32(-100)), uint64(u32(-1000))).
				AppendPackedVarintField(502, u64(-200), u64(-2000)).
				AppendPackedVarintField(503, 300, 3000).
				AppendPackedVarintField(504, 400, 4000).
				AppendPackedVarintField(505, 1, 0).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				PackedInt32s:  []int32{-100, -1000},
				PackedInt64s:  []int64{-200, -2000},
				PackedUint32s: []uint32{300, 3000},
				PackedUint64s: []uint64{400, 4000},
				PackedBools:   []bool{true, false},
			}
		},
	})
	runTest(t, &testcase{
		Name: "Fixed",
		Bytes: func() []byte {
			return x.Reset().
				AppendPackedFixed32Field(511, 100, 1000).
				AppendPackedFixed64Field(512, 200, 2000).
				AppendPackedFixed32Field(513, math.Float32bits(300), math.Float32bits(3000)).
				AppendPackedFixed64Field(514, math.Float64bits(400), math.Float64bits(4000)).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				PackedFixed32: []uint32{100, 1000},
				PackedFixed64: []uint64{200, 2000},
				PackedFloat:   []float32{300, 3000},
				PackedDouble:  []float64{400, 4000},
			}
		},
	})
	runTest(t, &testcase{
		Name: "ZigZag",
		Bytes: func() []byte {
			return x.Reset().
				AppendPackedZigZagField(521, -100, -1000).
				AppendPackedZigZagField(522, -200, -2000).Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				PackedZigZag32: []int32{-100, -1000},
				PackedZigZag64: []int64{-200, -2000},
			}
		},
	})
}

func TestEncoderDecoderRepeated(t *testing.T) {
	x := &wire.Builder{}
	tmpx := &wire.Builder{}
	sizes := []int{1, 2, 15, 20}
	for _, n := range sizes {
		{ // unpacked testcase
			b := []byte{}
			p := &TestStruct{Repeated: make([]uint64, 0, n)}

			x.Reset()
			for i := 0; i < n; i++ {
				x.AppendVarintField(601, uint64(i))
				p.Repeated = append(p.Repeated, uint64(i))
			}
			b = x.Bytes()

			runTest(t, &testcase{
				Name:   fmt.Sprintf("Unpacked_%d_Element", n),
				Bytes:  func() []byte { return b },
				Struct: func() interface{} { return p },
			})
		}

		{ // packed testcase
			b := []byte{}
			p := &TestStruct{PackedRepeated: make([]uint64, 0, n)}
			for i := 0; i < n; i++ {
				p.PackedRepeated = append(p.PackedRepeated, uint64(i))
			}
			b = x.Reset().AppendPackedVarintField(602, p.PackedRepeated...).Bytes()

			runTest(t, &testcase{
				Name:   fmt.Sprintf("Packed_%d_Element", n),
				Bytes:  func() []byte { return b },
				Struct: func() interface{} { return p },
			})
		}

		{ // struct testcase
			b := []byte{}
			p := &TestStruct{StructsA: make([]*TestStructS, 0, n), StructsB: make([]TestStructS, 0, n)}

			x.Reset()
			for i := 0; i < n; i++ {
				v := uint64(i + 1)
				x.AppendBytesField(611, tmpx.Reset().AppendVarintField(1, v).Bytes())
				p.StructsA = append(p.StructsA, &TestStructS{V: v})
			}
			for i := 0; i < n; i++ {
				v := uint64(i + 1)
				x.AppendBytesField(612, tmpx.Reset().AppendVarintField(1, v).Bytes())
				p.StructsB = append(p.StructsB, TestStructS{V: v})
			}
			b = x.Bytes()

			runTest(t, &testcase{
				Name:   fmt.Sprintf("Struct_%d_Element", n),
				Bytes:  func() []byte { return b },
				Struct: func() interface{} { return p },
			})
		}

	}
}

func TestEncoderDecoderMapVarint(t *testing.T) {
	x := &wire.Builder{}
	tmpx := &wire.Builder{}

	runTest(t, &testcase{
		Name: "Bool",
		Bytes: func() []byte {
			x.Reset()
			for i := 0; i < 3; i++ {
				x.AppendBytesField(701, tmpx.Reset().
					AppendVarintField(1, uint64(i&1)).
					AppendVarintField(2, uint64((i+1)&1)).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapBool: map[bool]bool{
					true:  false,
					false: true,
				},
			}
		},
	})
	runTest(t, &testcase{
		Name: "Int32", Bytes: func() []byte {
			x.Reset()
			for i := 0; i < 3; i++ {
				x.AppendBytesField(702, tmpx.Reset().
					AppendVarintField(1, uint64(100+i)).
					AppendVarintField(2, uint64(200+i)).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapInt32: map[int32]int32{
					100: 200,
					101: 201,
					102: 202,
				},
			}
		},
	})
	runTest(t, &testcase{
		Name: "Int64",
		Bytes: func() []byte {
			x.Reset()
			for i := 0; i < 3; i++ {
				x.AppendBytesField(703, tmpx.Reset().
					AppendVarintField(1, uint64(100+i)).
					AppendVarintField(2, uint64(200+i)).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapInt64: map[int64]int64{
					100: 200,
					101: 201,
					102: 202,
				},
			}
		},
	})
	runTest(t, &testcase{
		Name: "Uint32",
		Bytes: func() []byte {
			x.Reset()
			for i := 0; i < 3; i++ {
				x.AppendBytesField(704, tmpx.Reset().
					AppendVarintField(1, uint64(100+i)).
					AppendVarintField(2, uint64(200+i)).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapUint32: map[uint32]uint32{
					100: 200,
					101: 201,
					102: 202,
				},
			}
		},
	})
	runTest(t, &testcase{
		Name: "Uint64",
		Bytes: func() []byte {
			x.Reset()
			for i := 0; i < 3; i++ {
				x.AppendBytesField(705, tmpx.Reset().
					AppendVarintField(1, uint64(100+i)).
					AppendVarintField(2, uint64(200+i)).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapUint64: map[uint64]uint64{
					100: 200,
					101: 201,
					102: 202,
				},
			}
		},
	})
	runTest(t, &testcase{
		Name: "Zigzag32",
		Bytes: func() []byte {
			x.Reset()
			for i := 0; i < 3; i++ {
				x.AppendBytesField(706, tmpx.Reset().
					AppendZigZagField(1, int64(-100-i)).
					AppendZigZagField(2, int64(-200-i)).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapZigZag32: map[int32]int32{
					-100: -200,
					-101: -201,
					-102: -202,
				},
			}
		},
	})
	runTest(t, &testcase{
		Name: "zigzag64",
		Bytes: func() []byte {
			x.Reset()
			for i := 0; i < 3; i++ {
				x.AppendBytesField(707, tmpx.Reset().
					AppendZigZagField(1, int64(-100-i)).
					AppendZigZagField(2, int64(-200-i)).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapZigZag64: map[int64]int64{
					-100: -200,
					-101: -201,
					-102: -202,
				},
			}
		},
	})
}

func TestEncoderDecoderMapFixed(t *testing.T) {
	x := &wire.Builder{}
	tmpx := &wire.Builder{}

	runTest(t, &testcase{
		Name: "Uint32",
		Bytes: func() []byte {
			x.Reset()
			for i := uint32(0); i < 3; i++ {
				x.AppendBytesField(801, tmpx.Reset().
					AppendFixed32Field(1, 100+i).
					AppendFixed32Field(2, 200+i).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapFixed32: map[uint32]uint32{
					100: 200,
					101: 201,
					102: 202,
				},
			}
		},
	})
	runTest(t, &testcase{
		Name: "Uint64",
		Bytes: func() []byte {
			x.Reset()
			for i := uint64(0); i < 3; i++ {
				x.AppendBytesField(802, tmpx.Reset().
					AppendFixed64Field(1, 100+i).
					AppendFixed64Field(2, 200+i).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapFixed64: map[uint64]uint64{
					100: 200,
					101: 201,
					102: 202,
				},
			}
		},
	})
	runTest(t, &testcase{
		Name: "Int32",
		Bytes: func() []byte {
			x.Reset()
			for i := int32(0); i < 3; i++ {
				x.AppendBytesField(803, tmpx.Reset().
					AppendFixed32Field(1, u32(-100-i)).
					AppendFixed32Field(2, u32(-200-i)).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapSfixed32: map[int32]int32{
					-100: -200,
					-101: -201,
					-102: -202,
				},
			}
		},
	})
	runTest(t, &testcase{
		Name: "Int64",
		Bytes: func() []byte {
			x.Reset()
			for i := 0; i < 3; i++ {
				x.AppendBytesField(804, tmpx.Reset().
					AppendFixed64Field(1, uint64(-100-i)).
					AppendFixed64Field(2, uint64(-200-i)).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapSfixed64: map[int64]int64{
					-100: -200,
					-101: -201,
					-102: -202,
				},
			}
		},
	})
	runTest(t, &testcase{
		Name: "Float",
		Bytes: func() []byte {
			x.Reset()
			for i := float32(0); i < 3; i++ {
				x.AppendBytesField(805, tmpx.Reset().
					AppendFixed32Field(1, math.Float32bits(100+i)).
					AppendFixed32Field(2, math.Float32bits(200+i)).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapFloat: map[float32]float32{
					100: 200,
					101: 201,
					102: 202,
				},
			}
		},
	})
	runTest(t, &testcase{
		Name: "Double",
		Bytes: func() []byte {
			x.Reset()
			for i := float64(0); i < 3; i++ {
				x.AppendBytesField(806, tmpx.Reset().
					AppendFixed64Field(1, math.Float64bits(100+i)).
					AppendFixed64Field(2, math.Float64bits(200+i)).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapDouble: map[float64]float64{
					100: 200,
					101: 201,
					102: 202,
				},
			}
		},
	})
}

func TestEncoderDecoderMapBytes(t *testing.T) {
	x := &wire.Builder{}
	tmpx := &wire.Builder{}

	runTest(t, &testcase{
		Name: "String2String",
		Bytes: func() []byte {
			x.Reset()
			x.AppendBytesField(901, tmpx.Reset().
				AppendStringField(1, "hello").
				AppendStringField(2, "world").Bytes())
			x.AppendBytesField(901, tmpx.Reset().
				AppendStringField(1, "").
				AppendStringField(2, "").Bytes())
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapStringString: map[string]string{
					"hello": "world",
					"":      "",
				},
			}
		},
	})
	runTest(t, &testcase{
		Name: "String2Bytes",
		Bytes: func() []byte {
			x.Reset()
			x.AppendBytesField(902, tmpx.Reset().
				AppendStringField(1, "hello").
				AppendStringField(2, "world").Bytes())
			x.AppendBytesField(902, tmpx.Reset().
				AppendStringField(1, "").
				AppendStringField(2, "").Bytes())
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapStringBytes: map[string][]byte{
					"hello": []byte("world"),
					"":      []byte{},
				},
			}
		},
	})

	runTest(t, &testcase{
		Name: "String2Struct",
		Bytes: func() []byte {
			x.Reset()
			for i := 0; i < 3; i++ {
				v := tmpx.Reset().AppendVarintField(1, uint64(i)).Bytes() // TestStructS
				x.AppendBytesField(903, tmpx.Reset().
					AppendStringField(1, "k-"+strconv.Itoa(i)).
					AppendBytesField(2, v).Bytes())
				x.AppendBytesField(904, tmpx.Reset().
					AppendStringField(1, "k-"+strconv.Itoa(i)).
					AppendBytesField(2, v).Bytes())
			}
			return x.Bytes()
		},
		Struct: func() interface{} {
			return &TestStruct{
				MapStringStructA: map[string]*TestStructS{
					"k-0": &TestStructS{V: 0},
					"k-1": &TestStructS{V: 1},
					"k-2": &TestStructS{V: 2},
				},
				MapStringStructB: map[string]TestStructS{
					"k-0": TestStructS{V: 0},
					"k-1": TestStructS{V: 1},
					"k-2": TestStructS{V: 2},
				},
			}
		},
	})
}

func TestUnknownFields(t *testing.T) {
	x := &wire.Builder{}
	x.Reset().AppendStringField(1, "hello").
		AppendStringField(2, "world")
	b := x.Bytes()

	type TestUnknownFieldsStruct struct {
		unknownFields []byte
	}
	type TestUnknownFieldsStruct2 struct {
		unknownFields *[]byte
	}

	{ // no pointer
		p := &TestUnknownFieldsStruct{}
		p.unknownFields = append(p.unknownFields, byte(9)) // will reset by Unmarshal
		err := Unmarshal(b, p)
		assert.NoError(t, err)
		assert.BytesEqual(t, b, p.unknownFields)

		newb, err := MarshalAppend([]byte{}, p)
		assert.NoError(t, err)
		assert.BytesEqual(t, b, newb)
	}
	{ // pointer=true
		p2 := &TestUnknownFieldsStruct2{}
		err := Unmarshal(b, p2)
		assert.NoError(t, err)
		assert.BytesEqual(t, b, *p2.unknownFields)

		newb, err := MarshalAppend([]byte{}, p2)
		assert.NoError(t, err)
		assert.BytesEqual(t, b, newb)

		// test again with p.unknownFields != nil
		err = Unmarshal(b, p2)
		assert.NoError(t, err)
		assert.BytesEqual(t, b, *p2.unknownFields)
	}

}
