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
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/prutal/prutalgen/internal/parser"
	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/strs"
)

type Message struct {
	HeadComment   string
	InlineComment string
	Directives    Directives

	Name string // original name in IDL

	GoName string // name used by generated code

	// embedded declarations
	Enums    []*Enum
	Messages []*Message

	// Fields and Options
	Fields  []*Field
	Oneofs  []*Oneof
	Options Options

	reserved reservedRanges

	Msg   *Message // for embedded message only
	Proto *Proto
}

func (m *Message) String() string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "Message %s {\n", m.FullName())
	fmt.Fprintf(b, "Options:%v\n", m.Options)
	for _, x := range m.Enums {
		fmt.Fprintf(b, "-> %s\n", x.String())
	}
	for _, x := range m.Messages {
		fmt.Fprintf(b, "-> %s\n", x.String())
	}
	for _, x := range m.Fields {
		fmt.Fprintf(b, "-> %s\n", x.String())
	}
	fmt.Fprintf(b, "}\n")
	return b.String()
}

func (m *Message) FullName() string {
	ss := make([]string, 0, 2)
	if m.Msg != nil {
		ss = append(ss, m.Msg.FullName())
	} else if m.Proto.Package != "" {
		ss = append(ss, m.Proto.Package)
	}
	ss = append(ss, m.Name)
	return strings.Join(ss, ".")
}

func (m *Message) IsReservedField(v int32) bool {
	return m.reserved.In(v)
}

func (m *Message) genUnknownFields() bool {
	if v, ok := m.Directives.IsSet(prutalUnknownFields); ok {
		return v
	}
	if v, ok := m.Proto.Directives.IsSet(prutalUnknownFields); ok {
		return v
	}
	if v, ok := m.Options.Get(gogoproto_goproto_unrecognized); ok {
		return istrue(v)
	} else if v, ok := m.Proto.Options.Get(gogoproto_goproto_unrecognized_all); ok {
		return istrue(v)
	}
	return false
}

func (m *Message) getType(name string) any {
	if m.Name == name {
		return m
	}
	if name, ok := trimPathPrefix(name, m.Name); ok {
		for _, x := range m.Enums {
			if x.Name == name {
				return x
			}
		}
		for _, x := range m.Messages {
			if v := x.getType(name); v != nil {
				return v
			}
		}
	}
	return nil
}

func (m *Message) resolve() {
	p := m.Proto
	name := strings.TrimPrefix(m.FullName(), p.Package+".")
	m.GoName = strs.GoCamelCase(name)

	// resolve declarations before fields,
	// coz fields may use these declarations
	for _, x := range m.Enums {
		x.resolve()
	}

	for _, x := range m.Messages {
		x.resolve()
	}

	for _, x := range m.Fields {
		x.resolve()
	}

	for _, o := range m.Oneofs {
		o.GoName = strs.GoCamelCase(o.Name)
	}

	m.resolveFieldConflicts()
	m.resolveOneofTypeConflicts()
}

// resolveOneofTypeConflicts ensures the wrapper struct name of each oneof
// field does not collide with a sibling nested message or enum GoName by
// appending '_' to the wrapper name (not the field's GoName).
//
// Mirrors the second pass in protoc-gen-go's protogen.newMessage.
func (m *Message) resolveOneofTypeConflicts() {
	nested := make(map[string]bool, len(m.Messages)+len(m.Enums))
	for _, x := range m.Messages {
		nested[x.GoName] = true
	}
	for _, x := range m.Enums {
		nested[x.GoName] = true
	}
	for _, f := range m.Fields {
		if f.Oneof == nil {
			continue
		}
		for nested[f.OneofStructName()] {
			f.oneofTypeSuffix += "_"
		}
	}
}

// resolveFieldConflicts appends '_' to field and oneof GoNames that collide
// with well-known method names attached to a generated message type, or with
// the 'Get*' getter of another field.
//
// Mirrors the algorithm in protoc-gen-go's protogen.newMessage. Any change to
// the set of names below is a potential incompatible API change, because it
// may change generated field names.
func (m *Message) resolveFieldConflicts() {
	usedNames := map[string]bool{
		"Reset":               true,
		"String":              true,
		"ProtoMessage":        true,
		"Marshal":             true,
		"Unmarshal":           true,
		"ExtensionRangeArray": true,
		"ExtensionMap":        true,
		"Descriptor":          true,
	}
	for _, f := range m.Fields {
		for usedNames[f.GoName] || usedNames["Get"+f.GoName] {
			f.GoName += "_"
		}
		usedNames[f.GoName] = true
		usedNames["Get"+f.GoName] = true
		if f.Oneof != nil && f.Oneof.Fields[0] == f {
			o := f.Oneof
			for usedNames[o.GoName] {
				o.GoName += "_"
			}
			usedNames[o.GoName] = true
		}
	}
}

func (m *Message) verify() error {
	var errs []error
	for _, x := range m.Enums {
		if err := x.verify(); err != nil {
			errs = append(errs, fmt.Errorf("enum %s verify err: %w", x.FullName(), err))
		}
	}
	for _, x := range m.Messages {
		if err := x.verify(); err != nil {
			errs = append(errs, fmt.Errorf("message %s verify err: %w", x.FullName(), err))
		}
	}
	exists := map[int32]bool{}
	for _, x := range m.Fields {
		if m.IsReservedField(x.FieldNumber) {
			errs = append(errs, fmt.Errorf("field %q field number = %d is reserved", x.Name, x.FieldNumber))
		} else if exists[x.FieldNumber] {
			errs = append(errs, fmt.Errorf("field %q field number = %d is duplicated", x.Name, x.FieldNumber))
		}
		exists[x.FieldNumber] = true
	}
	return errors.Join(errs...)
}

func (x *protoLoader) EnterMessageDef(c *parser.MessageDefContext) {
	m := &Message{}
	m.HeadComment = x.consumeHeadComment(c)
	m.InlineComment = x.consumeInlineComment(c)
	m.Directives.Parse(m.HeadComment, m.InlineComment)

	m.Name = c.MessageName().GetText()
	switch getRuleIndex(c.GetParent()) {
	case parser.ProtobufParserRULE_topLevelDef: // top level message
		p := x.currentProto()
		p.Messages = append(p.Messages, m)
		m.Proto = p

	case parser.ProtobufParserRULE_messageElement: // embedded message
		m0 := x.currentMsg()
		m0.Messages = append(m0.Messages, m)
		m.Msg = m0
		m.Proto = x.currentProto()
	default:
		panic("unknown parent rule")
	}
	push(&x.msgstack, m)
}

func (x *protoLoader) ExitMessageDef(c *parser.MessageDefContext) {
	pop(&x.msgstack)
}
