/*
 * Copyright 2026 CloudWeGo Authors
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

package prutal

import (
	"reflect"
	"runtime"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
	"github.com/cloudwego/prutal/internal/wire"
)

func TestMapAlignedEmptyValues(t *testing.T) {
	// The zero-length array keeps 8-byte alignment even on 32-bit systems.
	// Such a value cannot use the map[K]struct{} storage layout.
	type empty struct{ _ [0]atomic.Int64 }
	type message struct {
		U32 map[uint32]empty `protobuf:"bytes,1,rep" protobuf_key:"varint,1,opt" protobuf_val:"bytes,2,opt"`
		U64 map[uint64]empty `protobuf:"bytes,2,rep" protobuf_key:"varint,1,opt" protobuf_val:"bytes,2,opt"`
		Str map[string]empty `protobuf:"bytes,3,rep" protobuf_key:"bytes,1,opt" protobuf_val:"bytes,2,opt"`
	}
	for _, n := range []int{2, 64} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			src := message{
				U32: make(map[uint32]empty),
				U64: make(map[uint64]empty),
				Str: make(map[string]empty),
			}
			for i := 1; i <= n; i++ {
				src.U32[uint32(i)] = empty{}
				src.U64[uint64(i)] = empty{}
				src.Str[strconv.Itoa(i)] = empty{}
			}
			b, err := MarshalAppend(nil, &src)
			assert.NoError(t, err)
			var dst message
			assert.NoError(t, Unmarshal(b, &dst))
			runtime.GC()
			assert.DeepEqual(t, src, dst)
		})
	}
}

func TestDecodeDynamicTypes(t *testing.T) {
	child := reflect.StructOf([]reflect.StructField{
		{Name: "Number", Type: reflect.TypeOf(int32(0)), Tag: `protobuf:"varint,1,opt"`},
		{Name: "Text", Type: reflect.TypeOf(""), Tag: `protobuf:"bytes,2,opt"`},
	})
	payload := (&wire.Builder{}).AppendVarintField(1, 7).AppendStringField(2, "alive").Bytes()
	for _, repeated := range []bool{false, true} {
		name, tag, fieldType := "pointer", `protobuf:"bytes,1,opt"`, reflect.PointerTo(child)
		if repeated {
			name, tag, fieldType = "slice", `protobuf:"bytes,1,rep"`, reflect.SliceOf(child)
		}
		t.Run(name, func(t *testing.T) {
			rt := reflect.StructOf([]reflect.StructField{
				{Name: "Child", Type: fieldType, Tag: reflect.StructTag(tag)},
			})
			dst := reflect.New(rt)
			b := (&wire.Builder{}).AppendBytesField(1, payload).Bytes()
			assert.NoError(t, Unmarshal(b, dst.Interface()))
			runtime.GC()
			v := dst.Elem().Field(0)
			if repeated {
				assert.Equal(t, 1, v.Len())
				v = v.Index(0)
			} else {
				v = v.Elem()
			}
			assert.Equal(t, int64(7), v.Field(0).Int())
			assert.Equal(t, "alive", v.Field(1).String())
		})
	}
}
