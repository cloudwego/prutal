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
	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/text"
)

type Enum struct {
	HeadComment   string
	InlineComment string
	Directives    Directives

	Name string // orignial name in proto

	GoName  string
	Fields  []*EnumField
	Options Options

	reserved reservedRanges

	Msg   *Message // for embedded enum only
	Proto *Proto
}

func (e *Enum) String() string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "Enum %s {\n", e.FullName())
	for _, f := range e.Fields {
		fmt.Fprintf(b, " %s = %d", f.GoName, f.Value)
		if len(f.Options) > 0 {
			fmt.Fprintf(b, " %+v\n", f.Options)
		} else {
			fmt.Fprintf(b, "\n")
		}
	}
	fmt.Fprintf(b, "}\n")
	return b.String()
}

func (e *Enum) FullName() string {
	ss := make([]string, 0, 2)
	if e.Msg != nil {
		ss = append(ss, e.Msg.FullName())
	} else if e.Proto.Package != "" {
		ss = append(ss, e.Proto.Package)
	}
	ss = append(ss, e.Name)
	return strings.Join(ss, ".")
}

func (e *Enum) IsReservedField(v int32) bool {
	return e.reserved.In(v)
}

func (e *Enum) genNoPrefix() bool {
	if e.Directives.Has(prutalNoEnumPrefix) {
		return true
	} else if e.Proto.Directives.Has(prutalNoEnumPrefix) {
		return true
	} else if v, ok := e.Options.Get(gogoproto_enum_prefix); ok {
		return isfalse(v)
	} else if v, ok := e.Proto.Options.Get(gogoproto_enum_prefix_all); ok {
		return isfalse(v)
	}
	return false
}

func (e *Enum) genMapping() bool {
	if e.Directives.Has(prutalNoEnumMapping) {
		return false
	}
	if e.Proto.Directives.Has(prutalNoEnumMapping) {
		return false
	}
	return true
}

func (e *Enum) resolve() {
	p := e.Proto
	e.GoName = strs.GoCamelCase(strings.TrimPrefix(e.Name, p.Package+"."))
	if e.Msg != nil {
		e.GoName = e.Msg.GoName + "_" + e.GoName
	}
	for _, f := range e.Fields {
		if f.genNoPrefix() {
			f.GoName = strs.GoCamelCase(f.Name)
		} else if e.Msg != nil {
			f.GoName = e.Msg.GoName + "_" + f.Name
		} else {
			f.GoName = e.GoName + "_" + f.Name
		}
	}
}

func (e *Enum) verify() error {
	errs := []error{}
	m := map[int32]bool{}
	for _, f := range e.Fields {
		if e.IsReservedField(f.Value) {
			errs = append(errs, fmt.Errorf("%q = %d is reserved", f.Name, f.Value))
		} else if m[f.Value] {
			errs = append(errs, fmt.Errorf("%q = %d is duplicated", f.Name, f.Value))
		}
		m[f.Value] = true
	}
	return errors.Join(errs...)
}

type EnumField struct {
	HeadComment   string
	InlineComment string
	Directives    Directives

	Name string // orignial name in proto

	GoName  string
	Value   int32
	Options Options

	Enum *Enum
}

func (x *EnumField) genNoPrefix() bool {
	if x.Directives.Has(prutalNoEnumPrefix) {
		return true
	}
	return x.Enum.genNoPrefix()
}

func (x *protoLoader) EnterEnumDef(c *parser.EnumDefContext) {
	// ENUM enumName enumBody
	e := &Enum{
		HeadComment: x.consumeHeadComment(c), InlineComment: x.consumeInlineComment(c),
		Name: c.EnumName().GetText()}
	e.Directives.Parse(e.HeadComment, e.InlineComment)
	switch getRuleIndex(c.GetParent()) {
	case parser.ProtobufParserRULE_topLevelDef: // top level message
		p := x.currentProto()
		p.Enums = append(p.Enums, e)
		e.Proto = p
	case parser.ProtobufParserRULE_messageElement: // embedded message
		m := x.currentMsg()
		m.Enums = append(m.Enums, e)
		e.Msg = m
		e.Proto = x.currentProto()
	default:
		panic("unknown parent rule")
	}
	x.enum = e // for options, see ExitOptionStatement
}

func (x *protoLoader) ExitEnumDef(c *parser.EnumDefContext) {
	x.enum = nil
}

func (x *protoLoader) ExitEnumField(c *parser.EnumFieldContext) {
	// ident EQ (MINUS)? intLit enumValueOptions? SEMI
	f := &EnumField{
		HeadComment: x.consumeHeadComment(c), InlineComment: x.consumeInlineComment(c),
		Name: c.Ident().GetText()}
	f.Directives.Parse(f.HeadComment, f.InlineComment)

	//  (MINUS)? intLit
	if num, ok := text.UnmarshalI32(c.IntLit().GetText()); !ok {
		t := c.IntLit()
		x.Fatalf("%s - parse enum %q err", getTokenPos(t), t.GetText())
	} else {
		f.Value = num
	}
	if t := c.MINUS(); t != nil {
		f.Value = -f.Value
	}

	// enumValueOptions
	if oo := c.EnumValueOptions(); oo != nil {
		for _, o := range oo.AllEnumValueOption() {
			v, err := unmarshalConst(o.Constant().GetText())
			if err != nil {
				x.Fatalf("%s - enum field option syntax err: %s", getTokenPos(o), err)
			}
			f.Options = append(f.Options, &Option{Name: o.OptionName().GetText(), Value: v})
		}
	}
	f.Enum = x.enum
	f.Enum.Fields = append(f.Enum.Fields, f)
}
