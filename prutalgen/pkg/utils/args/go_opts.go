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

package args

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/prutal/prutalgen/pkg/prutalgen"
)

// GoOpts represents args of go_opt for protobuf go compatibility purposes
type GoOpts struct {
	StringArgs

	annotateCode    bool
	defaultAPILevel string
	apiLevels       map[string]string
}

// Parse parses paths, module, and M options using protoc-gen-go's parameter rules.
func (o *GoOpts) Parse() (prutalgen.GenPathType, map[string]string, string, error) {
	return o.parse(true)
}

func (o *GoOpts) parse(strict bool) (prutalgen.GenPathType, map[string]string, string, error) {
	if strict {
		o.annotateCode = false
		o.defaultAPILevel = ""
		o.apiLevels = nil
	}
	pathType := prutalgen.GenByImport
	module := ""
	importPaths := make(map[string]string)
	packageNames := make(map[string]string)
	apiLevels := make(map[string]string)
	annotateCode := false
	defaultAPILevel := ""
	plugins := ""
	for _, arg := range o.StringArgs {
		for _, opt := range strings.Split(arg, ",") {
			name, value, _ := strings.Cut(opt, "=")
			switch name {
			case "":
				continue
			case "annotate_code":
				switch value {
				case "true", "":
					annotateCode = true
				case "false":
				default:
					if strict {
						return pathType, nil, module, fmt.Errorf(
							`bad value for parameter %q: want "true" or "false"`, name)
					}
				}
			case "default_api_level":
				if !validAPILevel(value) {
					if strict {
						return pathType, nil, module, fmt.Errorf("unknown API level %q", value)
					}
					continue
				}
				defaultAPILevel = value
			case "experimental_strip_nonfunctional_codegen":
				if _, err := strconv.ParseBool(value); err != nil && strict {
					return pathType, nil, module, fmt.Errorf("invalid value %q for %s", value, name)
				}
			case "module":
				module = value
			case "paths":
				switch value {
				case "import":
					pathType = prutalgen.GenByImport
				case "source_relative":
					pathType = prutalgen.GenBySourceRelative
				default:
					if strict {
						return pathType, nil, module, fmt.Errorf(
							`unknown path type %q: want "import" or "source_relative"`, value)
					}
				}
			case "plugins":
				plugins = value
			default:
				if strings.HasPrefix(name, "apilevelM") {
					if !validAPILevel(value) {
						if strict {
							return pathType, nil, module, fmt.Errorf("unknown API level %q", value)
						}
						continue
					}
					apiLevels[strings.TrimPrefix(name, "apilevelM")] = value
					continue
				}
				if name[0] != 'M' {
					if strict {
						return pathType, nil, module, fmt.Errorf("no such flag -%s", name)
					}
					continue
				}
				imp, pkg, _ := strings.Cut(value, ";")
				filename := name[1:]
				if imp != "" {
					importPaths[filename] = imp
				}
				if pkg != "" {
					packageNames[filename] = pkg
				}
			}
		}
	}
	if strict {
		if plugins != "" {
			return pathType, nil, module, fmt.Errorf("plugins are not supported")
		}
	}
	if strict && module != "" && pathType == prutalgen.GenBySourceRelative {
		return pathType, nil, module, fmt.Errorf("cannot use module= with paths=source_relative")
	}

	proto2pkg := make(map[string]string, len(importPaths)+len(packageNames))
	for filename, imp := range importPaths {
		proto2pkg[filename] = imp
	}
	for filename, pkg := range packageNames {
		proto2pkg[filename] += ";" + pkg
	}
	if strict {
		o.annotateCode = annotateCode
		o.defaultAPILevel = defaultAPILevel
		o.apiLevels = apiLevels
	}
	return pathType, proto2pkg, module, nil
}

func validAPILevel(value string) bool {
	return value == "API_OPEN" || value == "API_HYBRID" || value == "API_OPAQUE"
}

// ValidateFiles rejects supported protoc-gen-go options whose requested
// generation mode prutalgen cannot produce for the loaded files.
func (o *GoOpts) ValidateFiles(protos []*prutalgen.Proto) error {
	if len(protos) > 0 {
		if o.annotateCode {
			return fmt.Errorf("annotate_code=true is unsupported")
		}
	}

	for _, p := range protos {
		level, fileOverride := o.apiLevels[p.DescriptorName()]
		if !fileOverride {
			level = o.defaultAPILevel
		}
		if level == "" || level == "API_OPEN" {
			continue
		}
		if fileOverride {
			return fmt.Errorf("apilevelM%s=%s is unsupported", p.DescriptorName(), level)
		}
		return fmt.Errorf("default_api_level=%s is unsupported", level)
	}
	return nil
}

// GenPathType ... prutalgen.GenByImport or prutalgen.GenBySourceRelative
func (o *GoOpts) GenPathType() prutalgen.GenPathType {
	pathType, _, _, _ := o.parse(false)
	return pathType
}

// Proto2pkg ... for the M opt
// see: https://protobuf.dev/reference/go/go-generated/#package
func (o *GoOpts) Proto2pkg() map[string]string {
	_, proto2pkg, _, _ := o.parse(false)
	return proto2pkg
}
