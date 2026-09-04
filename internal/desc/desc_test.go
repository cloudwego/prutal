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

type legacyOneofMessage struct {
	One legacyOneof `protobuf_oneof:"one"`
}

type legacyOneof interface {
	isLegacyOneof()
}

type legacyOneofInt32 struct {
	Value int32 `protobuf:"varint,1,opt"`
}

func (*legacyOneofInt32) isLegacyOneof() {}

func (*legacyOneofMessage) XXX_OneofFuncs() (func(), func(), func(), []any) {
	return nil, nil, nil, []any{(*legacyOneofInt32)(nil)}
}

func TestLegacyOneofFuncs(t *testing.T) {
	sd, err := GetOrParse(reflect.ValueOf(&legacyOneofMessage{}))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(sd.Fields))
	assert.True(t, sd.Fields[0].IsOneof())
	assert.DeepEqual(t, reflect.TypeOf(&legacyOneofInt32{}), sd.Fields[0].OneofType)
}

type invalidCyclicMessage struct {
	Child *invalidCyclicChild `protobuf:"bytes,1,opt"`
	Bad   int32               `protobuf:"varint,0,opt"`
}

type invalidCyclicChild struct {
	Root *invalidCyclicMessage `protobuf:"bytes,1,opt"`
}

type sharedDescriptorRoot struct {
	Child *sharedDescriptorChild `protobuf:"bytes,1,opt"`
}

type sharedDescriptorChild struct {
	Root *sharedDescriptorRoot `protobuf:"bytes,1,opt"`
}

type sharedDescriptorWrapper struct {
	Shared *sharedDescriptorRoot `protobuf:"bytes,1,opt"`
}

type panickingCyclicMessage struct {
	Child *panickingCyclicChild `protobuf:"bytes,1,opt"`
	One   panickingCyclicOneof  `protobuf_oneof:"one"`
}

type panickingCyclicChild struct {
	Root *panickingCyclicMessage `protobuf:"bytes,1,opt"`
}

type panickingCyclicOneof interface {
	isPanickingCyclicOneof()
}

type invalidOneofScalarPointerMessage struct {
	One invalidOneofScalarPointer `protobuf_oneof:"one"`
}

type invalidOneofScalarPointer interface {
	isInvalidOneofScalarPointerMessage()
}

type invalidOneofScalarPointerWrapper struct {
	Value *int32 `protobuf:"varint,1,opt"`
}

func (*invalidOneofScalarPointerMessage) XXX_OneofWrappers() []any {
	return []any{(*invalidOneofScalarPointerWrapper)(nil)}
}

func (*invalidOneofScalarPointerWrapper) isInvalidOneofScalarPointerMessage() {}

func (*panickingCyclicMessage) XXX_OneofWrappers() []any {
	// panics after the cyclic fields above have been parsed and cached,
	// so the parse scope must roll them back
	panic("XXX_OneofWrappers panic")
}

type invalidOneofNoTagMessage struct {
	One invalidOneofNoTag `protobuf_oneof:"one"`
}

type invalidOneofNoTag interface {
	isInvalidOneofNoTagMessage()
}

// the wrapper field is missing the protobuf tag
type invalidOneofNoTagWrapper struct {
	Value int32
}

func (*invalidOneofNoTagMessage) XXX_OneofWrappers() []any {
	return []any{(*invalidOneofNoTagWrapper)(nil)}
}

func (*invalidOneofNoTagWrapper) isInvalidOneofNoTagMessage() {}

type invalidOneofRepeatedMessage struct {
	One invalidOneofRepeated `protobuf_oneof:"one"`
}

type invalidOneofRepeated interface {
	isInvalidOneofRepeatedMessage()
}

type invalidOneofRepeatedWrapper struct {
	Value []int32 `protobuf:"varint,1,rep"`
}

func (*invalidOneofRepeatedMessage) XXX_OneofWrappers() []any {
	return []any{(*invalidOneofRepeatedWrapper)(nil)}
}

