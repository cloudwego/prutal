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
	"strconv"
)

// FieldType represents types used in struct tag
type TagType uint8

const (
	TypeVarint TagType = iota + 1
	TypeZigZag32
	TypeZigZag64
	TypeFixed32
	TypeFixed64
	TypeBytes
	TypeUnknown
)

var typeNames = []string{
	TypeVarint:   "varint",
	TypeZigZag32: "zigzag32",
	TypeZigZag64: "zigzag64",
	TypeFixed32:  "fixed32",
	TypeFixed64:  "fixed64",
	TypeBytes:    "bytes",
}

func (t TagType) String() string {
	ret := ""
	if uint(t) < uint(len(typeNames)) {
		ret = typeNames[uint(t)]
	}
	if ret == "" {
		ret = "FieldType-" + strconv.Itoa(int(t))
	}
	return ret
}
