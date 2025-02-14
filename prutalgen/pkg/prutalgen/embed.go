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

import _ "embed"

var (
	//go:embed wellknowns/any.proto
	any_proto []byte

	//go:embed wellknowns/api.proto
	api_proto []byte

	//go:embed wellknowns/descriptor.proto
	descriptor_proto []byte

	//go:embed wellknowns/duration.proto
	duration_proto []byte

	//go:embed wellknowns/empty.proto
	empty_proto []byte

	//go:embed wellknowns/field_mask.proto
	field_mask_proto []byte

	//go:embed wellknowns/source_context.proto
	source_context_proto []byte

	//go:embed wellknowns/struct.proto
	struct_proto []byte

	//go:embed wellknowns/timestamp.proto
	timestamp_proto []byte

	//go:embed wellknowns/type.proto
	type_proto []byte

	//go:embed wellknowns/wrappers.proto
	wrappers_proto []byte
)

var embeddedProtos = map[string][]byte{}

func RegisterEmbeddedProto(proto string, b []byte) {
	embeddedProtos[proto] = b
}

func init() {
	type protoFile struct {
		Name string
		Data []byte
	}
	wellknowns := []protoFile{
		{"google/protobuf/any.proto", any_proto},
		{"google/protobuf/api.proto", api_proto},
		{"google/protobuf/descriptor.proto", descriptor_proto},
		{"google/protobuf/duration.proto", duration_proto},
		{"google/protobuf/empty.proto", empty_proto},
		{"google/protobuf/field_mask.proto", field_mask_proto},
		{"google/protobuf/source_context.proto", source_context_proto},
		{"google/protobuf/struct.proto", struct_proto},
		{"google/protobuf/timestamp.proto", timestamp_proto},
		{"google/protobuf/type.proto", type_proto},
		{"google/protobuf/wrappers.proto", wrappers_proto},
	}
	for _, f := range wellknowns {
		RegisterEmbeddedProto(f.Name, f.Data)
	}
}
