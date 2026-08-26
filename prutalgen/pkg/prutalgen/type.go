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
	"path"
	"strings"

	"github.com/cloudwego/prutal/prutalgen/internal/antlr"
	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/strs"
)

type Type struct {
	Name string

	p             *Proto
	goPackageName string

	// *Enum or *Message
	typ any

	// the type belongs to Field or RPC
	f *Field
	m *Method

	// for logging when resolve
	rule antlr.ParserRuleContext
}

// GoName returns the GoName of underlying type
//
// Coz when calling resolve(), the underlying type may remains unresolved,
// that's why GoName is a method instead of a property.
func (t *Type) GoName() string {
	if t.typ == nil {
		return scalar2GoTypes[t.Name]
	}
	ret := ""
	switch ft := t.typ.(type) {
	case *Enum:
		ret = ft.GoName
	case *Message:
		ret = ft.GoName
	default:
		panic("[BUG] unknown type")
	}
	if t.p == nil {
		return ret
	}
	return t.GoPackageName() + "." + ret
}

// GoPackageName returns the package name used to qualify references to types
// from the package this type belongs to. It is also used as the import alias.
func (t *Type) GoPackageName() string {
	if t.goPackageName != "" {
		return t.goPackageName
	}
	return strs.GoSanitized(path.Base(t.p.GoImport))
}

func (t *Type) IsExternalType() bool {
	return t.p != nil
}

func (t *Type) GoImport() string {
	if t.p != nil {
		return t.p.GoImport
	}
	if t.f != nil {
		return t.f.Msg.Proto.GoImport
	} else if t.m != nil {
		return t.m.Service.Proto.GoImport
	}
	return ""
}

func (t *Type) String() string {
	return t.GoName()
}

func (t *Type) EncodingType() string {
	if t.typ != nil {
		switch t.typ.(type) {
		case *Message:
			return "bytes"
		case *Enum:
			return "varint"
		default:
			panic("[BUG] unknown type")
		}
	}
	ret, ok := scalar2encodingType[t.Name]
	if !ok {
		panic("[BUG] unknown type name")
	}
	return ret
}

func (t *Type) Message() *Message {
	m, _ := t.typ.(*Message)
	return m
}

func (t *Type) IsMessage() bool {
	return t.Message() != nil
}

func (t *Type) Enum() *Enum {
	e, _ := t.typ.(*Enum)
	return e
}

func (t *Type) IsEnum() bool {
	return t.Enum() != nil
}

func (t *Type) resolve(allowScalar bool) {
	var p *Proto
	var scope string
	if t.f != nil {
		p = t.f.Msg.Proto
		scope = t.f.Msg.FullName()
	} else if t.m != nil {
		p = t.m.Service.Proto
		scope = fullProtoName(p.Package, t.m.Service.Name)
	}

	t.p = nil
	t.typ = nil
	t.goPackageName = ""

	if allowScalar {
		if _, ok := scalar2GoTypes[t.Name]; ok {
			return
		}
	}

	visible := visibleProtos(p)
	ref := t.Name
	if strings.HasPrefix(ref, ".") {
		ref = strings.TrimPrefix(ref, ".")
		if typ, owner := findVisibleType(visible, ref); typ != nil {
			t.setResolvedType(p, owner, typ)
			return
		}
		t.typeNotFound(p)
		return
	}

	for {
		candidate := fullProtoName(scope, ref)
		if typ, owner := findVisibleType(visible, candidate); typ != nil {
			t.setResolvedType(p, owner, typ)
			return
		}
		// RPC types use symbol lookup before checking that the symbol is a
		// message. Field types ignore non-type symbols and keep walking outward.
		if t.m != nil && hasVisibleSymbol(visible, candidate) {
			t.typeNotFound(p)
			return
		}

		// For compound names, finding the first component as an aggregate fixes
		// the lookup scope even when the remaining components do not resolve.
		if first, _, compound := strings.Cut(ref, "."); compound {
			prefix := fullProtoName(scope, first)
			if isVisibleAggregate(visible, prefix) {
				t.typeNotFound(p)
				return
			}
		}
		if scope == "" {
			break
		}
		scope = parentProtoName(scope)
	}
	t.typeNotFound(p)
}

func (t *Type) setResolvedType(current, owner *Proto, typ any) {
	t.typ = typ
	if owner.GoImport != current.GoImport {
		t.p = owner
	}
}

func (t *Type) typeNotFound(p *Proto) {
	line := 0
	if t.rule != nil && t.rule.GetStart() != nil {
		line = t.rule.GetStart().GetLine()
	}
	p.Fatalf("line %d: type %q not found.", line, t.Name)
}

func findVisibleType(protos []*Proto, fullName string) (any, *Proto) {
	for _, p := range protos {
		if typ := p.getTypeByFullName(fullName); typ != nil {
			return typ, p
		}
	}
	return nil, nil
}

func hasVisibleSymbol(protos []*Proto, fullName string) bool {
	for _, p := range protos {
		if p.hasSymbolFullName(fullName) {
			return true
		}
	}
	return false
}

func isVisibleAggregate(protos []*Proto, fullName string) bool {
	for _, p := range protos {
		if p.getTypeByFullName(fullName) != nil || p.hasServiceFullName(fullName) {
			return true
		}
		if p.Package == fullName || strings.HasPrefix(p.Package, fullName+".") {
			return true
		}
	}
	return false
}

// visibleProtos returns the current file, its direct imports, and the public
// import closure reachable from every direct import.
func visibleProtos(p *Proto) []*Proto {
	ret := []*Proto{p}
	seen := map[*Proto]bool{p: true}
	var addPublic func(*Proto)
	addPublic = func(current *Proto) {
		if !seen[current] {
			seen[current] = true
			ret = append(ret, current)
		}
		for _, imp := range current.Imports {
			if imp.Public && !seen[imp.Proto] {
				addPublic(imp.Proto)
			}
		}
	}
	for _, imp := range p.Imports {
		addPublic(imp.Proto)
	}
	return ret
}

func fullProtoName(scope, name string) string {
	if scope == "" {
		return name
	}
	if name == "" {
		return scope
	}
	return scope + "." + name
}

func parentProtoName(name string) string {
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		return name[:i]
	}
	return ""
}
