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
	"strings"

	"github.com/cloudwego/prutal/prutalgen/internal/antlr"

	"github.com/cloudwego/prutal/prutalgen/internal/parser"
	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/strs"
	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/text"
)

type Field struct {
	HeadComment   string
	InlineComment string
	Directives    Directives

	// names in original IDL
	Name string
	Key  *Type // for map only
	Type *Type

	GoName string

	FieldNumber int32
	Repeated    bool
	Required    bool
	Optional    bool

	Options Options

	Oneof *Oneof // if not nil, it's from oneof
	Msg   *Message
	rule  antlr.ParserRuleContext

	// oneofTypeSuffix is appended to the oneof wrapper struct name to
	// resolve collisions with sibling nested message/enum GoNames. Only
	// set for fields inside an oneof, during Message.resolve.
	oneofTypeSuffix string
}

// String returns a string representation of the Field, including its number, Go name, and type(s).
func (f *Field) String() string {
	b := &strings.Builder{}
	if f.IsMap() {
		fmt.Fprintf(b, "%d - %s <%s, %s>", f.FieldNumber, f.GoName, f.Key, f.Type)
	} else {
		fmt.Fprintf(b, "%d - %s %s", f.FieldNumber, f.GoName, f.Type)
	}
	return b.String()
}

// IsMap reports whether the field is a map type.
func (f *Field) IsMap() bool {
	return f.Key != nil
}

// IsEnum reports whether the field is an enum type
func (f *Field) IsEnum() bool {
	return !f.IsMap() && f.Type.IsEnum()
}

// IsMessage reports whether the field is a message type
func (f *Field) IsMessage() bool {
	return !f.IsMap() && f.Type.IsMessage()
}

// IsPackedEncoding reports whether the field uses packed encoding
func (f *Field) IsPackedEncoding() bool {
	if f.IsMap() {
		return false
	}
	if !scalarPackedTypes[f.Type.Name] && !f.Type.IsEnum() {
		return false
	}

	p := f.Msg.Proto
	if p.IsProto2() {
		return f.Options.Is(option_packed, "true")
	}

	if p.IsEdition2023() {
		v := "PACKED"
		if s, ok := f.Options.Get(f_repeated_field_encoding); ok {
			v = s
		} else if s, ok = p.Options.Get(f_repeated_field_encoding); ok {
			v = s
		}
		return v == "PACKED"
	}
	return true // proto3
}

func (f *Field) resolve() {
	f.GoName = strs.GoCamelCase(f.Name)
	f.Type.resolve(true)
	// f.IsMap()
	// no need to call resolve, map is alaways scalar type
}

func (f *Field) genNoPointer() bool {
	if f.Directives.Has(prutalNoPointer) {
		return true
	}
	if f.Msg != nil && f.Msg.Directives.Has(prutalNoPointer) {
		return true
	}
	if v, ok := f.Options.Get(gogoproto_nullable); ok {
		return isfalse(v)
	}
	return false
}

func goTypeName(t *Type, noptr bool) string {
	if !noptr && t.IsMessage() {
		// message type is pointer by default
		return "*" + t.GoName()
	}
	return t.GoName()
}

// GoTypeName returns the Go type name for the field, considering pointer and repeated status.
func (f *Field) GoTypeName() string {
	noptr := f.genNoPointer()
	if f.IsMap() {
		kt := f.Key.GoName()
		vt := goTypeName(f.Type, noptr)
		return fmt.Sprintf("map[%s]%s", kt, vt)
	}
	if f.Repeated {
		return "[]" + goTypeName(f.Type, noptr)
	}
	if f.IsPointer() {
		return "*" + f.Type.GoName()
	}
	return f.Type.GoName()
}

// IsPointer reports whether the field is pointer type,
// and it means the zero value of the field is nil.
//
// It returns true for map, slice, optional fields and oneof types.
func (f *Field) IsPointer() bool {
	if f.genNoPointer() {
		return false
	}
	if f.IsMap() || f.Repeated || f.Type.GoName() == "[]byte" {
		// map or slice can be nil, so it's always NOT pointer
		// the "[]byte" case is same as f.Repeated
		return false
	}

	if f.IsMessage() {
		// message type is pointer by default
		return true
	}
	if f.Optional {
		// if optional it's pointer
		return true
	}
	if f.Oneof != nil {
		// no need pointer for oneof fields
		// we already have an interface for it.
		return false
	}

	p := f.Msg.Proto
	if p.IsProto2() {
		return true // proto2 is pointer by default
	}
	if p.IsEdition2023() {
		if s, ok := f.fieldPresence(); ok {
			return s != "IMPLICIT" // EXPLICIT or LEGACY_REQUIRED
		}
		return true // Default: EXPLICIT, which is same as proto2
	}
	return false // proto3?
}

