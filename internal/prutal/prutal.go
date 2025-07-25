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

	"github.com/cloudwego/prutal/internal/desc"
	"github.com/cloudwego/prutal/internal/hack"
)

var errNotPointer = errors.New("input not pointer type")

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

func Unmarshal(b []byte, v interface{}) error {
	hack.PanicIfHackErr()
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return errNotPointer
	}
	desc, err := desc.GetOrParse(rv)
	if err != nil {
		return err
	}
	d := decoderPool.Get().(*Decoder)
	_, err = d.DecodeStruct(b, rv.UnsafePointer(), desc, defaultRecursionMaxDepth)
	decoderPool.Put(d)
	return err
}
