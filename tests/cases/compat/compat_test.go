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
package compat

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/cloudwego/prutal"
	"github.com/cloudwego/prutal/internal/testutils/assert"
	"github.com/cloudwego/prutal/tests/cases/compat/apihybrid"
	"github.com/cloudwego/prutal/tests/cases/compat/apiopaque"
	"github.com/cloudwego/prutal/tests/cases/compat/compat2"
	"github.com/cloudwego/prutal/tests/cases/compat/compat2023"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// check runs every cross-implementation comparison on m and returns what
// failed, so a whole matrix can be reported at once.
func check(m proto.Message) []string {
	var errs []string
	add := func(format string, a ...any) { errs = append(errs, fmt.Sprintf(format, a...)) }
	fresh := func() proto.Message { return m.ProtoReflect().New().Interface() }

	ref, err := proto.Marshal(m)
	if err != nil {
		add("proto.Marshal: %v", err)
		return errs
	}
	m1 := fresh()
	if err := prutal.Unmarshal(ref, m1); err != nil {
		add("prutal.Unmarshal(proto bytes): %v", err)
	} else if !proto.Equal(m, m1) {
		add("prutal.Unmarshal(proto bytes): result differs")
	}

	pb, err := prutal.Marshal(m)
	if err != nil {
		add("prutal.Marshal: %v", err)
		return errs
	}
	m2 := fresh()
	if err := proto.Unmarshal(pb, m2); err != nil {
		add("proto.Unmarshal(prutal bytes): %v", err)
	} else if !proto.Equal(m, m2) {
		add("proto.Unmarshal(prutal bytes): result differs")
	}
	if sz, err := prutal.Size(m); err != nil {
		add("prutal.Size: %v", err)
	} else if sz != proto.Size(m) || len(pb) != sz {
		add("size: prutal.Size=%d len(prutal.Marshal)=%d proto.Size=%d", sz, len(pb), proto.Size(m))
	}
	// both write fields in field number order, so the bytes match unless
	// something reorders them: map entries come out in whatever order the
	// map iterates, and protobuf-go writes oneof members after every other
	// field of their message, to keep its historic wire output
	if !mayReorder(m.ProtoReflect().Descriptor(), map[protoreflect.FullName]bool{}) && !bytes.Equal(pb, ref) {
		add("prutal.Marshal: bytes differ from proto.Marshal\n  prutal: %x\n  proto:  %x", pb, ref)
	}
	m3 := fresh()
	if err := prutal.Unmarshal(pb, m3); err != nil {
		add("prutal self round-trip: %v", err)
	} else if !proto.Equal(m, m3) {
		add("prutal self round-trip: result differs")
	}
	return errs
}

// mayReorder reports whether md or any message reachable from it has a map
// or a oneof, whose fields prutal and protobuf-go may write in different orders.
func mayReorder(md protoreflect.MessageDescriptor, seen map[protoreflect.FullName]bool) bool {
	if seen[md.FullName()] {
		return false
	}
	seen[md.FullName()] = true
	fds := md.Fields()
	for i := 0; i < fds.Len(); i++ {
		fd := fds.Get(i)
		if fd.IsMap() || fd.ContainingOneof() != nil && !fd.ContainingOneof().IsSynthetic() {
			return true
		}
		if fd.Message() != nil && mayReorder(fd.Message(), seen) {
			return true
		}
	}
	return false
}

// hybridOpaque reports whether the hybrid API was built with the protoopaque
// tag, which gives it the opaque layout with its presence bitmap instead of
// the open one; run.sh tests both builds.
var hybridOpaque = func() bool {
	_, open := reflect.TypeOf((*apihybrid.Msg)(nil)).Elem().FieldByName("A")
	return !open
}()

// namedMessage is a matrix entry: a fresh message of the type to cover.
type namedMessage struct {
	name string
	m    proto.Message
}

