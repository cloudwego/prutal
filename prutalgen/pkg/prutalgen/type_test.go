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
	"path/filepath"
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestType(t *testing.T) {
	p := &Type{}

	m := &Message{}
	m.GoName = "MessageType"

	e := &Enum{}
	m.GoName = "EnumType"

	// test GoName
	p.Name = "sint64"
	assert.Equal(t, "int64", p.GoName())
	assert.Equal(t, "int64", p.String()) // same

	p.typ = m
	assert.Equal(t, m.GoName, p.GoName())

	p.typ = e
	assert.Equal(t, e.GoName, p.GoName())
	p.p = &Proto{GoImport: "prutal/base", GoPackage: "base"}
	assert.Equal(t, "base."+e.GoName, p.GoName())

	// test EncodingType
	p.typ = nil
	p.Name = "sint64"
	assert.Equal(t, "zigzag64", p.EncodingType())
	p.typ = m
	assert.Equal(t, "bytes", p.EncodingType())
	p.typ = e
	assert.Equal(t, "varint", p.EncodingType())

	// Message
	p.typ = nil
	assert.False(t, p.IsMessage())
	p.typ = m
	assert.True(t, p.IsMessage())
	assert.Same(t, m, p.Message())

	// Enum
	p.typ = nil
	assert.False(t, p.IsEnum())
	p.typ = e
	assert.True(t, p.IsEnum())
	assert.Same(t, e, p.Enum())

	// resolve:scalar
	p.typ, p.f, p.m = nil, nil, nil
	p.Name = "sint64"
	p.resolve(true)
	assert.Equal(t, "int64", p.GoName())

	// resolve: field type (nested type) of a message
	// NestedType in m's parent
	localProto := &Proto{}
	parent := &Message{Name: "parent", Proto: localProto}
	nested := &Message{Name: "nested_type", GoName: "NestedType", Msg: parent, Proto: localProto}
	m = &Message{Name: "child", Msg: parent, Proto: localProto}
	parent.Messages = []*Message{nested, m}
	localProto.Messages = []*Message{parent}
	p.typ, p.f, p.m = nil, nil, nil
	p.f = &Field{Msg: m}
	p.Name = nested.Name
	p.resolve(false)
	assert.Equal(t, nested.GoName, p.GoName())

	// resolve: field type of a message
	m = &Message{Proto: &Proto{Messages: []*Message{{Name: "message_type", GoName: "MessageType"}}}}
	p.typ, p.f, p.m = nil, nil, nil
	p.f = &Field{Msg: m}
	p.Name = m.Proto.Messages[0].Name
	p.resolve(false)
	assert.Equal(t, m.Proto.Messages[0].GoName, p.GoName())

	// resolve: args / returns type of a service
	method := &Method{Service: &Service{Proto: m.Proto}}
	p.typ, p.f, p.m = nil, nil, nil
	p.m = method
	p.Name = method.Service.Proto.Messages[0].Name
	p.resolve(false)
	assert.Equal(t, method.Service.Proto.Messages[0].GoName, p.GoName())

	// resolve: not in same package
	m = &Message{Proto: &Proto{
		Imports: []*Import{{Proto: &Proto{
			Package:   "base",
			GoImport:  "gobase",
			GoPackage: "gobase",
			Messages:  []*Message{{Name: "response", GoName: "Response"}},
		}}},
	}}
	p.typ, p.f, p.m = nil, nil, nil
	p.f = &Field{Msg: m}
	p.Name = m.Proto.Imports[0].Package + "." + m.Proto.Imports[0].Messages[0].Name
	p.resolve(false)
	assert.Equal(t,
		m.Proto.Imports[0].GoImport+"."+m.Proto.Imports[0].Messages[0].GoName,
		p.GoName())

}