// fieldPresence returns features.field_presence as set on the field, else on
// the file; protoc allows the feature nowhere else.
func (f *Field) fieldPresence() (string, bool) {
	if s, ok := f.Options.Get(f_field_presence); ok {
		return s, true
	}
	return f.Msg.Proto.Options.Get(f_field_presence)
}

// verifyFieldPresence checks direct option applicability and restrictions that
// depend on the resolved field kind and inherited file features.
func (f *Field) verifyFieldPresence() error {
	for _, o := range f.Options {
		if o.Name != f_field_presence {
			continue
		}
		switch {
		case f.IsMap():
			return fmt.Errorf("option %q cannot be set on map field %q", o.Name, f.Name)
		case f.Repeated:
			return fmt.Errorf("option %q cannot be set on repeated field %q", o.Name, f.Name)
		case f.Oneof != nil:
			return fmt.Errorf("option %q cannot be set on oneof field %q", o.Name, f.Name)
		case f.Type != nil && f.IsMessage() && o.Value == "IMPLICIT":
			return fmt.Errorf("option %q cannot use %q on message field %q", o.Name, o.Value, f.Name)
		}
	}
	if f.IsMap() && f.Msg != nil && f.Msg.Proto != nil && f.Msg.Proto.IsEdition2023() &&
		f.Msg.Proto.Options.Is(f_field_presence, "IMPLICIT") &&
		f.Type != nil && f.Type.IsEnum() && !f.Type.Enum().isOpen() {
		return fmt.Errorf("map field %q with implicit file presence must use an open enum value", f.Name)
	}
	if !f.isImplicitPresence() {
		return nil
	}
	if f.IsEnum() && !f.Type.Enum().isOpen() {
		return fmt.Errorf("implicit presence enum field %q must use an open enum", f.Name)
	}
	if _, ok := f.Options.Get(option_default); ok {
		return fmt.Errorf("implicit presence field %q cannot specify a default", f.Name)
	}
	return nil
}

// isImplicitPresence reports whether a singular scalar or enum field of an
// edition 2023 file has features.field_presence = IMPLICIT, set on the field
// or on the file. Such a field is a plain Go value like a proto3 field, but
// its struct tag has no "proto3" marker to say so.
func (f *Field) isImplicitPresence() bool {
	if f.Msg == nil || f.Msg.Proto == nil || !f.Msg.Proto.IsEdition2023() {
		return false
	}
	if f.Type == nil || f.IsMap() || f.Repeated || f.Optional || f.Oneof != nil || f.IsMessage() {
		return false
	}
	s, ok := f.fieldPresence()
	return ok && s == "IMPLICIT"
}

// GoZero returns the Go zero value for the field's type.
func (f *Field) GoZero() string {
	if f.IsMap() || f.Repeated {
		return "nil"
	}
	if f.IsEnum() { // search for the const var
		tp := f.Type
		e := tp.Enum()
		for _, f := range e.Fields {
			if f.Value != 0 {
				continue
			}
			zero := f.GoName
			if !tp.IsExternalType() {
				return zero
			}
			return tp.GoPackageName() + "." + zero
		}
		return "0"
	}
	ft := f.Type.GoName()
	switch ft { // scalar types
	case "float32", "float64",
		"int32", "int64",
		"uint32", "uint64":
		return "0"
	case "bool":
		return "false"
	case "string":
		return `""`
	case "[]byte":
		return "nil"
	}
	if f.IsMessage() && f.genNoPointer() {
		return ft + "{}" // zero struct
	}
	return "nil"
}

// OneofStructName returns the struct name of each field in oneof
//
// For each field in oneof, it will be defined as a struct.
// like for
//
//	message Example {
//	 oneof contact_info {
//	   string email = 3;
//	   string phone = 4;
//	 }
//
// for email, there will be a struct:
//
//	type Example_Email struct {
//			Email string
//	}
//
// for message Example, the field will be:
//
//	 type Example_Email struct {
//			ContactInfo isExample_ContactInfo
//		}
func (f *Field) OneofStructName() string {
	if f.Oneof == nil {
		panic("not oneof")
	}
	return f.Msg.GoName + "_" + f.GoName + f.oneofTypeSuffix
}

