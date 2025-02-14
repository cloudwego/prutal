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

func (x *protoLoader) ExitOptionStatement(c *parser.OptionStatementContext) {
	v, err := unmarshalConst(c.Constant().GetText())
	if err != nil {
		x.Fatalf("%s - option syntax err: %s", getTokenPos(c), err)
	}

	// name may include extensions like (gogoproto.goproto_unrecognized_all)
	// we ignore parsing it, coz it's only used by protoc.
	// but we will keep it for optimization cases like `[(gogoproto.nullable) = false];`
	o := &Option{Name: c.OptionName().GetText(), Value: v}

	if !verifyOption(o.Name, o.Value) {
		x.Fatalf("%s - option %q unsupported value %q", getTokenPos(c), o.Name, o.Value)
	}

	switch getRuleIndex(c.GetParent()) {
	case parser.ProtobufParserRULE_proto:
		p := x.currentProto()
		p.Options = append(p.Options, o)

	case parser.ProtobufParserRULE_messageElement:
		m := x.currentMsg()
		m.Options = append(m.Options, o)

	case parser.ProtobufParserRULE_oneof:
		of := x.currentOneof()
		of.Options = append(of.Options, o)

	case parser.ProtobufParserRULE_enumElement:
		x.enum.Options = append(x.enum.Options, o)

	case parser.ProtobufParserRULE_serviceElement:
		s := x.currentService()
		s.Options = append(s.Options, o)

	case parser.ProtobufParserRULE_rpc:
		s := x.currentService()
		rpc := last(s.Methods)
		rpc.Options = append(rpc.Options, o)

	default:
		return
	}
}

func verifyOption(name, v string) bool {
	switch name {
	case f_repeated_field_encoding:
		return v == "EXPANDED" || v == "PACKED"

	case f_field_presence:
		return v == "EXPLICIT" || v == "IMPLICIT"

	case option_packed:
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