// Types re-exported through (nested) public imports must resolve, including
// via a plain forwarding file, and be attributed to the defining proto.
func TestTypeResolve_PublicImportForwarding(t *testing.T) {
	dir := t.TempDir()
	_ = writeFileUnderDir(t, dir, "a.proto", []byte(
		`package a;`+`option go_package = "mod/a";`+"\nmessage T {}",
	))
	_ = writeFileUnderDir(t, dir, "c.proto", []byte(
		`import public "a.proto";`+`package c;`+`option go_package = "mod/c";`,
	))
	_ = writeFileUnderDir(t, dir, "b.proto", []byte(
		`import "c.proto";`+`package b;`+`option go_package = "mod/b";`+"\n"+
			`message B { a.T t = 1; }`,
	))

	x := NewLoader([]string{dir}, nil)
	x.SetLogger(&testLogger{t})
	ff := x.LoadProto(filepath.Join(dir, "b.proto"))
	b := ff[0]
	assert.NoError(t, b.verify())

	ft := b.Messages[0].Fields[0].Type
	assert.True(t, ft.IsExternalType())
	assert.Equal(t, "mod/a", ft.GoImport())
	assert.Equal(t, "a.T", ft.GoName())
}

// A bare name is not visible when none of the lexical parent scopes match the
// imported package.
func TestTypeResolve_PublicImportForwarding_BareNameNotVisible(t *testing.T) {
	dir := t.TempDir()
	_ = writeFileUnderDir(t, dir, "a.proto", []byte(
		`package a;`+`option go_package = "mod/a";`+"\nmessage T {}",
	))
	_ = writeFileUnderDir(t, dir, "c.proto", []byte(
		`import public "a.proto";`+`package c;`+`option go_package = "mod/c";`,
	))
	_ = writeFileUnderDir(t, dir, "b.proto", []byte(
		`import "c.proto";`+`package b;`+`option go_package = "mod/b";`+"\n"+
			`message B { T t = 1; }`,
	))

	x := NewLoader([]string{dir}, nil)
	x.SetLogger(&expectLogger{t: t, FatalContains: `type "T" not found`})
	x.LoadProto(filepath.Join(dir, "b.proto"))
	t.Fatal("never goes here. logger Fatalf in LoadProto")
}

func TestTypeResolveLexicalScopesThroughPublicImports(t *testing.T) {
	dir := t.TempDir()
	_ = writeFileUnderDir(t, dir, "a.proto", []byte(
		`package foo.baz; option go_package = "mod/target"; message T {}`,
	))
	_ = writeFileUnderDir(t, dir, "c.proto", []byte(
		`package forward.c; option go_package = "mod/c"; import public "a.proto";`,
	))
	_ = writeFileUnderDir(t, dir, "d.proto", []byte(
		`package forward.d; option go_package = "mod/d"; import public "c.proto";`,
	))
	_ = writeFileUnderDir(t, dir, "b.proto", []byte(
		`package foo.bar; option go_package = "mod/b"; import "d.proto";`+
			`message B { optional baz.T value = 1; }`,
	))

	x := NewLoader([]string{dir}, nil)
	x.SetLogger(&testLogger{t})
	files := x.LoadProto("b.proto")
	typ := files[0].Messages[0].Fields[0].Type
	assert.Equal(t, "mod/target", typ.GoImport())
	assert.Equal(t, "target.T", typ.GoName())
}

func TestTypeResolveFromAncestorPackage(t *testing.T) {
	dir := t.TempDir()
	_ = writeFileUnderDir(t, dir, "a.proto", []byte(
		`package foo; option go_package = "mod/ancestor"; message T {}`,
	))
	_ = writeFileUnderDir(t, dir, "b.proto", []byte(
		`package foo.bar; option go_package = "mod/b"; import "a.proto";`+
			`message B { optional T value = 1; }`,
	))

	x := NewLoader([]string{dir}, nil)
	x.SetLogger(&testLogger{t})
	files := x.LoadProto("b.proto")
	typ := files[0].Messages[0].Fields[0].Type
	assert.Equal(t, "mod/ancestor", typ.GoImport())
	assert.Equal(t, "ancestor.T", typ.GoName())
}

