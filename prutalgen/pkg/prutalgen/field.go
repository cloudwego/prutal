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

	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/strs"
	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/text"
	"github.com/cloudwego/prutal/prutalgen/internal/protoparser"
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

	// extProto is set for extend fields so type resolution and feature
	// validation share the normal Field machinery even when Msg is nil
	// (top-level extend).
	extProto *Proto

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
	return !f.Options.Is(option_packed, "false") // proto3
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

// isDelimitedEncoding reports whether a non-map message field in an edition 2023
// file resolves to features.message_encoding = DELIMITED (group wire encoding).
func (f *Field) isDelimitedEncoding() bool {
	if f.Msg == nil || f.Msg.Proto == nil || !f.Msg.Proto.IsEdition2023() {
		return false
	}
	if !f.IsMessage() {
		return false
	}
	if s, ok := f.Options.Get(f_message_encoding); ok {
		return s == "DELIMITED"
	}
	return f.Msg.Proto.Options.Is(f_message_encoding, "DELIMITED")
}

// verifyFieldFeatureOptions validates direct feature options on f using its
// resolved field kind and type. Shared between normal message fields and
// extension fields.
func (f *Field) verifyFieldFeatureOptions() error {
	for _, o := range f.Options {
		switch o.Name {
		case f_field_presence:
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
		case f_repeated_field_encoding:
			if !f.Repeated && !f.IsMap() {
				return fmt.Errorf("option %q can only be set on repeated fields, not %q", o.Name, f.Name)
			}
			if o.Value == "PACKED" && (f.IsMap() || (f.Type != nil && !scalarPackedTypes[f.Type.Name] && !f.Type.IsEnum())) {
				return fmt.Errorf("option %q cannot use %q on non-packable field %q", o.Name, o.Value, f.Name)
			}
		case f_utf8_validation:
			isString := f.Type != nil && ((!f.IsMap() && f.Type.Name == "string") ||
				(f.IsMap() && (f.Key.Name == "string" || f.Type.Name == "string")))
			if !isString {
				return fmt.Errorf("option %q can only be set on string fields, not %q", o.Name, f.Name)
			}
		case f_message_encoding:
			if !f.IsMessage() {
				return fmt.Errorf("option %q can only be set on message fields, not %q", o.Name, f.Name)
			}
		}
	}
	return nil
}

// verifyFeatures checks edition feature applicability and restrictions that
// depend on the resolved field kind and inherited file features.
func (f *Field) verifyFeatures() error {
	if err := f.verifyFieldFeatureOptions(); err != nil {
		return err
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

func (x *protoLoader) loadBracketOptions(opts []*protoparser.OptionDecl, target optionTarget) Options {
	if len(opts) == 0 {
		return nil
	}
	var options Options
	for _, o := range opts {
		x.appendScopedOption(&options, o, target)
	}
	return options
}

func (x *protoLoader) checkFieldLabel(label string, pos protoparser.Pos) {
	switch label {
	case "required":
		if x.currentProto().Edition != editionProto2 {
			x.Fatalf("%s - `required` keyword only available for proto2", pos)
		}
	case "optional":
		if x.currentProto().IsEdition2023() {
			x.Fatalf("%s - `optional` keyword is not available in editions", pos)
		}
	}
}

func (x *protoLoader) newFieldFromDecl(decl *protoparser.FieldDecl) *Field {
	f := &Field{
		HeadComment:   decl.HeadComment,
		InlineComment: decl.InlineComment,
		Name:          decl.Name,
		Type:          &Type{Name: decl.Type, pos: decl.TypePos},
	}
	f.Type.f = f
	f.Directives.Parse(f.HeadComment, f.InlineComment)

	num, ok := text.UnmarshalI32(decl.Number)
	if !ok {
		x.Fatalf("%s - parse field number %q err", decl.NumberPos, decl.Number)
	}
	f.FieldNumber = num
	f.Options = x.loadBracketOptions(decl.Options, optionTargetField)
	return f
}

func (x *protoLoader) loadExtend(p *Proto, parent *Message, decl *protoparser.ExtendDecl) {
	for _, fd := range decl.Fields {
		x.checkFieldLabel(fd.Label, fd.LabelPos)
		options := x.loadBracketOptions(fd.Options, optionTargetExtensionField)
		f := &Field{
			Name:     fd.Name,
			Type:     &Type{Name: fd.Type, pos: fd.TypePos},
			Repeated: fd.Label == "repeated",
			Options:  options,
			Msg:      parent,
			extProto: p,
		}
		f.Type.f = f
		if parent != nil {
			parent.extFields = append(parent.extFields, f)
		} else {
			p.extFields = append(p.extFields, f)
		}
	}
}

func (x *protoLoader) loadField(m *Message, decl *protoparser.FieldDecl) {
	x.checkFieldLabel(decl.Label, decl.LabelPos)
	f := x.newFieldFromDecl(decl)
	switch decl.Label {
	case "repeated":
		f.Repeated = true
	case "required":
		f.Required = true
	case "optional":
		f.Optional = true
	}
	m.Fields = append(m.Fields, f)
	f.Msg = m
	if m.Proto.IsEdition2023() {
		// edition 2023 spells required as a feature
		if s, ok := f.fieldPresence(); ok && s == "LEGACY_REQUIRED" {
			f.Required = true
		}
	}
}

func (x *protoLoader) loadOneofField(m *Message, o *Oneof, decl *protoparser.FieldDecl) {
	f := x.newFieldFromDecl(decl)
	f.Oneof = o
	o.Fields = append(o.Fields, f)
	f.Msg = m
	m.Fields = append(m.Fields, f)
}

func (x *protoLoader) loadMapField(m *Message, decl *protoparser.MapFieldDecl) {
	f := &Field{
		HeadComment:   decl.HeadComment,
		InlineComment: decl.InlineComment,
		Name:          decl.Name,
		Key:           &Type{Name: decl.KeyType, pos: decl.KeyPos},
		Type:          &Type{Name: decl.ValType, pos: decl.ValPos},
		Msg:           m,
	}
	f.Type.f = f
	f.Directives.Parse(f.HeadComment, f.InlineComment)

	num, ok := text.UnmarshalI32(decl.Number)
	if !ok {
		x.Fatalf("%s - parse field number %q err", decl.NumberPos, decl.Number)
	}
	f.FieldNumber = num
	f.Options = x.loadBracketOptions(decl.Options, optionTargetField)
	m.Fields = append(m.Fields, f)
}
