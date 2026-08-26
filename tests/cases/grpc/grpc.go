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

package echo

import (
	context "context"
	"sync/atomic"

	"github.com/cloudwego/prutal/pkg/grpccodec"
	"google.golang.org/grpc/encoding"
)

type EchoServer struct {
	UnimplementedEchoServiceServer
}

var _ encoding.Codec = grpccodec.PrutalCodec{}

// CountingCodec verifies that gRPC actually routes message encoding through
// grpccodec.PrutalCodec instead of the default protobuf codec.
type CountingCodec struct {
	grpccodec.PrutalCodec
	marshalCount   atomic.Int32
	unmarshalCount atomic.Int32
}

func (c *CountingCodec) Marshal(v any) ([]byte, error) {
	c.marshalCount.Add(1)
	return c.PrutalCodec.Marshal(v)
}

func (c *CountingCodec) Unmarshal(data []byte, v any) error {
	c.unmarshalCount.Add(1)
	return c.PrutalCodec.Unmarshal(data, v)
}

func (s *EchoServer) Echo(ctx context.Context, req *EchoRequest) (*EchoResponse, error) {
	return &EchoResponse{Message: req.Message}, nil
}
