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
)

type Option struct {
	Name  string
	Value string
}

type Options []*Option

func (oo Options) String() string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "[")
	for i, o := range oo {
		if i != 0 {
			fmt.Fprintf(b, ", ")
		}
		fmt.Fprintf(b, "%q: %q", o.Name, o.Value)
	}
	fmt.Fprintf(b, "]")
	return b.String()
}

func (oo Options) Get(name string) (string, bool) {
	// return the last one in case we have multiple values
	for i := len(oo) - 1; i >= 0; i-- {
		o := oo[i]
		if o.Name == name {
			return o.Value, true
		}
	}
	return "", false
}

func (oo Options) Is(name string, value string) bool {
	s, ok := oo.Get(name)
	return ok && s == value
}

// parseOptions returns the option name = constant as parsed, and for the
// aggregate form of the editions features, "features = { a: X, b: Y }", the
// entries "features.a" and "features.b" on top, so that a feature is found by
// its dotted name however it was spelled. Every value is verified.
func (x *protoLoader) parseOptions(name string, c parser.IConstantContext) Options {
	v, err := unmarshalConst(c.GetText())
	if err != nil {
		x.Fatalf("%s - option syntax err: %s", getTokenPos(c), err)
	}

	// name may include extensions like (gogoproto.goproto_unrecognized_all)
	// we ignore parsing it, coz it's only used by protoc.
	// but we will keep it for optimization cases like `[(gogoproto.nullable) = false];`
	options := Options{{Name: name, Value: v}}
	if name == "features" {
		if block := c.BlockLit(); block != nil {
			idents := block.AllIdent()
			values := block.AllConstant()
			for i, ident := range idents {
				v, err := unmarshalConst(values[i].GetText())
				if err != nil {
					x.Fatalf("%s - feature option syntax err: %s", getTokenPos(values[i]), err)
				}
				options = append(options, &Option{Name: "features." + ident.GetText(), Value: v})
			}
		}
	}

	for _, o := range options {
		if !verifyOption(o.Name, o.Value) {
			x.Fatalf("%s - option %q unsupported value %q", getTokenPos(c), o.Name, o.Value)
		}
		if o.Name == f_field_presence && !x.currentProto().IsEdition2023() {
			x.Fatalf("%s - option %q is only available in editions", getTokenPos(c), o.Name)
		}
	}
	return options
}

func (x *protoLoader) rejectFieldPresenceOption(
	c antlr.ParserRuleContext, options Options, target string,
) {
	for _, o := range options {
		if o.Name == f_field_presence {
			x.Fatalf("%s - option %q cannot be set on %s", getTokenPos(c), o.Name, target)
		}
	}
}

func (x *protoLoader) ExitOptionStatement(c *parser.OptionStatementContext) {
	options := x.parseOptions(c.OptionName().GetText(), c.Constant())

	switch getRuleIndex(c.GetParent()) {
	case parser.ProtobufParserRULE_proto:
		p := x.currentProto()
		for _, o := range options {
			if o.Name == f_field_presence && o.Value == "LEGACY_REQUIRED" {
				x.Fatalf("%s - option %q cannot default to %q on a file",
					getTokenPos(c), o.Name, o.Value)
			}
		}
		p.Options = append(p.Options, options...)

	case parser.ProtobufParserRULE_messageElement:
		x.rejectFieldPresenceOption(c, options, "a message")
		m := x.currentMsg()
		m.Options = append(m.Options, options...)

	case parser.ProtobufParserRULE_oneof:
		x.rejectFieldPresenceOption(c, options, "a oneof")
		of := x.currentOneof()
		of.Options = append(of.Options, options...)

	case parser.ProtobufParserRULE_enumElement:
		x.rejectFieldPresenceOption(c, options, "an enum")
		x.enum.Options = append(x.enum.Options, options...)

	case parser.ProtobufParserRULE_serviceElement:
		x.rejectFieldPresenceOption(c, options, "a service")
		s := x.currentService()
		s.Options = append(s.Options, options...)

	case parser.ProtobufParserRULE_rpc:
		x.rejectFieldPresenceOption(c, options, "an RPC")
		s := x.currentService()
		rpc := last(s.Methods)
		rpc.Options = append(rpc.Options, options...)

	default:
		return
	}
}

func verifyOption(name, v string) bool {
	switch name {
	case f_repeated_field_encoding:
		return v == "EXPANDED" || v == "PACKED"

	case f_field_presence:
		return v == "EXPLICIT" || v == "IMPLICIT" || v == "LEGACY_REQUIRED"

	case f_enum_type:
		return v == "OPEN" || v == "CLOSED"

	case option_packed, option_allow_alias:
		return verifyTrueOrFalse(v)

	default:
		return true
	}
}

func verifyTrueOrFalse(v string) bool {
	switch v {
	case "true", "false":
		return true
	}
	return false
}