func (*invalidOneofRepeatedWrapper) isInvalidOneofRepeatedMessage() {}

type invalidOneofMapMessage struct {
	One invalidOneofMap `protobuf_oneof:"one"`
}

type invalidOneofMap interface {
	isInvalidOneofMapMessage()
}

type invalidOneofMapWrapper struct {
	Value map[string]int32 `protobuf:"bytes,1,rep" protobuf_key:"bytes,1,opt" protobuf_val:"varint,2,opt"`
}

func (*invalidOneofMapMessage) XXX_OneofWrappers() []any {
	return []any{(*invalidOneofMapWrapper)(nil)}
}

func (*invalidOneofMapWrapper) isInvalidOneofMapMessage() {}

type invalidOneofNotIfaceMessage struct {
	One int32 `protobuf_oneof:"one"`
}

func (*invalidOneofNotIfaceMessage) XXX_OneofWrappers() []any {
	return []any{(*invalidOneofScalarPointerWrapper)(nil)}
}

type invalidOneofNoWrappersMessage struct {
	One invalidOneofNoWrappers `protobuf_oneof:"one"`
}

type invalidOneofNoWrappers interface {
	isInvalidOneofNoWrappers()
}

func (*invalidOneofNoWrappersMessage) XXX_OneofWrappers() int {
	return 1
}

type invalidMapKeyEnum int32

type validMapValueEnum int32

type invalidOneofBadWrapperMessage struct {
	One invalidOneofBadWrapper `protobuf_oneof:"one"`
}

type invalidOneofBadWrapper interface {
	isInvalidOneofBadWrapper()
}

func (*invalidOneofBadWrapperMessage) XXX_OneofWrappers() []any {
	return []any{42}
}

// a single wrapper implementing two oneof interfaces must error, not panic
type invalidOneofDoubleMatchMessage struct {
	A invalidOneofDoubleMatchA `protobuf_oneof:"a"`
	B invalidOneofDoubleMatchB `protobuf_oneof:"b"`
}

type invalidOneofDoubleMatchA interface {
	isInvalidOneofDoubleMatchA()
}

type invalidOneofDoubleMatchB interface {
	isInvalidOneofDoubleMatchB()
}

func (*invalidOneofDoubleMatchMessage) XXX_OneofWrappers() []any {
	return []any{(*invalidOneofDoubleMatchWrapper)(nil)}
}

type invalidOneofDoubleMatchWrapper struct {
	Value int32 `protobuf:"varint,1,opt"`
}

func (*invalidOneofDoubleMatchWrapper) isInvalidOneofDoubleMatchA() {}

func (*invalidOneofDoubleMatchWrapper) isInvalidOneofDoubleMatchB() {}

type invalidOneofEmptyInterfaceMessage struct {
	One any `protobuf_oneof:"one"`
}

type invalidOneofEmptyInterfaceWrapper struct {
	Value int32 `protobuf:"varint,1,opt"`
}

func (*invalidOneofEmptyInterfaceMessage) XXX_OneofWrappers() []any {
	return []any{(*invalidOneofEmptyInterfaceWrapper)(nil)}
}

type invalidOneofWrapperMethodArityMessage struct {
	One invalidOneofWrapperMethodArity `protobuf_oneof:"one"`
}

type invalidOneofWrapperMethodArity interface {
	isInvalidOneofWrapperMethodArity()
}

func (*invalidOneofWrapperMethodArityMessage) XXX_OneofWrappers(int) []any {
	return nil
}

type invalidOneofProtoReflectArityMessage struct {
	One invalidOneofProtoReflectArity `protobuf_oneof:"one"`
}

type invalidOneofProtoReflectArity interface {
	isInvalidOneofProtoReflectArity()
}

func (*invalidOneofProtoReflectArityMessage) ProtoReflect(int) any {
	return nil
}

type invalidOneofProtoReflectResultMessage struct {
	One invalidOneofProtoReflectResult `protobuf_oneof:"one"`
}

type invalidOneofProtoReflectResult interface {
	isInvalidOneofProtoReflectResult()
}

