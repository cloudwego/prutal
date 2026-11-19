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
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/text"
	"github.com/cloudwego/prutal/prutalgen/internal/protoparser"
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

func unmarshalStringParts(parts []string) (string, error) {
	if len(parts) == 1 {
		return text.UnmarshalString(parts[0])
	}
	var b strings.Builder
	for _, part := range parts {
		s, err := text.UnmarshalString(part)
		if err != nil {
			return "", err
		}
		b.WriteString(s)
	}
	return b.String(), nil
}

func unmarshalValue(v *protoparser.Value) (string, error) {
	if v.Kind == protoparser.ValueString {
		return unmarshalStringParts(v.StringParts)
	}
	return v.Raw, nil
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