func TestTypeResolveWithoutProtoPackage(t *testing.T) {
	dir := t.TempDir()
	_ = writeFileUnderDir(t, dir, "a.proto", []byte(
		`option go_package = "mod/target"; message T {} message Outer { message N {} }`,
	))
	_ = writeFileUnderDir(t, dir, "b.proto", []byte(
		`option go_package = "mod/b"; import "a.proto";`+
			`message B { optional .T first = 1; optional Outer.N second = 2; }`,
	))

	x := NewLoader([]string{dir}, nil)
	x.SetLogger(&testLogger{t})
	files := x.LoadProto("b.proto")
	fields := files[0].Messages[0].Fields
	assert.Equal(t, "target.T", fields[0].Type.GoName())
	assert.Equal(t, "target.Outer_N", fields[1].Type.GoName())
}

func TestTypeResolveHonorsInnermostAggregate(t *testing.T) {
	dir := t.TempDir()
	_ = writeFileUnderDir(t, dir, "a.proto", []byte(
		`package foo; option go_package = "mod/a"; message T {}`,
	))
	_ = writeFileUnderDir(t, dir, "b.proto", []byte(
		`package other; option go_package = "mod/b"; import "a.proto";`+
			`message foo {} message B { optional foo.T value = 1; }`,
	))

	x := NewLoader([]string{dir}, nil)
	x.SetLogger(&expectLogger{t: t, FatalContains: `type "foo.T" not found`})
	x.LoadProto("b.proto")
	t.Fatal("never goes here. logger Fatalf in LoadProto")
}

func TestTypeResolveHonorsInnermostPackage(t *testing.T) {
	dir := t.TempDir()
	_ = writeFileUnderDir(t, dir, "inner.proto", []byte(
		`package other.foo; option go_package = "mod/inner"; message X {}`,
	))
	_ = writeFileUnderDir(t, dir, "outer.proto", []byte(
		`package foo; option go_package = "mod/outer"; message T {}`,
	))
	_ = writeFileUnderDir(t, dir, "b.proto", []byte(
		`package other; option go_package = "mod/b";`+
			`import "inner.proto"; import "outer.proto";`+
			`message B { optional foo.T value = 1; }`,
	))

	x := NewLoader([]string{dir}, nil)
	x.SetLogger(&expectLogger{t: t, FatalContains: `type "foo.T" not found`})
	x.LoadProto("b.proto")
	t.Fatal("never goes here. logger Fatalf in LoadProto")
}

func TestTypeResolveBareNameIgnoresInnermostService(t *testing.T) {
	dir := t.TempDir()
	_ = writeFileUnderDir(t, dir, "a.proto", []byte(
		`package foo; option go_package = "mod/a"; message T {}`,
	))
	_ = writeFileUnderDir(t, dir, "b.proto", []byte(
		`package foo.bar; option go_package = "mod/b"; import "a.proto";`+
			`service T {} message B { optional T value = 1; }`,
	))

	x := NewLoader([]string{dir}, nil)
	x.SetLogger(&testLogger{t})
	files := x.LoadProto("b.proto")
	typ := files[0].Messages[0].Fields[0].Type
	assert.Equal(t, "mod/a", typ.GoImport())
	assert.Equal(t, "a.T", typ.GoName())
}

func TestTypeResolveBareNameIgnoresInnermostPackage(t *testing.T) {
	dir := t.TempDir()
	_ = writeFileUnderDir(t, dir, "a.proto", []byte(
		`package foo; option go_package = "mod/a"; message T {}`,
	))
	_ = writeFileUnderDir(t, dir, "inner.proto", []byte(
		`package foo.bar.T; option go_package = "mod/inner"; message X {}`,
	))
	_ = writeFileUnderDir(t, dir, "b.proto", []byte(
		`package foo.bar; option go_package = "mod/b";`+
			`import "a.proto"; import "inner.proto";`+
			`message B { optional T value = 1; }`,
	))

	x := NewLoader([]string{dir}, nil)
	x.SetLogger(&testLogger{t})
	files := x.LoadProto("b.proto")
	typ := files[0].Messages[0].Fields[0].Type
	assert.Equal(t, "mod/a", typ.GoImport())
	assert.Equal(t, "a.T", typ.GoName())
}

