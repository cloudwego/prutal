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

package edition2023

import (
	"testing"

	"github.com/cloudwego/prutal"
	"github.com/cloudwego/prutal/internal/testutils/assert"
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"google.golang.org/protobuf/proto"
)

func TestEdition2023(t *testing.T) {
	p := &TestAllTypesEdition2023{}
	err := faker.FakeData(p, options.WithIgnoreInterface(true), options.WithRandomMapAndSliceMaxSize(33))
	assert.NoError(t, err)

	bs, err := proto.Marshal(p)
	assert.NoError(t, err)

	p0 := &TestAllTypesEdition2023{}
	err = prutal.Unmarshal(bs, p0)
	assert.NoError(t, err)

	assert.DeepEqual(t, p, p0)

	bs, err = prutal.Marshal(p)
	assert.NoError(t, err)

	p0 = &TestAllTypesEdition2023{}
	err = proto.Unmarshal(bs, p0)
	assert.NoError(t, err)

	assert.DeepEqual(t, p, p0)

	// Size test: prutal.Size must match len(prutal.Marshal)
	bs, err = prutal.Marshal(p)
	assert.NoError(t, err)
	sz, err := prutal.Size(p)
	assert.NoError(t, err)
	assert.Equal(t, len(bs), sz)
}
