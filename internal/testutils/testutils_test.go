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
