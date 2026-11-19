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

func TestEdition2023FeaturesValid(t *testing.T) {
	// Valid for protoc v29.3: imports embedded go_features.proto and uses
	// standard and Go feature spellings (including file-level DELIMITED when
	// all generated message fields override LENGTH_PREFIXED, and file-level
	// API_OPAQUE when all root messages override API_OPEN).
	fn := writeFile(t, "ed2023.proto", []byte(`
edition = "2023";
option go_package = "example.com/ed2023";
import "google/protobuf/go_features.proto";

option features.message_encoding = DELIMITED;
option features.utf8_validation = NONE;
option features.json_format = LEGACY_BEST_EFFORT;
option features.(pb.go).api_level = API_OPAQUE;
option features.(.pb.go).legacy_unmarshal_json_enum = true;

enum E {
  option features.enum_type = CLOSED;
  option features.json_format = ALLOW;
  option features.(pb.go).legacy_unmarshal_json_enum = false;
  E_A = 1;
}

message Outer {
  option features.(pb.go) = { api_level: API_OPEN };
  option features.json_format = ALLOW;

  message Inner {
    // inherits API_OPEN from Outer
    int32 v = 1;
  }

  Inner sub = 1 [features.message_encoding = LENGTH_PREFIXED];
  repeated Inner subs = 2 [features.message_encoding = LENGTH_PREFIXED, features.repeated_field_encoding = EXPANDED];
  oneof choice {
    Inner one_sub = 3 [features.message_encoding = LENGTH_PREFIXED];
    string one_str = 4 [features.utf8_validation = VERIFY];
  }
  map<string, Inner> m_sub = 5 [features.utf8_validation = VERIFY, features.repeated_field_encoding = EXPANDED];
  map<int32, string> m_str = 6 [features.utf8_validation = NONE];
  repeated string strs = 7 [features.utf8_validation = VERIFY, features.repeated_field_encoding = EXPANDED];
  repeated E enums = 8 [features.repeated_field_encoding = PACKED];

  extensions 100 to max;
}

message OpenAggMsg {
  option features = { [pb.go]: { api_level: API_OPEN } };
  int32 x = 1;
}

extend Outer {
  repeated int32 ext_rep = 100 [features.repeated_field_encoding = PACKED];
  Outer.Inner ext_msg = 101 [features.message_encoding = DELIMITED];
}
`))
	loader := NewLoader([]string{filepath.Dir(fn)}, nil)
	loader.SetLogger(&testLogger{t})
	p := loader.LoadProtos([]string{fn})[0]
	assert.Equal(t, 2, len(p.Messages))
	assert.Equal(t, "API_OPEN", p.Messages[0].effectiveGoAPILevel())
	assert.Equal(t, "API_OPEN", p.Messages[0].Messages[0].effectiveGoAPILevel())
	assert.Equal(t, "API_OPEN", p.Messages[1].effectiveGoAPILevel())

	// Aggregate-first followed by non-overlapping dotted or Go-extension options
	// is valid in protoc v29.3.
	aggOrderCases := []struct {
		name  string
		proto string
	}{
		{
			name: "features aggregate then dotted",
			proto: `edition = "2023";
option go_package = "example.com/o1";
option features = { enum_type: CLOSED };
option features.utf8_validation = NONE;
message M {
  string a = 1 [features = { field_presence: IMPLICIT }, features.utf8_validation = NONE];
}
`,
		},
		{
			name: "go features aggregate then dotted",
			proto: `edition = "2023";
option go_package = "example.com/o2";
import "google/protobuf/go_features.proto";
option features.(pb.go) = { api_level: API_OPEN };
option features.(pb.go).legacy_unmarshal_json_enum = true;
message M { int32 a = 1; }
`,
		},
		{
			name: "features aggregate with pb.go then dotted go feature",
			proto: `edition = "2023";
option go_package = "example.com/o3";
import "google/protobuf/go_features.proto";
option features = { [pb.go]: { api_level: API_OPEN } };
option features.(pb.go).legacy_unmarshal_json_enum = true;
message M { int32 a = 1; }
`,
		},
		{
			name: "features aggregate then go features aggregate",
			proto: `edition = "2023";
option go_package = "example.com/o4";
import "google/protobuf/go_features.proto";
option features = { enum_type: CLOSED };
option features.(pb.go) = { api_level: API_OPEN };
message M { int32 a = 1; }
`,
		},
	}
	for _, tc := range aggOrderCases {
		t.Run(tc.name, func(t *testing.T) {
			f := writeFile(t, "order.proto", []byte(tc.proto))
			l := NewLoader([]string{filepath.Dir(f)}, nil)
			l.SetLogger(&testLogger{t})
			roots := l.LoadProtos([]string{f})
			assert.Equal(t, 1, len(roots))
		})
	}
}

