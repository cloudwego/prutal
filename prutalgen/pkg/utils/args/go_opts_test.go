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
		"--go_opt=paths=import,paths=source_relative",
		"--go_opt=Mprotos/buzz.proto=example.com/project/protos/first;firstpkg",
		"--go_opt=Mprotos/buzz.proto=example.com/project/protos/fizz",
		"--go_opt=Mprotos/buzz.proto=;buzzpkg",
		"--go_opt=Mprotos/buzz.proto=",
		"--go_opt=Mprotos/ignored.proto",
		"--go_opt=Mprotos/bar.proto=example.com/project/protos/foo",
		"--go_opt=M./protos/bar.proto=example.com/project/protos/different",
	}
	assert.NoError(t, f.Parse(inputs))
	gt, m, module, err := o.Parse()
	assert.NoError(t, err)
	assert.Equal(t, prutalgen.GenBySourceRelative, gt)
	assert.Equal(t, "", module)
	assert.MapEqual(t, map[string]string{
		"protos/buzz.proto":  "example.com/project/protos/fizz;buzzpkg",
		"protos/bar.proto":   "example.com/project/protos/foo",
		"./protos/bar.proto": "example.com/project/protos/different",
	}, m)

	gt = o.GenPathType()
	assert.Equal(t, prutalgen.GenBySourceRelative, gt)
	m = o.Proto2pkg()
	assert.Equal(t, 3, len(m))
}

func TestGoOpts_InvalidPathType(t *testing.T) {
	for _, opt := range []string{"paths", "paths=legacy", "paths=import,paths=bad"} {
		t.Run(opt, func(t *testing.T) {
			o := GoOpts{StringArgs: []string{opt}}
			_, _, _, err := o.Parse()
			assert.ErrorContains(t, err, "unknown path type")
		})
	}

	o := GoOpts{StringArgs: []string{"paths=legacy,paths=source_relative"}}
	assert.Equal(t, prutalgen.GenBySourceRelative, o.GenPathType())
}

func TestGoOpts_RejectsUnknownOption(t *testing.T) {
	o := GoOpts{StringArgs: []string{"bogus=value"}}
	_, _, _, err := o.Parse()
	assert.ErrorContains(t, err, "no such flag -bogus")

	// Keep the legacy helpers permissive.
	assert.Equal(t, prutalgen.GenByImport, o.GenPathType())
	assert.Equal(t, 0, len(o.Proto2pkg()))
}

func TestGoOpts_KnownOptions(t *testing.T) {
	o := GoOpts{StringArgs: []string{
		"annotate_code=false",
		"default_api_level=API_OPEN",
		"apilevelMtest.proto=API_OPEN",
		"experimental_strip_nonfunctional_codegen=true",
		"plugins=",
	}}
	_, _, _, err := o.Parse()
	assert.NoError(t, err)
	assert.NoError(t, o.ValidateFiles([]*prutalgen.Proto{{ProtoFile: "test.proto"}}))

	for _, opt := range []string{
		"experimental_strip_nonfunctional_codegen=maybe",
		"plugins=grpc",
	} {
		t.Run(opt, func(t *testing.T) {
			o := GoOpts{StringArgs: []string{opt}}
			_, _, _, err := o.Parse()
			if err == nil {
				t.Fatal("Parse succeeded, want error")
			}
		})
	}

	for _, opt := range []string{
		"annotate_code=true",
		"default_api_level=API_OPAQUE",
		"apilevelMtest.proto=API_HYBRID",
	} {
		t.Run(opt, func(t *testing.T) {
			o := GoOpts{StringArgs: []string{opt}}
			_, _, _, err := o.Parse()
			assert.NoError(t, err)
			if err := o.ValidateFiles([]*prutalgen.Proto{{ProtoFile: "test.proto"}}); err == nil {
				t.Fatal("ValidateFiles succeeded, want error")
			}
		})
	}

	o = GoOpts{StringArgs: []string{"apilevelMmissing.proto=API_OPAQUE"}}
	_, _, _, err = o.Parse()
	assert.NoError(t, err)
	assert.NoError(t, o.ValidateFiles([]*prutalgen.Proto{{ProtoFile: "test.proto"}}))

	dep := &prutalgen.Proto{ProtoFile: "dep.proto"}
	root := &prutalgen.Proto{
		ProtoFile: "test.proto",
		Imports:   []*prutalgen.Import{{Proto: dep}},
	}
	o = GoOpts{StringArgs: []string{"apilevelMdep.proto=API_OPAQUE"}}
	_, _, _, err = o.Parse()
	assert.NoError(t, err)
	assert.NoError(t, o.ValidateFiles([]*prutalgen.Proto{root}))

	o = GoOpts{StringArgs: []string{
		"default_api_level=API_OPAQUE",
		"apilevelMtest.proto=API_OPEN",
	}}
	_, _, _, err = o.Parse()
	assert.NoError(t, err)
	assert.NoError(t, o.ValidateFiles([]*prutalgen.Proto{{ProtoFile: "test.proto"}}))

	// Pinned protogen uses the final API/plugin values, while annotate_code
	// stays enabled once requested.
	o = GoOpts{StringArgs: []string{
		"default_api_level=API_OPAQUE,default_api_level=API_OPEN",
		"plugins=grpc,plugins=",
	}}
	_, _, _, err = o.Parse()
	assert.NoError(t, err)
	assert.NoError(t, o.ValidateFiles([]*prutalgen.Proto{{ProtoFile: "test.proto"}}))
}

func TestGoOpts_Module(t *testing.T) {
	o := GoOpts{StringArgs: []string{
		"module=example.com/old,module=example.com/project",
		"paths=import",
	}}
	pathType, _, module, err := o.Parse()
	assert.NoError(t, err)
	assert.Equal(t, prutalgen.GenByImport, pathType)
	assert.Equal(t, "example.com/project", module)

	o = GoOpts{StringArgs: []string{"module=example.com/project,paths=source_relative"}}
	_, _, _, err = o.Parse()
	assert.ErrorContains(t, err, "cannot use module= with paths=source_relative")

	o = GoOpts{StringArgs: []string{"module=example.com/project,module=,paths=source_relative"}}
	_, _, module, err = o.Parse()
	assert.NoError(t, err)
	assert.Equal(t, "", module)
}
