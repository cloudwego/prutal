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
)

type Type struct {
	Name     string
	GoImport string

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
	if len(t.GoImport) == 0 {
		return ret
	}
	return path.Base(t.GoImport) + "." + ret
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
	var m *Message // for checking embedded types
	if t.f != nil {
		m = t.f.Msg
		p = m.Proto
	} else if t.m != nil {
		p = t.m.Service.Proto
	}

	t.GoImport = ""
	t.typ = nil

	if allowScalar {
		if _, ok := scalar2GoTypes[t.Name]; ok {
			return
		}
	}

	if m != nil {
		if v := m.getType(t.Name); v != nil {
			t.typ = v
			return
		}
	}

	//  (DOT)? (ident DOT)* ident

	// search Message
	if m != nil {
		for x := m; x != nil; x = x.Msg {
			t.typ = x.getType(t.Name)
			if t.typ != nil {
				return
			}
			for _, e := range x.Enums {
				if e.Name == t.Name {
					t.typ = e
					return
				}
			}
			for _, m := range x.Messages {
				if v := m.getType(t.Name); v != nil {
					t.typ = v
					return
				}
			}
		}
	}

	// search proto and imports
	protos := make([]*Proto, 0, len(p.Imports)+1)
	protos = append(protos, p) // always check p
	for _, x := range p.Imports {
		if p.Package == x.Package { // same package
			protos = append(protos, x.Proto)
		}
	}

	// if name starts with "." it means the type is in local package
	if !strings.HasPrefix(t.Name, ".") {
		for _, x := range p.Imports {
			if hasPathPrefix(t.Name, x.Package) {
				protos = append(protos, x.Proto)
			}
		}
	} else {
		// trim "." for searching type
		t.Name = strings.TrimLeft(t.Name, ".")
	}
	for _, x := range protos {
		t.typ = x.getType(t.Name)
		if t.typ != nil {
			if x.GoImport != p.GoImport {
				t.GoImport = x.GoImport
			}
			return
		}
	}
	p.Fatalf("line %d: type %q not found.", t.rule.GetStart().GetLine(), t.Name)
}