func TestEdition2023ForeignFeaturePassthrough(t *testing.T) {
	// Non-Go language feature extensions pass through without error.
	p1 := loadTestProto(t, `
edition = "2023";
option go_package = "example.com/cpp1";
option features.(pb.cpp).string_type = VIEW;
message M { int32 a = 1; }
`)
	assert.Equal(t, 1, len(p1.Messages))

	p2 := loadTestProto(t, `
edition = "2023";
option go_package = "example.com/cpp2";
option features = { [pb.cpp]: { string_type: STRING } };
message M { int32 a = 1; }
`)
	assert.Equal(t, 1, len(p2.Messages))
}

func TestProto3UnpackedTag(t *testing.T) {
	p3 := loadTestProto(t, `
syntax = "proto3";
option go_package = "example.com/p3";
message M {
  repeated int32 packed_def = 1;
  repeated int32 unpacked = 2 [packed = false];
}
`)
	g := NewGoCodeGen()
	assert.Equal(t, `protobuf:"varint,1,rep,packed,name=packed_def,proto3" json:"packed_def,omitempty"`, string(g.FieldStructTag(p3.Messages[0].Fields[0])))
	assert.Equal(t, `protobuf:"varint,2,rep,name=unpacked,proto3" json:"unpacked,omitempty"`, string(g.FieldStructTag(p3.Messages[0].Fields[1])))
}

