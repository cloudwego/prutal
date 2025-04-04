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

package benchmark

import (
	"testing"

	"github.com/cloudwego/prutal"
	"github.com/cloudwego/prutal/internal/testutils/assert"

	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"google.golang.org/protobuf/proto"
)

type testmessage interface {
	proto.Message

	Reset()
}

func runBenchmark(p testmessage, b *testing.B) {
	var err error
	x := proto.MarshalOptions{}
	buf := make([]byte, 0, 16<<10)
	b.Run("encode-protobuf", func(b *testing.B) {
		for range b.N {
			buf, err = x.MarshalAppend(buf[:0], p)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("encode-prutal", func(b *testing.B) {
		for range b.N {
			buf, err = prutal.MarshalAppend(buf[:0], p)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	buf, err = x.MarshalAppend(buf[:0], p)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("decode-protobuf", func(b *testing.B) {
		for range b.N {
			p.Reset()
			err = proto.Unmarshal(buf, p)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("decode-prutal", func(b *testing.B) {
		for range b.N {
			p.Reset()
			err = prutal.Unmarshal(buf, p)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkScalarType(b *testing.B) {
	p := &ScalarTypeMessage{}
	err := faker.FakeData(p)
	assert.NoError(b, err)
	runBenchmark(p, b)
}

func BenchmarkScalarSlice(b *testing.B) {
	p := &ScalarSliceMessage{}
	err := faker.FakeData(p)
	assert.NoError(b, err)
	runBenchmark(p, b)
}

func BenchmarkScalarMap(b *testing.B) {
	p := &ScalarMapMessage{}
	err := faker.FakeData(p,
		options.WithRandomMapAndSliceMinSize(10),
		options.WithRandomMapAndSliceMaxSize(33))
	assert.NoError(b, err)
	runBenchmark(p, b)
}

func BenchmarkStringType(b *testing.B) {
	p := &StringTypeMessage{}
	err := faker.FakeData(p)
	assert.NoError(b, err)
	runBenchmark(p, b)
}

func BenchmarkStringSlice(b *testing.B) {
	p := &StringSliceMessage{}
	err := faker.FakeData(p,
		options.WithRandomMapAndSliceMinSize(10),
		options.WithRandomMapAndSliceMaxSize(33))
	assert.NoError(b, err)
	runBenchmark(p, b)
}

func BenchmarkStringMap(b *testing.B) {
	p := &StringMapMessage{}
	err := faker.FakeData(p,
		options.WithRandomMapAndSliceMinSize(10),
		options.WithRandomMapAndSliceMaxSize(33))
	assert.NoError(b, err)
	runBenchmark(p, b)
}

func BenchmarkStructSlice(b *testing.B) {
	p := &StructSliceMessage{}
	err := faker.FakeData(p,
		options.WithRandomMapAndSliceMinSize(10),
		options.WithRandomMapAndSliceMaxSize(33))
	assert.NoError(b, err)
	runBenchmark(p, b)
}

func BenchmarkStructMap(b *testing.B) {
	p := &StructMapMessage{}
	err := faker.FakeData(p,
		options.WithRandomMapAndSliceMinSize(10),
		options.WithRandomMapAndSliceMaxSize(33))
	assert.NoError(b, err)
	runBenchmark(p, b)
}
