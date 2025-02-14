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
	"crypto/rand"
	"errors"
	"io"
	"testing"
)

type mockTestingT struct {
	fatal bool
}

func (m *mockTestingT) Reset()  { *m = mockTestingT{} }
func (m *mockTestingT) Helper() {}
func (m *mockTestingT) Fatalf(format string, args ...interface{}) {
	m.fatal = true
}

func (m *mockTestingT) CheckPassed(t *testing.T) {
	if m.fatal {
		t.Helper()
		t.Fatal("Fatal called")
	}
}

func (m *mockTestingT) CheckFailed(t *testing.T) {
	if !m.fatal {
		t.Helper()
		t.Fatal("Fatal not called")
	}
}

func TestTrueFalse(t *testing.T) {
	m := &mockTestingT{}

	m.Reset()
	True(m, true)
	m.CheckPassed(t)

	m.Reset()
	True(m, false)
	m.CheckFailed(t)

	m.Reset()
	False(m, false)
	m.CheckPassed(t)

	m.Reset()
	False(m, true)
	m.CheckFailed(t)
}

func TestEqual(t *testing.T) {
	m := &mockTestingT{}

	m.Reset()
	Equal(m, 1, 1)
	m.CheckPassed(t)

	m.Reset()
	Equal(m, 1, 2)
	m.CheckFailed(t)

	m.Reset()
	DeepEqual(m, 1, 1)
	m.CheckPassed(t)

	m.Reset()
	DeepEqual(m, 1, 2)
	m.CheckFailed(t)
}

func TestSame(t *testing.T) {
	m := &mockTestingT{}

	m.Reset()
	Same(m, io.EOF, io.EOF)
	m.CheckPassed(t)

	m.Reset()
	v := 1
	var x interface{}
	var y *int
	x = &v
	y = &v
	Same(m, x, y)
	m.CheckPassed(t)

	m.Reset()
	x = t
	Same(m, x, y)
	m.CheckFailed(t)

	m.Reset()
	Same(m, 1, 1) // not pointer
	m.CheckFailed(t)
}

func TestStringContains(t *testing.T) {
	m := &mockTestingT{}

	m.Reset()
	StringContains(m, "helloworld", "world")
	m.CheckPassed(t)

	m.Reset()
	StringContains(m, "helloworld", "main")
	m.CheckFailed(t)
}

func TestErrorContains(t *testing.T) {
	m := &mockTestingT{}

	m.Reset()
	ErrorContains(m, errors.New("helloworld"), "world")
	m.CheckPassed(t)

	m.Reset()
	ErrorContains(m, nil, "main")
	m.CheckFailed(t)
}

func TestNoError(t *testing.T) {
	m := &mockTestingT{}

	m.Reset()
	NoError(m, nil)
	m.CheckPassed(t)

	m.Reset()
	NoError(m, errors.New(""))
	m.CheckFailed(t)
}

func TestSliceEqual(t *testing.T) {
	m := &mockTestingT{}

	m.Reset()
	SliceEqual(m, []int{1}, []int{1})
	m.CheckPassed(t)

	m.Reset()
	SliceEqual(m, []int{1}, []int{2})
	m.CheckFailed(t)

	m.Reset()
	SliceEqual(m, []int{1}, []int{1, 1})
	m.CheckFailed(t)
}

func TestSliceNotEqual(t *testing.T) {
	m := &mockTestingT{}

	m.Reset()
	SliceNotEqual(m, []int{1}, []int{2})
	m.CheckPassed(t)

	m.Reset()
	SliceNotEqual(m, []int{1}, []int{1, 1})
	m.CheckPassed(t)

	m.Reset()
	SliceNotEqual(m, []int{1, 1}, []int{1, 1})
	m.CheckFailed(t)
}

func TestBytesEqual(t *testing.T) {
	m := &mockTestingT{}

	m.Reset()
	BytesEqual(m, []byte{}, []byte{})
	m.CheckPassed(t)

	a := make([]byte, 100)
	b := make([]byte, 100)
	_, _ = rand.Read(a)
	_, _ = rand.Read(b)
	m.Reset()
	BytesEqual(m, a, b)
	m.CheckFailed(t)

	m.Reset()
	BytesEqual(m, []byte{}, []byte{1})
	m.CheckFailed(t)
}

func TestMapEqual(t *testing.T) {
	m := &mockTestingT{}

	m.Reset()
	MapEqual(m, map[int]int{1: 1}, map[int]int{1: 1})
	m.CheckPassed(t)

	m.Reset()
	MapEqual(m, map[int]int{1: 1}, map[int]int{})
	m.CheckFailed(t)

	m.Reset()
	MapEqual(m, map[int]int{1: 1}, map[int]int{2: 2})
	m.CheckFailed(t)

	m.Reset()
	MapEqual(m, map[int]int{1: 1}, map[int]int{1: 2})
	m.CheckFailed(t)
}
