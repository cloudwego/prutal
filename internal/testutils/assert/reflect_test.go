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

import "testing"

func TestDeepEqual(t *testing.T) {
	type Msg struct {
		Data string
	}
	type TestStruct struct {
		unexported int

		M *Msg

		List []int
		Map  map[int]int

		B bool
		U uint
		F float32
	}

	type testcase struct {
		name   string
		p0, p1 any
		ok     bool
	}

	m := &mockTestingT{}
	DeepEqual(m, nil, nil)
	m.CheckPassed(t)

	testcases := []testcase{
		{
			name: "both-nil",
			p0:   nil,
			p1:   nil,
			ok:   true,
		},
		{
			name: "one-nil",
			p0:   nil,
			p1:   &TestStruct{},
			ok:   false,
		},
		{
			name: "type-not-equal",
			p0:   &TestStruct{},
			p1:   &Msg{},
			ok:   false,
		},
		{
			name: "embbedstruct",
			p0:   &TestStruct{M: &Msg{Data: "hi"}},
			p1:   &TestStruct{M: &Msg{Data: "hi"}},
			ok:   true,
		},
		{
			name: "list-ok",
			p0:   &TestStruct{List: []int{1}},
			p1:   &TestStruct{List: []int{1}},
			ok:   true,
		},
		{
			name: "list-not-equal",
			p0:   &TestStruct{List: []int{1}},
			p1:   &TestStruct{List: []int{2}},
			ok:   false,
		},
		{
			name: "list-not-len",
			p0:   &TestStruct{List: []int{1}},
			p1:   &TestStruct{List: []int{1, 1}},
			ok:   false,
		},
		{
			name: "map-ok",
			p0:   &TestStruct{Map: map[int]int{1: 2}},
			p1:   &TestStruct{Map: map[int]int{1: 2}},
			ok:   true,
		},
		{
			name: "map-not-equal",
			p0:   &TestStruct{Map: map[int]int{1: 2}},
			p1:   &TestStruct{Map: map[int]int{1: 3}},
			ok:   false,
		},
		{
			name: "map-not-len",
			p0:   &TestStruct{Map: map[int]int{1: 1}},
			p1:   &TestStruct{Map: map[int]int{1: 1, 2: 2}},
			ok:   false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			m.Reset()
			DeepEqual(m, tc.p0, tc.p1)
			if tc.ok {
				m.CheckPassed(t)
			} else {
				m.CheckFailed(t)
			}
		})
	}

}
