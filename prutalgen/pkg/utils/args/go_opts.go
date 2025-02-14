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
	"path/filepath"
	"strings"

	"github.com/cloudwego/prutal/prutalgen/pkg/prutalgen"
)

// GoOpts represents args of go_opt for protobuf go compatibility purposes
type GoOpts struct {
	StringArgs
}

// GenPathType ... prutalgen.GenByImport or prutalgen.GenBySourceRelative
func (o *GoOpts) GenPathType() prutalgen.GenPathType {
	ret := prutalgen.GenByImport // default value
	for _, s := range o.StringArgs {
		if s == "paths=import" {
			ret = prutalgen.GenByImport
		} else if s == "paths=source_relative" {
			ret = prutalgen.GenBySourceRelative
		}
	}
	return ret
}

// Proto2pkg ... for the M opt
// see: https://protobuf.dev/reference/go/go-generated/#package
func (o *GoOpts) Proto2pkg() map[string]string {
	ret := make(map[string]string)
	for _, s := range o.StringArgs {
		if s == "" || s[0] != 'M' {
			continue
		}
		s = s[1:]
		a, b, ok := strings.Cut(s, "=")
		if ok {
			ret[filepath.Clean(a)] = b
		}
	}
	return ret
}
