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
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

type testLogger struct {
	*testing.T
}

type testFatal string

type panicLogger struct{}

func (panicLogger) Printf(string, ...any) {}

func (panicLogger) Fatalf(format string, v ...any) {
	panic(testFatal(fmt.Sprintf(format, v...)))
}

func expectProtoError(t *testing.T, payload, want string) {
	t.Helper()
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		fn := writeFile(t, "test.proto", []byte(payload))
		x := NewLoader([]string{filepath.Dir(fn)}, nil)
		x.SetLogger(panicLogger{})
		_ = x.LoadProto(fn)
	}()
	if recovered == nil {
		t.Fatal("expected loader error")
	}
	err, ok := recovered.(testFatal)
	if !ok {
		panic(recovered)
	}
	assert.StringContains(t, string(err), want)
}

func expectLoaderError(t *testing.T, x Loader, files []string, want string) {
	t.Helper()
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		x.SetLogger(panicLogger{})
		x.LoadProtos(files)
	}()
	if recovered == nil {
		t.Fatal("expected loader error")
	}
	err, ok := recovered.(testFatal)
	if !ok {
		panic(recovered)
	}
	assert.StringContains(t, string(err), want)
}

func (l testLogger) Printf(format string, v ...any) {
	l.Helper()
	l.Logf(format, v...)
}

func writeFile(t *testing.T, fn string, b []byte) string {
	t.Helper()
	return writeFileUnderDir(t, t.TempDir(), fn, b)
}

func writeFileUnderDir(t *testing.T, dir, fn string, b []byte) string {
	t.Helper()
	fn = filepath.Join(dir, fn)
	if err := os.WriteFile(fn, b, 0644); err != nil {
		t.Fatal(err)
	}
	return fn
}

type expectLogger struct {
	t *testing.T

	PrintContains []string
	FatalContains string
}

func (l *expectLogger) Printf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	l.t.Log(s)
	if len(l.PrintContains) > 0 {
		l.t.Helper()
		expect := l.PrintContains[0]
		l.PrintContains = l.PrintContains[1:]
		assert.StringContains(l.t, s, expect)
	}
}

func (l *expectLogger) Fatalf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	l.t.Log(s)
	if l.FatalContains == "" {
		l.t.FailNow()
	} else {
		l.t.Helper()
		assert.StringContains(l.t, s, l.FatalContains)
	}
	l.t.SkipNow() // must skip coz it should NOT continue to run
}

func loadTestProto(t *testing.T, payload string) *Proto {
	fn := writeFile(t, "test.proto", []byte(payload))
	x := NewLoader([]string{filepath.Dir(fn)}, nil)
	x.SetLogger(&testLogger{t})
	ff := x.LoadProto(fn)
	assert.Equal(t, 1, len(ff))
	assert.Equal(t, fn, ff[0].ProtoFile)
	return ff[0]
}

func expectFail(t *testing.T, payload string, l LoggerIface) {
	fn := writeFile(t, "test.proto", []byte(payload))
	x := NewLoader([]string{filepath.Dir(fn)}, nil)
	x.SetLogger(l)
	_ = x.LoadProto(fn)
	t.Helper()
	t.Fatal("didn't call Fatal")
}

func TestLoader(t *testing.T) {
	f := loadTestProto(t, `option go_package = "hello/prutal_test";`)
	assert.Equal(t, "prutal_test", f.GoPackage) // base path of go_package

	f = loadTestProto(t, `option go_package = "hello/prutal_test; prutal";`)
	assert.Equal(t, "prutal", f.GoPackage) // go_package with package name
}

