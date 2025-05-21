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
	"errors"
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"github.com/cloudwego/prutal/internal/hack"
)

const KindBytes reflect.Kind = 5000 // for []byte

var bytesType = reflect.TypeOf([]byte{})

func reflectTypeKind(t reflect.Type) reflect.Kind {
	if t == bytesType {
		return KindBytes // special case
	}
	return t.Kind()
}

type Type struct {
	T reflect.Type

	// true if t.Kind == reflect.Pointer
	IsPointer bool

	// true if t.Kind == reflect.Slice,
	// false for []byte which is considered to be scalar type
	IsSlice bool

	SliceLike bool // reflect.Slice, reflect.String, KindBytes

	// cache reflect.Type returns for performance
	Kind  reflect.Kind
	Size  uintptr
	Align int

	// for decoder
	MallocAbiType uintptr

	K *Type       // for map
	V *Type       // for pointer, slice or map
	S *StructDesc // struct

	// for map only
	MapTmpVarsPool sync.Pool // for decoder tmp vars

	finalized bool
}

func (t *Type) RealKind() reflect.Kind {
	if t.IsPointer || t.IsSlice {
		return t.V.RealKind()
	}
	return t.Kind
}

func (t *Type) finalizeType() error {
	if t.finalized {
		return nil
	}
	t.finalized = true
	if t.S != nil {
		if err := t.S.FinalizeFields(); err != nil {
			t.finalized = false
			return err
		}
	}
	if t.V != nil {
		if err := t.V.finalizeType(); err != nil {
			t.finalized = false
			return err
		}
	}
	return nil
}

func (t *Type) String() string {
	switch t.Kind {
	case reflect.Struct:
		return fmt.Sprintf("%+v", t.S)
	default:
		return fmt.Sprintf("%+v", t.T)
	}
}

var (
	cachedTypes = map[reflect.Type]*Type{}
)

func noopFinalizeField(_ *Type) error { return nil }

func parseType(rt reflect.Type) (t *Type, err error) {
	if t = cachedTypes[rt]; t != nil {
		return t, nil
	}

	t = &Type{}
	cachedTypes[rt] = t // reuse result and also fix cyclic refs

	t.T = rt
	t.Kind = rt.Kind()
	t.Size = rt.Size()
	t.Align = rt.Align()
	t.Kind = reflectTypeKind(rt)

	switch t.Kind {
	case reflect.Ptr, reflect.Slice, KindBytes, reflect.String,
		reflect.Map, reflect.Struct:
		// for these types, we can't use span mem allocator
		// coz then may contain pointer
		t.MallocAbiType = hack.ReflectTypePtr(t.T)
	}

	t.IsPointer = t.Kind == reflect.Pointer
	t.IsSlice = t.Kind == reflect.Slice

	t.SliceLike = t.Kind == reflect.Slice ||
		t.Kind == KindBytes ||
		t.Kind == reflect.String

	switch rt.Kind() {
	case reflect.Map:
		t.K, err = parseType(rt.Key())
		if err != nil {
			break
		}
		t.V, err = parseType(rt.Elem())
		if err != nil {
			break
		}
		t.MapTmpVarsPool.New = newTmpMapVarsFunc(rt)
	case reflect.Struct:
		t.S, err = parseStruct(rt)
	case reflect.Slice:
		t.V, err = parseType(rt.Elem())
	case reflect.Pointer:
		t.V, err = parseType(rt.Elem())
		if err == nil && t.V.IsPointer {
			err = errors.New("multilevel pointer")
		}
	default:
	}
	if err != nil {
		delete(cachedTypes, rt)
		return nil, err
	}
	return t, nil
}

// TmpMapVars contains key and value tmp vars used for updating associated map for a type
type TmpMapVars struct {
	m reflect.Value

	k  reflect.Value  // t.K.T
	kp unsafe.Pointer // *t.K.T

	v  reflect.Value  // t.V.T
	vp unsafe.Pointer // *t.V.T

	// zero value of v,
	// only used when non-pointer struct as map val
	// we need to zero the tmp var before using it
	z reflect.Value

	Update func(p *TmpMapVars, mp unsafe.Pointer) // see newTmpMapVarsFunc

}

func (p *TmpMapVars) MapWithPtr(x unsafe.Pointer) reflect.Value {
	return hack.ReflectValueWithPtr(p.m, x)
}
func (p *TmpMapVars) KeyPointer() unsafe.Pointer { return p.kp }
func (p *TmpMapVars) ValPointer() unsafe.Pointer { return p.vp }

