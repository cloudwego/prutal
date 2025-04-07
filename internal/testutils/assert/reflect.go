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

package assert

import (
	"errors"
	"fmt"
	"reflect"
)

// DeepEqual ...
func DeepEqual(t TestingT, a, b any, msgs ...interface{}) {
	if err := reflectEqual(reflect.ValueOf(a), reflect.ValueOf(b)); err != nil {
		t.Helper()
		t.Fatalf("not equal: %s %s", err, fmt.Sprint(msgs...))
	}
}

func checkValid(a, b reflect.Value) error {
	if a.IsValid() == b.IsValid() {
		return nil
	}
	return errors.New("not IsValid()")
}

func reflectEqual(a, b reflect.Value) error {
	if !a.IsValid() || !b.IsValid() {
		return checkValid(a, b)
	}
	if a.Type() != b.Type() {
		return errors.New("type not equal")
	}
	a = dereference(a)
	b = dereference(b)
	if !a.IsValid() || !b.IsValid() {
		return checkValid(a, b)
	}
	switch a.Kind() {
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			if !a.Type().Field(i).IsExported() {
				continue
			}
			f0 := a.Field(i)
			f1 := b.Field(i)
			if err := reflectEqual(f0, f1); err != nil {
				return fmt.Errorf("field %q: %w", a.Type().Field(i).Name, err)
			}
		}

	case reflect.Slice, reflect.Array:
		if a.Len() != b.Len() {
			return errors.New("len not equal")
		}
		for i := 0; i < a.Len(); i++ {
			if err := reflectEqual(a.Index(i), b.Index(i)); err != nil {
				return fmt.Errorf("elem[%d]: %w", i, err)
			}
		}

	case reflect.Map:
		if a.Len() != b.Len() {
			return errors.New("len not equal")
		}
		iter := a.MapRange()
		for iter.Next() {
			k := iter.Key()
			v0 := iter.Value()
			v1 := b.MapIndex(k)
			if err := reflectEqual(v0, v1); err != nil {
				return fmt.Errorf("elem[%v]: %w", k, err)
			}
		}

	case reflect.Bool:
		if a.Bool() != b.Bool() {
			return newNotEqual(a, b)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if a.Int() != b.Int() {
			return newNotEqual(a, b)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if a.Uint() != b.Uint() {
			return newNotEqual(a, b)
		}

	case reflect.Float32, reflect.Float64:
		if a.Float() != b.Float() {
			return newNotEqual(a, b)
		}

	case reflect.String:
		if a.String() != b.String() {
			return newNotEqual(a, b)
		}

	}

	return nil
}

func newNotEqual(a, b reflect.Value) error {
	return fmt.Errorf("%v != %v", a, b)
}

func dereference(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}
