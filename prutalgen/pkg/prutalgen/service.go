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
	"strings"

	"github.com/cloudwego/prutal/prutalgen/internal/parser"
	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/strs"
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

func (s *Service) verify() error { return nil }

type Method struct {
	HeadComment   string
	InlineComment string
	Directives    Directives

	Name   string // orignal name in IDL
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

func (x *protoLoader) EnterServiceDef(c *parser.ServiceDefContext) {
	// SERVICE serviceName LC serviceElement* RC
	s := &Service{
		HeadComment: x.consumeHeadComment(c), InlineComment: x.consumeInlineComment(c),
		Name:  c.ServiceName().GetText(),
		Proto: x.currentProto(),
	}
	s.Directives.Parse(s.HeadComment, s.InlineComment)

	p := x.currentProto()
	p.Services = append(p.Services, s)
}

func (x *protoLoader) EnterRpc(c *parser.RpcContext) {
	// rpc = "rpc" rpcName "(" [ "stream" ] messageType ")" "returns" "(" [ "stream" ]
	//	messageType ")" (( "{" {option | emptyStatement } "}" ) | ";")
	s := x.currentService()
	m := &Method{
		HeadComment: x.consumeHeadComment(c), InlineComment: x.consumeInlineComment(c),
		Name:    c.RpcName().GetText(),
		Service: s,
	}
	m.Directives.Parse(m.HeadComment, m.InlineComment)
	s.Methods = append(s.Methods, m)
}

func (x *protoLoader) ExitRpc(c *parser.RpcContext) {
	m := last(x.currentService().Methods)
	tt := c.AllMessageType() // 0 for request, 1 for return
	for _, s := range c.AllSTREAM() {
		if pos := s.GetSymbol().GetStart(); pos < tt[0].GetStart().GetStart() {
			m.RequestStream = true
		} else if pos < tt[1].GetStart().GetStart() {
			m.ReturnStream = true
		}
	}
	m.Request = &Type{
		Name: tt[0].GetText(),
		rule: tt[0],
		m:    m,
	}
	m.Return = &Type{
		Name: tt[1].GetText(),
		rule: tt[1],
		m:    m,
	}
}