// TestRandomMessages fills every message through protoreflect with random
// data and checks that prutal and protobuf-go agree in both directions.
func TestRandomMessages(t *testing.T) {
	cases := []namedMessage{
		{"proto3 Scalars", &Scalars{}},
		{"proto3 Optionals", &Optionals{}},
		{"proto3 Repeateds", &Repeateds{}},
		{"proto3 Maps", &Maps{}},
		{"proto3 Oneofs", &Oneofs{}},
		{"proto3 WellKnown fields", &WellKnown{}},
		{"proto3 Nested", &Nested{}},
		{"proto3 Empty", &Empty{}},
		{"proto3 OnlyRepeatedMsg", &OnlyRepeatedMsg{}},
		{"wkt Timestamp", &timestamppb.Timestamp{}},
		{"wkt Duration", &durationpb.Duration{}},
		{"wkt Any", &anypb.Any{}},
		{"wkt Struct", &structpb.Struct{}},
		{"wkt Value", &structpb.Value{}},
		{"wkt ListValue", &structpb.ListValue{}},
		{"wkt Empty", &emptypb.Empty{}},
		{"wkt FieldMask", &fieldmaskpb.FieldMask{}},
		{"wkt StringValue", &wrapperspb.StringValue{}},
		{"wkt BytesValue", &wrapperspb.BytesValue{}},
		{"wkt DoubleValue", &wrapperspb.DoubleValue{}},
		{"wkt BoolValue", &wrapperspb.BoolValue{}},
		{"proto2 Defaults", &compat2.Defaults{}},
		{"proto2 Proto2Message", &compat2.Proto2Message{}},
		{"edition 2023 Presence", &compat2023.Presence{}},
	}
	if !hybridOpaque {
		cases = append(cases, namedMessage{"hybrid api Msg", &apihybrid.Msg{}})
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for seed := int64(1); seed <= 5; seed++ {
				m := c.m.ProtoReflect().New()
				newFiller(seed).fill(m, 3)
				for _, e := range check(m.Interface()) {
					t.Errorf("seed %d: %s", seed, e)
				}
			}
		})
	}
}

