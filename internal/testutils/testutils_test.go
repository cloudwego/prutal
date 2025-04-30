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

package testutils

import (
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestRepeat(t *testing.T) {
	assert.SliceEqual(t, []int{1, 1, 1}, Repeat(3, 1))
}

func TestRandomBoolSlice(t *testing.T) {
	vv := RandomBoolSlice(100)
	assert.SliceNotEqual(t, vv, Repeat(100, true))
	assert.SliceNotEqual(t, vv, Repeat(100, false))
}

func TestRandomStr(t *testing.T) {
	_ = RandomStr(100)
}

func TestRandomBytes(t *testing.T) {
	vv := RandomBytes(100)
	assert.SliceNotEqual(t, vv, Repeat(100, byte(0)))
	assert.SliceNotEqual(t, vv, Repeat(100, byte(1)))
}

func TestRandFill(t *testing.T) {
	oo := DefaultFillOptions()

	// Basic types
	t.Run("BasicTypes", func(t *testing.T) {
		var i int
		var u uint
		var f float64
		var b bool
		var c complex128
		var s string

		RandFill(&i, oo)
		RandFill(&u, oo)
		RandFill(&f, oo)
		RandFill(&b, oo)
		RandFill(&c, oo)

		RandFill(&s, oo)
		assert.True(t, len(s) >= oo.StringMinLen && len(s) <= oo.StringMaxLen, len(s))
	})

	// Slices
	t.Run("Slices", func(t *testing.T) {
		var ii []int
		RandFill(&ii, oo)
		assert.True(t, len(ii) >= oo.SliceMinSize && len(ii) <= oo.SliceMaxSize, len(ii))

		var bb []byte
		RandFill(&bb, oo)
		assert.True(t, len(bb) >= oo.StringMinLen && len(bb) <= oo.StringMaxLen, len(bb))

		// Not all zeros
		if len(ii) > 0 {
			allZeros := true
			for _, v := range ii {
				if v != 0 {
					allZeros = false
					break
				}
			}
			assert.False(t, allZeros)
		}
	})

	// Arrays
	t.Run("Arrays", func(t *testing.T) {
		var arr [5]int
		RandFill(&arr, oo)

		// Not all zeros
		allZeros := true
		for i := 0; i < 5; i++ {
			if arr[i] != 0 {
				allZeros = false
				break
			}
		}
		assert.False(t, allZeros)
	})

	// Maps
	t.Run("Maps", func(t *testing.T) {
		var m map[string]int
		RandFill(&m, oo)
		assert.True(t, m != nil)
		assert.True(t, len(m) >= oo.MapMinSize && len(m) <= oo.MapMaxSize, len(m))

		var m2 map[bool]int
		RandFill(&m2, oo)
		assert.True(t, m2 != nil)
	})

	// Structs
	t.Run("Structs", func(t *testing.T) {
		type TestStruct struct {
			A int
			B string
			C []float64
			D map[string]bool
		}

		var s TestStruct
		RandFill(&s, oo)

		assert.True(t, len(s.B) >= oo.StringMinLen && len(s.B) <= oo.StringMaxLen, len(s.B))
		assert.True(t, len(s.C) >= oo.SliceMinSize && len(s.C) <= oo.SliceMaxSize, len(s.C))
		assert.True(t, len(s.D) >= oo.MapMinSize && len(s.D) <= oo.MapMaxSize, len(s.D))
	})

	// Pointers
	t.Run("Pointers", func(t *testing.T) {
		var p *int
		RandFill(&p, oo)
		assert.True(t, p != nil)
	})

	// Recursion depth
	t.Run("RecursionDepth", func(t *testing.T) {
		type Node struct {
			Value int
			Next  *Node
		}

		var root Node
		customOpts := DefaultFillOptions()
		customOpts.RecursionDepth = 3

		RandFill(&root, customOpts)

		// Count depth
		depth := 0
		current := &root
		for current.Next != nil {
			depth++
			current = current.Next
		}
		assert.Equal(t, depth, customOpts.RecursionDepth)
	})

	// Deterministic with same seed
	t.Run("DeterministicWithSeed", func(t *testing.T) {
		var a, b int
		o1 := DefaultFillOptions()
		o1.Seed = 12345
		RandFill(&a, o1)

		o2 := DefaultFillOptions()
		o2.Seed = 12345
		RandFill(&b, o2)

		assert.Equal(t, a, b)
	})
}
