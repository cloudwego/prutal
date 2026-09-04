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

package prutalgen

import (
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestGenCode(t *testing.T) {
	g := NewGoCodeGen()
	g.Getter = true
	g.Marshaler = MarshalerKitexProtobuf
	p := &Proto{ProtoFile: filepath.Join(t.TempDir(), "test.proto"), l: testLogger{t}}
	p.GoPackage = "test"

	p.Enums = []*Enum{{
		GoName: "Enum1",
		Name:   "enum1",
		Fields: []*EnumField{
			{GoName: "Enum1_Zero", Name: "enum1_zero", Value: 0},
			{GoName: "Enum1_N", Name: "enum1_n", Value: 10001},
		},
		Proto: p,
	}}

	m := &Message{
		GoName: "Message1",
		Fields: []*Field{
			{GoName: "Field1", Type: &Type{Name: "string"}, FieldNumber: 20001},
			{GoName: "Field2", Type: &Type{Name: "uint64"}, Repeated: true, FieldNumber: 20002},
			{GoName: "Field3", Type: &Type{Name: "string"}, Optional: true, FieldNumber: 20003},
			{GoName: "Field4", Key: &Type{Name: "string"}, Type: &Type{Name: "string"}, FieldNumber: 20004},
		},
		Proto: p,
	}
	p.Messages = []*Message{m}

	m.Oneofs = []*Oneof{{
		Name: "Oneof1",
		Fields: []*Field{
			{GoName: "OneofA", Type: &Type{Name: "string"}, FieldNumber: 20101},
			{GoName: "OneofB", Type: &Type{Name: "string"}, FieldNumber: 20102},
		},
		Msg: m,
	}}
	oneof := m.Oneofs[0]
	oneof.Fields[0].Oneof = oneof
	oneof.Fields[1].Oneof = oneof
	m.Fields = append(m.Fields, oneof.Fields...)
	for _, f := range m.Fields {
		f.Msg = m
	}

	m.Enums = []*Enum{{
		GoName: "NestedEnum",
		Name:   "nested_enum",
		Fields: []*EnumField{
			{GoName: "NestedEnum_Zero", Name: "nested_enum_zero", Value: 0},
			{GoName: "NestedEnum_N", Name: "nested_enum_n", Value: 30001},
		},
		Proto: p,
	}}

	m.Messages = []*Message{{
		GoName: "NestedMsg",
		Fields: []*Field{
			{GoName: "NField1", Type: &Type{Name: "string"}, FieldNumber: 40001},
		},
		Proto: p,
	}}
	for _, f := range m.Messages[0].Fields {
		f.Msg = m.Messages[0]
	}

	_ = g.Gen(p, GenBySourceRelative, "")

	outfn := filepath.Join(filepath.Dir(p.ProtoFile), "test.pb.go")
	b, err := os.ReadFile(outfn)
	assert.NoError(t, err)

	src := string(b)
	lines := strings.Split(src, "\n")
	assertLine := func(s string) {
		t.Helper()
		for _, l := range lines {
			if strings.Contains(l, s) {
				return
			}
		}
		t.Fatal("not match", s)
	}
	t.Log(src)
	_, err = format.Source(b)
	assert.NoError(t, err)

	assertLine("Enum1 = 0")
	assertLine("Enum1 = 10001")
	assertLine("NestedEnum = 0")
	assertLine("NestedEnum = 30001")
	assertLine("Field1 string")
	assertLine("Field2 []uint64")
	assertLine("Field3 *string")
	assertLine("Field4 map[string]string")
	assertLine("OneofA string")
	assertLine("OneofB string")
	assertLine("NField1 string")
	assertLine("prutal.MarshalAppend")
	assertLine("prutal.Unmarshal")
}

func TestSourcePathUsesLogicalProtoName(t *testing.T) {
	g := NewGoCodeGen()
	p := &Proto{
		ProtoFile: filepath.Join(t.TempDir(), "physical", "test.proto"),
		protoName: "logical/nested/test.proto",
		GoImport:  "example.com/project/testpb",
	}
	out := filepath.Join(t.TempDir(), "out")
	assert.Equal(t,
		filepath.Join(out, "logical", "nested", "test.pb.go"),
		g.SourcePath(p, GenBySourceRelative, out, ".pb.go"),
	)
	assert.Equal(t,
		filepath.Join(out, "example.com", "project", "testpb", "test.pb.go"),
		g.SourcePath(p, GenByImport, out, ".pb.go"),
	)

	g.Module = "example.com/project"
	assert.Equal(t,
		filepath.Join(out, "testpb", "test.pb.go"),
		g.SourcePath(p, GenByImport, out, ".pb.go"),
	)

	g.Module = "example.net/other"
	_, err := g.sourcePath(p, GenByImport, out, ".pb.go")
	assert.ErrorContains(t, err, `generated file does not match prefix "example.net/other"`)
	_, err = g.sourcePath(p, GenBySourceRelative, out, ".pb.go")
	assert.ErrorContains(t, err, "cannot use module= with paths=source_relative")

	g.Module = ""
	p.protoName = "logical/nested/test.protodevel"
	assert.Equal(t,
		filepath.Join(out, "logical", "nested", "test.pb.go"),
		g.SourcePath(p, GenBySourceRelative, out, ".pb.go"),
	)
	assert.Equal(t,
		filepath.Join(out, "example.com", "project", "testpb", "test.pb.go"),
		g.SourcePath(p, GenByImport, out, ".pb.go"),
	)
}

func TestValidateOutputPaths(t *testing.T) {
	out := t.TempDir()
	g := NewGoCodeGen()
	same := &Proto{protoName: "same.proto", GoImport: "example.com/project/same"}
	assert.NoError(t, g.ValidateOutputPaths([]*Proto{same, same}, GenByImport, out))

	protos := []*Proto{
		{protoName: "a/foo.proto", GoImport: "example.com/project/pkg"},
		{protoName: "b/foo.proto", GoImport: "example.com/project/pkg"},
	}
	assert.ErrorContains(t,
		g.ValidateOutputPaths(protos, GenByImport, out),
		"tried to write the same file twice",
	)

	g.Module = "example.com/project"
	protos[1].GoImport = "example.net/outside"
	assert.ErrorContains(t,
		g.ValidateOutputPaths(protos, GenByImport, out),
		`generated file does not match prefix "example.com/project"`,
	)

	g.Module = ""
	protos = []*Proto{{protoName: "../outside.proto", GoImport: "example.com/project/pkg"}}
	assert.ErrorContains(t,
		g.ValidateOutputPaths(protos, GenBySourceRelative, out),
		`invalid generated file path "../outside.pb.go"`,
	)
}

func TestProtoGenFlattenedDeclarationOrder(t *testing.T) {
	p := &Proto{GoPackage: "test", Directives: Directives{prutalNoEnumMapping}}
	outer := &Message{GoName: "Outer", Proto: p}
	other := &Message{GoName: "Other", Proto: p}
	inner := &Message{GoName: "OuterInner", Msg: outer, Proto: p}
	deep := &Message{GoName: "OuterInnerDeep", Msg: inner, Proto: p}
	outer.Messages = []*Message{inner}
	inner.Messages = []*Message{deep}
	p.Messages = []*Message{outer, other}

	p.Enums = []*Enum{{GoName: "TopEnum", Proto: p}}
	outer.Enums = []*Enum{{GoName: "OuterEnum", Msg: outer, Proto: p}}
	inner.Enums = []*Enum{{GoName: "InnerEnum", Msg: inner, Proto: p}}
	deep.Enums = []*Enum{{GoName: "DeepEnum", Msg: deep, Proto: p}}

	w := NewCodeWriter("", p.GoPackage)
	NewGoCodeGen().ProtoGen(p, w)
	src := string(w.Bytes())

	last := -1
	for _, declaration := range []string{
		"type TopEnum int32",
		"type OuterEnum int32",
		"type InnerEnum int32",
		"type DeepEnum int32",
		"type Outer struct",
		"type Other struct",
		"type OuterInner struct",
		"type OuterInnerDeep struct",
	} {
		pos := strings.Index(src, declaration)
		assert.True(t, pos > last, declaration)
		last = pos
	}
}

func TestProtoGenTrailingComments(t *testing.T) {
	p := &Proto{GoPackage: "test", Directives: Directives{prutalNoEnumMapping}}
	e := &Enum{
		GoName:        "E",
		InlineComment: "// enum trailing must be omitted",
		Fields: []*EnumField{
			{GoName: "E_ZERO", Value: 0, InlineComment: "// value is 100% valid"},
			{GoName: "E_ONE", Value: 1, InlineComment: "/* enum value multi\nline */"},
		},
		Proto: p,
	}
	m := &Message{
		GoName:        "M",
		InlineComment: "// message trailing must be omitted",
		Fields: []*Field{
			{GoName: "A", Type: &Type{Name: "string"}, FieldNumber: 1, InlineComment: "// field is 100% valid"},
			{GoName: "B", Type: &Type{Name: "string"}, FieldNumber: 2, InlineComment: "/* field multi\nline */"},
		},
		Proto: p,
	}
	for _, f := range m.Fields {
		f.Msg = m
	}
	p.Enums = []*Enum{e}
	p.Messages = []*Message{m}

	w := NewCodeWriter("", p.GoPackage)
	NewGoCodeGen().ProtoGen(p, w)
	src := string(w.Bytes())
	_, err := format.Source([]byte(src))
	assert.NoError(t, err)
	assert.True(t, strings.Contains(src, "E_ZERO E = 0 // value is 100% valid"))
	assert.True(t, strings.Contains(src, "A string `protobuf:"))
	assert.True(t, strings.Contains(src, "// field is 100% valid"))
	assert.False(t, strings.Contains(src, "enum trailing must be omitted"))
	assert.False(t, strings.Contains(src, "message trailing must be omitted"))
	assert.False(t, strings.Contains(src, "enum value multi"))
	assert.False(t, strings.Contains(src, "field multi"))
}

func TestProtoGenUsesAllocatedImportAliases(t *testing.T) {
	p := &Proto{GoPackage: "test"}
	targets := []*Proto{
		{GoImport: "example.com/a/bar", GoPackage: "first"},
		{GoImport: "example.net/b/bar", GoPackage: "second"},
		{GoImport: "example.org/string", GoPackage: "third"},
	}
	fields := make([]*Field, len(targets))
	for i, target := range targets {
		declaration := &Message{GoName: string(rune('A' + i)), Proto: target}
		target.Messages = []*Message{declaration}
		fields[i] = &Field{
			GoName:      "Field" + string(rune('A'+i)),
			FieldNumber: int32(i + 1),
			Type:        &Type{Name: declaration.Name, typ: declaration, p: target},
		}
	}
	m := &Message{GoName: "M", Fields: fields, Proto: p}
	for _, f := range fields {
		f.Msg = m
	}
	p.Messages = []*Message{m}

	w := NewCodeWriter("", p.GoPackage)
	NewGoCodeGen().ProtoGen(p, w)
	src := string(w.Bytes())
	_, err := format.Source([]byte(src))
	assert.NoError(t, err)
	assert.True(t, strings.Contains(src, "FieldA *bar.A"))
	assert.True(t, strings.Contains(src, "FieldB *bar1.B"))
	assert.True(t, strings.Contains(src, "FieldC *string1.C"))
	assert.True(t, strings.Contains(src, `bar "example.com/a/bar"`))
}

func TestProtoGenUsesAllocatedRuntimeAlias(t *testing.T) {
	p := &Proto{GoPackage: "test"}
	target := &Proto{GoImport: "example.com/acme/prutal", GoPackage: "different"}
	declaration := &Message{GoName: "External", Proto: target}
	field := &Field{
		GoName:      "External",
		FieldNumber: 1,
		Type:        &Type{Name: "External", typ: declaration, p: target},
	}
	message := &Message{GoName: "M", Fields: []*Field{field}, Proto: p}
	field.Msg = message
	p.Messages = []*Message{message}

	g := NewGoCodeGen()
	g.Marshaler = MarshalerKitexProtobuf
	w := NewCodeWriter("", p.GoPackage)
	g.ProtoGen(p, w)
	src := string(w.Bytes())
	_, err := format.Source([]byte(src))
	assert.NoError(t, err)
	assert.True(t, strings.Contains(src, `prutal "example.com/acme/prutal"`))
	assert.True(t, strings.Contains(src, `prutal1 "github.com/cloudwego/prutal"`))
	assert.True(t, strings.Contains(src, "return prutal1.MarshalAppend(in, x)"))
	assert.True(t, strings.Contains(src, "return prutal1.Unmarshal(in, x)"))
}

func TestProtoGenUsesAllocatedStrconvAlias(t *testing.T) {
	p := &Proto{GoPackage: "test"}
	p.Enums = []*Enum{{
		GoName: "E",
		Fields: []*EnumField{{
			GoName: "E_ZERO",
		}},
		Proto: p,
	}}
	target := &Proto{GoImport: "example.com/acme/strconv", GoPackage: "different"}
	declaration := &Message{GoName: "External", Proto: target}
	field := &Field{
		GoName:      "External",
		FieldNumber: 1,
		Type:        &Type{Name: "External", typ: declaration, p: target},
	}
	message := &Message{GoName: "M", Fields: []*Field{field}, Proto: p}
	field.Msg = message
	p.Messages = []*Message{message}

	w := NewCodeWriter("", p.GoPackage)
	NewGoCodeGen().ProtoGen(p, w)
	src := string(w.Bytes())
	_, err := format.Source([]byte(src))
	assert.NoError(t, err)
	assert.True(t, strings.Contains(src, `strconv "strconv"`))
	assert.True(t, strings.Contains(src, `strconv1 "example.com/acme/strconv"`))
	assert.True(t, strings.Contains(src, "return strconv.Itoa(int(x))"))
	assert.True(t, strings.Contains(src, "External *strconv1.External"))
}

func TestFieldAndOneofGenTrackExternalPackages(t *testing.T) {
	target := &Proto{GoImport: "time", GoPackage: "time"}
	declaration := &Message{GoName: "Duration", Proto: target}
	local := &Message{GoName: "Local", Proto: &Proto{GoPackage: "test"}}
	newField := func() *Field {
		return &Field{
			GoName:      "Value",
			FieldNumber: 1,
			Type: &Type{
				Name: "Duration",
				typ:  declaration,
				p:    target,
			},
			Msg: local,
		}
	}

	g := NewGoCodeGen()
	w := NewCodeWriter("", "test")
	w.F("type Local struct {")
	g.FieldGen(newField(), w)
	w.F("}")
	src := w.Bytes()
	assert.StringContains(t, string(src), `time "time"`)
	typeCheckSource(t, src)

	oneof := &Oneof{Name: "choice", Msg: local}
	field := newField()
	field.Oneof = oneof
	oneof.Fields = []*Field{field}
	w = NewCodeWriter("", "test")
	g.OneofGen(oneof, w)
	src = w.Bytes()
	assert.StringContains(t, string(src), `time "time"`)
	typeCheckSource(t, src)
}

func TestFieldStructTagPresenceMarkers(t *testing.T) {
	jsonTag := regexp.MustCompile(` json:"[^"]*"`)
	fieldTag := func(f *Field) string {
		return jsonTag.ReplaceAllString(string((&GoCodeGen{}).FieldStructTag(f)), "")
	}
	tag := func(p *Proto, msg, field int) string { return fieldTag(p.Messages[msg].Fields[field]) }

	p := loadTestProto(t, `
syntax = "proto3";
option go_package = "example.com/tag";
message M {
  int32 a = 1;
  optional int32 b = 2;
  repeated int32 c = 3;
  bytes d = 4;
  oneof k { int32 e = 5; }
  map<string, int32> f = 6;
}
`)
	assert.Equal(t, `protobuf:"varint,1,opt,name=a,proto3"`, tag(p, 0, 0))
	assert.Equal(t, `protobuf:"varint,2,opt,name=b,proto3,oneof"`, tag(p, 0, 1))
	assert.Equal(t, `protobuf:"varint,3,rep,packed,name=c,proto3"`, tag(p, 0, 2))
	assert.Equal(t, `protobuf:"bytes,4,opt,name=d,proto3"`, tag(p, 0, 3))
	assert.Equal(t, `protobuf:"varint,5,opt,name=e,proto3,oneof"`, tag(p, 0, 4))
	assert.Equal(t, `protobuf:"bytes,6,rep,name=f,proto3" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`, tag(p, 0, 5))

	p = loadTestProto(t, `
syntax = "proto2";
option go_package = "example.com/tag";
message M {
  optional bytes a = 1;
  required int32 b = 2;
  oneof k { int32 c = 3; }
}
`)
	assert.Equal(t, `protobuf:"bytes,1,opt,name=a"`, tag(p, 0, 0))
	assert.Equal(t, `protobuf:"varint,2,req,name=b"`, tag(p, 0, 1))
	assert.Equal(t, `protobuf:"varint,3,opt,name=c,oneof"`, tag(p, 0, 2))

	// edition 2023: presence is a feature, set on the field or on the file,
	// in the dotted or the aggregate form
	p = loadTestProto(t, `
edition = "2023";
option go_package = "example.com/tag";
option features.field_presence = EXPLICIT;
enum E { E_UNSPECIFIED = 0; }
message M {
  bytes a = 1;
  bytes b = 2 [features.field_presence = IMPLICIT];
  repeated bytes c = 3;
  M d = 4 [features.field_presence = EXPLICIT];
  bytes e = 5 [features = { field_presence: IMPLICIT }];
  int32 f = 6 [features = { field_presence: IMPLICIT }];
  bytes g = 7 [features.field_presence = LEGACY_REQUIRED];
  int32 h = 8 [features.field_presence = LEGACY_REQUIRED];
  E i = 9 [features.field_presence = IMPLICIT];
  M j = 10 [features.field_presence = LEGACY_REQUIRED];
}
`)
	assert.True(t, p.Options.Is(f_field_presence, "EXPLICIT"))
	assert.Equal(t, `protobuf:"bytes,1,opt,name=a"`, tag(p, 0, 0))
	assert.Equal(t, `protobuf:"bytes,2,opt,name=b,implicit"`, tag(p, 0, 1))
	assert.Equal(t, `protobuf:"bytes,3,rep,name=c"`, tag(p, 0, 2))
	assert.Equal(t, `protobuf:"bytes,4,opt,name=d"`, tag(p, 0, 3))
	assert.Equal(t, `protobuf:"bytes,5,opt,name=e,implicit"`, tag(p, 0, 4))
	assert.Equal(t, `protobuf:"varint,6,opt,name=f,implicit"`, tag(p, 0, 5))
	assert.False(t, p.Messages[0].Fields[5].IsPointer())
	assert.Equal(t, `protobuf:"bytes,7,req,name=g"`, tag(p, 0, 6))
	assert.Equal(t, `protobuf:"varint,8,req,name=h"`, tag(p, 0, 7))
	assert.True(t, p.Messages[0].Fields[7].IsPointer())
	assert.True(t, p.Messages[0].Fields[8].isImplicitPresence())
	assert.False(t, p.Messages[0].Fields[8].IsPointer())
	assert.True(t, p.Messages[0].Fields[9].Required)
	assert.True(t, p.Messages[0].Fields[9].IsPointer())

	p = loadTestProto(t, `
edition = "2023";
option go_package = "example.com/tag";
option features = { field_presence: IMPLICIT };
message M {
  bytes a = 1;
  bytes b = 2 [features.field_presence = EXPLICIT];
  oneof k { bytes c = 3; }
  int32 d = 4;
  int32 e = 5 [features = { field_presence: EXPLICIT }];
  M f = 6;
  int32 g = 7 [features.field_presence = EXPLICIT, default = 1];
  repeated int32 h = 8;
  map<string, int32> i = 9;
  message Inner { bytes a = 1; }
}
`)
	assert.Equal(t, `protobuf:"bytes,1,opt,name=a,implicit"`, tag(p, 0, 0))
	assert.Equal(t, `protobuf:"bytes,2,opt,name=b"`, tag(p, 0, 1))
	assert.Equal(t, `protobuf:"bytes,3,opt,name=c,oneof"`, tag(p, 0, 2))
	assert.Equal(t, `protobuf:"varint,4,opt,name=d,implicit"`, tag(p, 0, 3))
	assert.False(t, p.Messages[0].Fields[3].IsPointer())
	assert.Equal(t, `protobuf:"varint,5,opt,name=e"`, tag(p, 0, 4))
	assert.True(t, p.Messages[0].Fields[4].IsPointer())
	assert.Equal(t, `protobuf:"bytes,6,opt,name=f"`, tag(p, 0, 5))
	assert.False(t, p.Messages[0].Fields[5].isImplicitPresence())
	assert.True(t, p.Messages[0].Fields[5].IsPointer())
	assert.True(t, p.Messages[0].Fields[6].Options.Is(option_default, "1"))
	assert.False(t, p.Messages[0].Fields[6].isImplicitPresence())
	assert.True(t, p.Messages[0].Fields[6].IsPointer())
	assert.False(t, p.Messages[0].Fields[7].isImplicitPresence())
	assert.False(t, p.Messages[0].Fields[8].isImplicitPresence())
	assert.Equal(t, `protobuf:"bytes,1,opt,name=a,implicit"`, fieldTag(p.Messages[0].Messages[0].Fields[0]))

	// A file-level implicit default does not apply to repeated, oneof, or
	// extension fields, even when their enum type is closed.
	p = loadTestProto(t, `
edition = "2023";
option go_package = "example.com/tag";
option features.field_presence = IMPLICIT;
option features.enum_type = CLOSED;
enum E { E_UNSPECIFIED = 0; }
enum OpenE {
  option features.enum_type = OPEN;
  OPEN_E_UNSPECIFIED = 0;
}
message M {
  repeated E a = 1;
  oneof choice { E b = 2; }
  map<string, OpenE> c = 3;
  extensions 100 to max;
}
extend M { E ext = 100; }
`)
	assert.False(t, p.Messages[0].Fields[0].isImplicitPresence())
	assert.False(t, p.Messages[0].Fields[1].isImplicitPresence())
	assert.False(t, p.Messages[0].Fields[2].isImplicitPresence())
}

func TestInvalidFieldPresenceRejected(t *testing.T) {
	tests := []struct {
		name  string
		proto string
		want  string
	}{
		{
			name: "proto2",
			proto: `syntax = "proto2";
option go_package = "example.com/tag";
option features.field_presence = EXPLICIT;
`,
			want: `"features.field_presence" is only available in editions`,
		},
		{
			name: "proto3 field",
			proto: `syntax = "proto3";
option go_package = "example.com/tag";
message M { int32 a = 1 [features.field_presence = IMPLICIT]; }
`,
			want: `"features.field_presence" is only available in editions`,
		},
		{
			name: "file legacy required",
			proto: `edition = "2023";
option go_package = "example.com/tag";
option features.field_presence = LEGACY_REQUIRED;
`,
			want: `"features.field_presence" cannot default to "LEGACY_REQUIRED" on a file`,
		},
		{
			name: "message",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M {
  option features.field_presence = IMPLICIT;
  int32 a = 1;
}
`,
			want: `"features.field_presence" cannot be set on a message`,
		},
		{
			name: "oneof",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M {
  oneof choice {
    option features.field_presence = EXPLICIT;
    int32 a = 1;
  }
}
`,
			want: `"features.field_presence" cannot be set on a oneof`,
		},
		{
			name: "enum",
			proto: `edition = "2023";
option go_package = "example.com/tag";
enum E {
  option features.field_presence = EXPLICIT;
  E_UNSPECIFIED = 0;
}
`,
			want: `"features.field_presence" cannot be set on an enum`,
		},
		{
			name: "enum entry",
			proto: `edition = "2023";
option go_package = "example.com/tag";
enum E { E_UNSPECIFIED = 0 [features.field_presence = EXPLICIT]; }
`,
			want: `"features.field_presence" cannot be set on an enum entry`,
		},
		{
			name: "service",
			proto: `edition = "2023";
option go_package = "example.com/tag";
service S { option features.field_presence = EXPLICIT; }
`,
			want: `"features.field_presence" cannot be set on a service`,
		},
		{
			name: "RPC",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M {}
service S {
  rpc Call(M) returns (M) { option features.field_presence = EXPLICIT; }
}
`,
			want: `"features.field_presence" cannot be set on an RPC`,
		},
		{
			name: "repeated field",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M { repeated int32 a = 1 [features.field_presence = EXPLICIT]; }
`,
			want: `cannot be set on repeated field "a"`,
		},
		{
			name: "map field",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M { map<string, int32> a = 1 [features.field_presence = EXPLICIT]; }
`,
			want: `cannot be set on map field "a"`,
		},
		{
			name: "oneof field",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M { oneof choice { int32 a = 1 [features.field_presence = EXPLICIT]; } }
`,
			want: `cannot be set on oneof field "a"`,
		},
		{
			name: "implicit message field",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M { M a = 1 [features.field_presence = IMPLICIT]; }
`,
			want: `cannot use "IMPLICIT" on message field "a"`,
		},
		{
			name: "extension field",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M { extensions 100 to max; }
extend M { int32 a = 100 [features = { field_presence: EXPLICIT }]; }
`,
			want: `"features.field_presence" cannot be set on an extension field`,
		},
		{
			name: "extension range",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M { extensions 100 to max [features.field_presence = EXPLICIT]; }
`,
			want: `"features.field_presence" cannot be set on an extension range`,
		},
		{
			name: "unknown value",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M { int32 a = 1 [features.field_presence = REQUIRED]; }
`,
			want: `unsupported value "REQUIRED"`,
		},
		{
			name: "implicit closed enum from enum",
			proto: `edition = "2023";
option go_package = "example.com/tag";
enum E {
  option features.enum_type = CLOSED;
  E_UNSPECIFIED = 0;
}
message M { E a = 1 [features.field_presence = IMPLICIT]; }
`,
			want: `implicit presence enum field "a" must use an open enum`,
		},
		{
			name: "implicit closed enum from file",
			proto: `edition = "2023";
option go_package = "example.com/tag";
option features = { field_presence: IMPLICIT enum_type: CLOSED };
enum E { E_UNSPECIFIED = 0; }
message M { E a = 1; }
`,
			want: `implicit presence enum field "a" must use an open enum`,
		},
		{
			name: "implicit closed enum map value",
			proto: `edition = "2023";
option go_package = "example.com/tag";
option features.field_presence = IMPLICIT;
enum E {
  option features.enum_type = CLOSED;
  E_UNSPECIFIED = 0;
}
message M { map<string, E> a = 1; }
`,
			want: `map field "a" with implicit file presence must use an open enum value`,
		},
		{
			name: "direct implicit default",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M {
  int32 a = 1 [features.field_presence = IMPLICIT, default = 1];
}
`,
			want: `implicit presence field "a" cannot specify a default`,
		},
		{
			name: "inherited implicit default",
			proto: `edition = "2023";
option go_package = "example.com/tag";
option features.field_presence = IMPLICIT;
message M { int32 a = 1 [default = 1]; }
`,
			want: `implicit presence field "a" cannot specify a default`,
		},
		{
			name: "edition optional label",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M { optional int32 a = 1; }
`,
			want: "`optional` keyword is not available in editions",
		},
		{
			name: "edition extension optional label",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M { extensions 100 to max; }
extend M { optional int32 a = 100; }
`,
			want: "`optional` keyword is not available in editions",
		},
		{
			name: "edition extension required label",
			proto: `edition = "2023";
option go_package = "example.com/tag";
message M { extensions 100 to max; }
extend M { required int32 a = 100; }
`,
			want: "`required` keyword only available for proto2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectProtoError(t, tt.proto, tt.want)
		})
	}
}
