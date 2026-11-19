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
		replacement := map[int]string{8: "world"}
		rv := reflect.NewAt(reflect.TypeOf(m), unsafe.Pointer(&m)).Elem()
		rv1 := ReflectValueWithPtr(rv, unsafe.Pointer(&replacement))

		if rv.UnsafePointer() != reflect.ValueOf(m).UnsafePointer() ||
			rv1.UnsafePointer() != reflect.ValueOf(replacement).UnsafePointer() {
			return errors.New("compatibility issue found: ReflectValueWithPtr")
		}

		m1, ok := rv1.Interface().(map[int]string)
		if !ok || !reflect.DeepEqual(replacement, m1) {
			return errors.New("compatibility issue found: ReflectValueWithPtr (Interface())")
		}

		sliceTemplate := reflect.New(reflect.TypeOf([]string(nil))).Elem()
		s := []string{"a"}
		ReflectValueWithPtr(sliceTemplate, unsafe.Pointer(&s)).Grow(4)
		if cap(s) < 5 || len(s) != 1 || s[0] != "a" || sliceTemplate.Cap() != 0 {
			return errors.New("compatibility issue found: ReflectValueWithPtr (Grow)")
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

	{ // Test SliceHeader - the encoder tells a present but empty []byte
		// from an unset one by the data pointer, which it reads as the first
		// word of the header without looking at the type
		var unset []byte
		empty := []byte{}
		if unsafe.Sizeof(SliceHeader{}) != unsafe.Sizeof(unset) ||
			unsafe.Alignof(SliceHeader{}) != unsafe.Alignof(unset) {
			return errors.New("compatibility issue found: SliceHeader size or alignment")
		}
		if unsafe.Offsetof(SliceHeader{}.Data) != 0 || unsafe.Sizeof(SliceHeader{}.Data) != unsafe.Sizeof(uintptr(0)) {
			return errors.New("compatibility issue found: SliceHeader.Data is not the first word")
		}
		if (*SliceHeader)(unsafe.Pointer(&unset)).Data != nil {
			return errors.New("compatibility issue found: nil slice has a data pointer")
		}
		if (*SliceHeader)(unsafe.Pointer(&empty)).Data == nil {
			return errors.New("compatibility issue found: empty slice has no data pointer")
		}
		b := make([]byte, 3, 5)
		h := (*SliceHeader)(unsafe.Pointer(&b))
		if h.Data != unsafe.Pointer(&b[0]) || h.Len != len(b) || h.Cap != cap(b) {
			return errors.New("compatibility issue found: SliceHeader fields")
		}
	}

	{ // StringHeader also reads the shared Data/Len prefix of slice headers.
		s := "hello"
		if unsafe.Sizeof(StringHeader{}) != unsafe.Sizeof(s) ||
			unsafe.Alignof(StringHeader{}) != unsafe.Alignof(s) ||
			unsafe.Offsetof(StringHeader{}.Data) != unsafe.Offsetof(SliceHeader{}.Data) ||
			unsafe.Offsetof(StringHeader{}.Len) != unsafe.Offsetof(SliceHeader{}.Len) {
			return errors.New("compatibility issue found: StringHeader layout")
		}
		h := (*StringHeader)(unsafe.Pointer(&s))
		if h.Data != unsafe.Pointer(unsafe.StringData(s)) || h.Len != len(s) {
			return errors.New("compatibility issue found: StringHeader fields")
		}
	}

	{ // Scalar and packed bool coders read bool values as one byte.
		f, t := false, true
		if unsafe.Sizeof(f) != 1 || *(*byte)(unsafe.Pointer(&f)) != 0 ||
			*(*byte)(unsafe.Pointer(&t)) != 1 {
			return errors.New("compatibility issue found: bool representation")
		}
	}

	return nil
}

// hackMapIter mirrors the internal structure of reflect.MapIter for unsafe access.
// This allows us to directly access the key and value pointers without allocations.
type hackMapIter struct {
	m      reflect.Value
	hitter struct {
		// k and v must be the first two fields of the runtime iterator.
		// testhack verifies this layout before the codec uses it.
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
func NewMapIter(rv reflect.Value) MapIter {
	return MapIter{*rv.MapRange()}
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
// rv must hold an indirect value, as returned by reflect.New(t).Elem().
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

// Iface mirrors Go's internal non-empty interface structure for direct manipulation.
// tab is stored as uintptr because itab is never allocated on heap, safe for GC.
type Iface struct {
	tab  uintptr        // pointer to itab
	data unsafe.Pointer // pointer to actual data
}

// IfaceData extracts the data pointer from a non-empty interface value.
// This provides direct access to the underlying concrete value without reflection overhead.
//
// p MUST point to a non-empty interface value.
func IfaceData(p unsafe.Pointer) unsafe.Pointer {
	return (*Iface)(p).data
}

// IfaceTypePtr extracts the type pointer from a non-empty interface value.
// This allows fast type checking and comparison without reflection.
//
// p MUST point to a non-empty interface value with a non-nil type table.
func IfaceTypePtr(p unsafe.Pointer) uintptr {
	tab := *(*unsafe.Pointer)(p)
	return (*itab)(tab).typ
}

// IfaceUpdate constructs a non-empty interface value by setting its type table and data pointer.
// The table must match the destination interface type and the concrete type at data.
func IfaceUpdate(p unsafe.Pointer, tab uintptr, data unsafe.Pointer) {
	(*Iface)(p).tab = tab
	(*Iface)(p).data = data
}

// IfaceTab generates the interface table (itab) for concrete type `v` implementing interface `t`.
// This is used for fast non-empty interface construction without repeated type assertions.
func IfaceTab(t reflect.Type, v reflect.Type) uintptr {
	if t.Kind() != reflect.Interface || v.Kind() != reflect.Struct {
		panic("input type mismatch")
	}
	i := &Iface{}
	reflect.NewAt(t, unsafe.Pointer(i)).Elem().Set(reflect.New(v))
	return i.tab
}

// ReflectValueTypePtr returns the abi.Type address as an integer for cache keys
// and type comparisons. Use ReflectTypePointer when the result will be dereferenced.
func ReflectValueTypePtr(rv reflect.Value) uintptr {
	return (*rvtype)(unsafe.Pointer(&rv)).abiType
}

// ReflectTypePtr returns the abi.Type address as an integer for type comparisons.
// Use ReflectTypePointer when the result will be dereferenced.
func ReflectTypePtr(rt reflect.Type) uintptr {
	return uintptr(ReflectTypePointer(rt))
}

// ReflectTypePointer extracts the abi.Type pointer from a reflect.Type.
// The *rtype from reflect package shares the same data structure as *abi.Type,
// allowing direct access to the underlying type information. Keep it as a pointer:
// types constructed by reflect can live on the heap.
func ReflectTypePointer(rt reflect.Type) unsafe.Pointer {
	return (*Iface)(unsafe.Pointer(&rt)).data
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
