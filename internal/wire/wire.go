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

import "github.com/cloudwego/prutal/internal/protowire"

type Type = protowire.Type

const ( // align with protowire.Type
	TypeVarint  Type = 0
	TypeFixed32 Type = 5
	TypeFixed64 Type = 1
	TypeBytes   Type = 2
	TypeSGroup  Type = 3
	TypeEGroup  Type = 4
)

// ConsumeKVTag implements ConsumeTag for key and value of map
//
// for map pairs, num=1 for key, and num=2 for value.
// the max int should be 2<<3 + 15 = 31 which always < 0x80 (128)
func ConsumeKVTag(b []byte) (int32, Type) {
	if len(b) > 0 && uint64(b[0]) < 0x80 {
		return DecodeTag(uint64(b[0]))
	}
	return -1, -1
}

// EncodeTag ...
//
// see: https://protobuf.dev/programming-guides/encoding/#structure
func EncodeTag(num int32, t Type) uint64 {
	return uint64(num)<<3 | uint64(t)
}

// DecodeTag ...
func DecodeTag(v uint64) (int32, Type) {
	return int32(uint32(v >> 3)), Type(v & 0b111)
}

// AppendKeyTag ... faster version of AppendVarint(b, EncodeTag(1, t))
func AppendKeyTag(b []byte, t Type) []byte {
	return append(b, byte(1)<<3|byte(t))
}

// AppendValTag ... faster version of AppendVarint(b, EncodeTag(2, t))
func AppendValTag(b []byte, t Type) []byte {
	return append(b, byte(2)<<3|byte(t))
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
