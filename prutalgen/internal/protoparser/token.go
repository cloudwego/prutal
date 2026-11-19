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

import "fmt"

// Pos identifies a source position in a proto file. Line and Column are 1-based;
// Offset is a 0-based byte offset.
type Pos struct {
	Offset int
	Line   int
	Column int
}

// String formats p as "line X column Y".
func (p Pos) String() string {
	return fmt.Sprintf("line %d column %d", p.Line, p.Column)
}

// SyntaxError describes a lexical or grammatical error at a source position.
type SyntaxError struct {
	Pos Pos
	Msg string
}

// Error implements the error interface.
func (e SyntaxError) Error() string {
	return fmt.Sprintf("%s - %s", e.Pos.String(), e.Msg)
}

// TokenKind classifies a lexical token produced by Lex.
type TokenKind uint8

const (
	TokenEOF TokenKind = iota
	TokenError

	TokenIdent
	TokenInt
	TokenFloat
	TokenString

	TokenSemi     // ;
	TokenEqual    // =
	TokenLParen   // (
	TokenRParen   // )
	TokenLBracket // [
	TokenRBracket // ]
	TokenLBrace   // {
	TokenRBrace   // }
	TokenLAngle   // <
	TokenRAngle   // >
	TokenDot      // .
	TokenComma    // ,
	TokenColon    // :
	TokenPlus     // +
	TokenMinus    // -
	TokenSlash    // /

	TokenWhitespace
	TokenLineComment
	TokenBlockComment
)

var tokenKindNames = [...]string{
	TokenEOF:          "EOF",
	TokenError:        "error",
	TokenIdent:        "identifier",
	TokenInt:          "integer",
	TokenFloat:        "float",
	TokenString:       "string",
	TokenSemi:         "';'",
	TokenEqual:        "'='",
	TokenLParen:       "'('",
	TokenRParen:       "')'",
	TokenLBracket:     "'['",
	TokenRBracket:     "']'",
	TokenLBrace:       "'{'",
	TokenRBrace:       "'}'",
	TokenLAngle:       "'<'",
	TokenRAngle:       "'>'",
	TokenDot:          "'.'",
	TokenComma:        "','",
	TokenColon:        "':'",
	TokenPlus:         "'+'",
	TokenMinus:        "'-'",
	TokenSlash:        "'/'",
	TokenWhitespace:   "whitespace",
	TokenLineComment:  "line comment",
	TokenBlockComment: "block comment",
}

// String returns a human-readable description of k.
func (k TokenKind) String() string {
	if int(k) < len(tokenKindNames) && tokenKindNames[k] != "" {
		return tokenKindNames[k]
	}
	return fmt.Sprintf("TokenKind(%d)", k)
}

// IsTrivia reports whether k is whitespace or a comment.
func (k TokenKind) IsTrivia() bool {
	return k == TokenWhitespace || k == TokenLineComment || k == TokenBlockComment
}

// Token represents one lexical token in a TokenStream.
type Token struct {
	Kind   TokenKind
	Text   string
	Pos    Pos
	EndPos Pos
	Index  int // index within the full token stream returned by TokenStream.All
}
