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
	"github.com/cloudwego/prutal/internal/prutal"
)

// MarshalAppend appends the protobuf encoding of v to b and returns the new bytes
func MarshalAppend(b []byte, v interface{}) ([]byte, error) {
	return prutal.MarshalAppend(b, v)
}

// Marshal is alias of MarshalAppend(nil, v).
//
// You should consider using MarshalAppend for performance concerns
func Marshal(v interface{}) ([]byte, error) {
	return prutal.MarshalAppend(nil, v)
}

// Unmarshal parses the protobuf-encoded data and stores the result in the value pointed to by v.
func Unmarshal(b []byte, v interface{}) error {
	return prutal.Unmarshal(b, v)
}
