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
	"strings"
	"unicode/utf8"
)

// TokenStream holds the full token stream (including trivia for comment
// attachment), the filtered non-trivia code token index table, and any lex errors.
type TokenStream struct {
	all      []Token
	code     []int
	consumed []bool
	Errors   []SyntaxError
}

// Len returns the number of non-trivia tokens (including the trailing TokenEOF).
func (s *TokenStream) Len() int {
	return len(s.code)
}

// At returns the i-th non-trivia token, returning TokenEOF when i is out of range.
func (s *TokenStream) At(i int) Token {
	if len(s.code) == 0 {
		return Token{Kind: TokenEOF}
	}
	if i < 0 || i >= len(s.code) {
		return s.all[s.code[len(s.code)-1]]
	}
	return s.all[s.code[i]]
}

// All returns a copy of every token in the stream, including whitespace and comments.
func (s *TokenStream) All() []Token {
	out := make([]Token, len(s.all))
	copy(out, s.all)
	return out
}

// ConsumeHeadComment returns leading comments attached to the token at tokenIndex
// in s.all, walking backward until a non-trivia token, an already-consumed trivia
// token, or a blank line (>1 newline in a whitespace run).
func (s *TokenStream) ConsumeHeadComment(tokenIndex int) string {
	if tokenIndex <= 0 || tokenIndex >= len(s.all) {
		return ""
	}
	var comments []string
	for i := tokenIndex - 1; i >= 0; i-- {
		t := s.all[i]
		if !t.Kind.IsTrivia() || s.consumed[i] {
			break
		}
		if t.Kind == TokenWhitespace && strings.Count(t.Text, "\n") > 1 {
			// A blank line separates unrelated comments from this declaration.
			break
		}
		s.consumed[i] = true
		if t.Kind != TokenWhitespace {
			if text := strings.TrimSpace(t.Text); text != "" {
				comments = append(comments, text)
			}
		}
	}
	if len(comments) == 0 {
		return ""
	}
	for i, j := 0, len(comments)-1; i < j; i, j = i+1, j-1 {
		comments[i], comments[j] = comments[j], comments[i]
	}
	return strings.Join(comments, "\n")
}

// ConsumeInlineComment returns the trailing comment on the same line immediately
// following the token at tokenIndex in s.all.
func (s *TokenStream) ConsumeInlineComment(tokenIndex int) string {
	if tokenIndex < 0 || tokenIndex >= len(s.all)-1 {
		return ""
	}
	for i := tokenIndex + 1; i < len(s.all); i++ {
		t := s.all[i]
		if !t.Kind.IsTrivia() || s.consumed[i] {
			return ""
		}
		if t.Kind == TokenWhitespace {
			if strings.ContainsRune(t.Text, '\n') {
				return ""
			}
			continue
		}
		s.consumed[i] = true
		return strings.TrimSpace(t.Text)
	}
	return ""
}

// Lex tokenizes src according to Protocol Buffers lexical rules.
func Lex(src []byte) *TokenStream {
	l := &lexer{
		src: src,
		pos: Pos{Offset: 0, Line: 1, Column: 1},
	}
	// Strip a leading UTF-8 BOM if present without shifting line/column.
	if len(src) >= 3 && src[0] == 0xEF && src[1] == 0xBB && src[2] == 0xBF {
		l.pos.Offset = 3
	}
	l.scanAll()
	l.stream.consumed = make([]bool, len(l.stream.all))
	return &l.stream
}

type lexer struct {
	src    []byte
	pos    Pos
	stream TokenStream
}

var singleCharTokens = [256]TokenKind{
	';': TokenSemi,
	'=': TokenEqual,
	'(': TokenLParen,
	')': TokenRParen,
	'[': TokenLBracket,
	']': TokenRBracket,
	'{': TokenLBrace,
	'}': TokenRBrace,
	'<': TokenLAngle,
	'>': TokenRAngle,
	'.': TokenDot,
	',': TokenComma,
	':': TokenColon,
	'+': TokenPlus,
	'-': TokenMinus,
	'/': TokenSlash,
}

func (l *lexer) emit(kind TokenKind, start Pos, text string) {
	idx := len(l.stream.all)
	tok := Token{
		Kind:   kind,
		Text:   text,
		Pos:    start,
		EndPos: l.pos,
		Index:  idx,
	}
	l.stream.all = append(l.stream.all, tok)
	if !kind.IsTrivia() {
		l.stream.code = append(l.stream.code, idx)
	}
}

