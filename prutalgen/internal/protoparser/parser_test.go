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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func mustParse(t *testing.T, src string) *File {
	t.Helper()
	file, errs := Parse([]byte(src))
	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}
	return file
}

func TestParseTopLevelAndOptions(t *testing.T) {
	src := `//prutalgen:no_pointer
edition = "2023";
import public "a.proto";
import weak "b" ".proto";
import "c.proto";
package foo.bar;

option go_package = "example.com/foo/bar;bar";
option features.(pb.go).api_level = API_OPEN;
option (google.api.http).get = "/v1/items";
option features = {
  field_presence: EXPLICIT,
  repeated_field_encoding: PACKED;
  [pb.go]: { api_level: API_HYBRID }
  items: ["x", 'y']
  [type.googleapis.com/foo.Bar] { count: -5, ratio: +1.5, infv: -inf, nanv: nan, flag: true }
  [type.my-host.com/foo.Bar] [{ nested < x: 1 > }]
};
option qual = true.bar;
;
`
	f := mustParse(t, src)
	assert.Equal(t, "//prutalgen:no_pointer", f.HeadComment)
	assert.True(t, f.Syntax != nil && f.Syntax.IsEdition)
	assert.Equal(t, `"2023"`, f.Syntax.Value.Raw())
	assert.Equal(t, 3, len(f.Imports))
	assert.Equal(t, ImportPublic, f.Imports[0].Kind)
	assert.Equal(t, ImportWeak, f.Imports[1].Kind)
	assert.Equal(t, `"b"".proto"`, f.Imports[1].Path.Raw())
	assert.Equal(t, 2, len(f.Imports[1].Path.Parts))
	assert.Equal(t, ImportRegular, f.Imports[2].Kind)
	assert.Equal(t, 1, len(f.Packages))
	assert.Equal(t, "foo.bar", f.Packages[0].Name)
	assert.Equal(t, 5, len(f.Options))
	assert.Equal(t, "features.(pb.go).api_level", f.Options[1].Name)
	assert.Equal(t, ValueIdent, f.Options[1].Value.Kind)
	assert.Equal(t, "API_OPEN", f.Options[1].Value.Raw)
	assert.Equal(t, "(google.api.http).get", f.Options[2].Name)
	assert.Equal(t, ValueString, f.Options[2].Value.Kind)
	blk := f.Options[3].Value
	assert.Equal(t, ValueBlock, blk.Kind)
	assert.Equal(t, 6, len(blk.Fields))
	assert.True(t, blk.Fields[2].IsExtension)
	assert.Equal(t, "pb.go", blk.Fields[2].Name)
	assert.Equal(t, ValueList, blk.Fields[3].Value.Kind)
	assert.Equal(t, 2, len(blk.Fields[3].Value.List))
	assert.Equal(t, "type.googleapis.com/foo.Bar", blk.Fields[4].Name)
	assert.Equal(t, "type.my-host.com/foo.Bar", blk.Fields[5].Name)
	assert.Equal(t, ValueList, blk.Fields[5].Value.Kind)
	assert.Equal(t, 1, len(blk.Fields[5].Value.List))
	assert.Equal(t, ValueBlock, blk.Fields[5].Value.List[0].Kind)
	assert.Equal(t, ValueIdent, f.Options[4].Value.Kind)
	assert.Equal(t, "true.bar", f.Options[4].Value.Raw)
}

