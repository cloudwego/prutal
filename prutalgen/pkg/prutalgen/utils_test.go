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
	"io"
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestHasPathPrefix(t *testing.T) {
	assert.True(t, hasPathPrefix("a.b.c", "a.b"))
	assert.False(t, hasPathPrefix("a.b", "a.b"))
	assert.False(t, hasPathPrefix("a.bc", "a.b"))
	assert.False(t, hasPathPrefix("a", "b"))
}

func TestJoinErrs(t *testing.T) {
	assert.NoError(t, joinErrs())
	assert.Same(t, io.EOF, joinErrs(io.EOF))
	assert.Equal(t,
		io.EOF.Error()+"\n"+io.ErrUnexpectedEOF.Error(),
		joinErrs(io.EOF, io.ErrUnexpectedEOF).Error())
}
