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
	"errors"
	"fmt"

	"github.com/cloudwego/prutal/internal/desc"
	"github.com/cloudwego/prutal/internal/wire"
)

const defaultRecursionMaxDepth = 1000

var (
	errMaxDepthExceeded = errors.New("max depth exceeded")
)

func newWireTypeNotMatch(t0 wire.Type, t1 desc.TagType) error {
	return fmt.Errorf("wire type %s not match %s", t0, t1)
}