func (*invalidOneofProtoReflectResultMessage) ProtoReflect() any {
	return struct{ OneofWrappers int }{}
}

type nilPanickingCyclicMessage struct {
	Child *nilPanickingCyclicChild `protobuf:"bytes,1,opt"`
	One   nilPanickingCyclicOneof  `protobuf_oneof:"one"`
}

type nilPanickingCyclicChild struct {
	Root *nilPanickingCyclicMessage `protobuf:"bytes,1,opt"`
}

type nilPanickingCyclicOneof interface {
	isNilPanickingCyclicOneof()
}

func (*nilPanickingCyclicMessage) XXX_OneofWrappers() []any {
	panic(nil)
}

func recoverPanic(f func()) (recovered bool) {
	returned := false
	defer func() {
		if !returned {
			_ = recover()
			recovered = true
		}
	}()
	f()
	returned = true
	return false
}

func TestGetOrParseCyclicErrorDoesNotPoisonCache(t *testing.T) {
	_, err := GetOrParse(reflect.ValueOf(&invalidCyclicMessage{}))
	assert.True(t, err != nil, "invalid cyclic root")

	// The successfully parsed child retains a reference to the failed parent.
	// It must not be reusable until the parent has parsed successfully.
	_, err = GetOrParse(reflect.ValueOf(&invalidCyclicChild{}))
	assert.True(t, err != nil, "invalid cyclic child")
}

func TestGetOrParseRecoveredPanicDoesNotPoisonCache(t *testing.T) {
	if !recoverPanic(func() {
		_, _ = GetOrParse(reflect.ValueOf(&panickingCyclicMessage{}))
	}) {
		t.Fatal("expected the first parse to panic")
	}
	if !recoverPanic(func() {
		_, _ = GetOrParse(reflect.ValueOf(&panickingCyclicChild{}))
	}) {
		t.Fatal("expected the second parse to panic")
	}
}

func TestGetOrParseRecoveredNilPanicDoesNotPoisonCache(t *testing.T) {
	if !recoverPanic(func() {
		_, _ = GetOrParse(reflect.ValueOf(&nilPanickingCyclicMessage{}))
	}) {
		t.Fatal("expected the first parse to panic")
	}
	if !recoverPanic(func() {
		_, _ = GetOrParse(reflect.ValueOf(&nilPanickingCyclicChild{}))
	}) {
		t.Fatal("expected the second parse to panic")
	}
}

func TestParseErrorKeepsSuccessfulDescriptors(t *testing.T) {
	root, err := GetOrParse(reflect.ValueOf(&sharedDescriptorRoot{}))
	assert.NoError(t, err)

	_, err = GetOrParse(reflect.ValueOf(&invalidCyclicMessage{}))
	assert.True(t, err != nil, "invalid cyclic root")

	wrapper, err := GetOrParse(reflect.ValueOf(&sharedDescriptorWrapper{}))
	assert.NoError(t, err)
	assert.Equal(t, root, wrapper.GetField(1).T.V.S)
}

