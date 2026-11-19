/*
 * Copyright 2026 CloudWeGo Authors
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

package protoparser

import "strings"

// File represents a parsed .proto source file.
type File struct {
	Syntax      *SyntaxDecl
	HeadComment string

	Imports  []*ImportDecl
	Packages []*PackageDecl
	Options  []*OptionDecl
	Decls    []FileDecl

	Messages []*MessageDecl
	Enums    []*EnumDecl
	Extends  []*ExtendDecl
	Services []*ServiceDecl
}

// FileDecl is implemented by top-level declarations in a File.
type FileDecl interface {
	isFileDecl()
}

// MessageElement is implemented by elements allowed inside a MessageDecl body.
type MessageElement interface {
	isMessageElement()
}

// SyntaxDecl represents a syntax = "..." or edition = "..." header.
type SyntaxDecl struct {
	IsEdition bool
	Value     StringLit
	Pos       Pos
}

// ImportKind classifies an import statement.
type ImportKind uint8

const (
	ImportRegular ImportKind = iota
	ImportPublic
	ImportWeak
)

// ImportDecl represents an import statement.
type ImportDecl struct {
	Kind ImportKind
	Path StringLit
	Pos  Pos
}

// PackageDecl represents a package statement.
type PackageDecl struct {
	Name string
	Pos  Pos
}

// StringLit represents one or more adjacent quoted string literals.
type StringLit struct {
	Parts []string
	Pos   Pos
}

// Raw returns the concatenated raw source literals.
func (s StringLit) Raw() string {
	if len(s.Parts) == 1 {
		return s.Parts[0]
	}
	return strings.Join(s.Parts, "")
}

// OptionDecl represents an option statement or a field/range/enum-value option.
type OptionDecl struct {
	Name  string
	Value *Value
	Pos   Pos
}

func (*OptionDecl) isMessageElement() {}

// ValueKind classifies an option constant value.
type ValueKind uint8

const (
	ValueIdent ValueKind = iota
	ValueInt
	ValueFloat
	ValueString
	ValueBool
	ValueBlock
	ValueList
)

// Value represents an option value (scalar, identifier, aggregate block, or list).
type Value struct {
	Kind        ValueKind
	Raw         string
	StringParts []string
	Fields      []*BlockField
	List        []*Value
	Pos         Pos
}

// BlockField represents a key-value entry inside an aggregate option block.
type BlockField struct {
	Name        string
	IsExtension bool
	Value       *Value
	Pos         Pos
}

// MessageDecl represents a message definition.
type MessageDecl struct {
	Name          string
	HeadComment   string
	InlineComment string
	Elements      []MessageElement
	Pos           Pos
}

func (*MessageDecl) isFileDecl()       {}
func (*MessageDecl) isMessageElement() {}

// FieldDecl represents a message field, oneof field, or extend field.
type FieldDecl struct {
	Label         string
	LabelPos      Pos
	Type          string
	TypePos       Pos
	Name          string
	Number        string
	NumberPos     Pos
	Options       []*OptionDecl
	HeadComment   string
	InlineComment string
	Pos           Pos
}

func (*FieldDecl) isMessageElement() {}

// MapFieldDecl represents a map<K, V> field.
type MapFieldDecl struct {
	KeyType       string
	KeyPos        Pos
	ValType       string
	ValPos        Pos
	Name          string
	Number        string
	NumberPos     Pos
	Options       []*OptionDecl
	HeadComment   string
	InlineComment string
	Pos           Pos
}

func (*MapFieldDecl) isMessageElement() {}

// OneofDecl represents a oneof definition inside a message.
type OneofDecl struct {
	Name    string
	Options []*OptionDecl
	Fields  []*FieldDecl
	Pos     Pos
}

func (*OneofDecl) isMessageElement() {}

// EnumDecl represents an enum definition.
type EnumDecl struct {
	Name          string
	HeadComment   string
	InlineComment string
	Options       []*OptionDecl
	Values        []*EnumValueDecl
	Reserved      []*ReservedDecl
	Pos           Pos
}

func (*EnumDecl) isFileDecl()       {}
func (*EnumDecl) isMessageElement() {}

// EnumValueDecl represents a single enum constant declaration.
type EnumValueDecl struct {
	Name          string
	Number        string
	NumberPos     Pos
	Options       []*OptionDecl
	HeadComment   string
	InlineComment string
	Pos           Pos
}

// ReservedDecl represents a reserved statement in a message or enum.
type ReservedDecl struct {
	Ranges []*RangeDecl
	Names  []ReservedName
	Pos    Pos
}

func (*ReservedDecl) isMessageElement() {}

// ReservedName represents either a quoted or identifier reserved field/enum name.
type ReservedName struct {
	Ident     string
	StringLit *StringLit
	Pos       Pos
}

// RangeDecl represents a single number or "start to end|max" range.
type RangeDecl struct {
	Start    string
	StartPos Pos
	End      string
	EndPos   Pos
	IsRange  bool
	IsMax    bool
	Pos      Pos
}

// ExtensionsDecl represents an extensions range declaration in a message.
type ExtensionsDecl struct {
	Ranges  []*RangeDecl
	Options []*OptionDecl
	Pos     Pos
}

func (*ExtensionsDecl) isMessageElement() {}

// ExtendDecl represents an extend block.
type ExtendDecl struct {
	Extendee    string
	ExtendeePos Pos
	Fields      []*FieldDecl
	Pos         Pos
}

func (*ExtendDecl) isFileDecl()       {}
func (*ExtendDecl) isMessageElement() {}

// ServiceDecl represents a service definition.
type ServiceDecl struct {
	Name          string
	HeadComment   string
	InlineComment string
	Options       []*OptionDecl
	Methods       []*MethodDecl
	Pos           Pos
}

func (*ServiceDecl) isFileDecl() {}

// MethodDecl represents an rpc method declaration in a service.
type MethodDecl struct {
	Name          string
	RequestStream bool
	RequestType   string
	RequestPos    Pos
	ReturnStream  bool
	ReturnType    string
	ReturnPos     Pos
	Options       []*OptionDecl
	HeadComment   string
	InlineComment string
	Pos           Pos
}
