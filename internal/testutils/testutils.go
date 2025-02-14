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
	"math/rand"
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
