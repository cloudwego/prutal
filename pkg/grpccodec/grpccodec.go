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
	"google.golang.org/grpc/encoding"
	_ "google.golang.org/grpc/encoding/proto" // force the pkg init before this package

	"github.com/cloudwego/prutal"
)

func init() {
	encoding.RegisterCodec(&protoCodec{})
}

// protoCodec implements google.golang.org/grpc/encoding.Codec
type protoCodec struct{}

var _ encoding.Codec = (*protoCodec)(nil)

func (p *protoCodec) Marshal(v any) ([]byte, error) {
	return prutal.MarshalAppend(nil, v)
}

func (p *protoCodec) Unmarshal(data []byte, v any) error {
	return prutal.Unmarshal(data, v)
}

func (p *protoCodec) Name() string {
	// same as https://pkg.go.dev/google.golang.org/grpc/encoding/proto
	// so that we can overwrites the default proto codec
	return "proto"
}
