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

package desc

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

type TestMessage struct {
	Ptr      *int32 `protobuf:"varint,1,opt"`
	Varint32 int32  `protobuf:"varint,2,opt"`
	Varint64 int64  `protobuf:"varint,3,opt"`
	Bool     bool   `protobuf:"varint,4,opt"`
	Fixed32  uint32 `protobuf:"fixed32,5,opt"`
	Fixed64  uint64 `protobuf:"fixed64,6,opt"`
	ZigZag32 int32  `protobuf:"zigzag32,7,opt"`
	ZigZag64 int64  `protobuf:"zigzag64,8,opt"`

	Str           string   `protobuf:"bytes,101,opt"`
	Bytes         []byte   `protobuf:"bytes,102,opt"`
	PtrBytes      *[]byte  `protobuf:"bytes,103,opt"`
	RepeatedBytes [][]byte `protobuf:"bytes,104,rep"`
	PackedVarint  []int32  `protobuf:"varint,105,rep,packed"`

	MapVarint   map[int32]int32   `protobuf:"bytes,201,rep" protobuf_key:"varint,1,opt" protobuf_val:"varint,2,opt"`
	MapFixed32  map[uint32]uint32 `protobuf:"bytes,202,rep" protobuf_key:"fixed32,1,opt" protobuf_val:"fixed32,2,opt"`
	MapFixed64  map[uint64]uint64 `protobuf:"bytes,203,rep" protobuf_key:"fixed64,1,opt" protobuf_val:"fixed64,2,opt"`
	MapZigZag32 map[int32]int32   `protobuf:"bytes,204,rep" protobuf_key:"zigzag32,1,opt" protobuf_val:"zigzag32,2,opt"`
	MapZigZag64 map[int64]int64   `protobuf:"bytes,205,rep" protobuf_key:"zigzag64,1,opt" protobuf_val:"zigzag64,2,opt"`

	MapStringString map[string]string         `protobuf:"bytes,211,rep" protobuf_key:"bytes,1,opt" protobuf_val:"bytes,2,opt"`
	MapStringBytes  map[string][]byte         `protobuf:"bytes,212,rep" protobuf_key:"bytes,1,opt" protobuf_val:"bytes,2,opt"`
	MapStringStruct map[string]*NestedMessage `protobuf:"bytes,213,rep" protobuf_key:"bytes,1,opt" protobuf_val:"bytes,2,opt"`

	Nested1 *NestedMessage `protobuf:"bytes,301,opt"`
	Nested2 *TestMessage   `protobuf:"bytes,302,opt"`
}

type NestedMessage struct {
	X *NestedMessage `protobuf:"bytes,1,opt"`
	V *TestMessage   `protobuf:"bytes,2,opt"`
	Y *NestedMessage `protobuf:"bytes,3,opt"`
}

