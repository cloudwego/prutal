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

package prutal

import (
	"errors"
	"reflect"
	"unsafe"

	"github.com/cloudwego/prutal/internal/desc"
	"github.com/cloudwego/prutal/internal/hack"
)

var (
	errNotPointer = errors.New("input not pointer type")
	errNilPointer = errors.New("cannot unmarshal into nil message")
)

func MarshalAppend(b []byte, v interface{}) ([]byte, error) {
	hack.PanicIfHackErr()
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return nil, errNotPointer
	}
	s := desc.CacheGet(rv)
	if s == nil {
		var err error
		s, err = desc.GetOrParse(rv)
		if err != nil {
			return nil, err
		}
	}
	enc := encoderPool.Get().(*Encoder)
	b, err := enc.AppendStruct(b, rv.UnsafePointer(), s, false, defaultRecursionMaxDepth)
	encoderPool.Put(enc)
	return b, err
}

// Size returns the size in bytes of the wire-format encoding of v.
func Size(v interface{}) (int, error) {
	if v == nil {
		return 0, nil
	}
	hack.PanicIfHackErr()
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return 0, errNotPointer
	}
	if rv.IsNil() {
		return 0, nil
	}
	s := desc.CacheGet(rv)
	if s == nil {
		var err error
		s, err = desc.GetOrParse(rv)
		if err != nil {
			return 0, err
		}
	}
	return SizeStruct(rv.UnsafePointer(), s, defaultRecursionMaxDepth)
}

// resetStruct clears the struct at base before Unmarshal decodes into it.
// Pointer-free structs zero directly without write barriers, whereas structs
// with pointers delegate to SetZero (runtime typedmemclr) so GC write barriers
// stay intact without relying on go:linkname. Callers passing a zeroed or
// pre-reset struct can use UnmarshalMerge to skip this reset cost.
func resetStruct(s *desc.StructDesc, base unsafe.Pointer) {
	if s.Size == 0 {
		return
	}
	if !s.HasPointers {
		b := unsafe.Slice((*byte)(base), s.Size)
		for i := range b {
			b[i] = 0
		}
		return
	}
	hack.ReflectValueWithPtr(s.ElemValue, base).SetZero()
}

// Unmarshal resets v and decodes protobuf bytes b into the struct pointed to by v.
func Unmarshal(b []byte, v interface{}) error {
	return unmarshal(b, v, true)
}

// UnmarshalMerge decodes protobuf bytes b into the struct pointed to by v without resetting v first.
func UnmarshalMerge(b []byte, v interface{}) error {
	return unmarshal(b, v, false)
}

func unmarshal(b []byte, v interface{}, reset bool) error {
	hack.PanicIfHackErr()
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return errNotPointer
	}
	if rv.IsNil() {
		return errNilPointer
	}
	s := desc.CacheGet(rv)
	if s == nil {
		var err error
		s, err = desc.GetOrParse(rv)
		if err != nil {
			return err
		}
	}
	base := rv.UnsafePointer()
	if reset {
		resetStruct(s, base)
	}
	d := decoderPool.Get().(*Decoder)
	_, err := d.DecodeStruct(b, base, s, defaultRecursionMaxDepth)
	decoderPool.Put(d)
	return err
}
