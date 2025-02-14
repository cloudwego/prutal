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

package desc

import (
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestMapStructDescItem(t *testing.T) {
	m := newMapStructDesc()
	k0 := uintptr(1)
	v0 := &StructDesc{}
	m.Set(k0, v0)
	assert.Same(t, v0, m.Get(k0))

	k1 := uintptr(mapStructDescBuckets + 2) // k1 & mapStructDescBuckets == k0
	v1 := &StructDesc{}
	m.Set(k1, v1)
	assert.Same(t, v1, m.Get(k1))
	assert.Same(t, v0, m.Get(k0))

	k2 := uintptr(2)
	v2 := &StructDesc{}
	m.Set(uintptr(2), v2)
	assert.Same(t, v2, m.Get(k2))
	assert.Same(t, v1, m.Get(k1))
	assert.Same(t, v0, m.Get(k0))

}
