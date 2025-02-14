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
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cloudwego/prutal/prutalgen/internal/antlr"

	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/text"
)

func push[T any](vv *[]T, v T) {
	*vv = append(*vv, v)
}

func pop[T any](vv *[]*T) {
	(*vv)[len(*vv)-1] = nil
	*vv = (*vv)[:len(*vv)-1]
}

func last[T any](vv []T) T {
	return vv[len(vv)-1]
}

func unmarshalConst(s string) (string, error) {
	if len(s) > 0 && (s[0] == '\'' || s[0] == '"') {
		return text.UnmarshalString(s)
	}
	return s, nil
}

func getText(n antlr.TerminalNode) string {
	if n == nil {
		return ""
	}
	return n.GetText()
}

func getRuleIndex(v antlr.Tree) int {
	type RuleContext interface {
		GetRuleIndex() int
	}
	return v.(RuleContext).GetRuleIndex()
}

func getTokenPos(v antlr.ParserRuleContext) string {
	t := v.GetStart()
	return fmt.Sprintf("line %d column %d", t.GetLine(), t.GetColumn())
}

func parseI32(c antlr.ParserRuleContext) (int32, error) {
	v, err := strconv.ParseInt(c.GetText(), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s - %w", getTokenPos(c), err)
	}
	return int32(v), nil
}

func isfalse(v string) bool { return v == "false" }
func istrue(v string) bool  { return v == "true" }

func hasPathPrefix(a, b string) bool {
	// same as strings.HasPrefix(a, b + ".")
	return len(a) >= len(b)+1 &&
		a[:len(b)] == b && a[len(b)] == '.'
}

func trimPathPrefix(a, b string) (string, bool) {
	if hasPathPrefix(a, b) {
		return a[len(b)+1:], true
	}
	return a, false
}

func refPath(abs string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return abs
	}
	ret, err := filepath.Rel(cwd, abs)
	if err != nil {
		return abs
	}
	return ret
}

// NOTE: This implemenation doesn't works with errors.Is/As.
// Use errors.Join if >=go1.20
func joinErrs(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	ss := make([]string, 0, len(errs))
	for _, err := range errs {
		ss = append(ss, err.Error())
	}
	return errors.New(strings.Join(ss, "\n"))
}
