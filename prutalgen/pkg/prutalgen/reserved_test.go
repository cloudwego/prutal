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
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestReservedRange(t *testing.T) {
	rr := reservedRanges{}
	rr = append(rr, reservedRange{1, 2}, reservedRange{5, 6})
	assert.False(t, rr.In(0))
	assert.True(t, rr.In(1))
	assert.True(t, rr.In(2))
	assert.False(t, rr.In(3))
	assert.False(t, rr.In(4))
	assert.True(t, rr.In(5))
	assert.True(t, rr.In(6))
	assert.False(t, rr.In(7))
	assert.True(t, rr.In(19000))
	assert.True(t, rr.In(19999))
}

func TestLoader_Reserved(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "testmessage";
message M {
  string f = 1;

	reserved 3,5;
	reserved 7 to 10;
	reserved 100 to max;

  enum e {
	reserved 30,50;
	reserved 70 to 100;
	reserved 1000 to max;
  }
}
`)

	m := p.Messages[0]
	type testcase struct {
		f int32
		v bool
	}

	{ // (*Message) IsReservedField
		cases := []testcase{
			{2, false},
			{3, true},
			{4, false},
			{5, true},
			{6, false},
			{7, true},
			{8, true},
			{10, true},
			{11, false},
			{99, false},
			{100, true},
			{101, true},
			{1000000, true},
		}
		for _, c := range cases {
			assert.Equal(t, c.v, m.IsReservedField(c.f), c.f)
		}
	}

	{ // (*Enum) IsReservedField
		e := m.Enums[0]
		cases := []testcase{
			{20, false},
			{30, true},
			{40, false},
			{50, true},
			{60, false},
			{70, true},
			{80, true},
			{100, true},
			{110, false},
			{990, false},
			{1000, true},
			{1001, true},
			{10000000, true},
		}
		for _, c := range cases {
			assert.Equal(t, c.v, e.IsReservedField(c.f), c.f)
		}
	}

}
