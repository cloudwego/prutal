/*
 * Copyright 2026 CloudWeGo Authors
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

package prutal

import (
	"reflect"
	"testing"
)

type benchmarkRepeatedMessage struct {
	Value int64  `protobuf:"varint,1,opt"`
	Text  string `protobuf:"bytes,2,opt"`
}

func BenchmarkDecodeRepeated(b *testing.B) {
	numbers := make([]uint64, 128)
	strings := make([]string, len(numbers))
	messages := make([]benchmarkRepeatedMessage, len(numbers))
	pointers := make([]*benchmarkRepeatedMessage, len(numbers))
	for i := range numbers {
		numbers[i] = uint64(i + 1)
		strings[i] = "repeated value"
		messages[i] = benchmarkRepeatedMessage{Value: int64(i + 1), Text: strings[i]}
		pointers[i] = &messages[i]
	}
	cases := []struct {
		name string
		src  any
	}{
		{"Uint64", &struct {
			Values []uint64 `protobuf:"varint,1,rep"`
		}{numbers}},
		{"String", &struct {
			Values []string `protobuf:"bytes,1,rep"`
		}{strings}},
		{"Message", &struct {
			Values []benchmarkRepeatedMessage `protobuf:"bytes,1,rep"`
		}{messages}},
		{"MessagePointer", &struct {
			Values []*benchmarkRepeatedMessage `protobuf:"bytes,1,rep"`
		}{pointers}},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			data, err := MarshalAppend(nil, tc.src)
			if err != nil {
				b.Fatal(err)
			}
			typ := reflect.TypeOf(tc.src).Elem()
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := Unmarshal(data, reflect.New(typ).Interface()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
