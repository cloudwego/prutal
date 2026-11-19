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

type optionTarget uint8

const (
	optionTargetFile optionTarget = iota
	optionTargetMessage
	optionTargetEnum
	optionTargetEnumEntry
	optionTargetOneof
	optionTargetField
	optionTargetExtensionField
	optionTargetExtensionRange
	optionTargetService
	optionTargetMethod
)

func (t optionTarget) label() string {
	switch t {
	case optionTargetFile:
		return "a file"
	case optionTargetMessage:
		return "a message"
	case optionTargetEnum:
		return "an enum"
	case optionTargetEnumEntry:
		return "an enum entry"
	case optionTargetOneof:
		return "a oneof"
	case optionTargetField:
		return "a field"
	case optionTargetExtensionField:
		return "an extension field"
	case optionTargetExtensionRange:
		return "an extension range"
	case optionTargetService:
		return "a service"
	case optionTargetMethod:
		return "an RPC"
	default:
		return "this declaration"
	}
}

func isPBPackage(pkg string) bool {
	return pkg == "pb" || strings.HasPrefix(pkg, "pb.")
}

func canonicalOptionName(name, pkg string) string {
	switch {
	case name == "features.(.pb.go)":
		return f_go_features
	case strings.HasPrefix(name, "features.(.pb.go)."):
		return f_go_features + name[len("features.(.pb.go)"):]
	case isPBPackage(pkg) && name == "features.(go)":
		return f_go_features
	case isPBPackage(pkg) && strings.HasPrefix(name, "features.(go)."):
		return f_go_features + name[len("features.(go)"):]
	default:
		return name
	}
}

func hasGoFeaturesImport(p *Proto) bool {
	for _, vp := range visibleProtos(p) {
		if vp.DescriptorName() == "google/protobuf/go_features.proto" {
			return true
		}
	}
	return false
}

type parsedOption struct {
	opt  *Option
	kind protoparser.ValueKind
	pos  protoparser.Pos
}

func (x *protoLoader) flattenGoFeatures(val *protoparser.Value) []parsedOption {
	if val.Kind != protoparser.ValueBlock {
		v, _ := unmarshalValue(val)
		x.Fatalf("%s - option %q unsupported value %q", val.Pos, f_go_features, v)
	}
	seen := make(map[string]bool, len(val.Fields))
	out := make([]parsedOption, 0, len(val.Fields))
	for _, sf := range val.Fields {
		if sf.IsExtension {
			x.Fatalf("%s - unknown Go feature extension %q", sf.Pos, sf.Name)
		}
		key := f_go_features + "." + sf.Name
		if seen[key] {
			x.Fatalf("%s - option %q was already set", sf.Pos, key)
		}
		seen[key] = true
		fv, err := unmarshalValue(sf.Value)
		if err != nil {
			x.Fatalf("%s - feature option syntax err: %s", sf.Value.Pos, err)
		}
		out = append(out, parsedOption{
			opt:  &Option{Name: key, Value: fv},
			kind: sf.Value.Kind,
			pos:  sf.Value.Pos,
		})
	}
	return out
}

