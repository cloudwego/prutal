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

import "strings"

// StringArgs impelements https://pkg.go.dev/flag#Value for []string
type StringArgs []string

// String ...
func (a *StringArgs) String() string {
	return strings.Join(*a, ",")
}

// Set ...
func (a *StringArgs) Set(v string) error {
	*a = append(*a, v)
	return nil
}