func TestLoader_MOptionPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name        string
		fileOpt     string
		mOpt        string
		wantImport  string
		wantPackage string
	}{
		{
			name:        "M import and go_package derived package",
			fileOpt:     "golang.org/x/foo",
			mOpt:        "golang.org/x/bar",
			wantImport:  "golang.org/x/bar",
			wantPackage: "foo",
		},
		{
			name:        "M import and package",
			fileOpt:     "golang.org/x/foo;filepkg",
			mOpt:        "golang.org/x/bar;mpkg",
			wantImport:  "golang.org/x/bar",
			wantPackage: "mpkg",
		},
		{
			name:        "M package only",
			fileOpt:     "golang.org/x/foo;filepkg",
			mOpt:        ";mpkg",
			wantImport:  "golang.org/x/foo",
			wantPackage: "mpkg",
		},
		{
			name:        "go_package package with M import",
			fileOpt:     ";filepkg",
			mOpt:        "golang.org/x/bar",
			wantImport:  "golang.org/x/bar",
			wantPackage: "filepkg",
		},
		{
			name:        "M only",
			mOpt:        "golang.org/x/bar",
			wantImport:  "golang.org/x/bar",
			wantPackage: "bar",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			payload := ""
			if tc.fileOpt != "" {
				payload = fmt.Sprintf("option go_package = %q;", tc.fileOpt)
			}
			_ = writeFileUnderDir(t, dir, "test.proto", []byte(payload))
			x := NewLoader([]string{dir}, map[string]string{"test.proto": tc.mOpt})
			x.SetLogger(&testLogger{t})
			p := x.LoadProto("test.proto")[0]
			assert.Equal(t, "test.proto", p.protoName)
			assert.Equal(t, tc.wantImport, p.GoImport)
			assert.Equal(t, tc.wantPackage, p.GoPackage)
		})
	}
}

func TestLoader_MOptionUsesLogicalProtoName(t *testing.T) {
	dir := t.TempDir()
	protoPath := filepath.Join(dir, "cases", "oneof")
	assert.NoError(t, os.MkdirAll(protoPath, 0755))
	filename := writeFileUnderDir(t, protoPath, "oneof.proto", []byte(
		`option go_package = "example.com/original/foo";`,
	))

	for _, tc := range []struct {
		name       string
		mapping    map[string]string
		wantImport string
	}{
		{
			name:       "descriptor name matches",
			mapping:    map[string]string{"oneof.proto": "example.com/mapped/bar"},
			wantImport: "example.com/mapped/bar",
		},
		{
			name:       "M key is not cleaned",
			mapping:    map[string]string{"./oneof.proto": "example.com/mapped/bar"},
			wantImport: "example.com/original/foo",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			x := NewLoader([]string{protoPath}, tc.mapping)
			x.SetLogger(&testLogger{t})
			p := x.LoadProto(filename)[0]
			assert.Equal(t, "oneof.proto", p.protoName)
			assert.Equal(t, tc.wantImport, p.GoImport)
			assert.Equal(t, "foo", p.GoPackage)
		})
	}

	t.Run("first matching proto path", func(t *testing.T) {
		x := NewLoader([]string{dir, protoPath}, map[string]string{
			"cases/oneof/oneof.proto": "example.com/mapped/bar",
		})
		x.SetLogger(&testLogger{t})
		p := x.LoadProto(filename)[0]
		assert.Equal(t, "cases/oneof/oneof.proto", p.protoName)
		assert.Equal(t, "example.com/mapped/bar", p.GoImport)
	})

	t.Run("virtual input keeps requested name", func(t *testing.T) {
		x := NewLoader([]string{dir, protoPath}, map[string]string{
			"oneof.proto": "example.com/mapped/bar",
		})
		x.SetLogger(&testLogger{t})
		p := x.LoadProto("oneof.proto")[0]
		assert.Equal(t, "oneof.proto", p.protoName)
		assert.Equal(t, "example.com/mapped/bar", p.GoImport)
	})
}

