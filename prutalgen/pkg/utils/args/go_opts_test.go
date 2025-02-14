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

package args

import (
	"flag"
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
	"github.com/cloudwego/prutal/prutalgen/pkg/prutalgen"
)

func TestGoOpts(t *testing.T) {
	o := GoOpts{}
	f := flag.NewFlagSet(t.Name(), flag.PanicOnError)
	f.Var(&o, "go_opt", "")
	inputs := []string{
		"--go_opt=paths=import",
		"--go_opt=paths=source_relative",
		"--go_opt=Mprotos/buzz.proto=example.com/project/protos/fizz",
		"--go_opt=Mprotos/bar.proto=example.com/project/protos/foo",
	}
	assert.NoError(t, f.Parse(inputs))
	assert.Equal(t, prutalgen.GenBySourceRelative, o.GenPathType())
	m := o.Proto2pkg()
	assert.MapEqual(t, map[string]string{
		"protos/buzz.proto": "example.com/project/protos/fizz",
		"protos/bar.proto":  "example.com/project/protos/foo",
	}, m)
}
