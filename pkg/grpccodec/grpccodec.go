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

package grpccodec

import (
	"github.com/cloudwego/prutal"
)

// PrutalCodec is a protobuf codec backed by prutal.
//
// It intentionally works with raw byte slices so this package stays independent
// from google.golang.org/grpc. Its method set matches the v1
// google.golang.org/grpc/encoding.Codec interface, so applications can use it
// directly with APIs expecting that shape or wrap it for CodecV2.
type PrutalCodec struct{}

// Marshal encodes a protobuf message into the standard binary wire format.
func (PrutalCodec) Marshal(v any) ([]byte, error) {
	return prutal.MarshalAppend(nil, v)
}

// Unmarshal decodes a protobuf binary payload into v.
func (PrutalCodec) Unmarshal(data []byte, v any) error {
	return prutal.Unmarshal(data, v)
}

// Name returns "proto" so this codec can replace gRPC's default protobuf codec.
func (PrutalCodec) Name() string {
	return "proto"
}
