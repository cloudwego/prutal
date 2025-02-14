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
	"go/format"
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func sourceEqual(t *testing.T, a, b []byte) {
	t.Helper()
	if v, err := format.Source(a); err == nil {
		a = v
	}
	if v, err := format.Source(b); err == nil {
		b = v
	}
	assert.Equal(t, string(a), string(b))
}

func TestCodeWriter(t *testing.T) {
	w := NewCodeWriter("", "main")
	w.UsePkg("fmt", "")
	w.UsePkg("time", "")
	w.UsePkg("github.com/cloudwego/gopkg", "")
	w.F("func main() {}")

	sourceEqual(t, []byte(`
package main
import (
	"fmt"
	"time"

	"github.com/cloudwego/gopkg"
)
func main() {}`), w.Bytes())
}