type fieldContext interface {
	antlr.ParserRuleContext

	FieldType() parser.IFieldTypeContext
	FieldName() parser.IFieldNameContext
	FieldNumber() parser.IFieldNumberContext
	FieldOptions() parser.IFieldOptionsContext
}

type fieldWithKeyTypeContext interface {
	fieldContext

	KeyType() parser.IKeyTypeContext
}

type noKeyTypeContext struct{ fieldContext }

func (noKeyTypeContext) KeyType() parser.IKeyTypeContext { return nil }

func (x *protoLoader) parseFieldOptions(c parser.IFieldOptionsContext) Options {
	if c == nil {
		return nil
	}
	var options Options
	for _, o := range c.AllFieldOption() {
		options = append(options, x.parseOptions(o.OptionName().GetText(), o.Constant())...)
	}
	return options
}

func (x *protoLoader) newField(c fieldWithKeyTypeContext) *Field {
	ft := c.FieldType()
	f := &Field{
		HeadComment: x.consumeHeadComment(c), InlineComment: x.consumeInlineComment(c),
		Name: c.FieldName().GetText(),
		Type: &Type{Name: ft.GetText(), rule: ft},
		rule: c,
	}
	f.Type.f = f
	f.Directives.Parse(f.HeadComment, f.InlineComment)

	if kt := c.KeyType(); kt != nil {
		f.Key = &Type{Name: kt.GetText(), rule: kt}
	}
	fieldn := c.FieldNumber()
	num, ok := text.UnmarshalI32(fieldn.GetText())
	if !ok {
		x.Fatalf("%s - parse field number %q err", getTokenPos(fieldn), fieldn.GetText())
	}
	f.FieldNumber = num
	f.Options = x.parseFieldOptions(c.FieldOptions())
	return f
}

func (x *protoLoader) ExitField(c *parser.FieldContext) {
	label := c.FieldLabel()
	if label != nil {
		switch label.GetText() {
		case "required":
			if x.currentProto().Edition != editionProto2 {
				x.Fatalf("%s - `required` keyword only available for proto2", getTokenPos(label))
			}
		case "optional":
			if x.currentProto().IsEdition2023() {
				x.Fatalf("%s - `optional` keyword is not available in editions", getTokenPos(label))
			}
		}
	}
	switch getRuleIndex(c.GetParent()) {
	case parser.ProtobufParserRULE_extendDef: // only for protoc
		options := x.parseFieldOptions(c.FieldOptions())
		x.rejectFieldPresenceOption(c, options, "an extension field")
		return
	}
	// fieldLabel? type_ fieldName EQ fieldNumber (LB fieldOptions RB)? SEMI
	f := x.newField(noKeyTypeContext{c})
	if label != nil {
		switch label.GetText() {
		case "repeated":
			f.Repeated = true

		case "required":
			f.Required = true

		case "optional":
			f.Optional = true
		}
	}
	m := x.currentMsg()
	m.Fields = append(m.Fields, f)
	f.Msg = m
	if m.Proto.IsEdition2023() {
		// edition 2023 spells required as a feature
		if s, ok := f.fieldPresence(); ok && s == "LEGACY_REQUIRED" {
			f.Required = true
		}
	}
}

func (x *protoLoader) ExitOneofField(c *parser.OneofFieldContext) {
	// type_ fieldName EQ fieldNumber (LB fieldOptions RB)? SEMI
	f := x.newField(noKeyTypeContext{c})
	m := x.currentMsg()

	// oneof fields are normal fields with oneof define.
	f.Oneof = last(m.Oneofs)
	f.Oneof.Fields = append(f.Oneof.Fields, f)
	f.Msg = m
	m.Fields = append(m.Fields, f)
}

func (x *protoLoader) ExitMapField(c *parser.MapFieldContext) {
	// MAP LT keyType COMMA type_ GT fieldName EQ fieldNumber (LB fieldOptions RB)? SEMI
	f := x.newField(c)
	m := x.currentMsg()
	m.Fields = append(m.Fields, f)
	f.Msg = m
}
