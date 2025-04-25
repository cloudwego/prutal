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

package desc

import (
	"reflect"
	"unsafe"

	"github.com/cloudwego/prutal/internal/wire"
)

func getCoderType(t TagType, kd reflect.Kind) wire.CoderType {
	switch t {
	case TypeVarint:
		switch kd {
		case reflect.Int32, reflect.Uint32:
			return wire.CoderVarint32
		case reflect.Int64, reflect.Uint64:
			return wire.CoderVarint64
		case reflect.Bool:
			return wire.CoderBool
		}
	case TypeZigZag32:
		return wire.CoderZigZag32
	case TypeZigZag64:
		return wire.CoderZigZag64
	case TypeFixed32:
		return wire.CoderFixed32
	case TypeFixed64:
		return wire.CoderFixed64
	case TypeBytes:
		if kd == reflect.String {
			return wire.CoderString
		}
		if kd == KindBytes {
			return wire.CoderBytes
		}
		if kd == reflect.Struct {
			return wire.CoderStruct
		}
	}
	return wire.CoderUnknown
}

func getDecodeFunc(f *FieldDesc) wire.DecodeFunc {
	if f.IsMap {
		return getMapDecodeFunc(f)
	}
	if f.Packed {
		return getPackedDecodeFunc(f)
	}
	return nil
}

func getPackedDecodeFunc(f *FieldDesc) wire.DecodeFunc {
	c := getCoderType(f.TagType, dereferenceElemKind(f.T.V.T))
	return wire.GetPackedDecoderFunc(c)
}

func getMapDecodeFunc(f *FieldDesc) wire.DecodeFunc {
	kt := getCoderType(f.KeyType, dereferenceTypeKind(f.T.K.T))
	vt := getCoderType(f.ValType, dereferenceTypeKind(f.T.V.T))
	return wire.GetMapDecoderFunc(kt, vt)
}

func dereferenceTypeKind(t reflect.Type) reflect.Kind {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == bytesType {
		return KindBytes
	}
	return t.Kind()
}

func dereferenceElemKind(t reflect.Type) reflect.Kind {
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	if t == bytesType {
		return KindBytes
	}
	return t.Kind()
}

func getAppendFunc(t TagType, k reflect.Kind, packed bool) wire.AppendFunc {
	c := getCoderType(t, k)
	return wire.GetAppendFunc(c, packed)
}

func getAppendListFunc(t TagType, k reflect.Kind) func(b []byte, id int32, p unsafe.Pointer) []byte {
	c := getCoderType(t, k)
	return wire.GetAppendListFunc(c)
}
