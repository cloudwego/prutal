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

const (
	editionProto2 = "proto2"
	editionProto3 = "proto3"
	edition2023   = "2023"
)

type GenPathType string

const (
	GenByImport         GenPathType = "import"
	GenBySourceRelative GenPathType = "source_relative"
)

const ( // implemented options of github.com/gogo/protobuf
	gogoproto_nullable = "(gogoproto.nullable)"

	gogoproto_enum_prefix     = "(gogoproto.goproto_enum_prefix)"
	gogoproto_enum_prefix_all = "(gogoproto.goproto_enum_prefix_all)"

	gogoproto_goproto_unrecognized     = "(gogoproto.goproto_unrecognized)"
	gogoproto_goproto_unrecognized_all = "(gogoproto.goproto_unrecognized_all)"
)

const (
	// editions features
	// see: https://protobuf.dev/editions/features/

	f_repeated_field_encoding = "features.repeated_field_encoding"
	f_field_presence          = "features.field_presence"
)

const ( // proto2 options
	option_packed = "packed"
)

// https://protobuf.dev/programming-guides/proto3/#scalar
var scalar2GoTypes = map[string]string{
	"double":   "float64",
	"float":    "float32",
	"int32":    "int32",
	"int64":    "int64",
	"uint32":   "uint32",
	"uint64":   "uint64",
	"sint32":   "int32",
	"sint64":   "int64",
	"fixed32":  "uint32",
	"fixed64":  "uint64",
	"sfixed32": "int32",
	"sfixed64": "int64",
	"bool":     "bool",
	"string":   "string",
	"bytes":    "[]byte",
}

var scalar2encodingType = map[string]string{
	"double":   "fixed64",
	"float":    "fixed32",
	"int32":    "varint",
	"int64":    "varint",
	"uint32":   "varint",
	"uint64":   "varint",
	"sint32":   "zigzag32",
	"sint64":   "zigzag64",
	"fixed32":  "fixed32",
	"fixed64":  "fixed64",
	"sfixed32": "fixed32",
	"sfixed64": "fixed64",
	"bool":     "varint",
	"string":   "bytes",
	"bytes":    "bytes",
}

// `packed` for any scalar type that is not string or bytes
var scalarPackedTypes = map[string]bool{
	"double":   true,
	"float":    true,
	"int32":    true,
	"int64":    true,
	"uint32":   true,
	"uint64":   true,
	"sint32":   true,
	"sint64":   true,
	"fixed32":  true,
	"fixed64":  true,
	"sfixed32": true,
	"sfixed64": true,
	"bool":     true,
}
