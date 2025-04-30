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

package testutils

import (
	crand "crypto/rand"
	"math"
	"math/rand"
	"reflect"
	"time"
)

func Repeat[T any](n int, v T) []T {
	ret := make([]T, n)
	for i := 0; i < len(ret); i++ {
		ret[i] = v
	}
	return ret
}

func RandomBoolSlice(n int) []bool {
	v := uint64(0)
	ret := make([]bool, n)
	for i := range ret {
		if v == 0 {
			v = rand.Uint64()
		}
		ret[i] = ((v & 1) == 1)
		v = v >> 1
	}
	return ret
}

func RandomStr(n int) string {
	return string(RandomBytes(n))
}

func RandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := crand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

type FillOptions struct {
	Seed int64
	r    *rand.Rand

	MapMinSize   int
	MapMaxSize   int
	SliceMinSize int
	SliceMaxSize int
	StringMinLen int
	StringMaxLen int

	RecursionDepth int
}

func (o *FillOptions) RandRange(min, max int) int {
	if max <= min {
		return max
	}
	return min + o.r.Intn(max-min+1)
}

func DefaultFillOptions() FillOptions {
	return FillOptions{
		Seed: time.Now().UnixNano(),

		MapMinSize:   0,
		MapMaxSize:   5,
		SliceMinSize: 0,
		SliceMaxSize: 5,
		StringMinLen: 0,
		StringMaxLen: 10,

		RecursionDepth: 2,
	}
}

func RandFill(m any, oo FillOptions) {
	v := reflect.ValueOf(m)
	oo.r = rand.New(rand.NewSource(oo.Seed))
	fillRandomValue(v, oo)
}

func fillRandomValue(v reflect.Value, oo FillOptions) {
	if oo.RecursionDepth < 0 {
		return
	}
	oo.RecursionDepth--

	// Handle indirection for pointers
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			if !v.CanSet() {
				return
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	// If the value cannot be set, return early
	if !v.CanSet() {
		return
	}

	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(oo.r.Intn(2) == 1)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(oo.r.Int63())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v.SetUint(oo.r.Uint64())

	case reflect.Float32, reflect.Float64:
		v.SetFloat(oo.r.Float64())

	case reflect.Complex64, reflect.Complex128:
		v.SetComplex(complex(oo.r.Float64(), oo.r.Float64()))

	case reflect.String:
		length := oo.RandRange(oo.StringMinLen, oo.StringMaxLen)
		v.SetString(RandomStr(length))

	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 { // []byte
			length := oo.RandRange(oo.StringMinLen, oo.StringMaxLen)
			v.SetBytes(RandomBytes(length))
			return
		}

		length := oo.RandRange(oo.SliceMinSize, oo.SliceMaxSize)
		slice := reflect.MakeSlice(v.Type(), length, length)
		for i := 0; i < length; i++ {
			fillRandomValue(slice.Index(i), oo)
		}
		v.Set(slice)

	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			fillRandomValue(v.Index(i), oo)
		}

	case reflect.Map:
		mt := v.Type()
		if v.IsNil() {
			v.Set(reflect.MakeMap(mt))
		}

		// For bool keys, we can only have 2 entries max
		maxEntries := oo.MapMaxSize
		if mt.Key().Kind() == reflect.Bool {
			maxEntries = 2
		}
		count := oo.RandRange(oo.MapMinSize, maxEntries)
		for i := 0; i < count; i++ {
			key := reflect.New(mt.Key()).Elem()
			value := reflect.New(mt.Elem()).Elem()

			fillRandomValue(key, oo)
			fillRandomValue(value, oo)

			// Skip if key can't be used as map key (e.g., NaN for float)
			if mt.Key().Kind() == reflect.Float32 || mt.Key().Kind() == reflect.Float64 {
				if math.IsNaN(key.Float()) || math.IsInf(key.Float(), 0) {
					continue
				}
			}

			v.SetMapIndex(key, value)
		}

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.CanSet() {
				fillRandomValue(field, oo)
			}
		}
	}
}
