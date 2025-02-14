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

func TestLoader_Service(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "echo";
message M {
  string Msg = 1;
}

service echo_service {
	rpc echo (M) returns (M);
}`)
	s := p.Services[0]
	assert.Equal(t, "echo_service", s.Name)
	assert.Equal(t, "EchoService", s.GoName)

	rpc := s.Methods[0]
	assert.Equal(t, "echo", rpc.Name)
	assert.Equal(t, "Echo", rpc.GoName)
	assert.Same(t, p.Messages[0], rpc.Request.Message())
	assert.Same(t, p.Messages[0], rpc.Return.Message())
	assert.False(t, rpc.RequestStream)
	assert.False(t, rpc.ReturnStream)

	p = loadTestProto(t, `
option go_package = "echo";
message M {
  string Msg = 1;
}

service echo_service {
	rpc echo (stream M) returns (stream M);
}`)

	rpc = p.Services[0].Methods[0]
	assert.True(t, rpc.RequestStream)
	assert.True(t, rpc.ReturnStream)

}
