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
	"strings"
)

const importPlaceHolder = "{PRUTALGEN_GO_IMPORTS}"

type CodeWriter struct {
	buf  *bytes.Buffer
	pkgs map[string]string // import -> alias
}

func NewCodeWriter(header, pkg string) *CodeWriter {
	w := &CodeWriter{
		buf:  &bytes.Buffer{},
		pkgs: make(map[string]string),
	}
	w.F("%s", header)
	w.F("")
	w.F("package %s", pkg)
	w.F("")
	w.F(importPlaceHolder)
	return w
}

func (w *CodeWriter) UsePkg(p, a string) {
	if p == "" {
		return
	}
	if path.Base(p) == a {
		w.pkgs[p] = ""
	} else {
		w.pkgs[p] = a
	}
}

func (w *CodeWriter) genImports() []byte {
	const (
		cloudwegoRepoPrefix = "github.com/cloudwego/"
	)
	pp0 := make([]string, 0, len(w.pkgs))
	pp1 := make([]string, 0, len(w.pkgs)) // for cloudwego
	for pkg := range w.pkgs {             // grouping
		if strings.HasPrefix(pkg, cloudwegoRepoPrefix) {
			pp1 = append(pp1, pkg)
		} else {
			pp0 = append(pp0, pkg)
		}
	}

	// check if need an empty line between groups
	if len(pp0) != 0 && len(pp1) > 0 {
		pp0 = append(pp0, "")
	}

	// no imports?
	pp0 = append(pp0, pp1...)
	if len(pp0) == 0 {
		return nil
	}

	s := &bytes.Buffer{}
	if len(pp0) == 1 { // only imports one pkg?
		fmt.Fprintf(s, "import %s %q", w.pkgs[pp0[0]], pp0[0])
	} else { // more than one imports
		fmt.Fprintln(s, "import (")
		for _, p := range pp0 {
			if p == "" {
				fmt.Fprintln(s, "")
			} else {
				fmt.Fprintf(s, "%s %q\n", w.pkgs[p], p)
			}
		}
		fmt.Fprint(s, ")")
	}
	return s.Bytes()
}

func (w *CodeWriter) F(format string, a ...interface{}) {
	if len(a) == 0 {
		w.buf.WriteString(format)
	} else {
		fmt.Fprintf(w.buf, format, a...)
	}

	// always newline for each call
	if len(format) == 0 {
		w.buf.WriteByte('\n')
	} else if buf := w.buf.Bytes(); len(buf) == 0 || buf[len(buf)-1] != '\n' {
		w.buf.WriteByte('\n')
	}
}

func (w *CodeWriter) Bytes() []byte {
	b := w.buf.Bytes()
	b = bytes.Replace(b, []byte(importPlaceHolder), w.genImports(), 1)
	return b
}
