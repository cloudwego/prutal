/*
 * Copyright 2026 CloudWeGo Authors
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

package protoparser

import (
	"fmt"
	"sort"
	"strings"
)

const maxParseDepth = 256

var validMapKeyTypes = map[string]bool{
	"int32":    true,
	"int64":    true,
	"uint32":   true,
	"uint64":   true,
	"sint32":   true,
	"sint64":   true,
	"fixed32":  true,
	"fixed64":  true,
	"sfixed32": true,
	"sfixed64": true,
	"bool":     true,
	"string":   true,
}

func isScalarType(s string) bool {
	return validMapKeyTypes[s] || s == "double" || s == "float" || s == "bytes"
}

// Parse lexes and parses a Protocol Buffers source file into a File AST.
func Parse(src []byte) (*File, []SyntaxError) {
	stream := Lex(src)
	p := &parser{
		src:    src,
		stream: stream,
		errors: append([]SyntaxError(nil), stream.Errors...),
	}
	file := p.parseFile()
	sort.SliceStable(p.errors, func(i, j int) bool {
		return p.errors[i].Pos.Offset < p.errors[j].Pos.Offset
	})
	return file, p.errors
}

type parser struct {
	src    []byte
	stream *TokenStream
	pos    int
	errors []SyntaxError
}

func (p *parser) cur() Token {
	return p.stream.At(p.pos)
}

func (p *parser) peek(offset int) Token {
	return p.stream.At(p.pos + offset)
}

func (p *parser) advance() Token {
	tok := p.cur()
	if tok.Kind != TokenEOF {
		p.pos++
	}
	return tok
}

func (p *parser) errorf(pos Pos, format string, args ...any) {
	p.errors = append(p.errors, SyntaxError{
		Pos: pos,
		Msg: fmt.Sprintf(format, args...),
	})
}

// errorUnexpected reports an unexpected token error unless tok is already a lex
// TokenError (whose error was already recorded during Lex).
func (p *parser) errorUnexpected(tok Token, expected string) {
	if tok.Kind == TokenError {
		return
	}
	if tok.Kind == TokenEOF {
		p.errorf(tok.Pos, "expected %s, got EOF", expected)
		return
	}
	p.errorf(tok.Pos, "expected %s, got %q", expected, tok.Text)
}

func (p *parser) expect(kind TokenKind) (Token, bool) {
	tok := p.cur()
	if tok.Kind == kind {
		p.advance()
		return tok, true
	}
	if kind == TokenSemi {
		if tok.Kind != TokenError {
			p.errorf(tok.Pos, "missing ';'")
		}
	} else {
		p.errorUnexpected(tok, kind.String())
	}
	return Token{}, false
}

func (p *parser) expectIdent(what string) (Token, bool) {
	tok := p.cur()
	if tok.Kind == TokenIdent {
		p.advance()
		return tok, true
	}
	p.errorUnexpected(tok, what)
	return Token{}, false
}

// recoverStmt skips forward until a statement boundary (';', '}', or EOF) at
// zero brace depth, consuming ';' if reached so enclosing loops strictly make
// progress.
func (p *parser) recoverStmt() {
	depth := 0
	for {
		switch p.cur().Kind {
		case TokenEOF:
			return
		case TokenLBrace:
			depth++
			p.advance()
		case TokenRBrace:
			if depth == 0 {
				return
			}
			depth--
			p.advance()
		case TokenSemi:
			p.advance()
			if depth == 0 {
				return
			}
		default:
			p.advance()
		}
	}
}

// recoverInBlock skips forward to the next ',', ';', or closeKind at zero
// nested brace depth, stopping before an unmatched '}' when closeKind != '}'.
func (p *parser) recoverInBlock(closeKind TokenKind) {
	braceDepth := 0
	for {
		k := p.cur().Kind
		if k == TokenEOF {
			return
		}
		if braceDepth == 0 {
			if k == closeKind {
				return
			}
			if closeKind != TokenRBrace && k == TokenRBrace {
				return
			}
			if k == TokenComma || k == TokenSemi {
				p.advance()
				return
			}
		}
		switch k {
		case TokenLBrace:
			braceDepth++
		case TokenRBrace:
			if braceDepth > 0 {
				braceDepth--
			}
		}
		p.advance()
	}
}

func (p *parser) parseFile() *File {
	file := &File{}
	firstNonEmpty := true

	for p.cur().Kind != TokenEOF {
		if p.cur().Kind == TokenSemi {
			firstNonEmpty = false
			p.advance()
			continue
		}
		startPos := p.pos
		tok := p.cur()
		if tok.Kind != TokenIdent {
			p.errorUnexpected(tok, "top-level declaration")
			p.recoverStmt()
			if p.cur().Kind == TokenRBrace {
				p.advance()
			}
			firstNonEmpty = false
			continue
		}

		switch tok.Text {
		case "syntax", "edition":
			if !firstNonEmpty {
				p.errorf(tok.Pos, "%q statement must be the first statement in the file", tok.Text)
				p.recoverStmt()
				break
			}
			file.Syntax = p.parseSyntax()
		case "import":
			if imp := p.parseImport(); imp != nil {
				file.Imports = append(file.Imports, imp)
			}
		case "package":
			if pkg := p.parsePackage(); pkg != nil {
				file.Packages = append(file.Packages, pkg)
			}
		case "option":
			if opt := p.parseOptionStmt(); opt != nil {
				file.Options = append(file.Options, opt)
			}
		case "message":
			if msg := p.parseMessage(0); msg != nil {
				file.Messages = append(file.Messages, msg)
				file.Decls = append(file.Decls, msg)
			}
		case "enum":
			if en := p.parseEnum(); en != nil {
				file.Enums = append(file.Enums, en)
				file.Decls = append(file.Decls, en)
			}
		case "extend":
			if ext := p.parseExtend(); ext != nil {
				file.Extends = append(file.Extends, ext)
				file.Decls = append(file.Decls, ext)
			}
		case "service":
			if svc := p.parseService(); svc != nil {
				file.Services = append(file.Services, svc)
				file.Decls = append(file.Decls, svc)
			}
		default:
			p.errorUnexpected(tok, "top-level declaration")
			p.recoverStmt()
		}
		firstNonEmpty = false
		if p.pos == startPos && p.cur().Kind != TokenEOF {
			p.advance()
		}
	}

	file.HeadComment = p.stream.ConsumeHeadComment(p.stream.At(0).Index)
	return file
}

func (p *parser) parseSyntax() *SyntaxDecl {
	kw := p.advance()
	decl := &SyntaxDecl{
		IsEdition: kw.Text == "edition",
		Pos:       kw.Pos,
	}
	if _, ok := p.expect(TokenEqual); !ok {
		p.recoverStmt()
		return decl
	}
	lit, ok := p.parseStringLit()
	if !ok {
		p.recoverStmt()
		return decl
	}
	decl.Value = lit
	if _, ok := p.expect(TokenSemi); !ok {
		p.recoverStmt()
	}
	return decl
}

func (p *parser) parseImport() *ImportDecl {
	kw := p.advance()
	decl := &ImportDecl{Pos: kw.Pos}
	switch p.cur().Kind {
	case TokenIdent:
		switch p.cur().Text {
		case "public":
			decl.Kind = ImportPublic
			p.advance()
		case "weak":
			decl.Kind = ImportWeak
			p.advance()
		}
	}
	lit, ok := p.parseStringLit()
	if !ok {
		p.recoverStmt()
		return nil
	}
	decl.Path = lit
	if _, ok := p.expect(TokenSemi); !ok {
		p.recoverStmt()
	}
	return decl
}

func (p *parser) parsePackage() *PackageDecl {
	kw := p.advance()
	name, _, ok := p.parseFullIdent("package name")
	if !ok {
		p.recoverStmt()
		return nil
	}
	if _, ok := p.expect(TokenSemi); !ok {
		p.recoverStmt()
	}
	return &PackageDecl{Name: name, Pos: kw.Pos}
}

func (p *parser) parseOptionStmt() *OptionDecl {
	kw := p.advance()
	name, ok := p.parseOptionName()
	if !ok {
		p.recoverStmt()
		return nil
	}
	if _, ok := p.expect(TokenEqual); !ok {
		p.recoverStmt()
		return nil
	}
	val, ok := p.parseOptionValue(0)
	if !ok {
		p.recoverStmt()
		return nil
	}
	if _, ok := p.expect(TokenSemi); !ok {
		p.recoverStmt()
	}
	return &OptionDecl{Name: name, Value: val, Pos: kw.Pos}
}

func (p *parser) parseOptionName() (string, bool) {
	var b strings.Builder
	for i := 0; ; i++ {
		if i > 0 {
			b.WriteByte('.')
		}
		if p.cur().Kind == TokenLParen {
			p.advance()
			b.WriteByte('(')
			if p.cur().Kind == TokenDot {
				p.advance()
				b.WriteByte('.')
			}
			ident, _, ok := p.parseFullIdent("extension option name")
			if !ok {
				return "", false
			}
			b.WriteString(ident)
			if _, ok := p.expect(TokenRParen); !ok {
				return "", false
			}
			b.WriteByte(')')
		} else {
			tok, ok := p.expectIdent("option name")
			if !ok {
				return "", false
			}
			b.WriteString(tok.Text)
		}
		if p.cur().Kind != TokenDot {
			break
		}
		p.advance()
	}
	return b.String(), true
}

func (p *parser) parseOptionValue(depth int) (*Value, bool) {
	if depth > maxParseDepth {
		p.errorf(p.cur().Pos, "maximum nesting depth exceeded")
		return nil, false
	}
	tok := p.cur()
	switch tok.Kind {
	case TokenString:
		lit, ok := p.parseStringLit()
		if !ok {
			return nil, false
		}
		return &Value{
			Kind:        ValueString,
			Raw:         lit.Raw(),
			StringParts: lit.Parts,
			Pos:         lit.Pos,
		}, true
	case TokenLBrace:
		return p.parseBlock(TokenRBrace, depth+1)
	case TokenPlus, TokenMinus:
		return p.parseSignedNumber()
	case TokenInt:
		p.advance()
		return &Value{Kind: ValueInt, Raw: tok.Text, Pos: tok.Pos}, true
	case TokenFloat:
		p.advance()
		return &Value{Kind: ValueFloat, Raw: tok.Text, Pos: tok.Pos}, true
	case TokenIdent:
		if p.peek(1).Kind != TokenDot {
			switch tok.Text {
			case "true", "false":
				p.advance()
				return &Value{Kind: ValueBool, Raw: tok.Text, Pos: tok.Pos}, true
			case "inf", "nan":
				p.advance()
				return &Value{Kind: ValueFloat, Raw: tok.Text, Pos: tok.Pos}, true
			}
		}
		ident, pos, ok := p.parseFullIdent("identifier")
		if !ok {
			return nil, false
		}
		return &Value{Kind: ValueIdent, Raw: ident, Pos: pos}, true
	default:
		p.errorUnexpected(tok, "option value")
		return nil, false
	}
}

func (p *parser) parseSignedNumber() (*Value, bool) {
	sign := p.advance()
	tok := p.cur()
	switch tok.Kind {
	case TokenInt:
		p.advance()
		return &Value{Kind: ValueInt, Raw: sign.Text + tok.Text, Pos: sign.Pos}, true
	case TokenFloat:
		p.advance()
		return &Value{Kind: ValueFloat, Raw: sign.Text + tok.Text, Pos: sign.Pos}, true
	case TokenIdent:
		if (tok.Text == "inf" || tok.Text == "nan") && p.peek(1).Kind != TokenDot {
			p.advance()
			return &Value{Kind: ValueFloat, Raw: sign.Text + tok.Text, Pos: sign.Pos}, true
		}
	}
	p.errorUnexpected(tok, "number")
	return nil, false
}

func (p *parser) parseBlock(closeKind TokenKind, depth int) (*Value, bool) {
	if depth > maxParseDepth {
		p.errorf(p.cur().Pos, "maximum nesting depth exceeded")
		return nil, false
	}
	openTok := p.advance() // '{' or '<'
	val := &Value{Kind: ValueBlock, Pos: openTok.Pos}

	for p.cur().Kind != closeKind && p.cur().Kind != TokenEOF {
		if closeKind != TokenRBrace && p.cur().Kind == TokenRBrace {
			break
		}
		startPos := p.pos
		field, ok := p.parseBlockField(depth)
		if ok {
			val.Fields = append(val.Fields, field)
		} else {
			p.recoverInBlock(closeKind)
			if p.pos == startPos && p.cur().Kind != TokenEOF {
				p.advance()
			}
			continue
		}
		if p.cur().Kind == TokenComma || p.cur().Kind == TokenSemi {
			p.advance()
		}
	}

	closeTok, ok := p.expect(closeKind)
	if !ok {
		return nil, false
	}
	val.Raw = string(p.src[openTok.Pos.Offset:closeTok.EndPos.Offset])
	return val, true
}

func (p *parser) parseBlockField(depth int) (*BlockField, bool) {
	field := &BlockField{Pos: p.cur().Pos}
	if p.cur().Kind == TokenLBracket {
		p.advance()
		name, ok := p.parseBlockExtensionKey()
		if !ok {
			return nil, false
		}
		if _, ok := p.expect(TokenRBracket); !ok {
			return nil, false
		}
		field.Name = name
		field.IsExtension = true
	} else {
		tok, ok := p.expectIdent("option block field name")
		if !ok {
			return nil, false
		}
		field.Name = tok.Text
	}

	var val *Value
	var ok bool
	switch p.cur().Kind {
	case TokenColon:
		p.advance()
		val, ok = p.parseBlockValue(depth)
	case TokenLBrace, TokenLAngle, TokenLBracket:
		val, ok = p.parseBlockValue(depth)
	default:
		p.errorUnexpected(p.cur(), "':', '{', or '['")
		return nil, false
	}
	if !ok {
		return nil, false
	}
	field.Value = val
	return field, true
}

func (p *parser) parseBlockExtensionKey() (string, bool) {
	firstTok, ok := p.expectIdent("extension name")
	if !ok {
		return "", false
	}
	var b strings.Builder
	b.WriteString(firstTok.Text)
	var hyphenTok Token
	for p.cur().Kind == TokenDot || p.cur().Kind == TokenMinus {
		sep := p.advance()
		if sep.Kind == TokenMinus && hyphenTok.Kind == 0 {
			hyphenTok = sep
		}
		b.WriteString(sep.Text)
		seg, ok := p.expectIdent("extension name")
		if !ok {
			return "", false
		}
		b.WriteString(seg.Text)
	}
	if p.cur().Kind != TokenSlash {
		if hyphenTok.Kind != 0 {
			p.errorUnexpected(hyphenTok, "']' or '.'")
			return "", false
		}
		return b.String(), true
	}
	for p.cur().Kind == TokenSlash {
		p.advance()
		b.WriteByte('/')
		seg, _, ok := p.parseFullIdent("type URL")
		if !ok {
			return "", false
		}
		b.WriteString(seg)
	}
	return b.String(), true
}

func (p *parser) parseBlockValue(depth int) (*Value, bool) {
	if depth > maxParseDepth {
		p.errorf(p.cur().Pos, "maximum nesting depth exceeded")
		return nil, false
	}
	if p.cur().Kind == TokenLBracket {
		openTok := p.advance()
		val := &Value{Kind: ValueList, Pos: openTok.Pos}
		for p.cur().Kind != TokenRBracket && p.cur().Kind != TokenRBrace && p.cur().Kind != TokenEOF {
			startPos := p.pos
			elem, ok := p.parseBlockElemValue(depth + 1)
			if ok {
				val.List = append(val.List, elem)
			} else {
				p.recoverInBlock(TokenRBracket)
				if p.pos == startPos && p.cur().Kind != TokenEOF && p.cur().Kind != TokenRBrace {
					p.advance()
				}
				continue
			}
			if p.cur().Kind == TokenComma {
				p.advance()
			} else if p.cur().Kind != TokenRBracket {
				p.errorUnexpected(p.cur(), "',' or ']'")
				p.recoverInBlock(TokenRBracket)
			}
		}
		closeTok, ok := p.expect(TokenRBracket)
		if !ok {
			return nil, false
		}
		val.Raw = string(p.src[openTok.Pos.Offset:closeTok.EndPos.Offset])
		return val, true
	}
	return p.parseBlockElemValue(depth)
}

func (p *parser) parseBlockElemValue(depth int) (*Value, bool) {
	if depth > maxParseDepth {
		p.errorf(p.cur().Pos, "maximum nesting depth exceeded")
		return nil, false
	}
	if p.cur().Kind == TokenLAngle {
		return p.parseBlock(TokenRAngle, depth+1)
	}
	return p.parseOptionValue(depth)
}

func (p *parser) skipMessageBody() {
	for p.cur().Kind != TokenEOF && p.cur().Kind != TokenLBrace && p.cur().Kind != TokenSemi && p.cur().Kind != TokenRBrace {
		p.advance()
	}
	if p.cur().Kind == TokenSemi {
		p.advance()
		return
	}
	if p.cur().Kind != TokenLBrace {
		return
	}
	braces := 0
	for p.cur().Kind != TokenEOF {
		switch p.cur().Kind {
		case TokenLBrace:
			braces++
		case TokenRBrace:
			braces--
			p.advance()
			if braces <= 0 {
				return
			}
			continue
		}
		p.advance()
	}
}

func (p *parser) parseMessage(depth int) *MessageDecl {
	if depth > maxParseDepth {
		p.errorf(p.cur().Pos, "maximum nesting depth exceeded")
		p.skipMessageBody()
		return nil
	}
	kw := p.advance() // message
	msg := &MessageDecl{
		HeadComment: p.stream.ConsumeHeadComment(kw.Index),
		Pos:         kw.Pos,
	}
	nameTok, ok := p.expectIdent("message name")
	if !ok {
		p.recoverStmt()
		return nil
	}
	msg.Name = nameTok.Text
	if _, ok := p.expect(TokenLBrace); !ok {
		p.recoverStmt()
		return nil
	}

	for p.cur().Kind != TokenRBrace && p.cur().Kind != TokenEOF {
		startPos := p.pos
		tok := p.cur()
		if tok.Kind == TokenSemi {
			p.advance()
			continue
		}
		if tok.Kind != TokenIdent && tok.Kind != TokenDot {
			p.errorUnexpected(tok, "message element")
			p.recoverStmt()
			continue
		}

		switch {
		case tok.Kind == TokenIdent && tok.Text == "message":
			if nested := p.parseMessage(depth + 1); nested != nil {
				msg.Elements = append(msg.Elements, nested)
			}
		case tok.Kind == TokenIdent && tok.Text == "enum":
			if en := p.parseEnum(); en != nil {
				msg.Elements = append(msg.Elements, en)
			}
		case tok.Kind == TokenIdent && tok.Text == "oneof":
			if of := p.parseOneof(); of != nil {
				msg.Elements = append(msg.Elements, of)
			}
		case tok.Kind == TokenIdent && tok.Text == "option":
			if opt := p.parseOptionStmt(); opt != nil {
				msg.Elements = append(msg.Elements, opt)
			}
		case tok.Kind == TokenIdent && tok.Text == "reserved":
			if res := p.parseReserved(); res != nil {
				msg.Elements = append(msg.Elements, res)
			}
		case tok.Kind == TokenIdent && tok.Text == "extensions":
			if ext := p.parseExtensions(); ext != nil {
				msg.Elements = append(msg.Elements, ext)
			}
		case tok.Kind == TokenIdent && tok.Text == "extend":
			if ext := p.parseExtend(); ext != nil {
				msg.Elements = append(msg.Elements, ext)
			}
		case tok.Kind == TokenIdent && tok.Text == "map" && p.peek(1).Kind == TokenLAngle:
			if mf := p.parseMapField(); mf != nil {
				msg.Elements = append(msg.Elements, mf)
			}
		default:
			if f := p.parseField(true); f != nil {
				msg.Elements = append(msg.Elements, f)
			}
		}

		if p.pos == startPos && p.cur().Kind != TokenEOF {
			p.advance()
		}
	}

	if rbrace, ok := p.expect(TokenRBrace); ok {
		msg.InlineComment = p.stream.ConsumeInlineComment(rbrace.Index)
	}
	return msg
}

func (p *parser) parseField(allowLabel bool) *FieldDecl {
	startTok := p.cur()
	field := &FieldDecl{
		HeadComment: p.stream.ConsumeHeadComment(startTok.Index),
		Pos:         startTok.Pos,
	}

	if startTok.Kind == TokenIdent && isFieldLabel(startTok.Text) {
		// Disambiguate label from a message type named "optional"/"required"/"repeated"
		// (e.g., "optional foo = 1;"), but treat "<label> <scalarType> =" as a label
		// missing its field name (e.g., "optional int32 = 1;").
		if p.peek(1).Kind != TokenIdent || p.peek(2).Kind != TokenEqual || isScalarType(p.peek(1).Text) {
			if !allowLabel {
				p.errorf(startTok.Pos, "field label %q is not allowed here", startTok.Text)
				p.recoverStmt()
				return nil
			}
			lbl := p.advance()
			field.Label = lbl.Text
			field.LabelPos = lbl.Pos
		}
	}

	typName, typPos, ok := p.parseTypeName()
	if !ok {
		p.recoverStmt()
		return nil
	}
	field.Type = typName
	field.TypePos = typPos

	nameTok, ok := p.expectIdent("field name")
	if !ok {
		p.recoverStmt()
		return nil
	}
	field.Name = nameTok.Text

	if _, ok := p.expect(TokenEqual); !ok {
		p.recoverStmt()
		return nil
	}
	numTok, ok := p.expect(TokenInt)
	if !ok {
		p.recoverStmt()
		return nil
	}
	field.Number = numTok.Text
	field.NumberPos = numTok.Pos

	if p.cur().Kind == TokenLBracket {
		opts, ok := p.parseBracketOptions()
		if !ok {
			p.recoverStmt()
			return nil
		}
		field.Options = opts
	}

	if semi, ok := p.expect(TokenSemi); ok {
		field.InlineComment = p.stream.ConsumeInlineComment(semi.Index)
	} else {
		p.recoverStmt()
	}
	return field
}

func (p *parser) parseMapField() *MapFieldDecl {
	kw := p.advance() // map
	mf := &MapFieldDecl{
		HeadComment: p.stream.ConsumeHeadComment(kw.Index),
		Pos:         kw.Pos,
	}
	if _, ok := p.expect(TokenLAngle); !ok {
		p.recoverStmt()
		return nil
	}
	keyTok, ok := p.expectIdent("map key type")
	if !ok {
		p.recoverStmt()
		return nil
	}
	if !validMapKeyTypes[keyTok.Text] {
		p.errorf(keyTok.Pos, "invalid map key type %q", keyTok.Text)
		p.recoverStmt()
		return nil
	}
	mf.KeyType = keyTok.Text
	mf.KeyPos = keyTok.Pos

	if _, ok := p.expect(TokenComma); !ok {
		p.recoverStmt()
		return nil
	}
	valType, valPos, ok := p.parseTypeName()
	if !ok {
		p.recoverStmt()
		return nil
	}
	mf.ValType = valType
	mf.ValPos = valPos

	if _, ok := p.expect(TokenRAngle); !ok {
		p.recoverStmt()
		return nil
	}
	nameTok, ok := p.expectIdent("map field name")
	if !ok {
		p.recoverStmt()
		return nil
	}
	mf.Name = nameTok.Text

	if _, ok := p.expect(TokenEqual); !ok {
		p.recoverStmt()
		return nil
	}
	numTok, ok := p.expect(TokenInt)
	if !ok {
		p.recoverStmt()
		return nil
	}
	mf.Number = numTok.Text
	mf.NumberPos = numTok.Pos

	if p.cur().Kind == TokenLBracket {
		opts, ok := p.parseBracketOptions()
		if !ok {
			p.recoverStmt()
			return nil
		}
		mf.Options = opts
	}

	if semi, ok := p.expect(TokenSemi); ok {
		mf.InlineComment = p.stream.ConsumeInlineComment(semi.Index)
	} else {
		p.recoverStmt()
	}
	return mf
}

func (p *parser) parseOneof() *OneofDecl {
	kw := p.advance() // oneof
	of := &OneofDecl{
		Pos: kw.Pos,
	}
	nameTok, ok := p.expectIdent("oneof name")
	if !ok {
		p.recoverStmt()
		return nil
	}
	of.Name = nameTok.Text
	if _, ok := p.expect(TokenLBrace); !ok {
		p.recoverStmt()
		return nil
	}

	for p.cur().Kind != TokenRBrace && p.cur().Kind != TokenEOF {
		startPos := p.pos
		tok := p.cur()
		if tok.Kind == TokenSemi {
			p.advance()
			continue
		}
		if tok.Kind == TokenIdent && tok.Text == "option" {
			if opt := p.parseOptionStmt(); opt != nil {
				of.Options = append(of.Options, opt)
			}
		} else {
			if f := p.parseField(false); f != nil {
				of.Fields = append(of.Fields, f)
			}
		}
		if p.pos == startPos && p.cur().Kind != TokenEOF {
			p.advance()
		}
	}

	_, _ = p.expect(TokenRBrace)
	return of
}

func (p *parser) parseEnum() *EnumDecl {
	kw := p.advance() // enum
	en := &EnumDecl{
		HeadComment: p.stream.ConsumeHeadComment(kw.Index),
		Pos:         kw.Pos,
	}
	nameTok, ok := p.expectIdent("enum name")
	if !ok {
		p.recoverStmt()
		return nil
	}
	en.Name = nameTok.Text
	if _, ok := p.expect(TokenLBrace); !ok {
		p.recoverStmt()
		return nil
	}

	for p.cur().Kind != TokenRBrace && p.cur().Kind != TokenEOF {
		startPos := p.pos
		tok := p.cur()
		if tok.Kind == TokenSemi {
			p.advance()
			continue
		}
		if tok.Kind != TokenIdent {
			p.errorUnexpected(tok, "enum element")
			p.recoverStmt()
			continue
		}
		switch tok.Text {
		case "option":
			if opt := p.parseOptionStmt(); opt != nil {
				en.Options = append(en.Options, opt)
			}
		case "reserved":
			if res := p.parseReserved(); res != nil {
				en.Reserved = append(en.Reserved, res)
			}
		default:
			if val := p.parseEnumValue(); val != nil {
				en.Values = append(en.Values, val)
			}
		}
		if p.pos == startPos && p.cur().Kind != TokenEOF {
			p.advance()
		}
	}

	if rbrace, ok := p.expect(TokenRBrace); ok {
		en.InlineComment = p.stream.ConsumeInlineComment(rbrace.Index)
	}
	return en
}

func (p *parser) parseEnumValue() *EnumValueDecl {
	nameTok, ok := p.expectIdent("enum value name")
	if !ok {
		p.recoverStmt()
		return nil
	}
	ev := &EnumValueDecl{
		Name:        nameTok.Text,
		HeadComment: p.stream.ConsumeHeadComment(nameTok.Index),
		Pos:         nameTok.Pos,
	}
	if _, ok := p.expect(TokenEqual); !ok {
		p.recoverStmt()
		return nil
	}
	num, numPos, ok := p.parseSignedInt()
	if !ok {
		p.recoverStmt()
		return nil
	}
	ev.Number = num
	ev.NumberPos = numPos

	if p.cur().Kind == TokenLBracket {
		opts, ok := p.parseBracketOptions()
		if !ok {
			p.recoverStmt()
			return nil
		}
		ev.Options = opts
	}

	if semi, ok := p.expect(TokenSemi); ok {
		ev.InlineComment = p.stream.ConsumeInlineComment(semi.Index)
	} else {
		p.recoverStmt()
	}
	return ev
}

func (p *parser) parseReserved() *ReservedDecl {
	kw := p.advance() // reserved
	res := &ReservedDecl{Pos: kw.Pos}

	switch p.cur().Kind {
	case TokenInt, TokenMinus:
		ranges, ok := p.parseRanges()
		if !ok {
			p.recoverStmt()
			return nil
		}
		res.Ranges = ranges
	case TokenString:
		for {
			lit, ok := p.parseStringLit()
			if !ok {
				p.recoverStmt()
				return nil
			}
			res.Names = append(res.Names, ReservedName{
				StringLit: &lit,
				Pos:       lit.Pos,
			})
			if p.cur().Kind != TokenComma {
				break
			}
			p.advance()
		}
	case TokenIdent:
		for {
			tok, ok := p.expectIdent("reserved identifier")
			if !ok {
				p.recoverStmt()
				return nil
			}
			res.Names = append(res.Names, ReservedName{
				Ident: tok.Text,
				Pos:   tok.Pos,
			})
			if p.cur().Kind != TokenComma {
				break
			}
			p.advance()
		}
	default:
		p.errorUnexpected(p.cur(), "reserved range or name")
		p.recoverStmt()
		return nil
	}

	if _, ok := p.expect(TokenSemi); !ok {
		p.recoverStmt()
	}
	return res
}

func (p *parser) parseExtensions() *ExtensionsDecl {
	kw := p.advance() // extensions
	decl := &ExtensionsDecl{Pos: kw.Pos}
	ranges, ok := p.parseRanges()
	if !ok {
		p.recoverStmt()
		return nil
	}
	decl.Ranges = ranges
	if p.cur().Kind == TokenLBracket {
		opts, ok := p.parseBracketOptions()
		if !ok {
			p.recoverStmt()
			return nil
		}
		decl.Options = opts
	}
	if _, ok := p.expect(TokenSemi); !ok {
		p.recoverStmt()
	}
	return decl
}

func (p *parser) parseRanges() ([]*RangeDecl, bool) {
	var ranges []*RangeDecl
	for {
		r, ok := p.parseRange()
		if !ok {
			return nil, false
		}
		ranges = append(ranges, r)
		if p.cur().Kind != TokenComma {
			break
		}
		p.advance()
	}
	return ranges, true
}

func (p *parser) parseRange() (*RangeDecl, bool) {
	start, startPos, ok := p.parseSignedInt()
	if !ok {
		return nil, false
	}
	r := &RangeDecl{
		Start:    start,
		StartPos: startPos,
		Pos:      startPos,
	}
	if p.cur().Kind == TokenIdent && p.cur().Text == "to" {
		p.advance()
		r.IsRange = true
		if p.cur().Kind == TokenIdent && p.cur().Text == "max" {
			tok := p.advance()
			r.IsMax = true
			r.EndPos = tok.Pos
		} else {
			end, endPos, ok := p.parseSignedInt()
			if !ok {
				return nil, false
			}
			r.End = end
			r.EndPos = endPos
		}
	}
	return r, true
}

func (p *parser) parseExtend() *ExtendDecl {
	kw := p.advance() // extend
	target, targetPos, ok := p.parseTypeName()
	if !ok {
		p.recoverStmt()
		return nil
	}
	ext := &ExtendDecl{
		Extendee:    target,
		ExtendeePos: targetPos,
		Pos:         kw.Pos,
	}
	if _, ok := p.expect(TokenLBrace); !ok {
		p.recoverStmt()
		return nil
	}
	for p.cur().Kind != TokenRBrace && p.cur().Kind != TokenEOF {
		startPos := p.pos
		if p.cur().Kind == TokenSemi {
			p.advance()
			continue
		}
		if f := p.parseField(true); f != nil {
			ext.Fields = append(ext.Fields, f)
		}
		if p.pos == startPos && p.cur().Kind != TokenEOF {
			p.advance()
		}
	}
	_, _ = p.expect(TokenRBrace)
	return ext
}

func (p *parser) parseService() *ServiceDecl {
	kw := p.advance() // service
	svc := &ServiceDecl{
		HeadComment: p.stream.ConsumeHeadComment(kw.Index),
		Pos:         kw.Pos,
	}
	nameTok, ok := p.expectIdent("service name")
	if !ok {
		p.recoverStmt()
		return nil
	}
	svc.Name = nameTok.Text
	if _, ok := p.expect(TokenLBrace); !ok {
		p.recoverStmt()
		return nil
	}

	for p.cur().Kind != TokenRBrace && p.cur().Kind != TokenEOF {
		startPos := p.pos
		tok := p.cur()
		if tok.Kind == TokenSemi {
			p.advance()
			continue
		}
		if tok.Kind != TokenIdent {
			p.errorUnexpected(tok, "service element")
			p.recoverStmt()
			continue
		}
		switch tok.Text {
		case "option":
			if opt := p.parseOptionStmt(); opt != nil {
				svc.Options = append(svc.Options, opt)
			}
		case "rpc":
			if m := p.parseMethod(); m != nil {
				svc.Methods = append(svc.Methods, m)
			}
		default:
			p.errorUnexpected(tok, "'rpc' or 'option'")
			p.recoverStmt()
		}
		if p.pos == startPos && p.cur().Kind != TokenEOF {
			p.advance()
		}
	}

	if rbrace, ok := p.expect(TokenRBrace); ok {
		svc.InlineComment = p.stream.ConsumeInlineComment(rbrace.Index)
	}
	return svc
}

func (p *parser) parseMethod() *MethodDecl {
	kw := p.advance() // rpc
	m := &MethodDecl{
		HeadComment: p.stream.ConsumeHeadComment(kw.Index),
		Pos:         kw.Pos,
	}
	nameTok, ok := p.expectIdent("rpc method name")
	if !ok {
		p.recoverStmt()
		return nil
	}
	m.Name = nameTok.Text

	if _, ok := p.expect(TokenLParen); !ok {
		p.recoverStmt()
		return nil
	}
	if p.cur().Kind == TokenIdent && p.cur().Text == "stream" &&
		(p.peek(1).Kind == TokenIdent || p.peek(1).Kind == TokenDot) {
		p.advance()
		m.RequestStream = true
	}
	reqType, reqPos, ok := p.parseTypeName()
	if !ok {
		p.recoverStmt()
		return nil
	}
	m.RequestType = reqType
	m.RequestPos = reqPos
	if _, ok := p.expect(TokenRParen); !ok {
		p.recoverStmt()
		return nil
	}

	retKw, ok := p.expectIdent("'returns'")
	if !ok || retKw.Text != "returns" {
		if ok {
			p.errorUnexpected(retKw, "'returns'")
		}
		p.recoverStmt()
		return nil
	}

	if _, ok := p.expect(TokenLParen); !ok {
		p.recoverStmt()
		return nil
	}
	if p.cur().Kind == TokenIdent && p.cur().Text == "stream" &&
		(p.peek(1).Kind == TokenIdent || p.peek(1).Kind == TokenDot) {
		p.advance()
		m.ReturnStream = true
	}
	retType, retPos, ok := p.parseTypeName()
	if !ok {
		p.recoverStmt()
		return nil
	}
	m.ReturnType = retType
	m.ReturnPos = retPos
	if _, ok := p.expect(TokenRParen); !ok {
		p.recoverStmt()
		return nil
	}

	if p.cur().Kind == TokenLBrace {
		p.advance()
		for p.cur().Kind != TokenRBrace && p.cur().Kind != TokenEOF {
			startPos := p.pos
			if p.cur().Kind == TokenSemi {
				p.advance()
				continue
			}
			if p.cur().Kind == TokenIdent && p.cur().Text == "option" {
				if opt := p.parseOptionStmt(); opt != nil {
					m.Options = append(m.Options, opt)
				}
			} else {
				p.errorUnexpected(p.cur(), "'option' or '}'")
				p.recoverStmt()
			}
			if p.pos == startPos && p.cur().Kind != TokenEOF {
				p.advance()
			}
		}
		if rbrace, ok := p.expect(TokenRBrace); ok {
			m.InlineComment = p.stream.ConsumeInlineComment(rbrace.Index)
		}
		return m
	}

	if semi, ok := p.expect(TokenSemi); ok {
		m.InlineComment = p.stream.ConsumeInlineComment(semi.Index)
	} else {
		p.recoverStmt()
	}
	return m
}

func (p *parser) parseBracketOptions() ([]*OptionDecl, bool) {
	p.advance() // '['
	var opts []*OptionDecl
	for {
		optPos := p.cur().Pos
		name, ok := p.parseOptionName()
		if !ok {
			return nil, false
		}
		if _, ok := p.expect(TokenEqual); !ok {
			return nil, false
		}
		val, ok := p.parseOptionValue(0)
		if !ok {
			return nil, false
		}
		opts = append(opts, &OptionDecl{Name: name, Value: val, Pos: optPos})
		if p.cur().Kind != TokenComma {
			break
		}
		p.advance()
	}
	if _, ok := p.expect(TokenRBracket); !ok {
		return nil, false
	}
	return opts, true
}

func (p *parser) parseStringLit() (StringLit, bool) {
	tok := p.cur()
	if tok.Kind != TokenString {
		p.errorUnexpected(tok, "string literal")
		return StringLit{}, false
	}
	p.advance()
	parts := []string{tok.Text}
	for p.cur().Kind == TokenString {
		parts = append(parts, p.advance().Text)
	}
	return StringLit{Parts: parts, Pos: tok.Pos}, true
}

func (p *parser) parseFullIdent(what string) (string, Pos, bool) {
	first, ok := p.expectIdent(what)
	if !ok {
		return "", Pos{}, false
	}
	if p.cur().Kind != TokenDot {
		return first.Text, first.Pos, true
	}
	var b strings.Builder
	b.WriteString(first.Text)
	for p.cur().Kind == TokenDot {
		p.advance()
		b.WriteByte('.')
		seg, ok := p.expectIdent(what)
		if !ok {
			return "", Pos{}, false
		}
		b.WriteString(seg.Text)
	}
	return b.String(), first.Pos, true
}

func (p *parser) parseTypeName() (string, Pos, bool) {
	if p.cur().Kind == TokenDot {
		dot := p.advance()
		rest, _, ok := p.parseFullIdent("type name")
		if !ok {
			return "", Pos{}, false
		}
		return "." + rest, dot.Pos, true
	}
	return p.parseFullIdent("type name")
}

func (p *parser) parseSignedInt() (string, Pos, bool) {
	if p.cur().Kind == TokenMinus {
		minus := p.advance()
		num, ok := p.expect(TokenInt)
		if !ok {
			return "", Pos{}, false
		}
		return "-" + num.Text, minus.Pos, true
	}
	num, ok := p.expect(TokenInt)
	if !ok {
		return "", Pos{}, false
	}
	return num.Text, num.Pos, true
}

func isFieldLabel(s string) bool {
	return s == "optional" || s == "required" || s == "repeated"
}