func TestGetOrParse(t *testing.T) {
	type testcase struct {
		ID       int32
		Name     string
		TagType  TagType
		Kind     reflect.Kind
		RealKind reflect.Kind
		KKind    reflect.Kind
		VKind    reflect.Kind
	}

	runTest := func(name string, s *StructDesc, cases []testcase) {
		t.Helper()
		for _, p := range cases {
			t.Run(name+"_"+p.Name, func(t *testing.T) {
				f := s.GetField(p.ID)
				assert.Equal(t, p.ID, f.ID)
				assert.Equal(t, p.Name, f.Name)
				assert.Equal(t, p.TagType, f.TagType)
				assert.Equal(t, p.Kind.String(), f.T.Kind.String())
				assert.Equal(t, p.RealKind.String(), f.T.RealKind().String())
				if p.Kind == reflect.Map {
					assert.Equal(t, p.KKind.String(), f.T.K.Kind.String())
					assert.Equal(t, p.VKind.String(), f.T.V.Kind.String())
					assert.True(t, f.IsMap)
				}
				if p.Kind == reflect.Slice {
					assert.True(t, f.IsList)
				}
			})
		}
	}

	s, err := GetOrParse(reflect.ValueOf(&TestMessage{}))
	assert.NoError(t, err)

	expects := []testcase{
		{
			ID:       1,
			Name:     "Ptr",
			TagType:  TypeVarint,
			Kind:     reflect.Pointer,
			RealKind: reflect.Int32,
		},
		{
			ID:       2,
			Name:     "Varint32",
			TagType:  TypeVarint,
			Kind:     reflect.Int32,
			RealKind: reflect.Int32,
		},
		{
			ID:       3,
			Name:     "Varint64",
			TagType:  TypeVarint,
			Kind:     reflect.Int64,
			RealKind: reflect.Int64,
		},
		{
			ID:       4,
			Name:     "Bool",
			TagType:  TypeVarint,
			Kind:     reflect.Bool,
			RealKind: reflect.Bool,
		},
		{
			ID:       5,
			Name:     "Fixed32",
			TagType:  TypeFixed32,
			Kind:     reflect.Uint32,
			RealKind: reflect.Uint32,
		},
		{
			ID:       6,
			Name:     "Fixed64",
			TagType:  TypeFixed64,
			Kind:     reflect.Uint64,
			RealKind: reflect.Uint64,
		},
		{
			ID:       7,
			Name:     "ZigZag32",
			TagType:  TypeZigZag32,
			Kind:     reflect.Int32,
			RealKind: reflect.Int32,
		},
		{
			ID:       8,
			Name:     "ZigZag64",
			TagType:  TypeZigZag64,
			Kind:     reflect.Int64,
			RealKind: reflect.Int64,
		},
		{
			ID:       101,
			Name:     "Str",
			TagType:  TypeBytes,
			Kind:     reflect.String,
			RealKind: reflect.String,
		},
		{
			ID:       102,
			Name:     "Bytes",
			TagType:  TypeBytes,
			Kind:     KindBytes,
			RealKind: KindBytes,
		},
		{
			ID:       103,
			Name:     "PtrBytes",
			TagType:  TypeBytes,
			Kind:     reflect.Pointer,
			RealKind: KindBytes,
		},
		{
			ID:       104,
			Name:     "RepeatedBytes",
			TagType:  TypeBytes,
			Kind:     reflect.Slice,
			RealKind: KindBytes,
		},
		{
			ID:       105,
			Name:     "PackedVarint",
			TagType:  TypeVarint,
			Kind:     reflect.Slice,
			RealKind: reflect.Int32,
		},
		{
			ID:       201,
			Name:     "MapVarint",
			TagType:  TypeBytes,
			Kind:     reflect.Map,
			RealKind: reflect.Map,
			KKind:    reflect.Int32,
			VKind:    reflect.Int32,
		},
		{
			ID:       202,
			Name:     "MapFixed32",
			TagType:  TypeBytes,
			Kind:     reflect.Map,
			RealKind: reflect.Map,
			KKind:    reflect.Uint32,
			VKind:    reflect.Uint32,
		},
		{
			ID:       203,
			Name:     "MapFixed64",
			TagType:  TypeBytes,
			Kind:     reflect.Map,
			RealKind: reflect.Map,
			KKind:    reflect.Uint64,
			VKind:    reflect.Uint64,
		},
		{
			ID:       204,
			Name:     "MapZigZag32",
			TagType:  TypeBytes,
			Kind:     reflect.Map,
			RealKind: reflect.Map,
			KKind:    reflect.Int32,
			VKind:    reflect.Int32,
		},
		{
			ID:       205,
			Name:     "MapZigZag64",
			TagType:  TypeBytes,
			Kind:     reflect.Map,
			RealKind: reflect.Map,
			KKind:    reflect.Int64,
			VKind:    reflect.Int64,
		},
		{
			ID:       211,
			Name:     "MapStringString",
			TagType:  TypeBytes,
			Kind:     reflect.Map,
			RealKind: reflect.Map,
			KKind:    reflect.String,
			VKind:    reflect.String,
		},
		{
			ID:       212,
			Name:     "MapStringBytes",
			TagType:  TypeBytes,
			Kind:     reflect.Map,
			RealKind: reflect.Map,
			KKind:    reflect.String,
			VKind:    KindBytes,
		},
		{
			ID:       213,
			Name:     "MapStringStruct",
			TagType:  TypeBytes,
			Kind:     reflect.Map,
			RealKind: reflect.Map,
			KKind:    reflect.String,
			VKind:    reflect.Pointer,
		},
		{
			ID:       301,
			Name:     "Nested1",
			TagType:  TypeBytes,
			Kind:     reflect.Pointer,
			RealKind: reflect.Struct,
		},
		{
			ID:       302,
			Name:     "Nested2",
			TagType:  TypeBytes,
			Kind:     reflect.Pointer,
			RealKind: reflect.Struct,
		},
	}
	t.Log(s)
	runTest("TestMessage", s, expects)

	// NestedMessage
	expects = []testcase{
		{
			ID:       1,
			Name:     "X",
			TagType:  TypeBytes,
			Kind:     reflect.Pointer,
			RealKind: reflect.Struct,
		},
		{
			ID:       2,
			Name:     "V",
			TagType:  TypeBytes,
			Kind:     reflect.Pointer,
			RealKind: reflect.Struct,
		},
		{
			ID:       3,
			Name:     "Y",
			TagType:  TypeBytes,
			Kind:     reflect.Pointer,
			RealKind: reflect.Struct,
		},
	}
	s = s.GetField(301).T.V.S // NestedMessage desc
	runTest("NestedMessage", s, expects)

	f1 := s.GetField(1)
	assert.Equal(t, f1.T.V.S, s)           // same *StructDesc
	assert.Equal(t, f1.T, s.GetField(3).T) // same as Field 3

	// same for GetOrParse
	s0, err := GetOrParse(reflect.ValueOf(&NestedMessage{}))
	assert.NoError(t, err)
	assert.Equal(t, s, s0)

	s1 := s.GetField(2).T.V.S // TestMessage field
	s2, err := GetOrParse(reflect.ValueOf(&TestMessage{}))
	assert.NoError(t, err)
	assert.Equal(t, s1, s2)
}

