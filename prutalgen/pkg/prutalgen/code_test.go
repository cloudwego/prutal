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