func (p *TmpMapVars) Reset() {
	if p.z.IsValid() {
		p.v.Set(p.z)
	}
}

// for (*sync.Pool).New
func newTmpMapVarsFunc(rt reflect.Type) func() any {
	if rt.Kind() != reflect.Map {
		panic(rt.Kind())
	}
	kt := rt.Key()
	et := rt.Elem()
	updateFunc := defaultMapUpdateFunc
	switch kt.Kind() {
	case reflect.Int64, reflect.Uint64:
		switch et.Kind() {
		case reflect.Pointer:
			updateFunc = mapUpdateFunc_u64_unsafe
		case reflect.String:
			updateFunc = mapUpdateFunc_u64_string
		case reflect.Struct:
			if et.Size() == 0 {
				updateFunc = mapUpdateFunc_u64_empty
			}
		}
	case reflect.Int32, reflect.Uint32:
		switch et.Kind() {
		case reflect.Pointer:
			updateFunc = mapUpdateFunc_u32_unsafe
		case reflect.String:
			updateFunc = mapUpdateFunc_u32_string
		case reflect.Struct:
			if et.Size() == 0 {
				updateFunc = mapUpdateFunc_u32_empty
			}
		}
	case reflect.String:
		switch et.Kind() {
		case reflect.Pointer:
			updateFunc = mapUpdateFunc_string_unsafe
		case reflect.String:
			updateFunc = mapUpdateFunc_string_string
		case reflect.Struct:
			if et.Size() == 0 {
				updateFunc = mapUpdateFunc_string_empty
			}
		}
	}
	return func() any {
		p := &TmpMapVars{}
		p.m = reflect.New(rt).Elem()
		p.k = reflect.New(kt)
		p.kp = p.k.UnsafePointer()
		p.k = p.k.Elem()
		p.v = reflect.New(et)
		p.vp = p.v.UnsafePointer()
		p.v = p.v.Elem()
		if et.Kind() == reflect.Struct {
			p.z = reflect.Zero(et)
		}
		p.Update = updateFunc
		return p
	}
}

func defaultMapUpdateFunc(p *TmpMapVars, mp unsafe.Pointer) {
	m := hack.ReflectValueWithPtr(p.m, mp)
	m.SetMapIndex(p.k, p.v)
}

func mapUpdateFunc_u64_unsafe(p *TmpMapVars, mp unsafe.Pointer) {
	m := *(*map[uint64]unsafe.Pointer)(mp)
	m[*(*uint64)(p.kp)] = *(*unsafe.Pointer)(p.vp)
}

func mapUpdateFunc_u64_empty(p *TmpMapVars, mp unsafe.Pointer) {
	m := *(*map[uint64]struct{})(mp)
	m[*(*uint64)(p.kp)] = struct{}{}
}

func mapUpdateFunc_u64_string(p *TmpMapVars, mp unsafe.Pointer) {
	m := *(*map[uint64]string)(mp)
	m[*(*uint64)(p.kp)] = *(*string)(p.vp)
}

func mapUpdateFunc_u32_unsafe(p *TmpMapVars, mp unsafe.Pointer) {
	m := *(*map[uint32]unsafe.Pointer)(mp)
	m[*(*uint32)(p.kp)] = *(*unsafe.Pointer)(p.vp)
}

func mapUpdateFunc_u32_empty(p *TmpMapVars, mp unsafe.Pointer) {
	m := *(*map[uint32]struct{})(mp)
	m[*(*uint32)(p.kp)] = struct{}{}
}

func mapUpdateFunc_u32_string(p *TmpMapVars, mp unsafe.Pointer) {
	m := *(*map[uint32]string)(mp)
	m[*(*uint32)(p.kp)] = *(*string)(p.vp)
}

func mapUpdateFunc_string_unsafe(p *TmpMapVars, mp unsafe.Pointer) {
	m := *(*map[string]unsafe.Pointer)(mp)
	m[*(*string)(p.kp)] = *(*unsafe.Pointer)(p.vp)
}

func mapUpdateFunc_string_empty(p *TmpMapVars, mp unsafe.Pointer) {
	m := *(*map[string]struct{})(mp)
	m[*(*string)(p.kp)] = struct{}{}
}

func mapUpdateFunc_string_string(p *TmpMapVars, mp unsafe.Pointer) {
	m := *(*map[string]string)(mp)
	m[*(*string)(p.kp)] = *(*string)(p.vp)
}
