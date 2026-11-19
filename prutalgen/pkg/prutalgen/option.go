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

	"github.com/cloudwego/prutal/prutalgen/internal/protoparser"
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

func (x *protoLoader) validateOptionValue(val *protoparser.Value) {
	switch val.Kind {
	case protoparser.ValueString:
		if _, err := unmarshalStringParts(val.StringParts); err != nil {
			x.Fatalf("%s - option syntax err: %s", val.Pos, err)
		}
	case protoparser.ValueBlock:
		for _, f := range val.Fields {
			x.validateOptionValue(f.Value)
		}
	case protoparser.ValueList:
		for _, elem := range val.List {
			x.validateOptionValue(elem)
		}
	}
}

// parseOptions returns the option name = constant as parsed, and for the
// aggregate form of the editions features, "features = { a: X, b: Y }", the
// entries "features.a" and "features.b" on top, so that a feature is found by
// its dotted name however it was spelled. Every value is verified.
func (x *protoLoader) parseOptions(name string, val *protoparser.Value) Options {
	v, err := unmarshalValue(val)
	if err != nil {
		x.Fatalf("%s - option syntax err: %s", val.Pos, err)
	}

	// name may include extensions like (gogoproto.goproto_unrecognized_all)
	// we ignore parsing it, coz it's only used by protoc.
	// but we will keep it for optimization cases like `[(gogoproto.nullable) = false];`
	options := Options{{Name: name, Value: v}}
	if name == "features" && val.Kind == protoparser.ValueBlock {
		for _, f := range val.Fields {
			if f.IsExtension {
				continue
			}
			fv, err := unmarshalValue(f.Value)
			if err != nil {
				x.Fatalf("%s - feature option syntax err: %s", f.Value.Pos, err)
			}
			options = append(options, &Option{Name: "features." + f.Name, Value: fv})
		}
	}
	if val.Kind == protoparser.ValueBlock || val.Kind == protoparser.ValueList {
		x.validateOptionValue(val)
	}

	for _, o := range options {
		if !verifyOption(o.Name, o.Value) {
			x.Fatalf("%s - option %q unsupported value %q", val.Pos, o.Name, o.Value)
		}
		if o.Name == f_field_presence && !x.currentProto().IsEdition2023() {
			x.Fatalf("%s - option %q is only available in editions", val.Pos, o.Name)
		}
	}
	return options
}

func (x *protoLoader) rejectFieldPresenceOption(
	pos protoparser.Pos, options Options, target string,
) {
	for _, o := range options {
		if o.Name == f_field_presence {
			x.Fatalf("%s - option %q cannot be set on %s", pos, o.Name, target)
		}
	}
}

func (x *protoLoader) loadFileOption(p *Proto, decl *protoparser.OptionDecl) {
	options := x.parseOptions(decl.Name, decl.Value)
	for _, o := range options {
		if o.Name == f_field_presence && o.Value == "LEGACY_REQUIRED" {
			x.Fatalf("%s - option %q cannot default to %q on a file",
				decl.Pos, o.Name, o.Value)
		}
	}
	p.Options = append(p.Options, options...)
}

func (x *protoLoader) loadScopedOption(decl *protoparser.OptionDecl, target string) Options {
	options := x.parseOptions(decl.Name, decl.Value)
	x.rejectFieldPresenceOption(decl.Pos, options, target)
	return options
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
