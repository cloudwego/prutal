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

	"github.com/cloudwego/prutal/prutalgen/internal/antlr"

	"github.com/cloudwego/prutal/prutalgen/internal/parser"
)

type protoLoader struct {
	*parser.BaseProtobufListener

	includes    []string
	proto2gopkg map[string]string

	protos []*Proto // all proto files appended by order

	// state vars
	streams    []*streamContext // mainly for comment
	protostack []*Proto
	msgstack   []*Message
	enum       *Enum // current enum

	l LoggerIface
}

type Loader interface {
	SetLogger(LoggerIface)
	LoadProto(file string) []*Proto
}

func NewLoader(includes []string, proto2gopkg map[string]string) Loader {
	// add empty element to check filepath without include in searchProtoFile
	includes = append(includes, "")
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
	return filepath.FromSlash(path.Join(incl, file))
}

func (x *protoLoader) searchProtoFile(file string) string {
	for _, incl := range x.includes {
		fn := fullFilename(incl, file)
		fn, err := filepath.Abs(fn)
		if err != nil {
			continue
		}
		if _, err := os.Stat(fn); err == nil {
			return fn
		}
	}
	x.Fatalf("proto file %q not found in includes %v", file, x.includes)
	return "" // never goes here
}

func (x *protoLoader) LoadProto(file string) []*Proto {
	x.reset()
	_ = x.loadProto(file)
	x.protos = sortProtoFiles(x.protos)       // sort by topological order
	for i := len(x.protos) - 1; i >= 0; i-- { // resolve in reverse topological order
		p := x.protos[i]
		p.resolve()
		if err := p.verify(); err != nil {
			x.Fatalf("proto %s verify err: %s", p.ProtoFile, err)
		}
	}
	return x.protos
}

func (x *protoLoader) reset() {
	x.protos = nil
	x.streams = nil
	x.protostack = nil
	x.msgstack = nil
	x.enum = nil
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

func (x *protoLoader) currentStream() *streamContext {
	return last(x.streams)
}

func (x *protoLoader) currentProto() *Proto {
	return last(x.protostack)
}

func (x *protoLoader) currentMsg() *Message {
	return last(x.msgstack)
}

func (x *protoLoader) currentOneof() *Oneof {
	m := x.currentMsg()
	return last(m.Oneofs)
}

func (x *protoLoader) currentService() *Service {
	p := x.currentProto()
	return last(p.Services)
}

func (x *protoLoader) getByFile(fn string, stack bool) *Proto {
	if !stack {
		for _, p := range x.protos {
			if p.ProtoFile == fn {
				return p
			}
		}
		return nil
	}
	for _, p := range x.protostack {
		if p.ProtoFile == fn {
			return p
		}
	}
	return nil
}

func (x *protoLoader) loadProto(file string) *Proto {
	if embeddedProtos[file] != nil {
		return x.loadEmbeddedProto(file)
	}
	protofile := x.searchProtoFile(file)

	if proto := x.getByFile(protofile, true); proto != nil {
		files := make([]string, 0, len(x.protostack))
		for _, p := range x.protostack {
			files = append(files, p.ProtoFile)
		}
		x.l.Fatalf("cyclic import is NOT allowed: %s", strings.Join(files, " \n\t-> "))
		return proto
	}

	if proto := x.getByFile(protofile, false); proto != nil {
		return proto // parsed
	}

	p := &Proto{
		ProtoFile: protofile,
		Edition:   editionProto2,
		l:         x.l,
	}
	p.setGoPackage(x.proto2gopkg[file])
	push(&x.protostack, p)
	defer pop(&x.protostack)
	x.protos = append(x.protos, p)

	x.Infof("parsing")
	is, err := antlr.NewFileStream(p.ProtoFile)
	if err != nil {
		x.Fatalf("open file err: %s", err)
	}
	x.parseInput(is)
	return p
}

func (x *protoLoader) loadEmbeddedProto(file string) *Proto {
	if proto := x.getByFile(file, false); proto != nil {
		return proto // parsed
	}

	data := embeddedProtos[file]
	p := &Proto{
		ProtoFile: file,
		Edition:   editionProto2,

		l: x.l,
	}
	push(&x.protostack, p)
	defer pop(&x.protostack)
	x.protos = append(x.protos, p)
	is := antlr.NewIoStream(bytes.NewReader(data))
	x.parseInput(is)
	return p
}

func (x *protoLoader) parseInput(in antlr.CharStream) {
	p := x.currentProto()

	lexer := parser.NewProtobufLexer(in)
	s := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	push(&x.streams, newStreamContext(s))
	defer pop(&x.streams)

	e := &errorListener{l: x.l}
	ps := parser.NewProtobufParser(s)
	ps.RemoveErrorListeners()
	ps.AddErrorListener(e)
	proto := ps.Proto()
	if e.HasError() {
		x.Fatalf("error occurred during parsing proto file")
	}
	antlr.ParseTreeWalkerDefault.Walk(x, proto)

	gopkg, ok := p.Options.Get("go_package")
	if ok {
		p.setGoPackage(gopkg)
	}
	if p.GoPackage == "" {
		x.Fatalf(`option "go_package" not set`)
	}
}

func (x *protoLoader) consumeHeadComment(c antlr.ParserRuleContext) string {
	s := x.currentStream()
	return s.consumeHeadComment(c)
}

func (x *protoLoader) consumeInlineComment(c antlr.ParserRuleContext) string {
	s := x.currentStream()
	return s.consumeInlineComment(c)
}

type errorListener struct {
	*antlr.DefaultErrorListener
	hasError bool

	l LoggerIface // from protoLoader
}

func (x *errorListener) SyntaxError(_ antlr.Recognizer, _ any,
	line, column int, msg string, _ antlr.RecognitionException) {
	x.hasError = true
	x.l.Printf("[ERROR] syntax error at line %d column %d - %s\n", line, column, msg)
}

func (x *errorListener) HasError() bool { return x.hasError }
