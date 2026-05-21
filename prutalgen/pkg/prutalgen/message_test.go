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
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestMessage(t *testing.T) {
	m := &Message{Name: "m"}
	m.Proto = &Proto{Package: "test.message"}
	p := m.Proto

	// getType
	m.Enums = []*Enum{{Name: "e", Proto: p}}
	assert.Same(t, m.Enums[0], m.getType("m.e"))

	m.Messages = []*Message{{Name: "m", Proto: p}}
	assert.Same(t, m.Messages[0], m.getType("m.m"))

	// genUnknownFields
	assert.False(t, m.genUnknownFields())

	// case: Message Directives
	m.Directives = Directives{prutalUnknownFields}
	assert.True(t, m.genUnknownFields())
	m.Directives = nil

	// case: Message gogoproto
	m.Options = Options{{Name: gogoproto_goproto_unrecognized, Value: "true"}}
	assert.True(t, m.genUnknownFields())
	m.Options = nil

	// case: Proto Directives
	p.Directives = Directives{prutalUnknownFields}
	assert.True(t, m.genUnknownFields())
	p.Directives = nil

	// case: Proto gogoproto
	p.Options = Options{{Name: gogoproto_goproto_unrecognized_all, Value: "true"}}
	assert.True(t, m.genUnknownFields())

	// String
	m.Fields = []*Field{{Name: "f"}}
	t.Log(m.String())
}

func TestMessage_Verify(t *testing.T) {
	p := &Proto{Package: "test.message.verify"}
	m := &Message{Name: "m"}
	m.Proto = p
	p.Messages = []*Message{m}

	// reserved
	m.reserved = append(m.reserved, reservedRange{from: 1, to: 1})
	m.Fields = append(m.Fields, &Field{Name: "testfield", FieldNumber: 1})
	assert.ErrorContains(t, p.verify(), "field number = 1 is reserved")
	m.reserved = nil
	assert.NoError(t, p.verify())

	// duplicated
	m.Fields = append(m.Fields, &Field{Name: "testfield2", FieldNumber: 1})
	assert.ErrorContains(t, p.verify(), "field number = 1 is duplicated")
	m.Fields = nil
	assert.NoError(t, p.verify())

	// nested msg case
	mm := &Message{
		Name: "mm",
		Fields: []*Field{
			{Name: "testfield1", FieldNumber: 1},
			{Name: "testfield2", FieldNumber: 1},
		},
		Msg:   m,
		Proto: m.Proto,
	}
	m.Messages = []*Message{mm}
	assert.ErrorContains(t, p.verify(), "field number = 1 is duplicated")
	m.Messages = nil
	assert.NoError(t, p.verify())

	// nested enum case
	m.Enums = []*Enum{{
		Name:  "e",
		Proto: m.Proto,
		Msg:   m,
		Fields: []*EnumField{
			{Name: "ev1", Value: 2},
			{Name: "ev2", Value: 2}, // duplicated
		},
	}}
	assert.ErrorContains(t, p.verify(), "2 is duplicated")
	m.Enums = nil
	assert.NoError(t, p.verify())
}

func TestLoader_Message(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "testmessage";
message M {
	message m {
	}

	enum e {
	 v = 0;
	}

	string f = 1;
}
`)

	m := p.Messages[0]
	assert.Equal(t, "M", m.Name)
	assert.Equal(t, 1, len(m.Messages))
	assert.Equal(t, "m", m.Messages[0].Name)
	assert.Equal(t, "MM", m.Messages[0].GoName) // GoCamelCase("M.m") = "MM"
	assert.Equal(t, 1, len(m.Enums))
	assert.Equal(t, "e", m.Enums[0].Name)
	assert.Equal(t, "ME", m.Enums[0].GoName) // GoCamelCase("M.e") = "ME"
	assert.Equal(t, 1, len(m.Fields))
	assert.Equal(t, "f", m.Fields[0].Name)
	assert.Equal(t, "F", m.Fields[0].GoName)
}

func TestLoader_MessageNestedNaming(t *testing.T) {
	p := loadTestProto(t, `
package test;
option go_package = "testmessage";

message MyOuter {
	message MyInner {}
}

message my_outer {
	message my_inner {}
}
`)

	// CamelCase names: dot before uppercase → underscore separator
	m := p.Messages[0]
	assert.Equal(t, "MyOuter", m.GoName)
	assert.Equal(t, "MyOuter_MyInner", m.Messages[0].GoName)

	// snake_case names: dot before lowercase → seamless join (matches protoc-gen-go)
	m = p.Messages[1]
	assert.Equal(t, "MyOuter", m.GoName) // GoCamelCase("my_outer") = "MyOuter"
	assert.Equal(t, "MyOuterMyInner", m.Messages[0].GoName)
}

func TestLoader_MessageFieldConflict(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "testconflict";
message M {
	string reset = 1;
	string string = 2;
	string descriptor = 3;
	string marshal = 4;
	string ok = 5;
}
`)

	m := p.Messages[0]
	// Fields conflicting with well-known method names get "_" suffix
	assert.Equal(t, "Reset_", m.Fields[0].GoName)
	assert.Equal(t, "String_", m.Fields[1].GoName)
	assert.Equal(t, "Descriptor_", m.Fields[2].GoName)
	assert.Equal(t, "Marshal_", m.Fields[3].GoName)
	// Non-conflicting field stays unchanged
	assert.Equal(t, "Ok", m.Fields[4].GoName)
}

func TestLoader_OneofWrapperTypeConflict(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "testoneofconflict";
message M {
	message Foo {}
	enum Bar { BAR0 = 0; }
	oneof o {
		string foo = 1;
		string bar = 2;
		string baz = 3;
	}
}
`)

	m := p.Messages[0]
	f0, f1, f2 := m.Fields[0], m.Fields[1], m.Fields[2]
	// Field GoNames are unaffected by wrapper-type collisions.
	assert.Equal(t, "Foo", f0.GoName)
	assert.Equal(t, "Bar", f1.GoName)
	assert.Equal(t, "Baz", f2.GoName)
	// Wrapper struct names get "_" suffix when colliding with a sibling
	// nested message ("M_Foo") or enum ("M_Bar").
	assert.Equal(t, "M_Foo_", f0.OneofStructName())
	assert.Equal(t, "M_Bar_", f1.OneofStructName())
	// Non-colliding wrapper stays unchanged.
	assert.Equal(t, "M_Baz", f2.OneofStructName())
}