// parseOptions returns the option name = constant as parsed, and for the
// aggregate form of the editions features, "features = { a: X, b: Y }" or
// "features.(pb.go) = { ... }", the dotted entries on top, so that a feature is
// found by its canonical dotted name however it was spelled. Every value is verified.
func (x *protoLoader) parseOptions(name string, val *protoparser.Value) Options {
	v, err := unmarshalValue(val)
	if err != nil {
		x.Fatalf("%s - option syntax err: %s", val.Pos, err)
	}
	if val.Kind == protoparser.ValueBlock || val.Kind == protoparser.ValueList {
		x.validateOptionValue(val)
	}

	p := x.currentProto()
	canonicalName := canonicalOptionName(name, p.Package)

	// name may include extensions like (gogoproto.goproto_unrecognized_all)
	// we ignore parsing it, coz it's only used by protoc.
	// but we will keep it for optimization cases like `[(gogoproto.nullable) = false];`
	parsed := []parsedOption{{
		opt:  &Option{Name: canonicalName, Value: v},
		kind: val.Kind,
		pos:  val.Pos,
	}}
	switch canonicalName {
	case "features":
		if val.Kind != protoparser.ValueBlock {
			x.Fatalf("%s - option %q unsupported value %q", val.Pos, canonicalName, v)
		}
		seen := make(map[string]bool, len(val.Fields))
		seenGo := false
		for _, f := range val.Fields {
			if f.IsExtension {
				if f.Name == "pb.go" {
					if seenGo {
						x.Fatalf("%s - option %q was already set", f.Pos, f_go_features)
					}
					seenGo = true
					gv, err := unmarshalValue(f.Value)
					if err != nil {
						x.Fatalf("%s - feature option syntax err: %s", f.Value.Pos, err)
					}
					parsed = append(parsed, parsedOption{
						opt:  &Option{Name: f_go_features, Value: gv},
						kind: f.Value.Kind,
						pos:  f.Value.Pos,
					})
					parsed = append(parsed, x.flattenGoFeatures(f.Value)...)
				}
				continue
			}
			key := "features." + f.Name
			if seen[key] {
				x.Fatalf("%s - option %q was already set", f.Pos, key)
			}
			seen[key] = true
			fv, err := unmarshalValue(f.Value)
			if err != nil {
				x.Fatalf("%s - feature option syntax err: %s", f.Value.Pos, err)
			}
			parsed = append(parsed, parsedOption{
				opt:  &Option{Name: key, Value: fv},
				kind: f.Value.Kind,
				pos:  f.Value.Pos,
			})
		}
	case f_go_features:
		parsed = append(parsed, x.flattenGoFeatures(val)...)
	}

	isEd2023 := p.IsEdition2023()
	hasGoImport := false
	checkedGoImport := false
	options := make(Options, 0, len(parsed))
	for _, po := range parsed {
		o := po.opt
		if o.Name == "features" || strings.HasPrefix(o.Name, "features.") {
			if !isEd2023 {
				x.Fatalf("%s - option %q is only available in editions", po.pos, o.Name)
			}
			if o.Name == f_go_features || strings.HasPrefix(o.Name, f_go_features+".") {
				if !checkedGoImport {
					hasGoImport = hasGoFeaturesImport(p)
					checkedGoImport = true
				}
				if !hasGoImport {
					x.Fatalf("%s - option %q requires importing \"google/protobuf/go_features.proto\"", po.pos, o.Name)
				}
				if o.Name == f_go_strip_enum_prefix {
					x.Fatalf("%s - option %q is not available in edition 2023", po.pos, o.Name)
				}
				if o.Name != f_go_features && !isKnownGoEdition2023Feature(o.Name) {
					x.Fatalf("%s - unknown Go feature %q", po.pos, o.Name)
				}
			} else if strings.HasPrefix(o.Name, "features.") && !strings.HasPrefix(o.Name, "features.(") && !isKnownEdition2023Feature(o.Name) {
				x.Fatalf("%s - unknown edition 2023 feature %q", po.pos, o.Name)
			}
		}
		if o.Name == option_packed && isEd2023 {
			x.Fatalf("%s - option %q is not available in editions; use %q",
				po.pos, o.Name, f_repeated_field_encoding)
		}
		if !verifyOptionValue(o.Name, o.Value, po.kind) {
			x.Fatalf("%s - option %q unsupported value %q", po.pos, o.Name, o.Value)
		}
		options = append(options, o)
	}
	return options
}

func isKnownEdition2023Feature(name string) bool {
	switch name {
	case f_field_presence, f_enum_type, f_repeated_field_encoding,
		f_utf8_validation, f_message_encoding, f_json_format:
		return true
	default:
		return false
	}
}

func isKnownGoEdition2023Feature(name string) bool {
	switch name {
	case f_go_api_level, f_go_legacy_unmarshal_json_enum:
		return true
	default:
		return false
	}
}

func isFeatureAllowedOnTarget(name string, target optionTarget) bool {
	switch name {
	case f_field_presence:
		return target == optionTargetFile || target == optionTargetField
	case f_enum_type, f_go_legacy_unmarshal_json_enum:
		return target == optionTargetFile || target == optionTargetEnum
	case f_repeated_field_encoding, f_utf8_validation, f_message_encoding:
		return target == optionTargetFile || target == optionTargetField || target == optionTargetExtensionField
	case f_json_format:
		return target == optionTargetFile || target == optionTargetMessage || target == optionTargetEnum
	case f_go_api_level:
		return target == optionTargetFile || target == optionTargetMessage
	default:
		return true
	}
}

