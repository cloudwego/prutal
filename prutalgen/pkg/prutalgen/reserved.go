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

package prutalgen

import (
	"math"

	"github.com/cloudwego/prutal/prutalgen/internal/parser"
)

const (
	// reserved by protobuf implementation
	protobufReservedMin = 19000
	protobufReservedMax = 19999
)

type reservedRange struct {
	from int32
	to   int32
}

type reservedRanges []reservedRange

func (rr reservedRanges) In(v int32) bool {
	if v >= protobufReservedMin && v <= protobufReservedMax {
		return true
	}
	for _, r := range rr {
		if v >= r.from && v <= r.to {
			return true
		}
	}
	return false
}

func (x *protoLoader) ExitReserved(c *parser.ReservedContext) {
	rc := c.Ranges()
	if rc == nil {
		return
	}
	rr := rc.AllOneRange()
	ranges := make([]reservedRange, 0, len(rr))
	for _, r := range rr {
		i, err := parseI32(r.IntLit(0))
		if err != nil {
			x.Fatalf("%s", err)
		}
		v := reservedRange{from: int32(i)}
		if r.TO() == nil {
			v.to = v.from
		} else if r.MAX() != nil {
			v.to = math.MaxInt32
		} else {
			i, err = parseI32(r.IntLit(1))
			if err != nil {
				x.Fatalf("%s", err)
			}
			v.to = i
		}
		ranges = append(ranges, v)
	}

	switch getRuleIndex(c.GetParent()) {
	case parser.ProtobufParserRULE_messageElement:
		m := x.currentMsg()
		m.reserved = append(m.reserved, ranges...)

	case parser.ProtobufParserRULE_enumElement:
		e := x.enum
		e.reserved = append(e.reserved, ranges...)
	}
}
