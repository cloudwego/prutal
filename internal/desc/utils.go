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
	"fmt"
	"reflect"
	"unsafe"
)

// IsFieldKeyTypeMatchReflectKind returns nil if t match k for map key type
// It's same as IsFieldTypeMatchReflectKind except k can't be `[]byte` or `struct`,
// and floats are excluded for fixed32/fixed64 since protobuf forbids
// floating-point map keys
func IsFieldKeyTypeMatchReflectKind(t TagType, k reflect.Kind) error {
	ok := false
	switch t {
	case TypeVarint:
		ok = k == reflect.Int32 || k == reflect.Uint32 ||
			k == reflect.Int64 || k == reflect.Uint64 ||
			k == reflect.Bool
	case TypeZigZag32:
		ok = k == reflect.Int32
	case TypeZigZag64:
		ok = k == reflect.Int64
	case TypeFixed32:
		ok = k == reflect.Int32 || k == reflect.Uint32
	case TypeFixed64:
		ok = k == reflect.Int64 || k == reflect.Uint64
	case TypeBytes:
		ok = k == reflect.String
	}
	if ok {
		return nil
	}
	return fmt.Errorf("tag type %q not match field type %q", t, k)
}

// IsFieldTypeMatchReflectKind return nil if t match k
func IsFieldTypeMatchReflectKind(t TagType, k reflect.Kind) error {
	ok := false
	switch t {
	case TypeVarint:
		ok = k == reflect.Int32 || k == reflect.Uint32 ||
			k == reflect.Int64 || k == reflect.Uint64 ||
			k == reflect.Bool
	case TypeZigZag32:
		ok = k == reflect.Int32
	case TypeZigZag64:
		ok = k == reflect.Int64
	case TypeFixed32:
		ok = k == reflect.Int32 || k == reflect.Uint32 || k == reflect.Float32
	case TypeFixed64:
		ok = k == reflect.Int64 || k == reflect.Uint64 || k == reflect.Float64
	case TypeBytes:
		ok = k == reflect.String || k == reflect.Map || k == reflect.Struct || k == KindBytes
	}
	if ok {
		return nil
	}
	return fmt.Errorf("tag type %q not match field type %q", t, k)
}

type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

const (
	// methodProtoReflect is the method which returns description of a proto struct.
	methodProtoReflect = "ProtoReflect"

	// fieldOneofWrappers is the field name which contains oneof fields of a proto struct
	//
	// see: https://go-review.googlesource.com/c/protobuf/+/185239
	fieldOneofWrappers = "OneofWrappers"

	// methodOneofWrappers is the method which returns oneof fields of a proto struct
	//
	// It's created by old version protobuf or gogoprotobuf
	methodOneofWrappers = "XXX_OneofWrappers"

	// methodOneofFuncs is used by older versions of protoc-gen-go. One of its
	// return values contains the oneof wrapper types.
	methodOneofFuncs = "XXX_OneofFuncs"
)

func searchOneofWrappers(t reflect.Type) []any {
	kd := t.Kind()
	for kd == reflect.Pointer || kd == reflect.Interface {
		t = t.Elem()
		kd = t.Kind()
	}
	if kd != reflect.Struct {
		return nil
	}
	pt := reflect.PointerTo(t)
	m, ok := pt.MethodByName(methodProtoReflect)
	if ok && m.Type.NumIn() == 1 && m.Type.NumOut() == 1 && !m.Type.IsVariadic() {
		args := []reflect.Value{reflect.NewAt(t, nil)}
		if wrappers := searchFieldOneofWrappers(m.Func.Call(args)[0], 5); wrappers != nil {
			return wrappers
		}
	}
	var wrappers []any
	for _, name := range []string{methodOneofFuncs, methodOneofWrappers} {
		m, ok = pt.MethodByName(name)
		if !ok {
			continue
		}
		if m.Type.NumIn() != 1 || m.Type.IsVariadic() {
			continue
		}
		args := []reflect.Value{reflect.Zero(m.Type.In(0))}
		for _, v := range m.Func.Call(args) {
			if !v.CanInterface() {
				continue
			}
			if vs, ok := v.Interface().([]any); ok {
				wrappers = vs
			}
		}
	}
	return wrappers
}

func searchFieldOneofWrappers(v reflect.Value, maxdepth int) []any {
	if maxdepth <= 0 {
		return nil
	}
	kd := v.Kind()
	for kd == reflect.Pointer || kd == reflect.Interface {
		v = v.Elem()
		kd = v.Kind()
	}
	if kd != reflect.Struct {
		return nil
	}
	oneofs := v.FieldByName(fieldOneofWrappers)
	if oneofs.IsValid() {
		if oneofs.Type() != reflect.TypeOf([]any(nil)) {
			return nil
		}
		// same as oneofs.Interface().([]any)
		// fix: cannot return value obtained from unexported field or method
		ret := []any{}
		p := (*sliceHeader)(unsafe.Pointer(&ret))
		p.Data = oneofs.UnsafePointer()
		p.Len = oneofs.Len()
		p.Cap = p.Len
		return ret
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if ret := searchFieldOneofWrappers(f, maxdepth-1); ret != nil {
			return ret
		}
	}
	return nil
}