func hasOptionNamed(oo Options, name string) bool {
	for _, o := range oo {
		if o.Name == name {
			return true
		}
	}
	return false
}

func hasFeatureOption(oo Options) bool {
	for _, o := range oo {
		if o.Name == "features" || strings.HasPrefix(o.Name, "features.") {
			return true
		}
	}
	return false
}

func hasGoFeatureOption(oo Options) bool {
	for _, o := range oo {
		if o.Name == f_go_features || strings.HasPrefix(o.Name, f_go_features+".") {
			return true
		}
	}
	return false
}

func (x *protoLoader) checkFeatureDuplicates(pos protoparser.Pos, existing, incoming Options) {
	if len(existing) == 0 || len(incoming) == 0 {
		return
	}
	// Match protoc v29.3 option-merge semantics: an aggregate cannot follow any
	// previously set option under the same root, while a later dotted option is
	// allowed as long as its exact leaf was not already set.
	if hasFeatureOption(existing) && hasOptionNamed(incoming, "features") {
		x.Fatalf("%s - option %q was already set", pos, "features")
	}
	if hasGoFeatureOption(existing) && hasOptionNamed(incoming, f_go_features) {
		x.Fatalf("%s - option %q was already set", pos, f_go_features)
	}
	for _, o := range incoming {
		if !strings.HasPrefix(o.Name, "features.") {
			continue
		}
		if hasOptionNamed(existing, o.Name) {
			x.Fatalf("%s - option %q was already set", pos, o.Name)
		}
	}
}

func (x *protoLoader) validateOptionTargets(
	pos protoparser.Pos, options Options, target optionTarget,
) {
	for _, o := range options {
		if !isFeatureAllowedOnTarget(o.Name, target) {
			x.Fatalf("%s - option %q cannot be set on %s", pos, o.Name, target.label())
		}
		if target == optionTargetFile && o.Name == f_field_presence && o.Value == "LEGACY_REQUIRED" {
			x.Fatalf("%s - option %q cannot default to %q on a file", pos, o.Name, o.Value)
		}
	}
}

func (x *protoLoader) appendScopedOption(dst *Options, decl *protoparser.OptionDecl, target optionTarget) {
	options := x.parseOptions(decl.Name, decl.Value)
	x.validateOptionTargets(decl.Pos, options, target)
	x.checkFeatureDuplicates(decl.Pos, *dst, options)
	*dst = append(*dst, options...)
}

func (x *protoLoader) loadFileOption(p *Proto, decl *protoparser.OptionDecl) {
	x.appendScopedOption(&p.Options, decl, optionTargetFile)
}

func verifyOptionValue(name, v string, kind protoparser.ValueKind) bool {
	switch name {
	case f_repeated_field_encoding:
		return kind == protoparser.ValueIdent && (v == "EXPANDED" || v == "PACKED")

	case f_field_presence:
		return kind == protoparser.ValueIdent && (v == "EXPLICIT" || v == "IMPLICIT" || v == "LEGACY_REQUIRED")

	case f_enum_type:
		return kind == protoparser.ValueIdent && (v == "OPEN" || v == "CLOSED")

	case f_utf8_validation:
		return kind == protoparser.ValueIdent && (v == "VERIFY" || v == "NONE")

	case f_message_encoding:
		return kind == protoparser.ValueIdent && (v == "LENGTH_PREFIXED" || v == "DELIMITED")

	case f_json_format:
		return kind == protoparser.ValueIdent && (v == "ALLOW" || v == "LEGACY_BEST_EFFORT")

	case f_go_api_level:
		return kind == protoparser.ValueIdent && (v == "API_LEVEL_UNSPECIFIED" || v == "API_OPEN" || v == "API_HYBRID" || v == "API_OPAQUE")

	case f_go_legacy_unmarshal_json_enum, option_packed, option_allow_alias:
		return kind == protoparser.ValueBool && verifyTrueOrFalse(v)

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
