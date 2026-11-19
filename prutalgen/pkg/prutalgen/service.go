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
	"strings"

	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/strs"
	"github.com/cloudwego/prutal/prutalgen/internal/protoparser"
)

type Service struct {
	HeadComment   string
	InlineComment string
	Directives    Directives

	Name   string
	GoName string

	Methods []*Method
	Options Options

	Proto *Proto
}

func (s *Service) resolve() {
	p := s.Proto
	s.GoName = strs.GoCamelCase(strings.TrimPrefix(s.Name, p.Package+"."))
	for _, m := range s.Methods {
		m.resolve()
	}
}

func (s *Service) verify() error {
	var errs []error
	for _, method := range s.Methods {
		if !method.Request.IsMessage() {
			errs = append(errs, fmt.Errorf(
				"method %q request type %q is not a message type", method.Name, method.Request.Name))
		}
		if !method.Return.IsMessage() {
			errs = append(errs, fmt.Errorf(
				"method %q return type %q is not a message type", method.Name, method.Return.Name))
		}
	}
	return errors.Join(errs...)
}

type Method struct {
	HeadComment   string
	InlineComment string
	Directives    Directives

	Name   string // original name in IDL
	GoName string

	Request *Type
	Return  *Type

	RequestStream bool
	ReturnStream  bool

	Options Options

	Service *Service
}

func (r *Method) resolve() {
	r.GoName = strs.GoCamelCase(r.Name)
	r.Request.resolve(false)
	r.Return.resolve(false)
}

func (x *protoLoader) loadService(p *Proto, decl *protoparser.ServiceDecl) {
	s := &Service{
		HeadComment:   decl.HeadComment,
		InlineComment: decl.InlineComment,
		Name:          decl.Name,
		Proto:         p,
	}
	s.Directives.Parse(s.HeadComment, s.InlineComment)
	for _, opt := range decl.Options {
		s.Options = append(s.Options, x.loadScopedOption(opt, "a service")...)
	}
	for _, md := range decl.Methods {
		m := &Method{
			HeadComment:   md.HeadComment,
			InlineComment: md.InlineComment,
			Name:          md.Name,
			RequestStream: md.RequestStream,
			ReturnStream:  md.ReturnStream,
			Service:       s,
		}
		m.Directives.Parse(m.HeadComment, m.InlineComment)
		m.Request = &Type{Name: md.RequestType, pos: md.RequestPos, m: m}
		m.Return = &Type{Name: md.ReturnType, pos: md.ReturnPos, m: m}
		for _, opt := range md.Options {
			m.Options = append(m.Options, x.loadScopedOption(opt, "an RPC")...)
		}
		s.Methods = append(s.Methods, m)
	}
	p.Services = append(p.Services, s)
}
