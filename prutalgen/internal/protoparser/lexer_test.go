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
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func collectCodeTokens(s *TokenStream) []Token {
	out := make([]Token, s.Len())
	for i := 0; i < s.Len(); i++ {
		out[i] = s.At(i)
	}
	return out
}

func TestLexTokensAndPositions(t *testing.T) {
	src := "\xef\xbb\xbfsyntax = \"proto3\";\r\n" +
		"// cmt\r\n" +
		"message /* 中文 */ M {\n" +
		"  double x = .5e-2 + -1. - 08.5e+1 / 0x1A + 077;\n" +
		"  string s = 'a\\'b\\\"c\\\\d';\n" +
		"  repeated<int32, bool> m = [1, 2]: (a);\n" +
		"}\n"
	s := Lex([]byte(src))
	assert.Equal(t, 0, len(s.Errors))

	toks := collectCodeTokens(s)
	want := []struct {
		kind TokenKind
		text string
		line int
		col  int
	}{
		{TokenIdent, "syntax", 1, 1},
		{TokenEqual, "=", 1, 8},
		{TokenString, `"proto3"`, 1, 10},
		{TokenSemi, ";", 1, 18},
		{TokenIdent, "message", 3, 1},
		{TokenIdent, "M", 3, 18},
		{TokenLBrace, "{", 3, 20},
		{TokenIdent, "double", 4, 3},
		{TokenIdent, "x", 4, 10},
		{TokenEqual, "=", 4, 12},
		{TokenFloat, ".5e-2", 4, 14},
		{TokenPlus, "+", 4, 20},
		{TokenMinus, "-", 4, 22},
		{TokenFloat, "1.", 4, 23},
		{TokenMinus, "-", 4, 26},
		{TokenFloat, "08.5e+1", 4, 28},
		{TokenSlash, "/", 4, 36},
		{TokenInt, "0x1A", 4, 38},
		{TokenPlus, "+", 4, 43},
		{TokenInt, "077", 4, 45},
		{TokenSemi, ";", 4, 48},
		{TokenIdent, "string", 5, 3},
		{TokenIdent, "s", 5, 10},
		{TokenEqual, "=", 5, 12},
		{TokenString, `'a\'b\"c\\d'`, 5, 14},
		{TokenSemi, ";", 5, 26},
		{TokenIdent, "repeated", 6, 3},
		{TokenLAngle, "<", 6, 11},
		{TokenIdent, "int32", 6, 12},
		{TokenComma, ",", 6, 17},
		{TokenIdent, "bool", 6, 19},
		{TokenRAngle, ">", 6, 23},
		{TokenIdent, "m", 6, 25},
		{TokenEqual, "=", 6, 27},
		{TokenLBracket, "[", 6, 29},
		{TokenInt, "1", 6, 30},
		{TokenComma, ",", 6, 31},
		{TokenInt, "2", 6, 33},
		{TokenRBracket, "]", 6, 34},
		{TokenColon, ":", 6, 35},
		{TokenLParen, "(", 6, 37},
		{TokenIdent, "a", 6, 38},
		{TokenRParen, ")", 6, 39},
		{TokenSemi, ";", 6, 40},
		{TokenRBrace, "}", 7, 1},
		{TokenEOF, "", 8, 1},
	}

	assert.Equal(t, len(want), len(toks))
	for i, w := range want {
		assert.Equal(t, w.kind, toks[i].Kind, i)
		assert.Equal(t, w.text, toks[i].Text, i)
		assert.Equal(t, w.line, toks[i].Pos.Line, i)
		assert.Equal(t, w.col, toks[i].Pos.Column, i)
	}
	assert.Equal(t, TokenEOF, s.At(-1).Kind)
	assert.Equal(t, TokenEOF, s.At(len(toks)+10).Kind)
	assert.True(t, len(s.All()) > len(toks))

	// Bare CR inside a string literal is allowed; EOF checks exercise peekAt bounds.
	s = Lex([]byte("option s = \"a\rb\"; /"))
	assert.Equal(t, 0, len(s.Errors))
	assert.Equal(t, "\"a\rb\"", s.At(3).Text)
	assert.Equal(t, TokenSlash, s.At(5).Kind)

	// Null-byte error token starts at opening quote while error Pos points at the NUL.
	s = Lex([]byte("x \"a\x00b\""))
	assert.Equal(t, 1, len(s.Errors))
	assert.Equal(t, 5, s.Errors[0].Pos.Column)
	assert.Equal(t, TokenError, s.At(1).Kind)
	assert.Equal(t, 3, s.At(1).Pos.Column)
	assert.Equal(t, 2, s.At(1).Pos.Offset)
}

func TestLexErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"invalid octal 08", "08", "invalid octal integer literal"},
		{"invalid octal 079", "079", "invalid octal integer literal"},
		{"empty hex", "0x", "invalid hex integer literal"},
		{"empty hex before ident", "0xz", "invalid hex integer literal"},
		{"adjacent int ident", "reserved 2to 3;", "space is required between a number and identifier"},
		{"adjacent hex ident", "0x10g", "space is required between a number and identifier"},
		{"adjacent float ident", "1.5foo", "space is required between a number and identifier"},
		{"adjacent leading-dot float ident", ".5foo", "space is required between a number and identifier"},
		{"adjacent trailing-dot float ident", "1.foo", "space is required between a number and identifier"},
		{"bad exponent", "1e+", "invalid float exponent"},
		{"bad dot exponent", ".5e", "invalid float exponent"},
		{"unterminated string", "\"abc", "unterminated string literal"},
		{"newline in string", "\"abc\ndef\"", "unterminated string literal"},
		{"escaped newline in string", "\"abc\\\ndef\"", "unterminated string literal"},
		{"escaped CR in string", "\"abc\\\r\ndef\"", "unterminated string literal"},
		{"trailing backslash in string", "\"abc\\", "unterminated string literal"},
		{"null byte in string", "\"a\x00b\"", "null byte in string literal"},
		{"escaped null byte in string", "\"a\\\x00b\"", "null byte in string literal"},
		{"unterminated block comment", "/* hello\nworld", "unterminated block comment"},
		{"unexpected char", "@", "unexpected character"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := Lex([]byte(tc.src))
			assert.True(t, len(s.Errors) > 0, tc.name)
			assert.StringContains(t, s.Errors[0].Msg, tc.want)
			assert.Equal(t, TokenEOF, s.At(s.Len()-1).Kind)
		})
	}
}

func TestLexCommentsAndLineRecovery(t *testing.T) {
	src := `// comment line1
message Hello {
// comment line2
string a = 1; // comment line3
// comment line4
string b = 2; /*
comment line5
line5+
line5++
*/
// comment line6
// line7
// line8
string c = 3;
// comment line9

/*
comment line10
line11
line12
*/
string d = 4;
}
`
	s := Lex([]byte(src))
	assert.Equal(t, 0, len(s.Errors))

	// Find token indices for message and fields.
	toks := collectCodeTokens(s)
	findNth := func(text string, n int) Token {
		seen := 0
		for _, tok := range toks {
			if tok.Text == text {
				seen++
				if seen == n {
					return tok
				}
			}
		}
		t.Fatalf("token %q #%d not found", text, n)
		return Token{}
	}

	msgTok := findNth("message", 1)
	assert.Equal(t, "// comment line1", s.ConsumeHeadComment(msgTok.Index))
	// Calling twice does not double-consume.
	assert.Equal(t, "", s.ConsumeHeadComment(msgTok.Index))

	aStart := findNth("string", 1)
	aSemi := findNth(";", 1)
	assert.Equal(t, "// comment line2", s.ConsumeHeadComment(aStart.Index))
	assert.Equal(t, "// comment line3", s.ConsumeInlineComment(aSemi.Index))
	assert.Equal(t, "", s.ConsumeInlineComment(aSemi.Index))

	bStart := findNth("string", 2)
	bSemi := findNth(";", 2)
	assert.Equal(t, "// comment line4", s.ConsumeHeadComment(bStart.Index))
	assert.Equal(t, "/*\ncomment line5\nline5+\nline5++\n*/", s.ConsumeInlineComment(bSemi.Index))

	cStart := findNth("string", 3)
	cSemi := findNth(";", 3)
	assert.Equal(t, "// comment line6\n// line7\n// line8", s.ConsumeHeadComment(cStart.Index))
	assert.Equal(t, "", s.ConsumeInlineComment(cSemi.Index))

	dStart := findNth("string", 4)
	assert.Equal(t, "/*\ncomment line10\nline11\nline12\n*/", s.ConsumeHeadComment(dStart.Index))

	// Out-of-bounds comment lookups are safe.
	assert.Equal(t, "", s.ConsumeHeadComment(0))
	assert.Equal(t, "", s.ConsumeHeadComment(-1))
	assert.Equal(t, "", s.ConsumeHeadComment(len(s.All())))
	assert.Equal(t, "", s.ConsumeInlineComment(-1))
	assert.Equal(t, "", s.ConsumeInlineComment(len(s.All())))
}

func TestTokenKindString(t *testing.T) {
	assert.Equal(t, "identifier", TokenIdent.String())
	assert.Equal(t, "';'", TokenSemi.String())
	assert.True(t, TokenWhitespace.IsTrivia())
	assert.True(t, TokenLineComment.IsTrivia())
	assert.True(t, TokenBlockComment.IsTrivia())
	assert.False(t, TokenIdent.IsTrivia())
	assert.StringContains(t, TokenKind(255).String(), "255")
	assert.Equal(t, "line 2 column 3", Pos{Line: 2, Column: 3}.String())
	assert.Equal(t, "line 1 column 4 - bad", SyntaxError{Pos: Pos{Line: 1, Column: 4}, Msg: "bad"}.Error())
}