func TestEdition2023ImportedFeatureReachability(t *testing.T) {
	dir := t.TempDir()
	depPath := filepath.Join(dir, "dep.proto")
	depSrc := []byte(`
edition = "2023";
package dep;
option go_package = "example.com/dep";
import public "google/protobuf/go_features.proto";

enum DepEnum {
  DEP_UNSPECIFIED = 0;
  DEP_A = 1;
}

message Sub { int32 v = 1; }

message OpaqueMsg {
  option features.(pb.go).api_level = API_OPAQUE;
  int32 a = 1;
}

message HybridMsg {
  option features.(pb.go).api_level = API_HYBRID;
  int32 a = 1;
}

message DelimitedMsg {
  Sub s = 1 [features.message_encoding = DELIMITED];
}
`)
	assert.NoError(t, os.WriteFile(depPath, depSrc, 0o600))

	// Root referencing only DepEnum, HybridMsg, and using transitive public
	// import of go_features.proto must succeed despite unrelated OpaqueMsg and
	// DelimitedMsg in dep.proto.
	okRoot := filepath.Join(dir, "ok_root.proto")
	assert.NoError(t, os.WriteFile(okRoot, []byte(`
edition = "2023";
package root;
option go_package = "example.com/root";
import "dep.proto";
option features.(pb.go).api_level = API_OPEN;
message Root {
  dep.DepEnum e = 1;
  dep.HybridMsg h = 2;
}
`), 0o600))
	loader := NewLoader([]string{dir}, nil)
	loader.SetLogger(&testLogger{t})
	roots := loader.LoadProtos([]string{okRoot})
	assert.Equal(t, 1, len(roots))

	// Root referencing dep.OpaqueMsg must fail.
	opaqueRoot := filepath.Join(dir, "opaque_root.proto")
	assert.NoError(t, os.WriteFile(opaqueRoot, []byte(`
edition = "2023";
package root;
option go_package = "example.com/root";
import "dep.proto";
message Root { dep.OpaqueMsg m = 1; }
`), 0o600))
	t.Run("reachable imported opaque message", func(t *testing.T) {
		l := NewLoader([]string{dir}, nil)
		l.SetLogger(&expectLogger{t: t, FatalContains: `imported message "dep.OpaqueMsg" uses unsupported Go API level "API_OPAQUE"`})
		_ = l.LoadProtos([]string{opaqueRoot})
		t.Fatal("expected error")
	})

	// Root referencing dep.DelimitedMsg must fail.
	delimRoot := filepath.Join(dir, "delim_root.proto")
	assert.NoError(t, os.WriteFile(delimRoot, []byte(`
edition = "2023";
package root;
option go_package = "example.com/root";
import "dep.proto";
message Root { dep.DelimitedMsg m = 1; }
`), 0o600))
	t.Run("reachable imported delimited message", func(t *testing.T) {
		l := NewLoader([]string{dir}, nil)
		l.SetLogger(&expectLogger{t: t, FatalContains: `imported message "dep.DelimitedMsg" field "s" uses unsupported delimited message_encoding`})
		_ = l.LoadProtos([]string{delimRoot})
		t.Fatal("expected error")
	})

	// Transitive reachability: root -> mid.Wrapper -> dep.OpaqueMsg.
	midPath := filepath.Join(dir, "mid.proto")
	assert.NoError(t, os.WriteFile(midPath, []byte(`
edition = "2023";
package mid;
option go_package = "example.com/mid";
import "dep.proto";
message Wrapper { dep.OpaqueMsg inner = 1; }
`), 0o600))
	transRoot := filepath.Join(dir, "trans_root.proto")
	assert.NoError(t, os.WriteFile(transRoot, []byte(`
edition = "2023";
package root;
option go_package = "example.com/root";
import "mid.proto";
message Root { mid.Wrapper w = 1; }
`), 0o600))
	t.Run("transitively reachable imported opaque message", func(t *testing.T) {
		l := NewLoader([]string{dir}, nil)
		l.SetLogger(&expectLogger{t: t, FatalContains: `imported message "dep.OpaqueMsg" uses unsupported Go API level "API_OPAQUE"`})
		_ = l.LoadProtos([]string{transRoot})
		t.Fatal("expected error")
	})

	// Map-value reachability: map<string, dep.DelimitedMsg>.
	mapRoot := filepath.Join(dir, "map_root.proto")
	assert.NoError(t, os.WriteFile(mapRoot, []byte(`
edition = "2023";
package root;
option go_package = "example.com/root";
import "dep.proto";
message Root { map<string, dep.DelimitedMsg> items = 1; }
`), 0o600))
	t.Run("map-value reachable imported delimited message", func(t *testing.T) {
		l := NewLoader([]string{dir}, nil)
		l.SetLogger(&expectLogger{t: t, FatalContains: `imported message "dep.DelimitedMsg" field "s" uses unsupported delimited message_encoding`})
		_ = l.LoadProtos([]string{mapRoot})
		t.Fatal("expected error")
	})

	// Proto3 root referencing an imported Edition 2023 opaque message.
	p3Root := filepath.Join(dir, "p3_root.proto")
	assert.NoError(t, os.WriteFile(p3Root, []byte(`
syntax = "proto3";
package root;
option go_package = "example.com/root";
import "dep.proto";
message Root { dep.OpaqueMsg m = 1; }
`), 0o600))
	t.Run("proto3 root referencing imported opaque message", func(t *testing.T) {
		l := NewLoader([]string{dir}, nil)
		l.SetLogger(&expectLogger{t: t, FatalContains: `imported message "dep.OpaqueMsg" uses unsupported Go API level "API_OPAQUE"`})
		_ = l.LoadProtos([]string{p3Root})
		t.Fatal("expected error")
	})
}

