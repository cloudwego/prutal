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
	"fmt"
	"strings"
)

// hexdumpAt outputs like `hexdump -C`,
// but it only outputs 3 lines, 16 bytes per line, and 2nd line contains b[i].
//
// format: addr
// 00000870  75 72 6e 0a 09 7d 0a 09  66 6f 72 20 69 2c 20 76  |urn..}..for i, v|
func hexdumpAt(b []byte, i int) string {
	pos1 := i - (i % 16)
	pos0 := pos1 - 16
	pos2 := pos1 + 16

	f := &strings.Builder{}
	for _, pos := range []int{pos0, pos1, pos2} {
		if pos < 0 || pos >= len(b) {
			continue
		}
		fmt.Fprintf(f, "%08x", pos)
		for i := 0; i < 16; i++ {
			if pos+i >= len(b) {
				f.WriteString(" --")
			} else {
				fmt.Fprintf(f, " %02x", b[pos+i])
			}
		}
		f.WriteString(" |")
		for i := 0; i < 16; i++ {
			if pos+i >= len(b) {
				f.WriteString(".")
				continue
			}
			c := b[pos+i]
			if c >= 32 && c < 127 {
				f.WriteByte(c) // printable ascii
			} else {
				f.WriteByte('.')
			}
		}
		f.WriteString("|\n")
	}
	return f.String()
}
