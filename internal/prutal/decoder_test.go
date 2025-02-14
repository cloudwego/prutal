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

package prutal

import (
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
	"github.com/cloudwego/prutal/internal/wire"
)

func TestDecodeOneof(t *testing.T) {
	n := 0x0000ffff
	s := "helloworld"
	tmp := wire.Builder{}
	buf := wire.Builder{}
	buf.AppendVarintField(2, uint64(n)).
		AppendStringField(4, s).
		AppendBytesField(5, tmp.AppendVarintField(1, 1).Bytes())
	b := buf.Bytes()
	p := &TestOneofMessage{}
	err := Unmarshal(b, p)
	assert.NoError(t, err)

	f2, ok := p.OneOfFieldA.(*TestOneofMessage_Field2)
	assert.True(t, ok)
	assert.Equal(t, int64(n), f2.Field2)
	f4, ok := p.OneOfFieldB.(*TestOneofMessage_Field4)
	assert.True(t, ok)
	assert.Equal(t, s, f4.Field4)

	f5, ok := p.OneOfFieldC.(*TestOneofMessage_Field5)
	assert.True(t, ok)
	assert.Equal(t, true, f5.Field5.Field1)
}
