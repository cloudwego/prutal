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

package wire

import (
	"testing"
	"unsafe"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

// Map entry: key (field 1) = 1 (canonical), value (field 2) = the over-long
// varint 0x80 0x00 (= 0, i.e. false). The pre-fix code read only the first
// byte (0x80 != 0) and wrongly returned true. Bool map keys/values are
// varints on the wire, not single bytes.
func TestDecodeMap_Bool_Bool(t *testing.T) {
	b := []byte{0x08, 0x01, 0x10, 0x80, 0x00}

	m := map[bool]bool{}
	assert.NoError(t, DecodeMap_Bool_Bool(b, unsafe.Pointer(&m)))
	assert.MapEqual(t, map[bool]bool{true: false}, m)
}
