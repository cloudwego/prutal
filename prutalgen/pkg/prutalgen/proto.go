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
	"errors"
	"fmt"
	"go/token"
	"path"
	"path/filepath"
	"strings"

	"github.com/cloudwego/prutal/prutalgen/internal/parser"
	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/strs"
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
	protoName string // slash-separated descriptor name used by M options
	Edition   string
	Package   string

	GoImport  string // the full import path
	GoPackage string // package name without path
	// goPackageExplicit reports whether GoPackage came from the suffix after
	// a semicolon instead of being derived from GoImport.
	goPackageExplicit bool
	// goPackageIdentity preserves explicit whitespace for protogen's package
	// consistency check while GoPackage contains the valid Go identifier.
	goPackageIdentity string

	Directives Directives

	goPackageFromM string

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

// DescriptorName returns the slash-separated protobuf descriptor name.
func (p *Proto) DescriptorName() string {
	if p.protoName != "" {
		return p.protoName
	}
	return filepath.ToSlash(p.ProtoFile)
}

func (p *Proto) setGoPackage(s string) {
	p.setGoPackageOptions(s, "")
}

func (p *Proto) setGoPackageOptions(fileOpt, mOpt string) {
	fileImport, filePackage, _ := strings.Cut(fileOpt, ";")
	mImport, mPackage, _ := strings.Cut(mOpt, ";")
	filePackageSet := filePackage != ""
	mPackageSet := mPackage != ""
	filePackageIdentity := filePackage
	mPackageIdentity := mPackage
	fileImport = strings.TrimSpace(fileImport)
	filePackage = strings.TrimSpace(filePackage)
	mPackage = strings.TrimSpace(mPackage)

	p.GoImport = fileImport
	if mImport != "" {
		p.GoImport = mImport
	}

	p.goPackageExplicit = true
	switch {
	case mPackageSet:
		p.GoPackage = mPackage
		p.goPackageIdentity = mPackageIdentity
	case filePackageSet:
		p.GoPackage = filePackage
		p.goPackageIdentity = filePackageIdentity
	default:
		p.goPackageExplicit = false
		packageImport := fileImport
		if packageImport == "" {
			packageImport = p.GoImport
		}
		if packageImport == "" {
			p.GoPackage = ""
		} else {
			p.GoPackage = strs.GoSanitized(path.Base(packageImport))
		}
		p.goPackageIdentity = p.GoPackage
	}
}

func (p *Proto) validateGoPackage() error {
	if p.GoImport == "" {
		return fmt.Errorf("unable to determine Go import path for %q", p.ProtoFile)
	}
	if !strings.ContainsAny(p.GoImport, "./") {
		return fmt.Errorf("invalid Go import path %q for %q", p.GoImport, p.ProtoFile)
	}
	if p.GoPackage == "" {
		return fmt.Errorf("unable to determine Go package name for %q", p.ProtoFile)
	}
	if p.goPackageExplicit && !token.IsIdentifier(p.GoPackage) {
		return fmt.Errorf("invalid Go package name %q for %q", p.GoPackage, p.ProtoFile)
	}
	return nil
}

func validateGoPackageConsistency(protos []*Proto) error {
	packageFiles := make(map[string]*Proto, len(protos))
	for _, p := range protos {
		identity := p.goPackageIdentity
		if identity == "" {
			identity = p.GoPackage
		}
		if previous := packageFiles[p.GoImport]; previous != nil {
			previousIdentity := previous.goPackageIdentity
			if previousIdentity == "" {
				previousIdentity = previous.GoPackage
			}
			if previousIdentity != identity {
				//nolint:staticcheck // match protoc-gen-go's diagnostic wording
				return fmt.Errorf("Go package %q has inconsistent names %q (%s) and %q (%s)",
					p.GoImport, previousIdentity, previous.ProtoFile, identity, p.ProtoFile)
			}
		}
		packageFiles[p.GoImport] = p
	}
	return nil
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
		if t := x.getType(name); t != nil {
			return t
		}
	}
	if name1, ok := trimPathPrefix(name, p.Package); ok {
		return p.getType(name1) // try again without package prefix
	}
	return nil
}

func (p *Proto) getTypeByFullName(name string) any {
	if p.Package != "" {
		var ok bool
		name, ok = trimPathPrefix(name, p.Package)
		if !ok {
			return nil
		}
	}
	for _, m := range p.Messages {
		if typ := m.getType(name); typ != nil {
			return typ
		}
	}
	for _, e := range p.Enums {
		if e.Name == name {
			return e
		}
	}
	return nil
}