// Hand-written structs whose tag/type combination would be incorrectly encoded
// (pointer bits leaked, slice headers read as scalars, panics on invalid
// field numbers) must be rejected by GetOrParse.
func TestGetOrParseReject(t *testing.T) {
	for _, tc := range []struct {
		name string
		typ  any
	}{
		{"field number 0", &struct {
			F int32 `protobuf:"varint,0,opt"`
		}{}},
		{"field number missing", &struct {
			F int32 `protobuf:"varint,opt"`
		}{}},
		{"field number 2^29", &struct {
			F int32 `protobuf:"varint,536870912,opt"`
		}{}},
		{"field number 2^31 wraps int32", &struct {
			F int32 `protobuf:"varint,2147483648,opt"`
		}{}},
		{"reserved field number 19000", &struct {
			F int32 `protobuf:"varint,19000,opt"`
		}{}},
		{"reserved field number 19999", &struct {
			F int32 `protobuf:"varint,19999,opt"`
		}{}},
		{"slice field without rep", &struct {
			F []int32 `protobuf:"varint,1,opt"`
		}{}},
		{"pointer to map field", &struct {
			F *map[string]int32 `protobuf:"bytes,1,opt"`
		}{}},
		{"oneof pointer to scalar", &invalidOneofScalarPointerMessage{}},
		{"oneof wrapper missing protobuf tag", &invalidOneofNoTagMessage{}},
		{"oneof repeated field", &invalidOneofRepeatedMessage{}},
		{"oneof map field", &invalidOneofMapMessage{}},
		{"oneof field not an interface", &invalidOneofNotIfaceMessage{}},
		{"oneof field not an interface without wrappers", &struct {
			One int32 `protobuf_oneof:"one"`
		}{}},
		{"oneof has no wrappers", &invalidOneofNoWrappersMessage{}},
		{"oneof wrapper not pointer to struct", &invalidOneofBadWrapperMessage{}},
		{"oneof wrapper matching two oneofs", &invalidOneofDoubleMatchMessage{}},
		{"oneof empty interface", &invalidOneofEmptyInterfaceMessage{}},
		{"oneof wrapper method has arguments", &invalidOneofWrapperMethodArityMessage{}},
		{"oneof ProtoReflect method has arguments", &invalidOneofProtoReflectArityMessage{}},
		{"oneof ProtoReflect result has invalid wrappers", &invalidOneofProtoReflectResultMessage{}},
		{"duplicated field number in tag", &struct {
			F int32 `protobuf:"varint,1,7,opt"`
		}{}},
		{"repeated pointer to scalar", &struct {
			F []*int32 `protobuf:"varint,1,rep"`
		}{}},
		{"map value pointer to scalar", &struct {
			F map[string]*int32 `protobuf:"bytes,1,rep" protobuf_key:"bytes,1,opt" protobuf_val:"varint,2,opt"`
		}{}},
		{"map value slice", &struct {
			F map[string][]int32 `protobuf:"bytes,1,rep" protobuf_key:"bytes,1,opt" protobuf_val:"varint,2,opt"`
		}{}},
		{"map key pointer", &struct {
			F map[*int32]string `protobuf:"bytes,1,rep" protobuf_key:"varint,1,opt" protobuf_val:"bytes,2,opt"`
		}{}},
		{"float32 map key", &struct {
			F map[float32]string `protobuf:"bytes,1,rep" protobuf_key:"fixed32,1,opt" protobuf_val:"bytes,2,opt"`
		}{}},
		{"float64 map key", &struct {
			F map[float64]string `protobuf:"bytes,1,rep" protobuf_key:"fixed64,1,opt" protobuf_val:"bytes,2,opt"`
		}{}},
		{"map key field number", &struct {
			F map[string]int32 `protobuf:"bytes,1,rep" protobuf_key:"bytes,2,opt" protobuf_val:"varint,2,opt"`
		}{}},
		{"map value field number", &struct {
			F map[string]int32 `protobuf:"bytes,1,rep" protobuf_key:"bytes,1,opt" protobuf_val:"varint,1,opt"`
		}{}},
		{"map key missing field number", &struct {
			F map[string]int32 `protobuf:"bytes,1,rep" protobuf_key:"bytes,opt" protobuf_val:"varint,2,opt"`
		}{}},
		{"enum map key", &struct {
			F map[invalidMapKeyEnum]int32 `protobuf:"bytes,1,rep" protobuf_key:"varint,1,opt,enum=test.Enum" protobuf_val:"varint,2,opt"`
		}{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetOrParse(reflect.ValueOf(tc.typ))
			assert.True(t, err != nil, tc.name)
		})
	}

	// valid shapes that must keep passing validation
	for _, tc := range []struct {
		name string
		typ  any
	}{
		{"optional bytes", &struct {
			F []byte `protobuf:"bytes,1,opt"`
		}{}},
		{"field number before reserved range", &struct {
			F int32 `protobuf:"varint,18999,opt"`
		}{}},
		{"field number after reserved range", &struct {
			F int32 `protobuf:"varint,20000,opt"`
		}{}},
		{"maximum field number", &struct {
			F int32 `protobuf:"varint,536870911,opt"`
		}{}},
		{"repeated bytes", &struct {
			F [][]byte `protobuf:"bytes,1,rep"`
		}{}},
		{"repeated pointer to message", &struct {
			F []*NestedMessage `protobuf:"bytes,1,rep"`
		}{}},
		{"map to message", &struct {
			F map[string]*NestedMessage `protobuf:"bytes,1,rep" protobuf_key:"bytes,1,opt" protobuf_val:"bytes,2,opt"`
		}{}},
		{"map to bytes", &struct {
			F map[string][]byte `protobuf:"bytes,1,rep" protobuf_key:"bytes,1,opt" protobuf_val:"bytes,2,opt"`
		}{}},
		{"map to enum", &struct {
			F map[string]validMapValueEnum `protobuf:"bytes,1,rep" protobuf_key:"bytes,1,opt" protobuf_val:"varint,2,opt,enum=test.Enum"`
		}{}},
		{"oneof pointer to message", &TestOneofMessage{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetOrParse(reflect.ValueOf(tc.typ))
			assert.NoError(t, err)
		})
	}
}

