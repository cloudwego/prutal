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
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
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

func typeCheckSource(t *testing.T, src []byte) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "generated.go", src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	config := types.Config{Importer: importer.Default()}
	if _, err := config.Check("generated", fset, []*ast.File{file}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestCodeWriter(t *testing.T) {
	w := NewCodeWriter("// header", "main")
	sourceEqual(t, []byte("// header\n\npackage main"), w.Bytes())
	w.UsePkg("fmt", "")
	w.UsePkg("time", "")
	w.UsePkg("github.com/cloudwego/gopkg", "gopkg")
	_, _ = w.Write([]byte("// hello main\n"))
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

func TestCodeWriterAllocatesImportAliases(t *testing.T) {
	w := NewCodeWriter("", "test")
	if got := w.UsePkgAlias("example.com/a/bar", ""); got != "bar" {
		t.Fatalf("first alias = %q, want bar", got)
	}
	if got := w.UsePkgAlias("example.net/b/bar", ""); got != "bar1" {
		t.Fatalf("second alias = %q, want bar1", got)
	}
	if got := w.UsePkgAlias("example.org/string", ""); got != "string1" {
		t.Fatalf("predeclared-name alias = %q, want string1", got)
	}

	w.F("var _ = bar.Value")
	w.F("var _ = bar1.Value")
	w.F("var _ = string1.Value")
	sourceEqual(t, []byte(`package test

import (
	bar "example.com/a/bar"
	bar1 "example.net/b/bar"
	string1 "example.org/string"
)

var _ = bar.Value
var _ = bar1.Value
var _ = string1.Value
`), w.Bytes())
}

func TestCodeWriterMixesImplicitAndAllocatedImports(t *testing.T) {
	w := NewCodeWriter("", "test")
	w.UsePkg("crypto/rand", "")
	if got := w.UsePkgAlias("crypto/rand", ""); got != "rand" {
		t.Fatalf("existing implicit import name = %q, want rand", got)
	}
	if got := w.UsePkgAlias("math/rand", ""); got != "rand1" {
		t.Fatalf("colliding import alias = %q, want rand1", got)
	}
	w.F("var _ = rand.Reader")
	w.F("var _ = rand1.Int")

	src := w.Bytes()
	sourceEqual(t, []byte(`package test

import (
	"crypto/rand"
	rand1 "math/rand"
)

var _ = rand.Reader
var _ = rand1.Int
`), src)
	typeCheckSource(t, src)
}

func TestCodeWriterRejectsLatePackageQualifierCollision(t *testing.T) {
	w := NewCodeWriter("", "test")
	if got := w.UsePkgAlias("math/rand", ""); got != "rand" {
		t.Fatalf("allocated import alias = %q, want rand", got)
	}
	w.UsePkg("math/rand", "")
	defer func() {
		if recover() == nil {
			t.Fatal("UsePkg did not reject a conflicting package qualifier")
		}
	}()
	w.UsePkg("crypto/rand", "")
}

func TestCodeWriterAllowsRepeatedSpecialImports(t *testing.T) {
	w := NewCodeWriter("", "test")
	w.UsePkg("crypto/rand", "_")
	w.UsePkg("math/rand", "_")
	w.UsePkg("fmt", ".")
	w.UsePkg("strings", ".")

	sourceEqual(t, []byte(`package test

import (
	_ "crypto/rand"
	. "fmt"
	_ "math/rand"
	. "strings"
)
`), w.Bytes())
}
