/*
 * Copyright 2024 CloudWeGo Authors
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

package hack

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"unsafe"
)

var (
	hackErr    bool
	hackErrMsg string
)

func init() {
	err := testhack()
	if err != nil {
		hackErr = true
		hackErrMsg = fmt.Sprintf("[BUG] Please upgrade Prutal to latest version.\n"+
			"If the issue still exists kindly report to author.\n"+
			"Err: %s %s/%s %s", runtime.GOOS, runtime.Version(), runtime.GOARCH, err)
	}
}

func PanicIfHackErr() {
	if hackErr {
		panic(hackErrMsg)
	}
}

// this func should be called once to test compatibility with Go runtime
func testhack() error {
	{ // MapIter

		m := map[int]string{7: "hello", 8: "world"}
		rv := reflect.ValueOf(m)
		it := NewMapIter(rv)
		m0 := map[int]string{}
		for {
			kp, vp := it.Next()
			if kp == nil {
				break
			}
			m0[*(*int)(kp)] = *(*string)(vp)
		}
		if !reflect.DeepEqual(m, m0) {
			return errors.New("compatibility issue found: MapIter")
		}
	}

	{ // maplen
		m := map[int]string{}
		m[8] = "world"
		m[9] = "!"
		m[10] = "?"
		if maplen(reflect.ValueOf(m).UnsafePointer()) != 3 {
			return errors.New("compatibility issue found: maplen")
		}
	}

	{ // ReflectValueWithPtr
		m := map[int]string{7: "hello"}
		rv := reflect.NewAt(reflect.TypeOf(m), unsafe.Pointer(&m)).Elem()
		rv1 := ReflectValueWithPtr(rv, unsafe.Pointer(&m))
		if p0, p1 := rv.UnsafePointer(), rv1.UnsafePointer(); p0 != p1 {
			return fmt.Errorf("compatibility issue found: ReflectValueWithPtr %p -> %p m=%p", p0, p1, &m)
		}
		m1, ok := rv1.Interface().(map[int]string)
		if !ok || !reflect.DeepEqual(m, m1) {
			return errors.New("compatibility issue found: ReflectValueWithPtr (Interface())")
		}
	}

	{ // ReflectValueTypePtr, ReflectTypePtr
		m1 := map[int]string{}
		m2 := map[int]*string{}
		m3 := map[int]*string{}
		rv := reflect.New(reflect.TypeOf(m1)).Elem()

		if ReflectValueTypePtr(reflect.ValueOf(m1)) != ReflectValueTypePtr(rv) ||
			ReflectValueTypePtr(reflect.ValueOf(m2)) != ReflectValueTypePtr(reflect.ValueOf(m3)) ||
			ReflectValueTypePtr(reflect.ValueOf(m1)) == ReflectValueTypePtr(reflect.ValueOf(m2)) {
			return errors.New("compatibility issue found: ReflectValueTypePtr")
		}

		if ReflectTypePtr(reflect.TypeOf(m1)) != ReflectTypePtr(rv.Type()) ||
			ReflectTypePtr(reflect.TypeOf(m2)) != ReflectTypePtr(reflect.TypeOf(m3)) ||
			ReflectTypePtr(reflect.TypeOf(m1)) == ReflectTypePtr(reflect.TypeOf(m3)) {
			return errors.New("compatibility issue found: ReflectTypePtr")
		}

		if ReflectTypePtr(reflect.TypeOf(m1)) != ReflectValueTypePtr(rv) ||
			ReflectTypePtr(reflect.TypeOf(m2)) != ReflectValueTypePtr(reflect.ValueOf(m3)) {
			return errors.New("compatibility issue found: ReflectTypePtr<>ReflectValueTypePtr")
		}
	}

	{
		d0 := &dog{"woo0"}
		f0 := iFoo(d0)
		if IfaceTypePtr(unsafe.Pointer(&f0)) != ReflectTypePtr(reflect.TypeOf(d0)) {
			return fmt.Errorf("compatibility issue found: IfaceTypePtr wrong type")
		}
		d1 := &dog{"woo1"}
		f1 := iFoo(d1)
		f2 := iFoo(&cat{"meow"})

		if IfaceTypePtr(unsafe.Pointer(&f0)) != IfaceTypePtr(unsafe.Pointer(&f1)) {
			return fmt.Errorf("compatibility issue found: IfaceTypePtr same type must equal")
		}
		if IfaceTypePtr(unsafe.Pointer(&f1)) == IfaceTypePtr(unsafe.Pointer(&f2)) {
			return fmt.Errorf("compatibility issue found: IfaceTypePtr two types must not equal")
		}
		tab := IfaceTab(reflect.TypeOf((*iFoo)(nil)).Elem(), reflect.TypeOf(d0).Elem())
		if tab != *(*uintptr)(unsafe.Pointer(&f0)) {
			return fmt.Errorf("compatibility issue found: IfaceTab")
		}

		var i iFoo
		IfaceUpdate(unsafe.Pointer(&i), tab, unsafe.Pointer(&dog{"ha"}))
		if i.Foo() != "ha" {
			return fmt.Errorf("compatibility issue found: IfaceUpdate")
		}

	}

	return nil
}

type hackMapIter struct {
	m      reflect.Value
	hitter struct {
		// k and v is always the 1st two fields of hitter
		// it will not be changed easily even though in the future
		k unsafe.Pointer
		v unsafe.Pointer
	}
}

// MapIter wraps reflect.MapIter for faster unsafe Next()
type MapIter struct {
	reflect.MapIter
}

// NewMapIter creates reflect.MapIter for reflect.Value.
// for go1.18, rv.MapRange() will cause one more allocation
// for >=go1.19, can use rv.MapRange() directly.
// see: https://github.com/golang/go/commit/c5edd5f616b4ee4bbaefdb1579c6078e7ed7e84e
// TODO: remove this func, and use MapIter{rv.MapRange()} when >=go1.19
func NewMapIter(rv reflect.Value) MapIter {
	ret := MapIter{}
	(*hackMapIter)(unsafe.Pointer(&ret.MapIter)).m = rv
	return ret
}

func (m *MapIter) Next() (unsafe.Pointer, unsafe.Pointer) {
	// use reflect.Next to initialize hitter
	// then we no need to bind mapiterinit, mapiternext
	m.MapIter.Next()
	p := (*hackMapIter)(unsafe.Pointer(&m.MapIter))
	return p.hitter.k, p.hitter.v
}

func maplen(p unsafe.Pointer) int {
	// XXX: race detector not working with this func
	type hmap struct {
		count int // count is the 1st field
	}
	return (*hmap)(p).count
}

type rvtype struct { // reflect.Value
	abiType uintptr
	ptr     unsafe.Pointer // data pointer
}

// ReflectValueWithPtr returns reflect.Value with the unsafe.Pointer.
// Same reflect.NewAt().Elem() without the cost of getting abi.Type
func ReflectValueWithPtr(rv reflect.Value, p unsafe.Pointer) reflect.Value {
	(*rvtype)(unsafe.Pointer(&rv)).ptr = p
	return rv
}

type iFoo interface{ Foo() string }

type dog struct{ sound string }

func (d *dog) Foo() string { return d.sound }

type cat struct{ sound string }

func (c *cat) Foo() string { return c.sound }

type itab struct {
	_   uintptr
	typ uintptr
}

type Iface struct {
	tab  uintptr
	data unsafe.Pointer
}

// IfaceTypePtr returns the underlying type ptr of the given p.
//
// p MUST be an Iface
func IfaceTypePtr(p unsafe.Pointer) uintptr {
	return (*itab)(unsafe.Pointer((*Iface)(p).tab)).typ
}

// IfaceUpdate updates iface p with tab and data
func IfaceUpdate(p unsafe.Pointer, tab uintptr, data unsafe.Pointer) {
	(*Iface)(p).tab = tab
	(*Iface)(p).data = data
}

// IfaceTab returns iface tab of `v` for iface `t`
func IfaceTab(t reflect.Type, v reflect.Type) uintptr {
	if t.Kind() != reflect.Interface || v.Kind() != reflect.Struct {
		panic("input type mismatch")
	}
	i := &Iface{}
	reflect.NewAt(t, unsafe.Pointer(i)).Elem().Set(reflect.New(v))
	return uintptr(unsafe.Pointer(i.tab))
}

// ExtratIface returns the underlying type and data ptr of the given p.
//
// p MUST be an Iface
func ExtratIface(p unsafe.Pointer) (typ uintptr, data unsafe.Pointer) {
	return (*itab)(unsafe.Pointer((*Iface)(p).tab)).typ, unsafe.Pointer((*Iface)(p).data)
}

// ReflectValueTypePtr returns the abi.Type pointer of the given reflect.Value.
// It used by createOrGetStructDesc for mapping a struct type to *StructDesc,
// and also used when Malloc
func ReflectValueTypePtr(rv reflect.Value) uintptr {
	return (*rvtype)(unsafe.Pointer(&rv)).abiType
}

// ReflectTypePtr returns the abi.Type pointer of the given reflect.Type.
// *rtype of reflect pkg shares the same data struct with *abi.Type
func ReflectTypePtr(rt reflect.Type) uintptr {
	type iface struct {
		_    uintptr
		data uintptr
	}
	return (*iface)(unsafe.Pointer(&rt)).data
}

// StringHeader ...
type StringHeader struct {
	Data unsafe.Pointer
	Len  int
}

// SliceHeader ...
type SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}
