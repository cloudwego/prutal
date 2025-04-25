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

import "strings"

const ( // prutalgen directives
	prutalDirectivePrefix = "//prutalgen:"
	prutalNoPointer       = "no_pointer"
	prutalNoEnumPrefix    = "no_enum_prefix"
	prutalNoEnumMapping   = "no_enum_mapping"
	prutalUnknownFields   = "unknown_fields"
)

// Directives represents a list of Prutal directives.
//
// like: "//prutalgen:no_pointer,no_enum_prefix"
// Directives should be ["no_pointer", "no_enum_prefix"]
type Directives []string

// Has reports whether the list contains the given directive.
func (dd Directives) Has(d string) bool {
	for _, s := range dd {
		if s == d {
			return true
		}
	}
	return false
}

// IsSet reports whether the given directive is set.
//
// It returns:
//
// - (true, true) if the directive is set
// - (false, true) if the directive is not set
// - (false, false) if the directive is not found
//
// For Example:
//
// - IsSet("unknown_fields") returns (true, true) for ["unknown_fields"]
// - IsSet("unknown_fields") returns (false, true) for ["no_unknown_fields"]
func (dd Directives) IsSet(d string) (v bool, ok bool) {
	nod := "no_" + d
	for _, s := range dd {
		if s == d {
			return true, true
		}
		if s == nod {
			return false, true
		}
	}
	return false, false
}

func (dd *Directives) Reset() {
	*dd = (*dd)[:0]
}

func (dd *Directives) Parse(ss ...string) {
	dd.Reset()
	for _, s := range ss {
		dd.parse(s)
	}
}

func (dd *Directives) parse(input string) {
	for len(input) > 0 {
		var s string
		s, input, _ = strings.Cut(input, "\n")
		s = strings.TrimSpace(s)
		if !strings.HasPrefix(s, prutalDirectivePrefix) {
			continue
		}
		s = s[len(prutalDirectivePrefix):]
		for len(s) > 0 { // supports //prutalgen:feature1,feature2
			var d string
			d, s, _ = strings.Cut(s, ",")
			if d = strings.TrimSpace(d); len(d) > 0 {
				*dd = append(*dd, d)
			}
		}
	}
}
