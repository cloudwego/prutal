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
	p.GoImport = "prutal/base"
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
	m = &Message{Msg: &Message{Messages: []*Message{{Name: "nested_type", GoName: "NestedType"}}}}
	p.typ, p.f, p.m = nil, nil, nil
	p.f = &Field{Msg: m}
	p.Name = m.Msg.Messages[0].Name
	p.resolve(false)
	assert.Equal(t, m.Msg.Messages[0].GoName, p.GoName())

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
			Package:  "base",
			GoImport: "gobase",
			Messages: []*Message{{Name: "response", GoName: "Response"}},
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
