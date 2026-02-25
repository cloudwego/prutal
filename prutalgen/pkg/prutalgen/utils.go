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
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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


// sort protos by topological order
func sortProtoFiles(pp []*Proto) []*Proto {
	// Build dependency graph
	deps := make(map[*Proto][]*Proto, len(pp))
	inDegree := make(map[*Proto]int, len(pp))
	for _, p := range pp {
		deps[p] = []*Proto{}
		for _, imp := range p.Imports {
			deps[p] = append(deps[p], imp.Proto)
			inDegree[imp.Proto]++
		}
	}
	// Kahn's algorithm
	ret := make([]*Proto, 0, len(pp))
	var queue []*Proto
	for _, p := range pp {
		if inDegree[p] == 0 {
			queue = append(queue, p)
		}
	}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		ret = append(ret, p)
		for _, dep := range deps[p] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}
	// If not all protos are in result, there is a cycle
	if len(ret) != len(pp) {
		// fallback: return original order (or could panic), never happens?
		return pp
	}
	return ret
}