// A proto2 default is serialized verbatim as the trailing tag element and may
// contain commas (e.g. [default = "a,1"]); the digit run after the comma must
// not be parsed as a field number.
func TestParseStructTagDefaultWithComma(t *testing.T) {
	d, err := GetOrParse(reflect.ValueOf(&struct {
		F *string `protobuf:"bytes,3,opt,name=f,def=a,1"`
	}{}))
	assert.NoError(t, err)
	assert.True(t, d.GetField(3) != nil, "field must keep number 3")
}

type zeroKindMessage struct {
	I64   int64           `protobuf:"varint,1,opt"`
	I32   int32           `protobuf:"varint,2,opt"`
	B     bool            `protobuf:"varint,3,opt"`
	S     string          `protobuf:"bytes,4,opt"`
	Bytes []byte          `protobuf:"bytes,5,opt"`
	List  []int32         `protobuf:"varint,6,rep,packed"`
	M     map[int32]int32 `protobuf:"bytes,7,rep" protobuf_key:"varint,1,opt" protobuf_val:"varint,2,opt"`
	Ptr   *TestMessage    `protobuf:"bytes,8,opt"`
	F32   float32         `protobuf:"fixed32,9,opt"`
	F64   float64         `protobuf:"fixed64,10,opt"`
}

// The encoder and sizer skip zero values by the kind resolved here; a wrong
// kind either drops a set field or serializes an unset one.
func TestZeroKind(t *testing.T) {
	sd, err := GetOrParse(reflect.ValueOf(&zeroKindMessage{}))
	assert.NoError(t, err)

	ptr := ZeroKindU64 // pointers and maps are tested by their word
	if unsafe.Sizeof(uintptr(0)) == 4 {
		ptr = ZeroKindU32
	}
	want := map[string]ZeroKind{
		"I64": ZeroKindU64, "I32": ZeroKindU32, "B": ZeroKindU8,
		"S": ZeroKindLen, "Bytes": ZeroKindLen, "List": ZeroKindLen,
		"M": ptr, "Ptr": ptr, "F32": ZeroKindU32, "F64": ZeroKindU64,
	}
	assert.Equal(t, len(want), len(sd.Fields))
	for _, f := range sd.Fields {
		assert.Equal(t, want[f.Name], f.ZeroKind, f.Name)
	}

	// a set oneof member is serialized even when zero, so it is never skipped
	sd, err = GetOrParse(reflect.ValueOf(&TestOneofMessage{}))
	assert.NoError(t, err)
	assert.Equal(t, ZeroKindU32, sd.Fields[0].ZeroKind)
	assert.Equal(t, ZeroKindNone, sd.Fields[1].ZeroKind)
}
