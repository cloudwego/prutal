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
	// hackErr indicates whether the unsafe hacks are compatible with current Go runtime
	hackErr bool
	// hackErrMsg contains the error message if compatibility test fails
	hackErrMsg string
)

// init runs compatibility tests for all unsafe operations used in this package.
// If any test fails, hackErr is set to true and PanicIfHackErr() will panic when called.
func init() {
	err := testhack()
	if err != nil {
		hackErr = true
		hackErrMsg = fmt.Sprintf("[BUG] Please upgrade Prutal to latest version.\n"+
			"If the issue still exists kindly report to author.\n"+
			"Err: %s %s/%s %s", runtime.GOOS, runtime.Version(), runtime.GOARCH, err)
	}
}

// PanicIfHackErr panics if any of the unsafe operations are incompatible with current Go runtime.
// This should be called early in the application lifecycle to ensure safety.
func PanicIfHackErr() {
	if hackErr {
		panic(hackErrMsg)
	}
}

// testhack runs comprehensive compatibility tests for all unsafe operations.
// It verifies that the internal Go runtime structures match our assumptions.
func testhack() error {
	{ // Test MapIter - verify unsafe map iteration works correctly
		m := map[int]string{7: "hello", 8: "world"}
		it := NewMapIter(reflect.ValueOf(m))

		// Reconstruct the map using unsafe pointers to verify iteration works
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

	{ // Test ReflectValueWithPtr - verify unsafe reflect.Value pointer manipulation
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

	{ // Test ReflectValueTypePtr and ReflectTypePtr - verify type pointer extraction
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

	{ // Test interface manipulation - verify unsafe interface operations work correctly
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

		if IfaceData(unsafe.Pointer(&f0)) != unsafe.Pointer(d0) {
			return fmt.Errorf("compatibility issue found: IfaceData")
		}
	}

	return nil
}

// hackMapIter mirrors the internal structure of reflect.MapIter for unsafe access.
// This allows us to directly access the key and value pointers without allocations.
type hackMapIter struct {
	m      reflect.Value
	hitter struct {
		// k and v are always the 1st two fields of hitter in Go's runtime
		// This layout is stable and unlikely to change in future Go versions
		k unsafe.Pointer
		v unsafe.Pointer
	}
}

// MapIter wraps reflect.MapIter for faster unsafe Next() operations.
// It provides direct access to key and value pointers without reflection overhead.
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

// Next returns unsafe pointers to the current key and value in the map iteration.
// Returns (nil, nil) when iteration is complete.
func (m *MapIter) Next() (unsafe.Pointer, unsafe.Pointer) {
	// Use reflect.Next to initialize hitter fields,
	// then we can directly access k,v pointers without binding mapiterinit/mapiternext
	m.MapIter.Next()
	p := (*hackMapIter)(unsafe.Pointer(&m.MapIter))
	return p.hitter.k, p.hitter.v
}

// rvtype mirrors the internal structure of reflect.Value for unsafe manipulation
type rvtype struct {
	abiType uintptr        // pointer to type information (abi.Type)
	ptr     unsafe.Pointer // pointer to the actual data
}

// ReflectValueWithPtr creates a reflect.Value by replacing its data pointer.
// This is equivalent to reflect.NewAt().Elem() but avoids the cost of type lookup.
// The returned Value shares the same type as rv but points to data at p.
func ReflectValueWithPtr(rv reflect.Value, p unsafe.Pointer) reflect.Value {
	(*rvtype)(unsafe.Pointer(&rv)).ptr = p
	return rv
}

// Test types for interface manipulation compatibility testing
type iFoo interface{ Foo() string }

type dog struct{ sound string }

func (d *dog) Foo() string { return d.sound }

type cat struct{ sound string }

func (c *cat) Foo() string { return c.sound }

// itab mirrors Go's internal interface table structure
type itab struct {
	_   uintptr // interface type info
	typ uintptr // concrete type pointer
}

// Iface mirrors Go's internal interface{} structure for direct manipulation
type Iface struct {
	tab  uintptr        // pointer to itab
	data unsafe.Pointer // pointer to actual data
}

// IfaceData extracts the data pointer from an interface{} value.
// This provides direct access to the underlying concrete value without reflection overhead.
//
// p MUST point to an interface{} value
func IfaceData(p unsafe.Pointer) unsafe.Pointer {
	return (*Iface)(p).data
}

// IfaceTypePtr extracts the type pointer from an interface{} value.
// This allows fast type checking and comparison without reflection.
//
// p MUST point to an interface{} value
func IfaceTypePtr(p unsafe.Pointer) uintptr {
	return (*itab)(unsafe.Pointer((*Iface)(p).tab)).typ
}

// IfaceUpdate constructs an interface{} value by setting both the type table and data pointer.
// This allows efficient interface{} construction without reflection overhead.
func IfaceUpdate(p unsafe.Pointer, tab uintptr, data unsafe.Pointer) {
	(*Iface)(p).tab = tab
	(*Iface)(p).data = data
}

// IfaceTab generates the interface table (itab) for concrete type `v` implementing interface `t`.
// This is used for fast interface{} construction without repeated type assertions.
func IfaceTab(t reflect.Type, v reflect.Type) uintptr {
	if t.Kind() != reflect.Interface || v.Kind() != reflect.Struct {
		panic("input type mismatch")
	}
	i := &Iface{}
	reflect.NewAt(t, unsafe.Pointer(i)).Elem().Set(reflect.New(v))
	return uintptr(unsafe.Pointer(i.tab))
}

// ReflectValueTypePtr extracts the abi.Type pointer from a reflect.Value.
// This is used for efficient type mapping in struct descriptors and memory allocation.
// The returned pointer uniquely identifies the type and can be used for fast type comparisons.
func ReflectValueTypePtr(rv reflect.Value) uintptr {
	return (*rvtype)(unsafe.Pointer(&rv)).abiType
}

// ReflectTypePtr extracts the abi.Type pointer from a reflect.Type.
// The *rtype from reflect package shares the same data structure as *abi.Type,
// allowing direct access to the underlying type information.
func ReflectTypePtr(rt reflect.Type) uintptr {
	type iface struct {
		_    uintptr // type info
		data uintptr // actual type pointer
	}
	return (*iface)(unsafe.Pointer(&rt)).data
}

// StringHeader mirrors Go's internal string structure for unsafe string operations.
// This allows zero-copy string manipulation and conversion between string and []byte.
type StringHeader struct {
	Data unsafe.Pointer
	Len  int
}

// SliceHeader mirrors Go's internal slice structure for unsafe slice operations.
// This enables zero-copy slice manipulation and direct memory access.
type SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}
