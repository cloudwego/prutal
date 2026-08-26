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
	"fmt"
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestLoader_Proto(t *testing.T) {
	x := NewLoader([]string{"."}, nil)
	x.SetLogger(&testLogger{t})
	empty := writeFile(t, "empty.proto", []byte(`option go_package = "example.com/empty";`))
	fn := writeFile(t, "test.proto", []byte(fmt.Sprintf(`
syntax = "proto3";
package prutal_test;
import public "%s";
option go_package = "hello/prutal_test;prutal";`, empty)))
	ff := x.LoadProto(fn)
	assert.Equal(t, 2, len(ff))
	f := ff[0]
	assert.Equal(t, fn, f.ProtoFile)
	assert.Equal(t, "prutal", f.GoPackage)    // from `go_package`
	assert.Equal(t, "prutal_test", f.Package) // from `package`
	assert.True(t, f.IsProto3())
	assert.True(t, f.Imports[0].Public)
	assert.Same(t, f.Imports[0].Proto, ff[1])
	assert.Equal(t, empty, ff[1].ProtoFile)
	t.Log(f.String())
}

func TestProto(t *testing.T) {
	p0 := &Proto{Package: "p0"}
	p := &Proto{Package: "p"}
	p.Imports = []*Import{{
		Proto: p0,
	}}

	p.Messages = []*Message{{Name: "m"}}
	p.Enums = []*Enum{{Name: "e"}}
	p.Imports = []*Import{
		{Public: false, Proto: &Proto{Messages: []*Message{{Name: "m0"}}}},
		{Public: true, Proto: &Proto{Messages: []*Message{{Name: "m0"}}}},
	}

	assert.Same(t, p.Messages[0], p.getType("m"))
	assert.Same(t, p.Enums[0], p.getType("e"))
	assert.Same(t, p.Imports[1].Messages[0], p.getType("m0"))
	assert.True(t, p.getType("x") == nil)

}

func TestLoaderRejectsDuplicateImport(t *testing.T) {
	dir := t.TempDir()
	_ = writeFileUnderDir(t, dir, "a.proto", []byte(
		`option go_package = "mod/a"; message A {}`,
	))
	_ = writeFileUnderDir(t, dir, "b.proto", []byte(
		`option go_package = "mod/b"; import "a.proto"; import public "./a.proto";`,
	))

	x := NewLoader([]string{dir}, nil)
	x.SetLogger(&expectLogger{t: t, FatalContains: `was listed twice`})
	x.LoadProto("b.proto")
	t.Fatal("never goes here. logger Fatalf in LoadProto")
}

func TestLoaderRejectsDuplicateSymbolsAcrossRoots(t *testing.T) {
	dir := t.TempDir()
	payload := []byte(`package dup; option go_package = "example.com/dup"; message T {}`)
	_ = writeFileUnderDir(t, dir, "a.proto", payload)
	_ = writeFileUnderDir(t, dir, "b.proto", payload)

	x := NewLoader([]string{dir}, nil)
	x.SetLogger(&expectLogger{t: t, FatalContains: `"dup.T"`})
	x.LoadProtos([]string{"a.proto", "b.proto"})
	t.Fatal("never goes here. logger Fatalf in LoadProtos")
}

func TestValidateProtoSymbols(t *testing.T) {
	message := func(p *Proto, name string) *Message {
		m := &Message{Name: name, Proto: p}
		p.Messages = append(p.Messages, m)
		return m
	}

	a := &Proto{ProtoFile: "a.proto", Package: "clash"}
	b := &Proto{ProtoFile: "b.proto", Package: "clash"}
	message(a, "T")
	message(b, "T")
	assert.ErrorContains(t, validateProtoSymbols([]*Proto{a, b}), "already defined")

	a = &Proto{ProtoFile: "a.proto", Package: "foo"}
	b = &Proto{ProtoFile: "b.proto", Package: "foo.bar"}
	message(a, "bar")
	assert.ErrorContains(t, validateProtoSymbols([]*Proto{a, b}), "conflicts with a package name")

	a = &Proto{ProtoFile: "a.proto", Package: "shared"}
	b = &Proto{ProtoFile: "b.proto", Package: "shared"}
	message(a, "A")
	message(b, "B")
	assert.NoError(t, validateProtoSymbols([]*Proto{a, b}))

	a = &Proto{ProtoFile: "a.proto", Package: "shared"}
	b = &Proto{ProtoFile: "b.proto", Package: "shared"}
	a.Enums = []*Enum{{Name: "A", Proto: a, Fields: []*EnumField{{Name: "VALUE"}}}}
	b.Enums = []*Enum{{Name: "B", Proto: b, Fields: []*EnumField{{Name: "VALUE"}}}}
	assert.ErrorContains(t, validateProtoSymbols([]*Proto{a, b}), `enum value "shared.VALUE"`)

	a = &Proto{ProtoFile: "a.proto", Package: "shared"}
	m := message(a, "M")
	m.Fields = []*Field{{Name: "VALUE"}}
	m.Enums = []*Enum{{
		Name: "E", Proto: a, Msg: m, Fields: []*EnumField{{Name: "VALUE"}},
	}}
	assert.ErrorContains(t, validateProtoSymbols([]*Proto{a}), `enum value "shared.M.VALUE"`)
}
