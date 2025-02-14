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

func TestLoader_Oneof(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "testoneof";
message M {
	oneof test_oneof {
		string name = 1;
		string nick = 2;
	}
}
`)

	o := p.Messages[0].Oneofs[0]
	assert.Equal(t, "TestOneof", o.FieldName())
	assert.Equal(t, "isM_TestOneof", o.FieldType())
	assert.Equal(t, 2, len(o.Fields))
}
