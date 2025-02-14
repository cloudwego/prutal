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

func TestOption(t *testing.T) {
	p := Options{
		{Name: "k", Value: "v1"},
		{Name: "k", Value: "v2"},
	}
	s, ok := p.Get("k")
	assert.True(t, ok)
	assert.Equal(t, "v2", s)
	assert.True(t, p.Is("k", "v2"))
	assert.False(t, p.Is("k", "v1"))
	s, ok = p.Get("x")
	assert.False(t, ok)
	assert.Equal(t, "", s)
	assert.False(t, p.Is("x", ""))
	t.Log(p.String())
}

func TestLoader_Option(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "testoption";

option (prutal.test.proto) = "o1";

message M {
  option (prutal.test.message) = "o2";

	oneof test_oneof {
		option (prutal.test.oneof) = "o3";
		string name = 2;
		string nick = 3;
	}
}

enum TestEnum {
	option (prutal.test.enum) = "o4";
	Z = 0;
}

service echo_service {
	option (prutal.test.service) = "o5";
	rpc echo (M) returns (M) {
		option (prutal.test.rpc) = "o6";
	};
}

`)

	v, ok := p.Options.Get("(prutal.test.proto)")
	assert.True(t, ok)
	assert.Equal(t, "o1", v)

	m := p.Messages[0]
	v, ok = m.Options.Get("(prutal.test.message)")
	assert.True(t, ok)
	assert.Equal(t, "o2", v)

	o := m.Oneofs[0]
	v, ok = o.Options.Get("(prutal.test.oneof)")
	assert.True(t, ok)
	assert.Equal(t, "o3", v)

	e := p.Enums[0]
	v, ok = e.Options.Get("(prutal.test.enum)")
	assert.True(t, ok)
	assert.Equal(t, "o4", v)

	s := p.Services[0]
	v, ok = s.Options.Get("(prutal.test.service)")
	assert.True(t, ok)
	assert.Equal(t, "o5", v)

	r := s.Methods[0]
	v, ok = r.Options.Get("(prutal.test.rpc)")
	assert.True(t, ok)
	assert.Equal(t, "o6", v)

}
