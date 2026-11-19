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
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/cloudwego/prutal/prutalgen/internal/protoparser"
)

type protoLoader struct {
	includes    []string
	proto2gopkg map[string]string

	protos []*Proto // all proto files appended by order

	protostack []*Proto

	l LoggerIface
}

type Loader interface {
	SetLogger(LoggerIface)
	// LoadProto loads one root and returns its complete import graph.
	LoadProto(file string) []*Proto
	// LoadProtos loads multiple roots as one invocation and returns the roots
	// in input order.
	LoadProtos(files []string) []*Proto
}

func NewLoader(includes []string, proto2gopkg map[string]string) Loader {
	includes = append([]string(nil), includes...)
	if len(includes) == 0 {
		includes = append(includes, ".")
	}
	return &protoLoader{
		includes:    includes,
		proto2gopkg: proto2gopkg,

		l: defaultLogger,
	}
}

func (x *protoLoader) SetLogger(l LoggerIface) {
	if l == nil {
		x.l = defaultLogger
	} else {
		x.l = l
	}
}

func fullFilename(incl string, file string) string {
	// file path in proto would be in the form of unix style
	// need `filepath.FromSlash` for converting it on windows
	if filepath.IsAbs(file) {
		return filepath.Clean(file)
	}
	return filepath.Join(incl, filepath.FromSlash(file))
}

func (x *protoLoader) searchProtoFile(file string, root bool) (string, string) {
	if isCanonicalProtoName(file) {
		if fn := x.searchVirtualProto(file); fn != "" {
			return fn, file
		}
	}

	if root {
		if info, err := os.Stat(file); err == nil && !info.IsDir() {
			fn, err := filepath.Abs(file)
			if err == nil {
				protoName, ok := x.protoNameForPhysical(file, fn)
				if !ok {
					x.Fatalf("proto file %q does not reside within any include path %v", file, x.includes)
				}
				resolved := x.searchVirtualProto(protoName)
				if !sameFile(fn, resolved) {
					x.Fatalf("proto file %q is shadowed by %q in the include path", file, resolved)
				}
				return resolved, protoName
			}
		}
	}

	x.Fatalf("proto file %q not found in includes %v", file, x.includes)
	return "", "" // never goes here
}

func (x *protoLoader) searchVirtualProto(name string) string {
	if !isCanonicalProtoName(name) {
		return ""
	}
	for _, incl := range x.includes {
		fn := fullFilename(incl, name)
		fn, err := filepath.Abs(fn)
		if err != nil {
			continue
		}
		if info, err := os.Stat(fn); err == nil && !info.IsDir() {
			return fn
		}
	}
	return ""
}

func (x *protoLoader) protoNameForPhysical(requested, protofile string) (string, bool) {
	for _, incl := range x.includes {
		rel, ok := relativeToInclude(incl, requested)
		if !ok || !isCanonicalProtoName(rel) {
			continue
		}
		candidate, err := filepath.Abs(fullFilename(incl, rel))
		if err != nil || !sameFile(protofile, candidate) {
			continue
		}
		return filepath.ToSlash(rel), true
	}
	return "", false
}

func relativeToInclude(incl, requested string) (string, bool) {
	if incl == "" {
		incl = "."
	}
	incl = normalizePhysicalPath(incl)
	requested = normalizePhysicalPath(requested)
	if incl == "" {
		if requested == "" || path.IsAbs(requested) || filepath.VolumeName(requested) != "" {
			return "", false
		}
		return requested, true
	}
	prefix := strings.TrimSuffix(incl, "/") + "/"
	if !strings.HasPrefix(requested, prefix) {
		return "", false
	}
	return strings.TrimPrefix(requested, prefix), true
}

func normalizePhysicalPath(name string) string {
	volume := filepath.ToSlash(filepath.VolumeName(name))
	rest := name[len(filepath.VolumeName(name)):]
	absolute := len(rest) > 0 && os.IsPathSeparator(rest[0])
	parts := strings.FieldsFunc(rest, func(r rune) bool {
		return r < 128 && os.IsPathSeparator(uint8(r))
	})
	kept := parts[:0]
	for _, part := range parts {
		if part != "." {
			kept = append(kept, part)
		}
	}
	normalized := strings.Join(kept, "/")
	if absolute {
		normalized = "/" + normalized
	}
	return volume + normalized
}

func isCanonicalProtoName(name string) bool {
	return name != "" && !path.IsAbs(name) && !filepath.IsAbs(name) &&
		!strings.ContainsRune(name, '\\') && path.Clean(name) == name && name != "." &&
		name != ".." && !strings.HasPrefix(name, "../")
}

func sameFile(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	aInfo, err := os.Stat(a)
	if err != nil {
		return false
	}
	bInfo, err := os.Stat(b)
	return err == nil && os.SameFile(aInfo, bInfo)
}

func (x *protoLoader) LoadProto(file string) []*Proto {
	x.loadRoots([]string{file})
	return x.protos
}

func (x *protoLoader) LoadProtos(files []string) []*Proto {
	return x.loadRoots(files)
}

