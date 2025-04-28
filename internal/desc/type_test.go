package desc

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestMapTmpVars(t *testing.T) {
	type S struct {
		V int64
	}

	testcases := []struct {
		Name     string
		Type     func() reflect.Type
		UpdateKV func(k, v unsafe.Pointer)
		Expect   func() any
	}{
		{
			Name: "mapUpdateFunc_u64_unsafe",
			Type: func() reflect.Type { return reflect.TypeOf(map[uint64]*S{}) },
			UpdateKV: func(k, v unsafe.Pointer) {
				*(*uint64)(k) = 1
				*(**S)(v) = &S{V: 2}
			},
			Expect: func() any { return map[uint64]*S{1: &S{V: 2}} },
		},
		{
			Name: "mapUpdateFunc_u64_string",
			Type: func() reflect.Type { return reflect.TypeOf(map[uint64]string{}) },
			UpdateKV: func(k, v unsafe.Pointer) {
				*(*uint64)(k) = 42
				*(*string)(v) = "hello"
			},
			Expect: func() any { return map[uint64]string{42: "hello"} },
		},
		{
			Name: "mapUpdateFunc_u64_empty",
			Type: func() reflect.Type { return reflect.TypeOf(map[uint64]struct{}{}) },
			UpdateKV: func(k, v unsafe.Pointer) {
				*(*uint64)(k) = 7
			},
			Expect: func() any { return map[uint64]struct{}{7: {}} },
		},
		{
			Name: "mapUpdateFunc_u32_unsafe",
			Type: func() reflect.Type { return reflect.TypeOf(map[uint32]*S{}) },
			UpdateKV: func(k, v unsafe.Pointer) {
				*(*uint32)(k) = 3
				*(**S)(v) = &S{V: 9}
			},
			Expect: func() any { return map[uint32]*S{3: &S{V: 9}} },
		},
		{
			Name: "mapUpdateFunc_u32_string",
			Type: func() reflect.Type { return reflect.TypeOf(map[uint32]string{}) },
			UpdateKV: func(k, v unsafe.Pointer) {
				*(*uint32)(k) = 5
				*(*string)(v) = "world"
			},
			Expect: func() any { return map[uint32]string{5: "world"} },
		},
		{
			Name: "mapUpdateFunc_u32_empty",
			Type: func() reflect.Type { return reflect.TypeOf(map[uint32]struct{}{}) },
			UpdateKV: func(k, v unsafe.Pointer) {
				*(*uint32)(k) = 8
			},
			Expect: func() any { return map[uint32]struct{}{8: {}} },
		},
		{
			Name: "mapUpdateFunc_string_unsafe",
			Type: func() reflect.Type { return reflect.TypeOf(map[string]*S{}) },
			UpdateKV: func(k, v unsafe.Pointer) {
				*(*string)(k) = "foo"
				*(**S)(v) = &S{V: 11}
			},
			Expect: func() any { return map[string]*S{"foo": &S{V: 11}} },
		},
		{
			Name: "mapUpdateFunc_string_string",
			Type: func() reflect.Type { return reflect.TypeOf(map[string]string{}) },
			UpdateKV: func(k, v unsafe.Pointer) {
				*(*string)(k) = "bar"
				*(*string)(v) = "baz"
			},
			Expect: func() any { return map[string]string{"bar": "baz"} },
		},
		{
			Name: "mapUpdateFunc_string_empty",
			Type: func() reflect.Type { return reflect.TypeOf(map[string]struct{}{}) },
			UpdateKV: func(k, v unsafe.Pointer) {
				*(*string)(k) = "empty"
			},
			Expect: func() any { return map[string]struct{}{"empty": {}} },
		},
		{
			Name: "mapUpdateFunc_default",
			Type: func() reflect.Type { return reflect.TypeOf(map[string]S{}) },
			UpdateKV: func(k, v unsafe.Pointer) {
				*(*string)(k) = "foo"
				*(*S)(v) = S{V: 7}
			},
			Expect: func() any { return map[string]S{"foo": {V: 7}} },
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			rt := tc.Type()
			newfunc := newTmpMapVarsFunc(rt)
			p := newfunc().(*TmpMapVars)
			tc.UpdateKV(p.KeyPointer(), p.ValPointer())

			m := reflect.New(rt)
			m.Elem().Set(reflect.MakeMap(rt))
			p.Update(p, m.UnsafePointer())
			assert.DeepEqual(t, tc.Expect(), m.Elem().Interface())
		})
	}
}