// TestEdgeCases covers the values whose encoding depends on presence rules
// rather than on the value itself.
func TestEdgeCases(t *testing.T) {
	for _, c := range []struct {
		name string
		m    proto.Message
	}{
		{"proto3 optional: present zero values", &Optionals{
			I32: proto.Int32(0), I64: proto.Int64(0), U32: proto.Uint32(0), U64: proto.Uint64(0),
			S32: proto.Int32(0), S64: proto.Int64(0), F32: proto.Uint32(0), F64: proto.Uint64(0),
			Sf32: proto.Int32(0), Sf64: proto.Int64(0), Fl: proto.Float32(0), Db: proto.Float64(0),
			B: proto.Bool(false), S: proto.String(""), By: []byte{}, Color: Color_COLOR_UNSPECIFIED.Enum(),
			Msg: &Scalars{},
		}},
		{"proto3 optional: empty bytes only", &Optionals{By: []byte{}}},
		{"proto3 optional: nil bytes", &Optionals{By: nil}},
		{"proto3: empty bytes without presence", &Scalars{By: []byte{}}},
		{"proto2: present zero values", &compat2.Proto2Message{
			OptI32: proto.Int32(0), OptS: proto.String(""), OptBy: []byte{}, OptLevel: compat2.Level_LOW.Enum(),
			OptMsg: &compat2.Proto2Message{ReqI32: proto.Int32(1), ReqS: proto.String("x"), ReqMsg: &compat2.Defaults{}},
			ReqI32: proto.Int32(0), ReqS: proto.String(""), ReqMsg: &compat2.Defaults{},
		}},
		{"proto2: empty bytes only", &compat2.Proto2Message{OptBy: []byte{},
			ReqI32: proto.Int32(1), ReqS: proto.String("x"), ReqMsg: &compat2.Defaults{}}},
		{"proto2 defaults: unset", &compat2.Defaults{}},
		{"proto2 defaults: set to default values", &compat2.Defaults{
			I32: proto.Int32(-42), S: proto.String("hello, world"), By: []byte("a,b\000c"),
			Level: compat2.Level_HIGH.Enum(), B: proto.Bool(true), Fl: proto.Float32(1.5),
		}},
		{"oneof: zero int32", &Oneofs{Kind: &Oneofs_I32{}}},
		{"oneof: empty string", &Oneofs{Kind: &Oneofs_S{}}},
		{"oneof: nil bytes", &Oneofs{Kind: &Oneofs_By{}}},
		{"oneof: empty bytes", &Oneofs{Kind: &Oneofs_By{By: []byte{}}}},
		{"oneof: false bool", &Oneofs{Kind: &Oneofs_B{}}},
		{"oneof: zero enum", &Oneofs{Kind: &Oneofs_Color{}}},
		{"oneof: nil message", &Oneofs{Kind: &Oneofs_Msg{}}},
		{"oneof: empty message", &Oneofs{Kind: &Oneofs_Msg{Msg: &Scalars{}}}},
		{"oneof: nil wkt message", &Oneofs{Kind: &Oneofs_Ts{}}},
		{"oneof: both oneofs + plain", &Oneofs{Kind: &Oneofs_F64{F64: 7}, Second: &Oneofs_Msg2{Msg2: &Scalars{I32: 1}}, Plain: 3}},
		{"map: nil message value", &Maps{SMsg: map[string]*Scalars{"a": nil, "b": {}}}},
		{"map: empty key and value", &Maps{SS: map[string]string{"": ""}, SBy: map[string][]byte{"k": {}, "n": nil}}},
		{"map: zero scalar entries", &Maps{I32I32: map[int32]int32{0: 0}, BB: map[bool]bool{false: false}}},
		{"repeated: empty elements", &Repeateds{S: []string{"", "a", ""}, By: [][]byte{nil, {}, {1}}, Msg: []*Scalars{nil, {}}}},
		{"repeated: empty slices", &Repeateds{S: []string{}, I32: []int32{}, Msg: []*Scalars{}}},
		{"repeated: unpacked fields", &Repeateds{UnpackedI32: []int32{1, -1, 0}, UnpackedColor: []Color{Color_RED, Color_NEG}, UnpackedDb: []float64{0, 1.5}, UnpackedB: []bool{true, false}}},
		{"scalar: negative enum + huge field numbers", &Scalars{Color: Color_NEG, HugeNumber: 1, AfterThousand: 2, FarAfterThousand: 3, I32: -1}},
		{"scalar: +Inf/-Inf", &Scalars{Fl: float32(math.Inf(1)), Db: math.Inf(-1)}},
		{"scalar: NaN", &Scalars{Fl: float32(math.NaN()), Db: math.NaN()}},
		{"scalar: negative zero", &Scalars{Fl: float32(math.Copysign(0, -1)), Db: math.Copysign(0, -1)}},
		{"wkt: Any with packed message", mustAny(&Scalars{I32: 5, S: "x"})},
		{"wkt: Struct from map", mustStruct(map[string]any{"a": 1.5, "b": "s", "c": nil, "d": []any{true, "x"}, "e": map[string]any{"f": 1.0}})},
		{"wkt: Timestamp now", timestamppb.Now()},
		{"wkt: Duration negative", durationpb.New(-1500000000)},
		{"wkt: zero wrappers", &WellKnown{WDouble: wrapperspb.Double(0), WI32: wrapperspb.Int32(0), WString: wrapperspb.String(""), WBytes: wrapperspb.Bytes(nil), WBool: wrapperspb.Bool(false), Empty: &emptypb.Empty{}}},
		// the required fields must be set for protobuf-go to marshal at all
		{"edition 2023: explicit empty bytes", &compat2023.Presence{ByExplicit: []byte{}, ByRequired: []byte{}, IRequired: proto.Int32(1)}},
		{"edition 2023: present zero values", &compat2023.Presence{IExplicit: proto.Int32(0), IRequired: proto.Int32(0), ByRequired: []byte{}, Sub: &compat2023.Presence{IRequired: proto.Int32(0), ByRequired: []byte{}}}},
		{"edition 2023: implicit zero values", &compat2023.Presence{IImplicit: 0, SImplicit: "", ByImplicit: nil, ByAggregate: nil, ByRequired: []byte("r"), IRequired: proto.Int32(1)}},
		{"edition 2023: implicit set values", &compat2023.Presence{IImplicit: 1, SImplicit: "s", ByImplicit: []byte{1}, ByAggregate: []byte{2}, ByRequired: []byte("r"), IRequired: proto.Int32(1)}},
	} {
		t.Run(c.name, func(t *testing.T) {
			for _, e := range check(c.m) {
				t.Error(e)
			}
		})
	}

	t.Run("hybrid api: setters", func(t *testing.T) {
		if hybridOpaque {
			t.Skip("opaque layout, see TestUnsupportedOpaqueAPI")
		}
		m := &apihybrid.Msg{}
		m.SetA(0)
		m.SetSub(&apihybrid.Msg{})
		for _, e := range check(m) {
			t.Error(e)
		}
	})
}

func mustAny(m proto.Message) *anypb.Any {
	a, err := anypb.New(m)
	if err != nil {
		panic(err)
	}
	return a
}

func mustStruct(v map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(v)
	if err != nil {
		panic(err)
	}
	return s
}