func (l *lexer) emitErrorAt(errPos, tokPos Pos, text, msg string) {
	l.stream.Errors = append(l.stream.Errors, SyntaxError{Pos: errPos, Msg: msg})
	l.emit(TokenError, tokPos, text)
}

func (l *lexer) emitError(start Pos, text, msg string) {
	l.emitErrorAt(start, start, text, msg)
}

func (l *lexer) scanAll() {
	for l.pos.Offset < len(l.src) {
		b := l.src[l.pos.Offset]
		switch {
		case isWhitespace(b):
			l.scanWhitespace()
		case b == '/' && l.peekAt(1) == '/':
			l.scanLineComment()
		case b == '/' && l.peekAt(1) == '*':
			l.scanBlockComment()
		case b == '"' || b == '\'':
			l.scanString()
		case isIdentStart(b):
			l.scanIdent()
		case isDecimalDigit(b) || (b == '.' && isDecimalDigit(l.peekAt(1))):
			l.scanNumber()
		default:
			if kind := singleCharTokens[b]; kind != 0 {
				start := l.pos
				l.advanceASCII(1)
				l.emit(kind, start, string(b))
				continue
			}
			start := l.pos
			text := l.advanceRune()
			l.emitError(start, text, fmt.Sprintf("unexpected character %q", text))
		}
	}
	l.emit(TokenEOF, l.pos, "")
}

func (l *lexer) peekAt(delta int) byte {
	idx := l.pos.Offset + delta
	if idx < 0 || idx >= len(l.src) {
		return 0
	}
	return l.src[idx]
}

func (l *lexer) advanceASCII(n int) {
	for i := 0; i < n; i++ {
		if l.src[l.pos.Offset] == '\n' {
			l.pos.Line++
			l.pos.Column = 1
		} else {
			l.pos.Column++
		}
		l.pos.Offset++
	}
}

func (l *lexer) advanceRune() string {
	start := l.pos.Offset
	b := l.src[start]
	if b < utf8.RuneSelf {
		l.advanceASCII(1)
		return string(l.src[start:l.pos.Offset])
	}
	_, width := utf8.DecodeRune(l.src[start:])
	l.pos.Offset += width
	l.pos.Column++
	return string(l.src[start:l.pos.Offset])
}

func (l *lexer) scanWhitespace() {
	start := l.pos
	for l.pos.Offset < len(l.src) && isWhitespace(l.src[l.pos.Offset]) {
		l.advanceASCII(1)
	}
	l.emit(TokenWhitespace, start, string(l.src[start.Offset:l.pos.Offset]))
}

func (l *lexer) scanLineComment() {
	start := l.pos
	l.advanceASCII(2) // //
	for l.pos.Offset < len(l.src) {
		b := l.src[l.pos.Offset]
		if b == '\n' || b == '\r' {
			break
		}
		l.advanceRune()
	}
	l.emit(TokenLineComment, start, string(l.src[start.Offset:l.pos.Offset]))
}

func (l *lexer) scanBlockComment() {
	start := l.pos
	l.advanceASCII(2) // /*
	for l.pos.Offset < len(l.src) {
		if l.src[l.pos.Offset] == '*' && l.peekAt(1) == '/' {
			l.advanceASCII(2)
			l.emit(TokenBlockComment, start, string(l.src[start.Offset:l.pos.Offset]))
			return
		}
		l.advanceRune()
	}
	l.emitError(start, string(l.src[start.Offset:l.pos.Offset]), "unterminated block comment")
}

func (l *lexer) scanString() {
	start := l.pos
	quote := l.src[l.pos.Offset]
	l.advanceASCII(1)
	hasNUL := false
	nulPos := Pos{}
	for l.pos.Offset < len(l.src) {
		b := l.src[l.pos.Offset]
		if b == quote {
			l.advanceASCII(1)
			text := string(l.src[start.Offset:l.pos.Offset])
			if hasNUL {
				l.emitErrorAt(nulPos, start, text, "null byte in string literal")
			} else {
				l.emit(TokenString, start, text)
			}
			return
		}
		if b == '\n' {
			l.emitError(start, string(l.src[start.Offset:l.pos.Offset]), "unterminated string literal")
			return
		}
		if b == 0 {
			if !hasNUL {
				hasNUL = true
				nulPos = l.pos
			}
			l.advanceASCII(1)
			continue
		}
		if b == '\\' {
			l.advanceASCII(1)
			if l.pos.Offset >= len(l.src) {
				l.emitError(start, string(l.src[start.Offset:l.pos.Offset]), "unterminated string literal")
				return
			}
			next := l.src[l.pos.Offset]
			if next == '\n' || next == '\r' {
				l.emitError(start, string(l.src[start.Offset:l.pos.Offset]), "unterminated string literal")
				return
			}
			if next == 0 {
				if !hasNUL {
					hasNUL = true
					nulPos = l.pos
				}
				l.advanceASCII(1)
				continue
			}
			l.advanceRune()
			continue
		}
		l.advanceRune()
	}
	l.emitError(start, string(l.src[start.Offset:l.pos.Offset]), "unterminated string literal")
}

