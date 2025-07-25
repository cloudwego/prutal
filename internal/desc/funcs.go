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

// getDecodeFunc returns optimized decode functions for specific field types.
// Returns nil for fields that should use generic reflection-based decoding.
func getDecodeFunc(f *FieldDesc) wire.DecodeFunc {
	if f.IsMap {
		// Returns optimized decoder for maps with primitive types
		// Returns nil for maps with complex types (e.g., struct values, nested maps)
		return getMapDecodeFunc(f)
	}
	if f.Packed {
		// Returns optimized decoder for packed repeated scalar fields
		return getPackedDecodeFunc(f)
	}
	// Non-packed, non-map fields use generic decoding
	return nil
}

func getPackedDecodeFunc(f *FieldDesc) wire.DecodeFunc {
	c := getCoderType(f.TagType, dereferenceElemKind(f.T.V.T))
	return wire.GetPackedDecoderFunc(c)
}

// getMapDecodeFunc returns optimized decoder functions for map fields.
// Only returns a function if BOTH key and value types have optimized decoders.
//
// Optimized decoders are available for:
// - Keys: int32, uint32, int64, uint64, sint32, sint64, fixed32, fixed64, sfixed32, sfixed64, bool
// - Values: All key types above (scalar types only)
//
// Returns nil (uses fallback) for:
// - String keys (map[string]T)
// - Bytes values (map[K][]byte)
// - String values (map[K]string) when K is string
// - Struct/Message values (map[K]Message or map[K]*Message)
// - Any other complex value types
func getMapDecodeFunc(f *FieldDesc) wire.DecodeFunc {
	kt := getCoderType(f.KeyType, reflectTypeKind(f.T.K.T))
	vt := getCoderType(f.ValType, reflectTypeKind(f.T.V.T))
	// GetMapDecoderFunc returns nil if no optimized decoder exists for the key-value pair
	return wire.GetMapDecoderFunc(kt, vt)
}

func dereferenceTypeKind(t reflect.Type) reflect.Kind {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return reflectTypeKind(t)
}

func dereferenceElemKind(t reflect.Type) reflect.Kind {
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	return reflectTypeKind(t)
}

func getAppendFunc(t TagType, k reflect.Kind, packed bool) wire.AppendFunc {
	c := getCoderType(t, k)
	return wire.GetAppendFunc(c, packed)
}

func getAppendListFunc(f *FieldDesc) wire.AppendRepeatedFunc {
	t, k := f.TagType, f.T.RealKind()
	c := getCoderType(t, k)
	return wire.GetAppendListFunc(c)
}

func getAppendMapFunc(f *FieldDesc) wire.AppendRepeatedFunc {
	kt := getCoderType(f.KeyType, reflectTypeKind(f.T.K.T))
	vt := getCoderType(f.ValType, reflectTypeKind(f.T.V.T))
	return wire.GetMapEncoderFunc(kt, vt)
}
