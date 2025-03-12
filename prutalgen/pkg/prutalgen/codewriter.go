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
	"bytes"
	"fmt"
	"path"
	"sort"
	"strings"
)

const importPlaceHolder = "{PRUTALGEN_GO_IMPORTS}"

// ImportGroupingFunc returns the weight for grouping for the given pkg
//
// - Packages with higher weights will be placed in later groups.
// - Packages with the same weight will be placed together.
type ImportGroupingFunc func(pkg string) (weight int)

func defaultGroupingFunc(pkg string) int {
	if strings.HasPrefix(pkg, "github.com/cloudwego/") {
		return 10
	}
	return 0
}

// CodeWriter wraps a simple code writer used by prutal
type CodeWriter struct {
	buf  *bytes.Buffer
	pkgs map[string]string // import -> alias

	grouping ImportGroupingFunc
}

// NewCodeWriter
func NewCodeWriter(header, pkg string) *CodeWriter {
	w := &CodeWriter{
		buf:  &bytes.Buffer{},
		pkgs: make(map[string]string),

		grouping: defaultGroupingFunc,
	}
	w.Reset(header, pkg)
	return w
}

func (w *CodeWriter) Reset(header, pkg string) {
	w.buf.Reset()
	for k := range w.pkgs {
		delete(w.pkgs, k)
	}
	if header != "" {
		w.F("%s", header)
		w.F("")
	}
	w.F("package %s", pkg)
	w.F("")
	w.F(importPlaceHolder)
	w.F("")
}

// UsePkg records packages used, and import them when calling `Bytes()`
func (w *CodeWriter) UsePkg(p, a string) {
	if p == "" {
		return
	}
	if path.Base(p) == a {
		w.pkgs[p] = "" // remove alias if it's same as its path.Base
	} else {
		w.pkgs[p] = a
	}
}

// SetGroupingFunc updates the grouping func used when generating imports.
//
// see comments of `ImportGroupingFunc` for details
func (w *CodeWriter) SetGroupingFunc(f ImportGroupingFunc) {
	if f != nil {
		w.grouping = f
	}
}

func (w *CodeWriter) genImports() []byte {
	if len(w.pkgs) == 0 {
		return nil
	}

	// if only one pkg, no need grouping
	if len(w.pkgs) == 1 {
		for p, a := range w.pkgs {
			return []byte(fmt.Sprintf("import %s %q", a, p))
		}
		return nil
	}

	type importPkg struct {
		w int
		p string
	}

	// calc weights
	pp := make([]importPkg, 0, len(w.pkgs))
	for pkg := range w.pkgs {
		weight := w.grouping(pkg)
		pp = append(pp, importPkg{w: weight, p: pkg})
	}

	// sort by weight & package name
	sort.Slice(pp, func(i, j int) bool {
		a, b := &pp[i], &pp[j]
		if a.w != b.w {
			return a.w < b.w
		}
		return a.p < b.p
	})

	// gen imports
	s := &bytes.Buffer{}
	fmt.Fprintln(s, "import (")
	for i, p := range pp {
		if i > 0 && p.w != pp[i-1].w {
			fmt.Fprintln(s, "") // empty line between grouping
		}
		fmt.Fprintf(s, "%s %q\n", w.pkgs[p.p], p.p)
	}
	fmt.Fprint(s, ")")
	return s.Bytes()
}

// Write implements io.Writer
func (w *CodeWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *CodeWriter) F(format string, a ...interface{}) {
	if len(format) == 0 {
		w.buf.WriteByte('\n')
		return
	}

	if len(a) == 0 {
		w.buf.WriteString(format)
	} else {
		fmt.Fprintf(w.buf, format, a...)
	}

	// always newline for each call
	if b := w.buf.Bytes(); b[len(b)-1] != '\n' {
		w.buf.WriteByte('\n')
	}
}

func (w *CodeWriter) Bytes() []byte {
	b := w.buf.Bytes()
	b = bytes.Replace(b, []byte(importPlaceHolder), w.genImports(), 1)
	return b
}
