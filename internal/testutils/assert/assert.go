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

package assert

import (
	"fmt"
	"reflect"
	"strings"
)

// TestingT is an interface used by assert pkg implemented by *testing.T
type TestingT interface {
	Fatalf(format string, args ...interface{})
	Helper()
}

func True(t TestingT, v bool, msgs ...interface{}) {
	if !v {
		t.Helper()
		t.Fatalf("not true. %s", fmt.Sprint(msgs...))
	}
}

func False(t TestingT, v bool, msgs ...interface{}) {
	if v {
		t.Helper()
		t.Fatalf("not false. %s", fmt.Sprint(msgs...))
	}
}

func Same(t TestingT, a, b any, msgs ...interface{}) {
	x := reflect.ValueOf(a)
	y := reflect.ValueOf(b)
	if x.Kind() != reflect.Pointer || y.Kind() != reflect.Pointer {
		t.Helper()
		t.Fatalf("not pointer. %s", fmt.Sprint(msgs...))
		return
	}
	if x.Pointer() != y.Pointer() {
		t.Helper()
		t.Fatalf("not same. %s", fmt.Sprint(msgs...))
	}
}

func Equal[T comparable](t TestingT, a, b T, msgs ...interface{}) {
	if a != b {
		t.Helper()
		t.Fatalf("not equal: [ %#v != %#v ] %s", a, b, fmt.Sprint(msgs...))
	}
}

func Nil[T any](t TestingT, a *T, msgs ...interface{}) {
	if a != nil {
		t.Helper()
		t.Fatalf("expected nil. %s", fmt.Sprint(msgs...))
	}
}

func NotNil[T any](t TestingT, a *T, msgs ...interface{}) {
	if a == nil {
		t.Helper()
		t.Fatalf("expected not nil. %s", fmt.Sprint(msgs...))
	}
}

func BytesEqual(t TestingT, a, b []byte) {
	if string(a) == string(b) {
		return // fast path if equal
	}

	reason := ""
	if len(a) != len(b) {
		reason = fmt.Sprintf("len(a) != len(b), %d != %d", len(a), len(b))
		goto failed
	}

	for i, v := range a {
		if v != b[i] {
			reason = fmt.Sprintf("a[%d] != b[%d]\na\n%s\nb\n%s", i, i, hexdumpAt(a, i), hexdumpAt(b, i))
			goto failed
		}
	}
failed:
	t.Helper()
	t.Fatalf("bytes not equal: %s", reason)
}

func SliceEqual[T comparable](t TestingT, a, b []T) {
	reason := ""
	if len(a) != len(b) {
		reason = fmt.Sprintf("len(a) != len(b), %d != %d", len(a), len(b))
		goto failed
	}

	for i, v := range a {
		if v != b[i] {
			reason = fmt.Sprintf("a[%d] != b[%d], %v != %v", i, i, a[i], b[i])
			goto failed
		}
	}
	return

failed:
	t.Helper()
	t.Fatalf("slice not equal: %s", reason)
}

func SliceNotEqual[T comparable](t TestingT, a, b []T) {
	if len(a) != len(b) {
		return
	}
	for i, v := range a {
		if v != b[i] {
			return
		}
	}
	t.Helper()
	t.Fatalf("slice equal")
}

func MapEqual[K, V comparable](t TestingT, a, b map[K]V) {
	reason := ""
	if len(a) != len(b) {
		reason = fmt.Sprintf("len(a) != len(b), %d != %d", len(a), len(b))
		goto failed
	}

	for k, v := range a {
		vb, ok := b[k]
		if !ok {
			reason = fmt.Sprintf("%v not found in b", k)
			goto failed
		}
		if v != vb {
			reason = fmt.Sprintf("a[%v] != b[%v], %v != %v", k, k, v, vb)
			goto failed
		}
	}
	return

failed:
	t.Helper()
	t.Fatalf("map not equal: %s", reason)
}

func StringContains(t TestingT, s, substr string) {
	if !strings.Contains(s, substr) {
		t.Helper()
		t.Fatalf("string %q not contains: %q", s, substr)
	}
}

func ErrorContains(t TestingT, err error, s string) {
	if err == nil || !strings.Contains(err.Error(), s) {
		t.Helper()
		t.Fatalf("err %v not contains: %q", err, s)
	}
}

func NoError(t TestingT, err error) {
	if err != nil {
		t.Helper()
		t.Fatalf("err: %s", err)
	}
}