func (x *protoLoader) loadRoots(files []string) []*Proto {
	x.reset()
	roots := make([]*Proto, 0, len(files))
	seenRoots := make(map[*Proto]bool, len(files))
	for _, file := range files {
		p := x.loadProto(file, true)
		if !seenRoots[p] {
			roots = append(roots, p)
			seenRoots[p] = true
		}
	}
	if err := validateGoPackageConsistency(x.protos); err != nil {
		x.Fatalf("%s", err)
	}
	if err := validateProtoSymbols(x.protos); err != nil {
		x.Fatalf("%s", err)
	}
	x.protos = sortProtoFiles(x.protos)       // sort by topological order
	for i := len(x.protos) - 1; i >= 0; i-- { // resolve in reverse topological order
		p := x.protos[i]
		p.resolve()
		if err := p.verify(); err != nil {
			x.Fatalf("proto %s verify err: %s", p.ProtoFile, err)
		}
	}
	return roots
}

func (x *protoLoader) reset() {
	x.protos = nil
	x.protostack = nil
}

func (x *protoLoader) Fatalf(fm string, aa ...any) {
	if len(x.protostack) > 0 {
		x.currentProto().Fatalf(fm, aa...)
	} else {
		x.l.Fatalf("[FATAL] "+fm, aa...)
	}
}

func (x *protoLoader) Warnf(fm string, aa ...any) {
	if len(x.protostack) > 0 {
		x.currentProto().Warnf(fm, aa...)
	} else {
		x.l.Printf("[WARN ] "+fm, aa...)
	}
}

func (x *protoLoader) Infof(fm string, aa ...any) {
	if len(x.protostack) > 0 {
		x.currentProto().Infof(fm, aa...)
	} else {
		x.l.Printf("[INFO ] "+fm, aa...)
	}
}

func (x *protoLoader) currentProto() *Proto {
	return last(x.protostack)
}

func (x *protoLoader) getByName(name string, stack bool) *Proto {
	if !stack {
		for _, p := range x.protos {
			if p.protoName == name {
				return p
			}
		}
		return nil
	}
	for _, p := range x.protostack {
		if p.protoName == name {
			return p
		}
	}
	return nil
}

func (x *protoLoader) loadProto(file string, root bool) *Proto {
	if embeddedProtos[file] != nil {
		return x.loadEmbeddedProto(file)
	}
	protofile, protoName := x.searchProtoFile(file, root)

	if proto := x.getByName(protoName, true); proto != nil {
		files := make([]string, 0, len(x.protostack))
		for _, p := range x.protostack {
			files = append(files, p.ProtoFile)
		}
		x.l.Fatalf("cyclic import is NOT allowed: %s", strings.Join(files, " \n\t-> "))
		return proto
	}

	if proto := x.getByName(protoName, false); proto != nil {
		return proto // parsed
	}

	p := &Proto{
		ProtoFile:      protofile,
		protoName:      protoName,
		Edition:        editionProto2,
		goPackageFromM: x.proto2gopkg[protoName],
		l:              x.l,
	}
	push(&x.protostack, p)
	defer pop(&x.protostack)
	x.protos = append(x.protos, p)

	x.Infof("parsing")
	data, err := os.ReadFile(p.ProtoFile)
	if err != nil {
		x.Fatalf("open file err: %s", err)
	}
	x.parseInput(data)
	return p
}

func (x *protoLoader) loadEmbeddedProto(file string) *Proto {
	if proto := x.getByName(file, false); proto != nil {
		return proto // parsed
	}

	data := embeddedProtos[file]
	p := &Proto{
		ProtoFile:      file,
		protoName:      file,
		Edition:        editionProto2,
		goPackageFromM: x.proto2gopkg[file],

		l: x.l,
	}
	push(&x.protostack, p)
	defer pop(&x.protostack)
	x.protos = append(x.protos, p)
	x.parseInput(data)
	return p
}

func (x *protoLoader) parseInput(data []byte) {
	p := x.currentProto()
	if !utf8.Valid(data) {
		// Replace invalid UTF-8 sequences so comments and option values remain valid UTF-8.
		data = bytes.ToValidUTF8(data, []byte("\uFFFD"))
	}

	file, errs := protoparser.Parse(data)
	if len(errs) > 0 {
		for _, e := range errs {
			x.l.Printf("[ERROR] syntax error at line %d column %d - %s\n", e.Pos.Line, e.Pos.Column, e.Msg)
		}
		x.Fatalf("error occurred during parsing proto file: %s", errs[0].Msg)
	}
	x.loadFile(p, file)

	p.Directives.Parse(file.HeadComment)

	gopkg, _ := p.Options.Get("go_package")
	p.setGoPackageOptions(gopkg, p.goPackageFromM)
	if err := p.validateGoPackage(); err != nil {
		x.Fatalf("%s", err)
	}
}

func (x *protoLoader) loadFile(p *Proto, file *protoparser.File) {
	if file.Syntax != nil {
		x.loadSyntax(p, file.Syntax)
	}
	for _, pkg := range file.Packages {
		x.loadPackage(p, pkg)
	}
	for _, imp := range file.Imports {
		x.loadImport(p, imp)
	}
	for _, opt := range file.Options {
		x.loadFileOption(p, opt)
	}
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *protoparser.MessageDecl:
			x.loadMessage(p, nil, d)
		case *protoparser.EnumDecl:
			x.loadEnum(p, nil, d)
		case *protoparser.ExtendDecl:
			x.loadExtend(d)
		case *protoparser.ServiceDecl:
			x.loadService(p, d)
		}
	}
}
