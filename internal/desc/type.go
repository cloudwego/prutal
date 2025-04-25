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
}

func (p *TmpMapVars) MapWithPtr(x unsafe.Pointer) reflect.Value {
	return hack.ReflectValueWithPtr(p.m, x)
}

func (p *TmpMapVars) KeyPointer() unsafe.Pointer { return p.kp }
func (p *TmpMapVars) ValPointer() unsafe.Pointer { return p.vp }
func (p *TmpMapVars) Update(m reflect.Value)     { m.SetMapIndex(p.k, p.v) }
func (p *TmpMapVars) Reset() {
	if p.z.IsValid() {
		p.v.Set(p.z)
	}
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

	if rt == bytesType { // special case
		t.Kind = KindBytes
	}

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
		t.MapTmpVarsPool.New = func() interface{} {
			m := &TmpMapVars{}
			m.m = reflect.New(rt).Elem()
			m.k = reflect.New(rt.Key())
			m.kp = m.k.UnsafePointer()
			m.k = m.k.Elem()
			m.v = reflect.New(rt.Elem())
			m.vp = m.v.UnsafePointer()
			m.v = m.v.Elem()
			if rt.Elem().Kind() == reflect.Struct {
				m.z = reflect.Zero(rt.Elem())
			}
			return m
		}
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