// TestUnknownFields checks that unknown fields of every wire type, groups
// included, survive a prutal round-trip byte for byte, in order.
func TestUnknownFields(t *testing.T) {
	m := &Scalars{I32: 1, S: "s"}
	ref, err := proto.Marshal(m)
	assert.NoError(t, err)
	head := protowire.AppendTag(nil, 999, protowire.VarintType)
	head = protowire.AppendVarint(head, 42)
	head = protowire.AppendTag(head, 998, protowire.BytesType)
	head = protowire.AppendBytes(head, []byte("u"))
	tail := protowire.AppendTag(nil, 997, protowire.Fixed64Type)
	tail = protowire.AppendFixed64(tail, 5)
	tail = protowire.AppendTag(tail, 996, protowire.Fixed32Type)
	tail = protowire.AppendFixed32(tail, 6)
	tail = protowire.AppendTag(tail, 995, protowire.StartGroupType)
	tail = protowire.AppendTag(tail, 1, protowire.VarintType)
	tail = protowire.AppendVarint(tail, 7)
	tail = protowire.AppendTag(tail, 995, protowire.EndGroupType)
	unknown := append(append([]byte{}, head...), tail...)

	m1 := &Scalars{}
	assert.NoError(t, prutal.Unmarshal(append(append([]byte{}, ref...), unknown...), m1))
	assert.BytesEqual(t, unknown, m1.ProtoReflect().GetUnknown())
	pb, err := prutal.Marshal(m1)
	assert.NoError(t, err)
	m2 := &Scalars{}
	assert.NoError(t, proto.Unmarshal(pb, m2))
	assert.BytesEqual(t, unknown, m2.ProtoReflect().GetUnknown())
	assert.True(t, proto.Equal(m, m2) == false) // m has no unknown fields
	m2.ProtoReflect().SetUnknown(nil)
	assert.True(t, proto.Equal(m, m2))

	// unknown fields between known ones, and inside a nested message, are
	// collected in the order met and written back after the known fields of
	// their message, which is what protobuf-go does as well
	innerUnknown := protowire.AppendTag(nil, 500, protowire.VarintType)
	innerUnknown = protowire.AppendVarint(innerUnknown, 1)
	inner := protowire.AppendTag(nil, 1, protowire.VarintType) // Inner.kind = KIND_B
	inner = protowire.AppendVarint(inner, 2)
	inner = append(inner, innerUnknown...)
	nested := append([]byte{}, head...)
	nested = protowire.AppendTag(nested, 1, protowire.BytesType) // Nested.inner
	nested = protowire.AppendBytes(nested, inner)
	nested = append(nested, tail...)
	nested = protowire.AppendTag(nested, 6, protowire.VarintType) // Nested.kind = KIND_A
	nested = protowire.AppendVarint(nested, 1)

	n1 := &Nested{}
	assert.NoError(t, prutal.Unmarshal(nested, n1))
	assert.Equal(t, Nested_Inner_KIND_A, n1.GetKind())
	assert.Equal(t, Nested_Inner_KIND_B, n1.GetInner().GetKind())
	assert.BytesEqual(t, unknown, n1.ProtoReflect().GetUnknown())
	assert.BytesEqual(t, innerUnknown, n1.GetInner().ProtoReflect().GetUnknown())

	n2 := &Nested{}
	assert.NoError(t, proto.Unmarshal(nested, n2))
	assert.True(t, proto.Equal(n1, n2)) // proto.Equal compares unknown fields too
	pb, err = prutal.Marshal(n1)
	assert.NoError(t, err)
	ref, err = proto.Marshal(n2)
	assert.NoError(t, err)
	assert.BytesEqual(t, ref, pb)
}

// TestProto2Defaults checks that the generated getters return the declared
// defaults after prutal decoded a payload that leaves the fields unset.
func TestProto2Defaults(t *testing.T) {
	m := &compat2.Defaults{}
	assert.NoError(t, prutal.Unmarshal(nil, m))
	assert.Equal(t, int32(-42), m.GetI32())
	assert.Equal(t, "hello, world", m.GetS())
	assert.BytesEqual(t, []byte("a,b\000c"), m.GetBy())
	assert.Equal(t, compat2.Level_HIGH, m.GetLevel())
	assert.True(t, math.IsInf(float64(m.GetInf()), 1))
	assert.True(t, math.IsNaN(m.GetNan()))
}

// The cases below pin down what is not supported, so that a change in
// behaviour is noticed and README.md can be kept accurate.