type NestedMessageA struct {
	NestedA     *NestedMessageA           `protobuf:"bytes,1,opt,name=nested_a" json:"nested_a,omitempty"`
	NestedB     *NestedMessageB           `protobuf:"bytes,2,opt,name=nested_b" json:"nested_b,omitempty"`
	NestedListA []*NestedMessageA         `protobuf:"bytes,3,rep,name=nested_list1" json:"nested_list_a,omitempty"`
	NestedListB []*NestedMessageB         `protobuf:"bytes,4,rep,name=nested_list2" json:"nested_list_b,omitempty"`
	NestedMapA  map[int64]*NestedMessageA `protobuf:"bytes,5,rep,name=nested_map1" json:"nested_map_a,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	NestedMapB  map[int64]*NestedMessageB `protobuf:"bytes,6,rep,name=nested_map2" json:"nested_map_b,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

type NestedMessageB struct {
	NestedA     *NestedMessageA           `protobuf:"bytes,11,opt,name=nested_a" json:"nested_a,omitempty"`
	NestedB     *NestedMessageB           `protobuf:"bytes,12,opt,name=nested_b" json:"nested_b,omitempty"`
	NestedListA []*NestedMessageA         `protobuf:"bytes,13,rep,name=nested_list_a" json:"nested_list_a,omitempty"`
	NestedListB []*NestedMessageB         `protobuf:"bytes,14,rep,name=nested_list_b" json:"nested_list_b,omitempty"`
	NestedMapA  map[int64]*NestedMessageA `protobuf:"bytes,15,rep,name=nested_map1" json:"nested_map_a,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	NestedMapB  map[int64]*NestedMessageB `protobuf:"bytes,16,rep,name=nested_map2" json:"nested_map_b,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func TestNested(t *testing.T) {
	s, err := GetOrParse(reflect.ValueOf(&NestedMessageA{}))
	assert.NoError(t, err)
	assert.NotNil(t, s)

	f2 := s.GetField(2)
	sb := f2.T.V.S
	f13 := sb.GetField(13)
	t.Log(f13)
}

type TestOneofMessage struct {
	Int32 int32 `protobuf:"varint,1,opt"`

	// Types that are assignable to OneOfField1:
	//
	//  *TestOneofMessage_Field1
	OneOfField1 isTestOneofMessage_OneOfField1 `protobuf_oneof:"one_of_field1"`
}

// XXX_OneofWrappers is for the internal use of the prutal package.
func (*TestOneofMessage) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*TestOneofMessage_Field1)(nil),
	}
}

type isTestOneofMessage_OneOfField1 interface {
	isTestOneofMessage_OneOfField1()
}

type TestOneofMessage_Field1 struct {
	Field1 *TestOneofMessage `protobuf:"bytes,2,opt"`
}

func (*TestOneofMessage_Field1) isTestOneofMessage_OneOfField1() {}

func TestOneOf(t *testing.T) {
	p := &TestOneofMessage{}
	sd, err := GetOrParse(reflect.ValueOf(p))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(sd.Fields))
	assert.Equal(t, int32(1), sd.Fields[0].ID)
	assert.Equal(t, int32(2), sd.Fields[1].ID)
	assert.False(t, sd.Fields[0].IsOneof())
	assert.True(t, sd.Fields[1].IsOneof())

	f := sd.Fields[1]
	assert.DeepEqual(t, reflect.TypeOf(&TestOneofMessage_Field1{}), f.OneofType)
	assert.Equal(t, "OneOfField1", f.Name)
	assert.Equal(t, unsafe.Offsetof((*TestOneofMessage)(nil).OneOfField1), f.Offset)
}