func (p *Proto) hasServiceFullName(name string) bool {
	for _, s := range p.Services {
		if fullProtoName(p.Package, s.Name) == name {
			return true
		}
	}
	return false
}

func (p *Proto) hasSymbolFullName(name string) bool {
	if name != "" && (p.Package == name || strings.HasPrefix(p.Package, name+".")) {
		return true
	}
	hasEnumSymbol := func(e *Enum) bool {
		if e.FullName() == name {
			return true
		}
		scope := parentProtoName(e.FullName())
		for _, value := range e.Fields {
			if fullProtoName(scope, value.Name) == name {
				return true
			}
		}
		return false
	}

	var hasMessageSymbol func(*Message) bool
	hasMessageSymbol = func(m *Message) bool {
		messageName := m.FullName()
		if messageName == name {
			return true
		}
		for _, field := range m.Fields {
			if fullProtoName(messageName, field.Name) == name {
				return true
			}
		}
		for _, oneof := range m.Oneofs {
			if fullProtoName(messageName, oneof.Name) == name {
				return true
			}
		}
		for _, e := range m.Enums {
			if hasEnumSymbol(e) {
				return true
			}
		}
		for _, nested := range m.Messages {
			if hasMessageSymbol(nested) {
				return true
			}
		}
		return false
	}

	for _, e := range p.Enums {
		if hasEnumSymbol(e) {
			return true
		}
	}
	for _, m := range p.Messages {
		if hasMessageSymbol(m) {
			return true
		}
	}
	for _, s := range p.Services {
		serviceName := fullProtoName(p.Package, s.Name)
		if serviceName == name {
			return true
		}
		for _, method := range s.Methods {
			if fullProtoName(serviceName, method.Name) == name {
				return true
			}
		}
	}
	return false
}

func validateProtoSymbols(protos []*Proto) error {
	packageScopes := make(map[string]bool)
	for _, p := range protos {
		for name := p.Package; name != ""; name = parentProtoName(name) {
			packageScopes[name] = true
		}
	}

	type declaration struct {
		kind  string
		proto *Proto
	}
	declarations := make(map[string]declaration)
	register := func(name, kind string, p *Proto) error {
		if packageScopes[name] {
			return fmt.Errorf("%s %q in %s conflicts with a package name", kind, name, p.ProtoFile)
		}
		if previous, ok := declarations[name]; ok {
			return fmt.Errorf("%s %q in %s is already defined as a %s in %s",
				kind, name, p.ProtoFile, previous.kind, previous.proto.ProtoFile)
		}
		declarations[name] = declaration{kind: kind, proto: p}
		return nil
	}
	registerEnum := func(e *Enum, p *Proto) error {
		if err := register(e.FullName(), "enum", p); err != nil {
			return err
		}
		scope := parentProtoName(e.FullName())
		for _, value := range e.Fields {
			if err := register(fullProtoName(scope, value.Name), "enum value", p); err != nil {
				return err
			}
		}
		return nil
	}

	var registerMessage func(*Message, *Proto) error
	registerMessage = func(m *Message, p *Proto) error {
		if err := register(m.FullName(), "message", p); err != nil {
			return err
		}
		for _, field := range m.Fields {
			if err := register(fullProtoName(m.FullName(), field.Name), "field", p); err != nil {
				return err
			}
		}
		for _, oneof := range m.Oneofs {
			if err := register(fullProtoName(m.FullName(), oneof.Name), "oneof", p); err != nil {
				return err
			}
		}
		for _, e := range m.Enums {
			if err := registerEnum(e, p); err != nil {
				return err
			}
		}
		for _, nested := range m.Messages {
			if err := registerMessage(nested, p); err != nil {
				return err
			}
		}
		return nil
	}
	for _, p := range protos {
		for _, e := range p.Enums {
			if err := registerEnum(e, p); err != nil {
				return err
			}
		}
		for _, m := range p.Messages {
			if err := registerMessage(m, p); err != nil {
				return err
			}
		}
		for _, s := range p.Services {
			serviceName := fullProtoName(p.Package, s.Name)
			if err := register(serviceName, "service", p); err != nil {
				return err
			}
			for _, method := range s.Methods {
				if err := register(fullProtoName(serviceName, method.Name), "method", p); err != nil {
					return err
				}
			}
		}
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
	return errors.Join(errs...)
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
	imp.Proto = x.loadProto(importpath, false)
	for _, previous := range p.Imports {
		if previous.Proto == imp.Proto {
			x.Fatalf("%s - import %q was listed twice", getTokenPos(c), importpath)
		}
	}
	p.Imports = append(p.Imports, imp)
}