func TestParseMessagesEnumsOneofsExtendsServices(t *testing.T) {
	src := `syntax = "proto2";
package p;

// msg doc
message M {
  optional int32 a = 1 [default = -42, deprecated = true]; // inline a
  repeated .p.M b = 2;
  map<string, int64> c = 3 [features.repeated_field_encoding = EXPANDED];
  optional optional_msg = 4; // type named 'optional'
  string option = 5;         // field named 'option'
  reserved 10, 20 to 30, 40 to max;
  reserved "old" "1", "old2";
  extensions 100 to 199, 500 to max [verification = UNVERIFIED];

  oneof choice {
    option (custom) = { a: 1 b < c: 2 > };
    string x = 6;
    .p.M y = 7;
  }

  enum E {
    option allow_alias = true;
    reserved -10 to -2;
    ZERO = 0;
    NEG = -1 [deprecated = true];
  }

  message Nested {}
  extend M { optional int32 ext_in = 100; }
  ;
} // trailing M

extend M {
  repeated string ext_top = 101;
}

// svc doc
service S {
  option (sopt) = "v";
  // rpc doc
  rpc Call (stream M) returns (.p.M) {
    option (ropt) = 1;
    ;
  } // rpc inline
  rpc Ping (M) returns (M); // ping inline
  ;
}
`
	f := mustParse(t, src)
	assert.Equal(t, "", f.HeadComment) // consumed by message M
	assert.True(t, !f.Syntax.IsEdition)
	assert.Equal(t, 1, len(f.Messages))
	m := f.Messages[0]
	assert.Equal(t, "// msg doc", m.HeadComment)
	assert.Equal(t, "// trailing M", m.InlineComment)
	assert.Equal(t, 12, len(m.Elements))

	f0 := m.Elements[0].(*FieldDecl)
	assert.Equal(t, "optional", f0.Label)
	assert.Equal(t, "int32", f0.Type)
	assert.Equal(t, "a", f0.Name)
	assert.Equal(t, "// inline a", f0.InlineComment)
	assert.Equal(t, "-42", f0.Options[0].Value.Raw)

	f3 := m.Elements[3].(*FieldDecl)
	assert.Equal(t, "", f3.Label)
	assert.Equal(t, "optional", f3.Type)
	assert.Equal(t, "optional_msg", f3.Name)

	f4 := m.Elements[4].(*FieldDecl)
	assert.Equal(t, "string", f4.Type)
	assert.Equal(t, "option", f4.Name)

	mf := m.Elements[2].(*MapFieldDecl)
	assert.Equal(t, "string", mf.KeyType)
	assert.Equal(t, "int64", mf.ValType)
	assert.Equal(t, "c", mf.Name)

	resRanges := m.Elements[5].(*ReservedDecl)
	assert.Equal(t, 3, len(resRanges.Ranges))
	assert.True(t, resRanges.Ranges[2].IsMax)

	resNames := m.Elements[6].(*ReservedDecl)
	assert.Equal(t, 2, len(resNames.Names))
	assert.Equal(t, `"old""1"`, resNames.Names[0].StringLit.Raw())

	exts := m.Elements[7].(*ExtensionsDecl)
	assert.Equal(t, 2, len(exts.Ranges))
	assert.Equal(t, 1, len(exts.Options))

	of := m.Elements[8].(*OneofDecl)
	assert.Equal(t, "choice", of.Name)
	assert.Equal(t, 1, len(of.Options))
	assert.Equal(t, 2, len(of.Fields))

	en := m.Elements[9].(*EnumDecl)
	assert.Equal(t, "E", en.Name)
	assert.Equal(t, "-1", en.Values[1].Number)

	svc := f.Services[0]
	assert.Equal(t, "// svc doc", svc.HeadComment)
	assert.Equal(t, 2, len(svc.Methods))
	assert.True(t, svc.Methods[0].RequestStream)
	assert.False(t, svc.Methods[0].ReturnStream)
	assert.Equal(t, ".p.M", svc.Methods[0].ReturnType)
	assert.Equal(t, "// rpc doc", svc.Methods[0].HeadComment)
	assert.Equal(t, "// rpc inline", svc.Methods[0].InlineComment)
	assert.Equal(t, "// ping inline", svc.Methods[1].InlineComment)
}

func TestParseErrorsAndRecovery(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"missing semicolon", `import "a.proto"`, "missing ';'"},
		{"misplaced syntax", `package p; syntax = "proto3";`, "must be the first statement"},
		{"duplicate edition", `edition = "2023"; edition = "2023";`, "must be the first statement"},
		{"adjacent range", `message M { reserved 2to 3; }`, "space is required between a number"},
		{"bad map key", `message M { map<float, int32> m = 1; }`, "invalid map key type"},
		{"label in oneof", `message M { oneof o { optional int32 a = 1; } }`, "not allowed here"},
		{"bad rpc returns", `service S { rpc Call(M) return (M); }`, "expected 'returns'"},
		{"empty bracket options", `message M { int32 a = 1 []; }`, "expected option name"},
		{"mixed reserved", `message M { reserved "a", 1; }`, "expected string literal"},
		{"deep nesting", strings.Repeat("message M { ", 260) + strings.Repeat("}", 260), "maximum nesting depth exceeded"},
		{"stray top-level rbrace", `} message M {}`, "expected top-level declaration"},
		{"unclosed block", `option x = { a: 1`, "expected '}'"},
		{"unclosed message", `message M { int32 a = `, "expected integer"},
		{"semi before syntax", `; syntax = "proto3";`, "must be the first statement"},
		{"top-level angle option", `option (x) = < a: 1 >;`, "expected option value"},
		{"bracket angle option", `message M { int32 a = 1 [(x) = < a: 1 >]; }`, "expected option value"},
		{"leading dot option value", `option lead = .foo.Bar;`, "expected option value"},
		{"nested list", `option (x) = { a: [[1]] };`, "expected option value"},
		{"hyphen without slash in block key", `option (x) = { [foo-bar]: 1 };`, `got "-"`},
		{"optional scalar missing name", `message M { optional int32 = 1; }`, "expected field name"},
		{"deep option block nesting", "option (x) = " + strings.Repeat("{ a: ", 260) + "1" + strings.Repeat(" }", 260) + ";", "maximum nesting depth exceeded"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := Parse([]byte(tc.src))
			assert.True(t, len(errs) > 0, tc.name)
			assert.StringContains(t, errs[0].Error(), tc.want)
		})
	}
}

