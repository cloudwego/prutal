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

package nested

import (
	"testing"

	"github.com/cloudwego/prutal"
	"github.com/cloudwego/prutal/internal/testutils/assert"
	"google.golang.org/protobuf/proto"
)

func TestNested(t *testing.T) {
	a := &TestMessageA{V: 7}
	b := &TestMessageB{
		NestedA:     &TestMessageA{V: 100},
		NestedB:     &TestMessageB{V: 101},
		NestedListA: []*TestMessageA{&TestMessageA{V: 102}},
		NestedListB: []*TestMessageB{&TestMessageB{V: 103}},
		NestedMapA:  map[int64]*TestMessageA{104: &TestMessageA{V: 105}},
		NestedMapB:  map[int64]*TestMessageB{106: &TestMessageB{V: 107}},
		V:           110,
	}
	p := &TestMessageA{
		NestedA:     a,
		NestedB:     b,
		NestedListA: []*TestMessageA{a},
		NestedListB: []*TestMessageB{b},
		NestedMapA:  map[int64]*TestMessageA{201: a},
		NestedMapB:  map[int64]*TestMessageB{202: b},
		V:           203,
	}

	// Unmarshal test

	// use proto to encode
	bs, err := proto.Marshal(p)
	assert.NoError(t, err)

	// use prutal to decode
	p0 := &TestMessageA{}
	err = prutal.Unmarshal(bs, p0)
	assert.NoError(t, err)
	assert.DeepEqual(t, p, p0)

	// Marshal test

	// use prutal to encode
	bs, err = prutal.Marshal(p)
	assert.NoError(t, err)

	// use proto to code
	p0 = &TestMessageA{}
	err = proto.Unmarshal(bs, p0)
	assert.NoError(t, err)

	assert.DeepEqual(t, p, p0)

	// Size test
	bs, err = prutal.Marshal(p)
	assert.NoError(t, err)
	sz, err := prutal.Size(p)
	assert.NoError(t, err)
	assert.Equal(t, len(bs), sz)
}