func TestTypeResolveRejectsRegularTransitiveImport(t *testing.T) {
	dir := t.TempDir()
	_ = writeFileUnderDir(t, dir, "a.proto", []byte(
		`package a; option go_package = "mod/a"; message T {}`,
	))
	_ = writeFileUnderDir(t, dir, "c.proto", []byte(
		`package c; option go_package = "mod/c"; import "a.proto";`,
	))
	_ = writeFileUnderDir(t, dir, "b.proto", []byte(
		`package b; option go_package = "mod/b"; import "c.proto";`+
			`message B { optional .a.T value = 1; }`,
	))

	x := NewLoader([]string{dir}, nil)
	x.SetLogger(&expectLogger{t: t, FatalContains: `type ".a.T" not found`})
	x.LoadProto("b.proto")
	t.Fatal("never goes here. logger Fatalf in LoadProto")
}

func TestTypeResolveBareNameIgnoresNonTypeSymbols(t *testing.T) {
	for _, tc := range []struct {
		name        string
		declaration string
	}{
		{
			name:        "field",
			declaration: `message B { string T = 1; T value = 2; }`,
		},
		{
			name:        "oneof",
			declaration: `message B { oneof T { string x = 1; } T value = 2; }`,
		},
		{
			name:        "enum value",
			declaration: `message B { enum E { T = 0; } T value = 1; }`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := loadTestProto(t, `
syntax = "proto3";
package p;
option go_package = "example.com/p";
message T {}
`+tc.declaration)
			b := p.Messages[1]
			assert.Same(t, p.Messages[0], b.Fields[len(b.Fields)-1].Type.Message())
		})
	}
}

func TestTypeResolveRPCExactNonMessageShadowsParentType(t *testing.T) {
	for _, tc := range []struct {
		name          string
		declaration   string
		extraProto    string
		fatalContains string
	}{
		{
			name:          "method",
			declaration:   `service S { rpc T(T) returns (T); }`,
			fatalContains: `type "T" not found`,
		},
		{
			name:          "enum value",
			declaration:   `enum E { T = 0; } service S { rpc M(T) returns (T); }`,
			fatalContains: `type "T" not found`,
		},
		{
			name:          "enum",
			declaration:   `enum T { ZERO = 0; } service S { rpc M(T) returns (T); }`,
			fatalContains: `not a message type`,
		},
		{
			name:          "service",
			declaration:   `service T {} service S { rpc M(T) returns (T); }`,
			fatalContains: `type "T" not found`,
		},
		{
			name:          "package",
			declaration:   `import "inner.proto"; service S { rpc M(T) returns (T); }`,
			extraProto:    `package p.q.T; option go_package = "example.com/inner"; message X {}`,
			fatalContains: `type "T" not found`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			_ = writeFileUnderDir(t, dir, "a.proto", []byte(
				`package p; option go_package = "example.com/a"; message T {}`,
			))
			if tc.extraProto != "" {
				_ = writeFileUnderDir(t, dir, "inner.proto", []byte(tc.extraProto))
			}
			_ = writeFileUnderDir(t, dir, "b.proto", []byte(
				`syntax = "proto3"; package p.q; option go_package = "example.com/b";`+
					`import "a.proto";`+tc.declaration,
			))

			x := NewLoader([]string{dir}, nil)
			x.SetLogger(&expectLogger{t: t, FatalContains: tc.fatalContains})
			x.LoadProto("b.proto")
			t.Fatal("never goes here. logger Fatalf in LoadProto")
		})
	}
}

func TestTypeResolveCompoundNameIgnoresNonAggregatePrefix(t *testing.T) {
	p := loadTestProto(t, `
syntax = "proto3";
package p;
option go_package = "example.com/p";
message foo { message T {} }
message B {
  string foo = 1;
  foo.T value = 2;
}`)
	assert.Same(t, p.Messages[0].Messages[0], p.Messages[1].Fields[1].Type.Message())
}