// Extensions set through protobuf-go are not visible to prutal: Marshal drops
// them. On Unmarshal their bytes are kept as unknown fields, so a payload
// carrying extensions is written back unchanged.
func TestUnsupportedExtensions(t *testing.T) {
	m := &compat2.Proto2Message{ReqI32: proto.Int32(1), ReqS: proto.String("x"), ReqMsg: &compat2.Defaults{}}
	proto.SetExtension(m, compat2.E_ExtI32, int32(7))
	proto.SetExtension(m, compat2.E_ExtS, "ext")
	proto.SetExtension(m, compat2.E_ExtRep, []int64{1, 2})
	proto.SetExtension(m, compat2.E_ExtMsg, &compat2.Defaults{I32: proto.Int32(9)})
	ref, err := proto.Marshal(m)
	assert.NoError(t, err)

	pb, err := prutal.Marshal(m)
	assert.NoError(t, err)
	assert.True(t, len(pb) < len(ref), "extensions are dropped by prutal.Marshal")

	m1 := &compat2.Proto2Message{}
	assert.NoError(t, prutal.Unmarshal(ref, m1))
	assert.True(t, !proto.HasExtension(m1, compat2.E_ExtI32))
	assert.True(t, len(m1.ProtoReflect().GetUnknown()) > 0, "extensions are kept as unknown fields")
	pb, err = prutal.Marshal(m1)
	assert.NoError(t, err)
	m2 := &compat2.Proto2Message{}
	assert.NoError(t, proto.Unmarshal(pb, m2))
	assert.True(t, proto.Equal(m, m2), "extensions survive a prutal round-trip as unknown fields")
}

// A message with a group field is rejected as a whole, and so is any message
// that refers to it.
func TestUnsupportedGroups(t *testing.T) {
	_, err := prutal.Marshal(&compat2.WithGroup{Before: proto.Int32(1)})
	assert.ErrorContains(t, err, "group encoding not supported")
	_, err = prutal.Marshal(&compat2.NoGroupSibling{V: proto.Int32(1)})
	assert.ErrorContains(t, err, "group encoding not supported")
	err = prutal.Unmarshal(nil, &compat2.WithGroup{})
	assert.ErrorContains(t, err, "group encoding not supported")
}

// The runtime cannot tell an edition 2023 implicit bytes field in
// protoc-gen-go output from one with presence, so an empty non-nil value is
// written as an empty field where protobuf-go writes nothing; a decoder reads
// both as the default. README.md records it.
func TestKnownDivergenceEditionImplicitBytes(t *testing.T) {
	m := &compat2023.Presence{ByImplicit: []byte{}, ByRequired: []byte{}, IRequired: proto.Int32(0)}
	ref, err := proto.Marshal(m)
	assert.NoError(t, err)
	assert.BytesEqual(t, []byte{0x1a, 0x00, 0x38, 0x00}, ref)
	pb, err := prutal.Marshal(m)
	assert.NoError(t, err)
	assert.BytesEqual(t, []byte{0x12, 0x00, 0x1a, 0x00, 0x38, 0x00}, pb)
	m2 := &compat2023.Presence{}
	assert.NoError(t, proto.Unmarshal(pb, m2))
	assert.True(t, proto.Equal(m, m2))
}

// The opaque API keeps field presence in a bitmap that prutal does not
// maintain: a present zero value is not serialized, and nothing decoded is
// reported as present. The hybrid API has the same layout when built with
// the protoopaque tag.
func TestUnsupportedOpaqueAPI(t *testing.T) {
	t.Run("opaque api", func(t *testing.T) {
		checkOpaqueUnsupported(t, &apiopaque.Msg{}, &apiopaque.Msg{})
	})
	t.Run("hybrid api, protoopaque build", func(t *testing.T) {
		if !hybridOpaque {
			t.Skip("open layout, covered by the matrix")
		}
		checkOpaqueUnsupported(t, &apihybrid.Msg{}, &apihybrid.Msg{})
	})
}

type opaqueMessage interface {
	proto.Message
	SetA(int32)
	SetS(string)
	GetA() int32
	HasA() bool
	HasS() bool
}

func checkOpaqueUnsupported(t *testing.T, m, m1 opaqueMessage) {
	m.SetA(0)
	ref, err := proto.Marshal(m)
	assert.NoError(t, err)
	pb, err := prutal.Marshal(m)
	assert.NoError(t, err)
	assert.True(t, len(pb) < len(ref), "a present zero value is dropped")

	m.SetA(5)
	m.SetS("")
	ref, err = proto.Marshal(m)
	assert.NoError(t, err)
	assert.NoError(t, prutal.Unmarshal(ref, m1))
	assert.Equal(t, int32(5), m1.GetA())
	assert.True(t, !m1.HasA(), "presence is not restored")
	assert.True(t, !m1.HasS(), "presence of an empty string is not restored")
}
