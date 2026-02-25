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
	"sync"
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func sd() *StructDesc { return &StructDesc{} }

func TestMapStructDesc(t *testing.T) {
	m := &mapStructDesc{}

	// empty get
	assert.True(t, m.Get(0) == nil)
	assert.True(t, m.Get(1) == nil)

	// set & get
	v0, v1, v2 := sd(), sd(), sd()
	m.Set(1, v0)
	m.Set(2, v1)
	assert.Same(t, v0, m.Get(1))
	assert.Same(t, v1, m.Get(2))
	assert.True(t, m.Get(99) == nil)

	// hash collision: same bucket, different key
	k3 := uintptr(mapStructDescBuckets + 2) // bucket == 1, same as key 1
	m.Set(k3, v2)
	assert.Same(t, v2, m.Get(k3))
	assert.Same(t, v0, m.Get(1))

	// update existing key
	v0new := sd()
	m.Set(1, v0new)
	assert.Same(t, v0new, m.Get(1))
	assert.Same(t, v1, m.Get(2))
	assert.Same(t, v2, m.Get(k3))

	// set same value is a no-op
	m.Set(1, v0new)
	assert.Same(t, v0new, m.Get(1))
}

func TestMapStructDescConcurrent(t *testing.T) {
	m := &mapStructDesc{}
	n := 100
	vals := make([]*StructDesc, n)
	for i := range vals {
		vals[i] = sd()
		m.Set(uintptr(i), vals[i])
	}

	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < n; i++ {
				if got := m.Get(uintptr(i)); got != vals[i] {
					t.Errorf("key %d: got %p, want %p", i, got, vals[i])
				}
			}
		}()
	}
	wg.Wait()
}
