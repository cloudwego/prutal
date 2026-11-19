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

// Package prutal provides reflection-based protobuf marshaling and
// unmarshaling.
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

// Size returns the size in bytes of the wire-format encoding of v.
func Size(v interface{}) (int, error) {
	return prutal.Size(v)
}

// Unmarshal resets v and parses the protobuf-encoded data into the value pointed to by v.
// If decoding fails, v has been reset and may contain partially decoded fields.
func Unmarshal(b []byte, v interface{}) error {
	return prutal.Unmarshal(b, v)
}

// UnmarshalMerge parses the protobuf-encoded data into the value pointed to by v
// without resetting v first, matching proto.UnmarshalOptions{Merge: true}:
// singular fields present on the wire overwrite existing values, repeated fields
// append, map entries overwrite matching keys, singular and oneof message fields
// merge recursively, and unknown fields append. On error, v may contain partially
// merged fields.
//
// Callers passing a zero or pre-reset struct can use UnmarshalMerge to avoid the
// reset cost of Unmarshal. For repeated non-pointer struct fields ([]T), any spare
// capacity (len < cap) is decoded into without zeroing, so pre-reset slices must
// clear or zero spare capacity (e.g. s = nil, s = s[:0:0], or zeroing s[:cap(s)]
// before s = s[:0]).
func UnmarshalMerge(b []byte, v interface{}) error {
	return prutal.UnmarshalMerge(b, v)
}
