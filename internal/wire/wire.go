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

package wire

import (
	"errors"

	"github.com/cloudwego/prutal/internal/protowire"
)

type Type = protowire.Type

const ( // align with protowire.Type
	TypeVarint  Type = 0
	TypeFixed32 Type = 5
	TypeFixed64 Type = 1
	TypeBytes   Type = 2
	TypeSGroup  Type = 3
	TypeEGroup  Type = 4
)

// Map entries are encoded exactly like a message with two fields:
// #1 for the key and #2 for the value.
const (
	MapKeyFieldNum = 1
	MapValFieldNum = 2
)

var errFieldNumber = errors.New("invalid field number")

// ConsumeMapEntryTag reads one map entry field tag.
//
// protowire.ConsumeTag mirrors the reference package and only rejects field
// numbers below the minimum; the reference decoders range-check the upper
// bound themselves before using the number. Doing the same here rejects such
// an entry instead of quietly filing the field away as unknown.
func ConsumeMapEntryTag(b []byte) (protowire.Number, Type, int, error) {
	num, typ, n := protowire.ConsumeTag(b)
	if n < 0 {
		return 0, 0, 0, protowire.ParseError(n)
	}
	if num > protowire.MaxValidNumber {
		return 0, 0, 0, errFieldNumber
	}
	return num, typ, n, nil
}

// EncodeTag ...
//
// see: https://protobuf.dev/programming-guides/encoding/#structure
func EncodeTag(num int32, t Type) uint64 {
	return uint64(num)<<3 | uint64(t)
}

// AppendKeyTag ... faster version of AppendVarint(b, EncodeTag(1, t))
func AppendKeyTag(b []byte, t Type) []byte {
	return append(b, byte(MapKeyFieldNum)<<3|byte(t))
}

// AppendValTag ... faster version of AppendVarint(b, EncodeTag(2, t))
func AppendValTag(b []byte, t Type) []byte {
	return append(b, byte(MapValFieldNum)<<3|byte(t))
}

func SizeVarint(v uint64) int {
	return protowire.SizeVarint(v)
}

// CoderType identifies a protobuf field encoding strategy. It is used as a key
// to look up type-specialized encode, decode, and size functions (AppendFunc,
// SizeFunc, SizeMapFunc, etc.) so the hot path dispatches through a function
// pointer instead of a type switch.
type CoderType int8

const (
	CoderVarintU32 CoderType = 1 + iota // uint32 varint (also used for enum uint32)
	CoderVarintI32                      // int32 varint, sign-extended to 64-bit per protobuf spec
	CoderVarint64                       // uint64 varint
	CoderZigZag32                       // sint32 zigzag encoding
	CoderZigZag64                       // sint64 zigzag encoding
	CoderFixed32                        // fixed32 / sfixed32 / float (4-byte little-endian)
	CoderFixed64                        // fixed64 / sfixed64 / double (8-byte little-endian)
	CoderBytes                          // bytes (length-prefixed)
	CoderString                         // string (length-prefixed)
	CoderBool                           // bool (single varint byte)
	CoderStruct                         // embedded message (length-prefixed, recursive)
	CoderUnknown                        // unsupported or unresolved type
)
