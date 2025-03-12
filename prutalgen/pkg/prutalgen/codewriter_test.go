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
)

func sourceEqual(t *testing.T, a, b []byte) {
	t.Helper()
	if v, err := format.Source(a); err == nil {
		a = v
	}
	if v, err := format.Source(b); err == nil {
		b = v
	}
	s0, s1 := string(a), string(b)
	if s0 != s1 {
		t.Fatalf("source not equal"+
			"\n===============\n"+
			"%s"+
			"\n===============\n"+
			"%s"+
			"\n===============\n", a, b)
	}
}

func TestCodeWriter(t *testing.T) {
	w := NewCodeWriter("// header", "main")
	sourceEqual(t, []byte("// header\n\npackage main"), w.Bytes())
	w.UsePkg("fmt", "")
	w.UsePkg("time", "")
	w.UsePkg("github.com/cloudwego/gopkg", "gopkg")
	w.Write([]byte("// hello main\n"))
	w.F("func main() {}")

	sourceEqual(t, []byte(`// header

package main

import (
	"fmt"
	"time"

	"github.com/cloudwego/gopkg"
)

// hello main
func main() {}
`), w.Bytes())

	w.Reset("", "main")
	w.UsePkg("fmt", "")
	sourceEqual(t, []byte(`package main`+"\n"+`import "fmt"`), w.Bytes())
	w.UsePkg("time", "")
	w.UsePkg("net/http", "")
	sourceEqual(t, []byte(`package main

import (
	"fmt"
	"net/http"
	"time"
)`), w.Bytes())

	w.SetGroupingFunc(func(pkg string) int {
		if pkg == "fmt" {
			return 5
		}
		if pkg == "time" {
			return 3
		}
		return 0
	})

	sourceEqual(t, []byte(`package main

import (
  "net/http"

  "time"

  "fmt"
)`), w.Bytes())

}