func TestParseCommentsAndEdgeCases(t *testing.T) {
	t.Run("enum value crlf blank line and oneof", func(t *testing.T) {
		src := "// detached\r\n\r\n" +
			"// enum doc\r\n" +
			"enum E {\r\n" +
			"  // zero doc\r\n" +
			"  ZERO = 0; // zero inline\r\n" +
			"}\r\n" +
			"message M {\r\n" +
			"  // oneof comment\r\n" +
			"  oneof choice {\r\n" +
			"    // x doc\r\n" +
			"    int32 x = 1; // x inline\r\n" +
			"  }\r\n" +
			"  reserved foo, bar;\r\n" +
			"}\r\n"
		f := mustParse(t, src)
		assert.Equal(t, "", f.HeadComment)
		assert.Equal(t, "// enum doc", f.Enums[0].HeadComment)
		assert.Equal(t, "// zero doc", f.Enums[0].Values[0].HeadComment)
		assert.Equal(t, "// zero inline", f.Enums[0].Values[0].InlineComment)

		m := f.Messages[0]
		of := m.Elements[0].(*OneofDecl)
		assert.Equal(t, "// x doc", of.Fields[0].HeadComment)
		assert.Equal(t, "// x inline", of.Fields[0].InlineComment)
		res := m.Elements[1].(*ReservedDecl)
		assert.Equal(t, 2, len(res.Names))
		assert.Equal(t, "foo", res.Names[0].Ident)
		assert.True(t, res.Names[0].StringLit == nil)
	})

	t.Run("sorted error order", func(t *testing.T) {
		src := "import \"a.proto\"\nmessage M { reserved 2to 3; }"
		_, errs := Parse([]byte(src))
		assert.True(t, len(errs) >= 2)
		assert.True(t, errs[0].Pos.Offset < errs[1].Pos.Offset)
		assert.StringContains(t, errs[0].Msg, "missing ';'")
		assert.StringContains(t, errs[1].Msg, "space is required")
	})

	t.Run("brace aware recovery", func(t *testing.T) {
		cases := []string{
			`message M { int32 a = 1 [(x) = 5 { a: 1 }]; string ok = 2; }`,
			`message M { option (x) = { a: [1 bad }; string ok = 2; }`,
		}
		for _, src := range cases {
			f, errs := Parse([]byte(src))
			assert.True(t, len(errs) > 0)
			assert.Equal(t, 1, len(f.Messages))
			var foundOK bool
			for _, el := range f.Messages[0].Elements {
				if fd, ok := el.(*FieldDecl); ok && fd.Name == "ok" {
					foundOK = true
				}
			}
			assert.True(t, foundOK)
		}
	})

	t.Run("independent message and option depth", func(t *testing.T) {
		var b strings.Builder
		for i := 0; i < 250; i++ {
			b.WriteString("message M { ")
		}
		b.WriteString("option (x) = ")
		for i := 0; i < 10; i++ {
			b.WriteString("{ a: ")
		}
		b.WriteString("1")
		for i := 0; i < 10; i++ {
			b.WriteString(" }")
		}
		b.WriteString(";")
		for i := 0; i < 250; i++ {
			b.WriteString(" }")
		}
		f := mustParse(t, b.String())
		assert.Equal(t, 1, len(f.Messages))
	})

	t.Run("suppress duplicate semi error on lex error", func(t *testing.T) {
		_, errs := Parse([]byte("package foo ?"))
		assert.Equal(t, 1, len(errs))
		assert.StringContains(t, errs[0].Msg, "unexpected")
	})
}

func FuzzParse(f *testing.F) {
	f.Add([]byte(`syntax = "proto3"; message M { int32 a = 1; }`))
	f.Add([]byte(`edition = "2023"; option (x) = { [a-b.com/c.D]: [{k: [1, -inf]}] };`))
	f.Add([]byte(`} [ { < @ "unterminated`))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = Parse(data)
	})
}

func TestParseRepoProtoCorpus(t *testing.T) {
	patterns := []string{
		filepath.Join("..", "..", "pkg", "prutalgen", "wellknowns", "*.proto"),
		filepath.Join("..", "..", "..", "internal", "prutal", "*.proto"),
		filepath.Join("..", "..", "..", "tests", "cases", "compat", "compat3.proto"),
		filepath.Join("..", "..", "..", "tests", "cases", "compat", "compat2023", "*.proto"),
		filepath.Join("..", "..", "..", "tests", "cases", "edition2023", "*.proto"),
		filepath.Join("..", "..", "..", "tests", "cases", "grpc", "*.proto"),
		filepath.Join("..", "..", "..", "tests", "cases", "nested", "*.proto"),
		filepath.Join("..", "..", "..", "tests", "cases", "oneof", "*.proto"),
		filepath.Join("..", "..", "..", "tests", "cases", "others", "*.proto"),
		filepath.Join("..", "..", "..", "tests", "cases", "proto2", "*.proto"),
	}
	total := 0
	for _, pat := range patterns {
		files, err := filepath.Glob(pat)
		assert.NoError(t, err)
		for _, fn := range files {
			total++
			t.Run(filepath.ToSlash(fn), func(t *testing.T) {
				b, err := os.ReadFile(fn)
				assert.NoError(t, err)
				file, errs := Parse(b)
				if len(errs) > 0 {
					t.Fatalf("parse %s failed: %v", fn, errs)
				}
				assert.True(t, file != nil)
			})
		}
	}
	assert.True(t, total >= 20)
}
