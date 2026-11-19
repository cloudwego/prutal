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
	"errors"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"unsafe"

	"github.com/cloudwego/prutal/internal/desc"
	"github.com/cloudwego/prutal/internal/hack"
)

func dirtySpanDecoder() *Decoder {
	b := make([]byte, defaultDecoderSpanSize)
	for i := range b {
		b[i] = 0xa5
	}
	return &Decoder{s: span{b: unsafe.Pointer(&b[0]), n: len(b)}}
}

func decodeWithDirtySpan(b []byte, v any) error {
	rv := reflect.ValueOf(v)
	s, err := desc.GetOrParse(rv)
	if err != nil {
		return err
	}
	_, err = dirtySpanDecoder().DecodeStruct(b, rv.UnsafePointer(), s, defaultRecursionMaxDepth)
	return err
}

func TestDecodedSliceCapacityIsZeroed(t *testing.T) {
	var v struct {
		Values []uint32 `protobuf:"varint,1,rep"`
	}
	if err := decodeWithDirtySpan([]byte{8, 7}, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Values) != 1 || v.Values[0] != 7 {
		t.Fatalf("decoded values = %v", v.Values)
	}
	for i, value := range v.Values[:cap(v.Values)][1:] {
		if value != 0 {
			t.Fatalf("unused capacity[%d] = %#x", i+1, value)
		}
	}
}

func TestMalformedScalarAllocationsAreZeroed(t *testing.T) {
	t.Run("pointer", func(t *testing.T) {
		var v struct {
			Value *uint64 `protobuf:"varint,1,opt"`
		}
		if err := decodeWithDirtySpan([]byte{8}, &v); err == nil {
			t.Fatal("expected truncated varint error")
		}
		if v.Value != nil && *v.Value != 0 {
			t.Fatalf("uninitialized scalar = %#x", *v.Value)
		}
	})
	t.Run("oneof", func(t *testing.T) {
		var v TestOneofMessage
		if err := decodeWithDirtySpan([]byte{0x10}, &v); err == nil {
			t.Fatal("expected truncated varint error")
		}
		if value, ok := v.OneOfFieldA.(*TestOneofMessage_Field2); ok && value.Field2 != 0 {
			t.Fatalf("uninitialized oneof scalar = %#x", value.Field2)
		}
	})
}

func TestReallocSliceRejectsOverflow(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, tc := range []struct {
		capacity int
		size     uintptr
	}{
		{capacity: -1, size: 4},
		{capacity: maxInt/4 + 1, size: 4},
		{capacity: 8, size: uintptr(maxInt/4 + 1)},
	} {
		h := hack.SliceHeader{}
		d := dirtySpanDecoder()
		err := d.ReallocSlice(&h, &desc.Type{V: &desc.Type{Size: tc.size, Align: 4}}, tc.capacity)
		if !errors.Is(err, errSliceTooLarge) {
			t.Fatalf("capacity %d, element size %d: error = %v", tc.capacity, tc.size, err)
		}
		if h.Data != nil || h.Len != 0 || h.Cap != 0 {
			t.Fatalf("capacity %d changed the destination: %+v", tc.capacity, h)
		}
	}
}

func TestDecodedZeroSizeSliceGrowth(t *testing.T) {
	var v struct {
		Values []struct{} `protobuf:"bytes,1,rep"`
	}
	for i := 0; i < 20; i++ {
		if err := UnmarshalMerge([]byte{10, 0}, &v); err != nil {
			t.Fatal(err)
		}
	}
	if len(v.Values) != 20 {
		t.Fatalf("decoded element count = %d", len(v.Values))
	}
}

func TestDecodeSliceRejectsCapacityOverflow(t *testing.T) {
	var v struct {
		Values []struct{} `protobuf:"bytes,1,rep"`
	}
	// Zero-size elements make this capacity valid without a huge allocation.
	maxInt := int(^uint(0) >> 1)
	v.Values = make([]struct{}, maxInt/2+1)
	if err := UnmarshalMerge([]byte{10, 0}, &v); !errors.Is(err, errSliceTooLarge) {
		t.Fatalf("error = %v", err)
	}
}

func TestDecodedPointerSliceGrowth(t *testing.T) {
	type item struct {
		Text string `protobuf:"bytes,1,opt"`
	}
	type message struct {
		Values []item `protobuf:"bytes,1,rep"`
	}
	var v message
	for i := 0; i < 40; i++ {
		if err := UnmarshalMerge([]byte{10, 3, 10, 1, 'x'}, &v); err != nil {
			t.Fatal(err)
		}
	}
	runtime.GC()
	if len(v.Values) != 40 {
		t.Fatalf("decoded element count = %d", len(v.Values))
	}
	for i, value := range v.Values {
		if value.Text != "x" {
			t.Fatalf("element %d = %+v", i, value)
		}
	}
}

func TestDecodedPointerSliceGrowthConcurrent(t *testing.T) {
	type item struct {
		Text string `protobuf:"bytes,1,opt"`
	}
	type message struct {
		Values []item `protobuf:"bytes,1,rep"`
	}
	var b []byte
	for i := 0; i < 40; i++ {
		b = append(b, 10, 3, 10, 1, 'x')
	}
	const goroutines = 8
	const iterations = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				var v message
				if err := Unmarshal(b, &v); err != nil {
					t.Errorf("unmarshal failed: %v", err)
					return
				}
				if len(v.Values) != 40 {
					t.Errorf("decoded element count = %d", len(v.Values))
					return
				}
				for j, value := range v.Values {
					if value.Text != "x" {
						t.Errorf("element %d = %+v", j, value)
						return
					}
				}
			}
		}()
	}
	wg.Wait()
}

func TestDecodedPointerFreeSliceGrowthDuringGC(t *testing.T) {
	// The element size is not a multiple of the pointer size, so copying it
	// with pointer write barriers would crash while GC is marking.
	type point struct {
		X int32 `protobuf:"varint,1,opt"`
	}
	type message struct {
		Points []point `protobuf:"bytes,1,rep"`
	}
	var b []byte
	for i := 0; i < 64; i++ {
		b = append(b, 10, 2, 8, 1)
	}
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	defer func() {
		close(done)
		wg.Wait()
	}()
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				runtime.GC()
			}
		}
	}()
	for i := 0; i < 20000; i++ {
		var v message
		if err := Unmarshal(b, &v); err != nil {
			t.Fatal(err)
		}
		if len(v.Points) != 64 || v.Points[63].X != 1 {
			t.Fatalf("decoded points = %d", len(v.Points))
		}
	}
}
