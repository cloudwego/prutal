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
	"fmt"
	"path"
	"strings"

	"github.com/cloudwego/prutal/prutalgen/internal/parser"
	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/text"
)

// Import ...
// https://protobuf.com/docs/language-spec#imports
type Import struct {
	*Proto

	Public bool
}

// Proto represents a proto file
type Proto struct {
	ProtoFile string
	Edition   string
	Package   string

	GoImport  string // the full import path
	GoPackage string // package name without path

	Directives Directives

	Imports []*Import
	Options Options

	Enums    []*Enum
	Messages []*Message

	Services []*Service

	l LoggerIface
}

func (p *Proto) String() string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "Proto %s Edition %s Package %s\n", p.ProtoFile, p.Edition, p.Package)
	fmt.Fprintf(b, "Options: %v\n", p.Options)
	for _, e := range p.Enums {
		fmt.Fprintf(b, "- %s\n", e.String())
	}
	for _, m := range p.Messages {
		fmt.Fprintf(b, "- %s\n", m.String())
	}
	return b.String()
}

func (p *Proto) setGoPackage(s string) {
	imp, pkg, _ := strings.Cut(s, ";")
	imp = strings.TrimSpace(imp)
	pkg = strings.TrimSpace(pkg)
	p.GoImport = imp
	if pkg != "" {
		p.GoPackage = pkg
	} else if p.GoImport != "" {
		p.GoPackage = path.Base(p.GoImport)
	}
}

func (p *Proto) getType(name string) any {
	for _, m := range p.Messages {
		if v := m.getType(name); v != nil {
			return v
		}
	}
	for _, e := range p.Enums {
		if e.Name == name {
			return e
		}
	}
	for _, x := range p.Imports {
		if !x.Public {
			continue
		}
		if t := x.Proto.getType(name); t != nil {
			return t
		}
	}
	if name1, ok := trimPathPrefix(name, p.Package); ok {
		return p.getType(name1) // try again without package prefix
	}
	return nil
}

func (p *Proto) IsProto2() bool {
	return p.Edition == editionProto2
}

func (p *Proto) IsProto3() bool {
	return p.Edition == editionProto3
}

func (p *Proto) IsEdition2023() bool {
	return p.Edition == edition2023
}

func (p *Proto) refFile() string {
	return refPath(p.ProtoFile)
}

func (p *Proto) Fatalf(fm string, aa ...any) {
	p.l.Fatalf("[FATAL] "+p.refFile()+": "+fm, aa...)
}

func (p *Proto) Warnf(fm string, aa ...any) {
	p.l.Printf("[WARN ] "+p.refFile()+": "+fm, aa...)
}

func (p *Proto) Infof(fm string, aa ...any) {
	p.l.Printf("[INFO ] "+p.refFile()+": "+fm, aa...)
}

func (p *Proto) resolve() {
	for _, e := range p.Enums {
		e.resolve()
	}
	for _, m := range p.Messages {
		m.resolve()
	}
	for _, s := range p.Services {
		s.resolve()
	}
}

func (p *Proto) verify() error {
	var errs []error
	for _, e := range p.Enums {
		if err := e.verify(); err != nil {
			errs = append(errs, fmt.Errorf("enum %s verify err: %w", e.FullName(), err))
		}
	}
	for _, m := range p.Messages {
		if err := m.verify(); err != nil {
			errs = append(errs, fmt.Errorf("message %s verify err: %w", m.FullName(), err))
		}
	}
	for _, s := range p.Services {
		if err := s.verify(); err != nil {
			errs = append(errs, fmt.Errorf("service %q verify err: %w", s.Name, err))
		}
	}
	return joinErrs(errs...)
}

// listeners

func (x *protoLoader) ExitEdition(c *parser.EditionContext) {
	p := x.currentProto()
	s := c.StrLit()
	v, err := text.UnmarshalString(s.GetText())
	if err != nil {
		x.Fatalf("%s : %s", getTokenPos(s), err)
	}
	switch v {
	case editionProto2, editionProto3, edition2023:
		p.Edition = v
	default:
		x.Fatalf("%s : unknown syntax/edition %q", getTokenPos(s), v)
	}
}

func (x *protoLoader) ExitPackageStatement(c *parser.PackageStatementContext) {
	p := x.currentProto()
	if p.Package != "" {
		x.Fatalf("%s - Multiple package definitions.", getTokenPos(c))
	}
	p.Package = c.FullIdent().GetText()
}

func (x *protoLoader) ExitImportStatement(c *parser.ImportStatementContext) {
	// IMPORT (WEAK | PUBLIC)? strLit SEMI
	p := x.currentProto()
	if len(getText(c.WEAK())) > 0 {
		x.Warnf("%s - weak import is not supported", getTokenPos(c))
	}
	imp := &Import{}
	if len(getText(c.PUBLIC())) > 0 {
		imp.Public = true
	}
	s := c.StrLit().GetText()
	importpath, err := text.UnmarshalString(s)
	if err != nil {
		x.Fatalf("%s - import syntax err: %s", getTokenPos(c), err)
	}
	imp.Proto = x.loadProto(importpath)
	p.Imports = append(p.Imports, imp)
}