func TestLoaderRejectsInvalidPhysicalRoots(t *testing.T) {
	include := t.TempDir()
	outside := writeFileUnderDir(t, t.TempDir(), "outside.proto", []byte(
		`option go_package = "example.com/outside";`,
	))
	expectLoaderError(t, NewLoader([]string{include}, nil), []string{outside},
		"does not reside within any include path")

	first := filepath.Join(t.TempDir(), "first")
	second := filepath.Join(filepath.Dir(first), "second")
	assert.NoError(t, os.MkdirAll(first, 0755))
	assert.NoError(t, os.MkdirAll(second, 0755))
	_ = writeFileUnderDir(t, first, "shadow.proto", []byte(
		`option go_package = "example.com/first"; message First {}`,
	))
	shadowed := writeFileUnderDir(t, second, "shadow.proto", []byte(
		`option go_package = "example.com/second"; message Second {}`,
	))
	expectLoaderError(t, NewLoader([]string{first, second}, nil), []string{shadowed},
		"is shadowed by")
}

func TestCanonicalProtoName(t *testing.T) {
	for _, name := range []string{"test.proto", "dir/test.proto"} {
		assert.True(t, isCanonicalProtoName(name), name)
	}
	for _, name := range []string{"", ".", "..", "./test.proto", "../test.proto", "dir/../test.proto", "/test.proto", `dir\test.proto`} {
		assert.False(t, isCanonicalProtoName(name), name)
	}
}

func TestRelativeToInclude(t *testing.T) {
	for _, tc := range []struct {
		include   string
		requested string
		want      string
	}{
		{"tests/cases", "./tests/cases/grpc/echo.proto", "grpc/echo.proto"},
		{"tests//cases", "tests/cases/./grpc/echo.proto", "grpc/echo.proto"},
		{"./tests/cases/.", "tests//cases/grpc/echo.proto", "grpc/echo.proto"},
		{"../other", "../other/foo.proto", "foo.proto"},
	} {
		got, ok := relativeToInclude(tc.include, tc.requested)
		assert.True(t, ok, tc)
		assert.Equal(t, tc.want, got, tc)
		assert.True(t, isCanonicalProtoName(got), tc)
	}

	got, ok := relativeToInclude("tests/cases", "tests/cases/sub/../echo.proto")
	assert.True(t, ok)
	assert.False(t, isCanonicalProtoName(got))
}

func TestLoaderUsesLogicalProtoIdentity(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "nested")
	assert.NoError(t, os.MkdirAll(nested, 0755))
	_ = writeFileUnderDir(t, nested, "same.proto", []byte(
		`package duplicate; option go_package = "example.com/duplicate"; message M {}`,
	))
	roots := NewLoader([]string{nested}, nil).LoadProtos([]string{"same.proto", "same.proto"})
	assert.Equal(t, 1, len(roots))

	x := NewLoader([]string{dir, nested}, nil)
	expectLoaderError(t, x, []string{"nested/same.proto", "same.proto"}, `"duplicate.M"`)
}

func TestGoPackageValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		opt  string
		err  string
	}{
		{"bare import path", "foo", "invalid Go import path"},
		{"missing import path", ";foo", "unable to determine Go import path"},
		{"keyword package name", "example.com/x;go", "invalid Go package name"},
		{"invalid package name", "example.com/x;bad-name", "invalid Go package name"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &Proto{ProtoFile: "test.proto"}
			p.setGoPackage(tc.opt)
			assert.ErrorContains(t, p.validateGoPackage(), tc.err)
		})
	}

	p := &Proto{ProtoFile: "test.proto"}
	p.setGoPackage("example.com/go")
	assert.NoError(t, p.validateGoPackage())
	assert.Equal(t, "_go", p.GoPackage)

	p.setGoPackage("example.com/x;_go")
	assert.NoError(t, p.validateGoPackage())
	assert.Equal(t, "_go", p.GoPackage)

	p.setGoPackageOptions("example.com/original/foo", "invalid")
	assert.ErrorContains(t, p.validateGoPackage(), "invalid Go import path")

	for _, mOpt := range []string{" ", " ;pkg", "; "} {
		p.setGoPackageOptions("example.com/original/foo", mOpt)
		if err := p.validateGoPackage(); err == nil {
			t.Fatalf("setGoPackageOptions(_, %q) succeeded, want error", mOpt)
		}
	}

	for _, mOpt := range []string{"; pkg", ";pkg ", "example.com/mapped/foo; pkg "} {
		p.setGoPackageOptions("example.com/original/foo", mOpt)
		assert.NoError(t, p.validateGoPackage())
		assert.Equal(t, "pkg", p.GoPackage)
	}
}