func (l *lexer) scanIdent() {
	start := l.pos
	l.advanceASCII(1)
	for l.pos.Offset < len(l.src) && isIdentPart(l.src[l.pos.Offset]) {
		l.advanceASCII(1)
	}
	l.emit(TokenIdent, start, string(l.src[start.Offset:l.pos.Offset]))
}

func (l *lexer) scanNumber() {
	start := l.pos
	first := l.src[l.pos.Offset]

	// Leading-dot float: .[0-9]+([eE][+-]?[0-9]+)?
	if first == '.' {
		l.advanceASCII(1) // '.'
		l.scanDecimalDigits()
		if l.pos.Offset < len(l.src) && (l.src[l.pos.Offset] == 'e' || l.src[l.pos.Offset] == 'E') {
			if !l.scanExponent(start) {
				return
			}
		}
		l.finishNumber(start, TokenFloat)
		return
	}

	// Hex integer: 0[xX][0-9A-Fa-f]+
	if first == '0' && (l.peekAt(1) == 'x' || l.peekAt(1) == 'X') {
		l.advanceASCII(2)
		hexStart := l.pos.Offset
		for l.pos.Offset < len(l.src) && isHexDigit(l.src[l.pos.Offset]) {
			l.advanceASCII(1)
		}
		if l.pos.Offset == hexStart {
			l.emitError(start, string(l.src[start.Offset:l.pos.Offset]), "invalid hex integer literal")
			return
		}
		l.finishNumber(start, TokenInt)
		return
	}

	// Decimal/octal digits: 0[0-9]* or [1-9][0-9]*
	l.scanDecimalDigits()
	isFloat := false
	if l.pos.Offset < len(l.src) && l.src[l.pos.Offset] == '.' {
		isFloat = true
		l.advanceASCII(1)
		l.scanDecimalDigits()
	}
	if l.pos.Offset < len(l.src) && (l.src[l.pos.Offset] == 'e' || l.src[l.pos.Offset] == 'E') {
		isFloat = true
		if !l.scanExponent(start) {
			return
		}
	}
	if isFloat {
		l.finishNumber(start, TokenFloat)
		return
	}

	text := string(l.src[start.Offset:l.pos.Offset])
	if l.pos.Offset < len(l.src) && isIdentStart(l.src[l.pos.Offset]) {
		l.emitError(start, text, "space is required between a number and identifier")
		return
	}
	if len(text) > 1 && text[0] == '0' {
		for i := 1; i < len(text); i++ {
			if text[i] == '8' || text[i] == '9' {
				l.emitError(start, text, fmt.Sprintf("invalid octal integer literal %q", text))
				return
			}
		}
	}
	l.emit(TokenInt, start, text)
}

func (l *lexer) scanDecimalDigits() {
	for l.pos.Offset < len(l.src) && isDecimalDigit(l.src[l.pos.Offset]) {
		l.advanceASCII(1)
	}
}

func (l *lexer) scanExponent(start Pos) bool {
	l.advanceASCII(1) // e or E
	if l.pos.Offset < len(l.src) && (l.src[l.pos.Offset] == '+' || l.src[l.pos.Offset] == '-') {
		l.advanceASCII(1)
	}
	expStart := l.pos.Offset
	l.scanDecimalDigits()
	if l.pos.Offset == expStart {
		l.emitError(start, string(l.src[start.Offset:l.pos.Offset]), "invalid float exponent")
		return false
	}
	return true
}

func (l *lexer) finishNumber(start Pos, kind TokenKind) {
	text := string(l.src[start.Offset:l.pos.Offset])
	if l.pos.Offset < len(l.src) && isIdentStart(l.src[l.pos.Offset]) {
		l.emitError(start, text, "space is required between a number and identifier")
		return
	}
	l.emit(kind, start, text)
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n' || b == '\f' || b == '\v'
}

func isIdentStart(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '_'
}

func isIdentPart(b byte) bool {
	return isIdentStart(b) || isDecimalDigit(b)
}

func isDecimalDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isHexDigit(b byte) bool {
	return isDecimalDigit(b) || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}
