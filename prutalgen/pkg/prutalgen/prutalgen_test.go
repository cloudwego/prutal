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
	x := NewLoader([]string{"."}, nil)
	x.SetLogger(&testLogger{t})
	fn := writeFile(t, "test.proto", []byte(payload))
	ff := x.LoadProto(fn)
	assert.Equal(t, 1, len(ff))
	assert.Equal(t, fn, ff[0].ProtoFile)
	return ff[0]
}

func expectFail(t *testing.T, payload string, l LoggerIface) {
	x := NewLoader([]string{"."}, nil)
	x.SetLogger(l)
	fn := writeFile(t, "test.proto", []byte(payload))
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

func TestLoader_SyntaxError(t *testing.T) {
	expectFail(t, `import "blabla"`, &expectLogger{t: t,
		PrintContains: []string{`parsing`, `missing ';'`},
		FatalContains: `error occurred`,
	})
}

func TestLoader_NoGoPackage(t *testing.T) {
	expectFail(t, ``, &expectLogger{t: t,
		FatalContains: `option "go_package" not set`,
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
