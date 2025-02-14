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

package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/prutal/prutalgen/pkg/prutalgen"
	"github.com/cloudwego/prutal/prutalgen/pkg/utils/args"
)

func main() {
	var (
		protoPath sliceArg
		out       string
		opts      args.GoOpts

		GenGetter bool
	)
	flags := flag.NewFlagSet("prutalgen", flag.ExitOnError)
	flags.Var(&protoPath, "proto_path", "")
	flags.Var(&protoPath, "I", "")
	flags.StringVar(&out, "go_out", "", "")
	flags.Var(&opts, "go_opt", "")
	flags.BoolVar(&GenGetter, "gen_getter", false, "")
	_ = flags.Parse(os.Args[1:])

	if len(protoPath) == 0 {
		protoPath = append(protoPath, ".")
	}
	if out == "" {
		out = "."
	}

	x := prutalgen.NewLoader([]string(protoPath), opts.Proto2pkg())
	g := prutalgen.NewGoCodeGen()
	g.Getter = GenGetter
	args := flags.Args()
	if len(args) == 0 {
		println("WARN: no proto file provided")
	}
	for _, a := range args {
		p := x.LoadProto(filepath.Clean(a))[0]
		if err := g.Gen(p, opts.GenPathType(), out); err != nil {
			p.Fatalf("generate code err: %s", err)
		}
	}
}

type sliceArg []string

func (a *sliceArg) String() string {
	return strings.Join(*a, ",")
}

func (a *sliceArg) Set(v string) error {
	*a = append(*a, v)
	return nil
}
