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

	"github.com/cloudwego/prutal/prutalgen/internal/antlr"

	"github.com/cloudwego/prutal/prutalgen/internal/parser"
)

type streamContext struct {
	s *antlr.CommonTokenStream

	// Token index -> true
	// a comment should be returned in
	// consumeHeadComment or consumeInlineComment, NOT both.
	commentConsumed map[int]bool
}

func newStreamContext(s *antlr.CommonTokenStream) *streamContext {
	return &streamContext{s: s, commentConsumed: map[int]bool{}}
}

func filterCommentOrWhitespace(tt []antlr.Token) (ret []antlr.Token, hasComment bool) {
	ret = tt[:0]
	for _, t := range tt {
		switch t.GetTokenType() {
		case parser.ProtobufParserLINE_COMMENT,
			parser.ProtobufParserCOMMENT:
			hasComment = true
			ret = append(ret, t)
		case parser.ProtobufParserWS:
			ret = append(ret, t)
		}
	}
	if !hasComment {
		return nil, hasComment
	}
	return ret, true
}

func (x *streamContext) consumeHeadComment(c antlr.ParserRuleContext) string {
	tt, ok := filterCommentOrWhitespace(
		x.s.GetHiddenTokensToLeft(c.GetStart().GetTokenIndex(), -1),
	)
	if !ok {
		return ""
	}
	ss := make([]string, 0, len(tt))
	for i := len(tt) - 1; i >= 0; i-- {
		t := tt[i]
		ti := t.GetTokenIndex()
		tp := t.GetTokenType()
		if x.commentConsumed[ti] {
			break
		}
		s := t.GetText()
		if tp == parser.ProtobufParserWS {
			if strings.Count(s, "\n") > 1 {
				// normally only one new line between 2 comment tokens are expected
				// if we got more than one \n, likely it's an empty line to seperate two definitions
				break
			}
		}
		x.commentConsumed[ti] = true
		s = strings.TrimSpace(s)
		if s != "" {
			ss = append(ss, s)
		}
	}
	if len(ss) == 0 {
		return ""
	}
	// reverse ss coz we scan backward above
	for i, j := 0, len(ss)-1; i < j; i, j = i+1, j-1 {
		ss[i], ss[j] = ss[j], ss[i]
	}
	return strings.Join(ss, "\n")
}

func (x *streamContext) consumeInlineComment(c antlr.ParserRuleContext) string {
	tt, ok := filterCommentOrWhitespace(
		x.s.GetHiddenTokensToRight(c.GetStop().GetTokenIndex(), -1),
	)
	if !ok {
		return ""
	}
	for _, t := range tt {
		s := t.GetText()
		tp := t.GetTokenType()
		ti := t.GetTokenIndex()
		if x.commentConsumed[ti] {
			return ""
		}
		if tp == parser.ProtobufParserWS {
			if strings.Contains(s, "\n") {
				return "" // newline? the commment may not belong to the given rule
			}
			continue // skip parser.ProtobufParserWS
		}
		// parser.ProtobufParserLINE_COMMENT or parser.ProtobufParserCOMMENT
		x.commentConsumed[t.GetTokenIndex()] = true
		return strings.TrimSpace(t.GetText())
	}
	return ""
}