func TestEdition2023FeaturesRejected(t *testing.T) {
	tests := []struct {
		name  string
		proto string
		want  string
	}{
		{
			name: "delimited singular message field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message Sub {}
message M { Sub a = 1 [features.message_encoding = DELIMITED]; }
`,
			want: `uses unsupported delimited message_encoding`,
		},
		{
			name: "delimited repeated message field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message Sub {}
message M { repeated Sub a = 1 [features.message_encoding = DELIMITED]; }
`,
			want: `uses unsupported delimited message_encoding`,
		},
		{
			name: "delimited oneof message field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message Sub {}
message M { oneof o { Sub a = 1 [features.message_encoding = DELIMITED]; } }
`,
			want: `uses unsupported delimited message_encoding`,
		},
		{
			name: "delimited inherited from file",
			proto: `edition = "2023";
option go_package = "example.com/ed";
option features.message_encoding = DELIMITED;
message Sub {}
message M { Sub a = 1; }
`,
			want: `uses unsupported delimited message_encoding`,
		},
		{
			name: "delimited inherited on nested message field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
option features = { message_encoding: DELIMITED };
message M {
  message Inner { M back = 1; }
}
`,
			want: `uses unsupported delimited message_encoding`,
		},
		{
			name: "message_encoding on scalar field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { int32 a = 1 [features.message_encoding = LENGTH_PREFIXED]; }
`,
			want: `can only be set on message fields`,
		},
		{
			name: "message_encoding on map field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message Sub {}
message M { map<string, Sub> a = 1 [features.message_encoding = LENGTH_PREFIXED]; }
`,
			want: `can only be set on message fields`,
		},
		{
			name: "message_encoding on message scope",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { option features.message_encoding = LENGTH_PREFIXED; }
`,
			want: `"features.message_encoding" cannot be set on a message`,
		},
		{
			name: "message_encoding invalid value",
			proto: `edition = "2023";
option go_package = "example.com/ed";
option features.message_encoding = GROUP;
`,
			want: `unsupported value "GROUP"`,
		},
		{
			name: "api_level missing go_features import",
			proto: `edition = "2023";
option go_package = "example.com/ed";
option features.(pb.go).api_level = API_OPEN;
message M { int32 a = 1; }
`,
			want: `requires importing "google/protobuf/go_features.proto"`,
		},
		{
			name: "api_level aggregate missing go_features import",
			proto: `edition = "2023";
option go_package = "example.com/ed";
option features = { [pb.go]: { api_level: API_OPEN } };
message M { int32 a = 1; }
`,
			want: `requires importing "google/protobuf/go_features.proto"`,
		},
		{
			name: "api_level opaque on file",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
option features.(pb.go).api_level = API_OPAQUE;
message M { int32 a = 1; }
`,
			want: `uses unsupported Go API level "API_OPAQUE"`,
		},
		{
			name: "api_level hybrid on root file",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
option features.(pb.go).api_level = API_HYBRID;
message M { int32 a = 1; }
`,
			want: `uses unsupported Go API level "API_HYBRID"`,
		},
		{
			name: "api_level opaque via leading dot",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
option features.(.pb.go).api_level = API_OPAQUE;
message M { int32 a = 1; }
`,
			want: `uses unsupported Go API level "API_OPAQUE"`,
		},
		{
			name: "api_level opaque via relative (go) in package pb",
			proto: `edition = "2023";
package pb;
option go_package = "example.com/pb";
import "google/protobuf/go_features.proto";
option features.(go).api_level = API_OPAQUE;
message M { int32 a = 1; }
`,
			want: `uses unsupported Go API level "API_OPAQUE"`,
		},
		{
			name: "api_level opaque via relative (go) in package pb.sub",
			proto: `edition = "2023";
package pb.sub;
option go_package = "example.com/pbsub";
import "google/protobuf/go_features.proto";
option features.(go).api_level = API_OPAQUE;
message M { int32 a = 1; }
`,
			want: `uses unsupported Go API level "API_OPAQUE"`,
		},
		{
			name: "api_level opaque via aggregate block",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
option features = { [pb.go]: { api_level: API_OPAQUE } };
message M { int32 a = 1; }
`,
			want: `uses unsupported Go API level "API_OPAQUE"`,
		},
		{
			name: "api_level opaque on nested message",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
option features.(pb.go).api_level = API_OPEN;
message M {
  message Inner {
    option features.(pb.go).api_level = API_OPAQUE;
    int32 a = 1;
  }
}
`,
			want: `uses unsupported Go API level "API_OPAQUE"`,
		},
		{
			name: "api_level unspecified does not clear file opaque",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
option features.(pb.go).api_level = API_OPAQUE;
message M {
  option features.(pb.go).api_level = API_LEVEL_UNSPECIFIED;
  int32 a = 1;
}
`,
			want: `uses unsupported Go API level "API_OPAQUE"`,
		},
		{
			name: "api_level on field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
message M { int32 a = 1 [features.(pb.go).api_level = API_OPEN]; }
`,
			want: `"features.(pb.go).api_level" cannot be set on a field`,
		},
		{
			name: "api_level invalid value",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
option features.(pb.go).api_level = API_INVALID;
`,
			want: `unsupported value "API_INVALID"`,
		},
		{
			name: "unknown GoFeatures subfield typo",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
option features.(pb.go).api_levle = API_OPAQUE;
`,
			want: `unknown Go feature "features.(pb.go).api_levle"`,
		},
		{
			name: "unknown GoFeatures nested extension",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
option features = { [pb.go]: { [ext]: 1 } };
`,
			want: `unknown Go feature extension "ext"`,
		},
		{
			name: "legacy_unmarshal_json_enum on message",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
message M { option features.(pb.go).legacy_unmarshal_json_enum = true; }
`,
			want: `"features.(pb.go).legacy_unmarshal_json_enum" cannot be set on a message`,
		},
		{
			name: "strip_enum_prefix not in edition 2023",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
option features.(pb.go).strip_enum_prefix = STRIP_ENUM_PREFIX_KEEP;
`,
			want: `"features.(pb.go).strip_enum_prefix" is not available in edition 2023`,
		},
		{
			name: "quoted enum value for message_encoding",
			proto: `edition = "2023";
option go_package = "example.com/ed";
option features.message_encoding = "LENGTH_PREFIXED";
`,
			want: `unsupported value "LENGTH_PREFIXED"`,
		},
		{
			name: "quoted enum value in features aggregate",
			proto: `edition = "2023";
option go_package = "example.com/ed";
option features = { field_presence: "IMPLICIT" };
`,
			want: `unsupported value "IMPLICIT"`,
		},
		{
			name: "quoted bool value for packed",
			proto: `syntax = "proto3";
option go_package = "example.com/p3";
message M { repeated int32 a = 1 [packed = "false"]; }
`,
			want: `unsupported value "false"`,
		},
		{
			name: "quoted bool value for legacy_unmarshal_json_enum",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
option features.(pb.go).legacy_unmarshal_json_enum = "true";
`,
			want: `unsupported value "true"`,
		},
		{
			name: "duplicate aggregate and dotted features",
			proto: `edition = "2023";
option go_package = "example.com/ed";
option features.utf8_validation = NONE;
option features = { enum_type: CLOSED };
`,
			want: `option "features" was already set`,
		},
		{
			name: "duplicate aggregate then overlapping dotted feature",
			proto: `edition = "2023";
option go_package = "example.com/ed";
option features = { enum_type: CLOSED };
option features.enum_type = OPEN;
`,
			want: `option "features.enum_type" was already set`,
		},
		{
			name: "duplicate aggregate pb.go then go features aggregate",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
option features = { [pb.go]: { api_level: API_OPEN } };
option features.(pb.go) = { legacy_unmarshal_json_enum: true };
`,
			want: `option "features.(pb.go)" was already set`,
		},
		{
			name: "duplicate dotted feature on field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message Sub {}
message M { Sub a = 1 [features.message_encoding = DELIMITED, features.message_encoding = LENGTH_PREFIXED]; }
`,
			want: `option "features.message_encoding" was already set`,
		},
		{
			name: "duplicate go api_level",
			proto: `edition = "2023";
option go_package = "example.com/ed";
import "google/protobuf/go_features.proto";
option features.(pb.go).api_level = API_OPAQUE;
option features.(pb.go).api_level = API_OPEN;
`,
			want: `option "features.(pb.go).api_level" was already set`,
		},
		{
			name: "duplicate key inside features block",
			proto: `edition = "2023";
option go_package = "example.com/ed";
option features = { enum_type: CLOSED, enum_type: OPEN };
`,
			want: `option "features.enum_type" was already set`,
		},
		{
			name: "packed option in edition 2023",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { repeated int32 a = 1 [packed = true]; }
`,
			want: `option "packed" is not available in editions`,
		},
		{
			name: "packed option on extend in edition 2023",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { extensions 100 to max; }
extend M { repeated int32 a = 100 [packed = false]; }
`,
			want: `option "packed" is not available in editions`,
		},
		{
			name: "repeated_field_encoding on message",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { option features.repeated_field_encoding = EXPANDED; }
`,
			want: `"features.repeated_field_encoding" cannot be set on a message`,
		},
		{
			name: "repeated_field_encoding on singular field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { int32 a = 1 [features.repeated_field_encoding = EXPANDED]; }
`,
			want: `can only be set on repeated fields`,
		},
		{
			name: "repeated_field_encoding PACKED on repeated string",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { repeated string a = 1 [features.repeated_field_encoding = PACKED]; }
`,
			want: `cannot use "PACKED" on non-packable field`,
		},
		{
			name: "repeated_field_encoding PACKED on map",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { map<string, int32> a = 1 [features.repeated_field_encoding = PACKED]; }
`,
			want: `cannot use "PACKED" on non-packable field`,
		},
		{
			name: "repeated_field_encoding on singular extend field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { extensions 100 to max; }
extend M { int32 a = 100 [features.repeated_field_encoding = EXPANDED]; }
`,
			want: `can only be set on repeated fields`,
		},
		{
			name: "repeated_field_encoding PACKED on repeated message extend field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message Sub {}
message M { extensions 100 to max; }
extend M { repeated Sub a = 100 [features.repeated_field_encoding = PACKED]; }
`,
			want: `cannot use "PACKED" on non-packable field`,
		},
		{
			name: "utf8_validation on message extend field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message Sub {}
message M { extensions 100 to max; }
extend M { Sub a = 100 [features.utf8_validation = VERIFY]; }
`,
			want: `can only be set on string fields`,
		},
		{
			name: "utf8_validation on enum extend field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
enum E { E0 = 0; }
message M { extensions 100 to max; }
extend M { E a = 100 [features.utf8_validation = NONE]; }
`,
			want: `can only be set on string fields`,
		},
		{
			name: "message_encoding on enum extend field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
enum E { E0 = 0; }
message M { extensions 100 to max; }
extend M { E a = 100 [features.message_encoding = LENGTH_PREFIXED]; }
`,
			want: `can only be set on message fields`,
		},
		{
			name: "utf8_validation on int32",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { int32 a = 1 [features.utf8_validation = NONE]; }
`,
			want: `can only be set on string fields`,
		},
		{
			name: "utf8_validation on bytes",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { bytes a = 1 [features.utf8_validation = NONE]; }
`,
			want: `can only be set on string fields`,
		},
		{
			name: "utf8_validation on non-string map",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { map<int32, int32> a = 1 [features.utf8_validation = NONE]; }
`,
			want: `can only be set on string fields`,
		},
		{
			name: "utf8_validation on message",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { option features.utf8_validation = NONE; }
`,
			want: `"features.utf8_validation" cannot be set on a message`,
		},
		{
			name: "enum_type on message",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { option features.enum_type = CLOSED; }
`,
			want: `"features.enum_type" cannot be set on a message`,
		},
		{
			name: "enum_type on field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
enum E { E0 = 0; }
message M { E a = 1 [features.enum_type = CLOSED]; }
`,
			want: `"features.enum_type" cannot be set on a field`,
		},
		{
			name: "json_format on field",
			proto: `edition = "2023";
option go_package = "example.com/ed";
message M { int32 a = 1 [features.json_format = ALLOW]; }
`,
			want: `"features.json_format" cannot be set on a field`,
		},
		{
			name: "enum_type in proto3",
			proto: `syntax = "proto3";
option go_package = "example.com/p3";
option features.enum_type = OPEN;
`,
			want: `is only available in editions`,
		},
		{
			name: "go api_level in proto3",
			proto: `syntax = "proto3";
option go_package = "example.com/p3";
option features.(pb.go).api_level = API_OPEN;
`,
			want: `is only available in editions`,
		},
		{
			name: "unknown edition 2023 feature",
			proto: `edition = "2023";
option go_package = "example.com/ed";
option features.unknown_feat = YES;
`,
			want: `unknown edition 2023 feature "features.unknown_feat"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectProtoError(t, tt.proto, tt.want)
		})
	}
}