func TestGoPackageConsistency(t *testing.T) {
	err := validateGoPackageConsistency([]*Proto{
		{ProtoFile: "a.proto", GoImport: "example.com/shared", GoPackage: "a"},
		{ProtoFile: "b.proto", GoImport: "example.com/shared", GoPackage: "b"},
	})
	assert.ErrorContains(t, err, "inconsistent names")

	err = validateGoPackageConsistency([]*Proto{
		{ProtoFile: "a.proto", GoImport: "example.com/shared", GoPackage: "shared"},
		{ProtoFile: "b.proto", GoImport: "example.com/shared", GoPackage: "shared"},
	})
	assert.NoError(t, err)

	err = validateGoPackageConsistency([]*Proto{
		{ProtoFile: "a.proto", GoImport: "example.com/shared", GoPackage: "shared", goPackageIdentity: "shared"},
		{ProtoFile: "b.proto", GoImport: "example.com/shared", GoPackage: "shared", goPackageIdentity: " shared"},
	})
	assert.ErrorContains(t, err, "inconsistent names")
}

func TestLoader_SyntaxError(t *testing.T) {
	expectFail(t, `import "blabla"`, &expectLogger{t: t,
		PrintContains: []string{`parsing`, `missing ';'`},
		FatalContains: `error occurred`,
	})
}

func TestLoader_NoGoPackage(t *testing.T) {
	expectFail(t, ``, &expectLogger{t: t,
		FatalContains: `unable to determine Go import path`,
	})
}

func TestLoader_FileNotFound(t *testing.T) {
	x := NewLoader([]string{"."}, nil)
	x.SetLogger(&expectLogger{t: t,
		FatalContains: `proto file "XXX" not found`,
	})
	_ = x.LoadProto("XXX")
	t.Fatal("never goes here. logger Fatalf in LoadProto")
}

func TestLoader_RAG(t *testing.T) {
	// a -> (b, c)
	// b -> d
	// c -> d
	// d -> e
	dir := t.TempDir()
	_ = writeFileUnderDir(t, dir, "e.proto", []byte(
		`package e;`+`option go_package = "./e";`,
	))
	_ = writeFileUnderDir(t, dir, "d.proto", []byte(
		`import "e.proto";`+`package d;`+`option go_package = "./d";`,
	))

	_ = writeFileUnderDir(t, dir, "c.proto", []byte(
		`import "d.proto";`+`package c;`+`option go_package = "./c";`,
	))

	_ = writeFileUnderDir(t, dir, "b.proto", []byte(
		`import "d.proto";`+`package b;`+`option go_package = "./b";`,
	))

	fn := writeFileUnderDir(t, dir, "a.proto", []byte(
		`import "b.proto";`+`import "c.proto";`+
			`package a;`+`option go_package = "./a";`,
	))
	x := NewLoader([]string{filepath.Dir(fn)}, nil)
	ff := x.LoadProto("a.proto")
	assert.Equal(t, len(ff), 5)
	assert.Equal(t, filepath.Base(ff[0].ProtoFile), "a.proto")
	assert.Equal(t, filepath.Base(ff[1].ProtoFile), "b.proto")
	assert.Equal(t, filepath.Base(ff[2].ProtoFile), "c.proto")
	assert.Equal(t, filepath.Base(ff[3].ProtoFile), "d.proto")
	assert.Equal(t, filepath.Base(ff[4].ProtoFile), "e.proto")
}

func TestLoader_CyclicImport(t *testing.T) {
	fn := writeFile(t, "test.proto", []byte(
		`import "test.proto";`,
	))
	x := NewLoader([]string{filepath.Dir(fn)}, nil)
	x.SetLogger(&expectLogger{t: t,
		FatalContains: `cyclic import`,
	})
	_ = x.LoadProto("test.proto")
	t.Fatal("never goes here. logger Fatalf in LoadProto")

}
