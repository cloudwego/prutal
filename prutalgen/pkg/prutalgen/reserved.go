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
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cloudwego/prutal/internal/protowire"
	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/text"
	"github.com/cloudwego/prutal/prutalgen/internal/protoparser"
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

func (r reservedRange) overlaps(other reservedRange) bool {
	return r.from <= other.to && other.from <= r.to
}

type reservedRanges []reservedRange

func (rr reservedRanges) In(v int32) bool {
	for _, r := range rr {
		if v >= r.from && v <= r.to {
			return true
		}
	}
	return false
}

type reservedNames []string

func (rn reservedNames) Has(n string) bool {
	for _, x := range rn {
		if x == n {
			return true
		}
	}
	return false
}

func (x *protoLoader) loadMessageReserved(m *Message, decl *protoparser.ReservedDecl) {
	ranges, names := x.loadReserved(decl, int32(protowire.MinValidNumber), int32(protowire.MaxValidNumber),
		m.reserved, m.reservedNames)
	m.reserved = append(m.reserved, ranges...)
	m.reservedNames = append(m.reservedNames, names...)
}

func (x *protoLoader) loadEnumReserved(e *Enum, decl *protoparser.ReservedDecl) {
	ranges, names := x.loadReserved(decl, int32(math.MinInt32), int32(math.MaxInt32),
		e.reserved, e.reservedNames)
	e.reserved = append(e.reserved, ranges...)
	e.reservedNames = append(e.reservedNames, names...)
}

func (x *protoLoader) loadReserved(
	decl *protoparser.ReservedDecl,
	minValue, maxValue int32,
	existingRanges reservedRanges,
	existingNames reservedNames,
) (reservedRanges, reservedNames) {
	if len(decl.Names) > 0 {
		names := x.parseReservedNames(decl)
		seen := make(map[string]bool, len(existingNames)+len(names))
		for _, name := range existingNames {
			seen[name] = true
		}
		for i, name := range names {
			if seen[name] {
				x.Fatalf("%s - reserved name %q is duplicated", decl.Names[i].Pos, name)
			}
			seen[name] = true
		}
		return nil, names
	}
	if len(decl.Ranges) > 0 {
		return x.parseRanges(decl.Ranges, minValue, maxValue, existingRanges, "reserved range", true), nil
	}
	return nil, nil
}

func (x *protoLoader) loadExtensions(m *Message, decl *protoparser.ExtensionsDecl) {
	if x.currentProto().IsProto3() {
		x.Fatalf("%s - extension ranges are not allowed in proto3", decl.Pos)
	}
	_ = x.loadBracketOptions(decl.Options, optionTargetExtensionRange)
	ranges := x.parseRanges(decl.Ranges, int32(protowire.MinValidNumber), int32(protowire.MaxValidNumber),
		m.extensionRanges, "extension range", false)
	m.extensionRanges = append(m.extensionRanges, ranges...)
}

func (x *protoLoader) parseReservedNames(decl *protoparser.ReservedDecl) reservedNames {
	if x.currentProto().IsEdition2023() {
		for _, rn := range decl.Names {
			if rn.StringLit != nil {
				x.Fatalf("%s - reserved names must be identifiers in editions", rn.Pos)
			}
		}
		names := make(reservedNames, 0, len(decl.Names))
		for _, rn := range decl.Names {
			names = append(names, rn.Ident)
		}
		return names
	}
	for _, rn := range decl.Names {
		if rn.StringLit == nil {
			x.Fatalf("%s - reserved names must be string literals before editions", rn.Pos)
		}
	}
	names := make(reservedNames, 0, len(decl.Names))
	for _, rn := range decl.Names {
		var value strings.Builder
		for _, part := range rn.StringLit.Parts {
			s, err := unmarshalProtoSourceString(part)
			if err != nil {
				x.Fatalf("%s - reserved name syntax err: %s", rn.Pos, err)
			}
			value.WriteString(s)
		}
		names = append(names, value.String())
	}
	return names
}

// unmarshalProtoSourceString normalizes source-language escapes that differ
// from textproto escapes before delegating the common string decoder.
func unmarshalProtoSourceString(s string) (string, error) {
	if len(s) < 2 {
		return text.UnmarshalString(s)
	}
	end := len(s) - 1
	var normalized strings.Builder
	normalized.Grow(len(s))
	normalized.WriteByte(s[0])
	for i := 1; i < end; {
		if s[i] != '\\' || i+1 >= end {
			normalized.WriteByte(s[i])
			i++
			continue
		}

		switch s[i+1] {
		case '0', '1', '2', '3', '4', '5', '6', '7':
			j := i + 1
			for j < end && j < i+4 && s[j] >= '0' && s[j] <= '7' {
				j++
			}
			v, _ := strconv.ParseUint(s[i+1:j], 8, 16)
			fmt.Fprintf(&normalized, "\\%03o", byte(v))
			i = j
		case 'x', 'X':
			j := i + 2
			for j < end && j < i+4 && isHexDigit(s[j]) {
				j++
			}
			v, _ := strconv.ParseUint(s[i+2:j], 16, 8)
			fmt.Fprintf(&normalized, "\\%03o", byte(v))
			i = j
		case 'u', 'U':
			v, next, _, ok := parseProtoUnicodeEscape(s, i, end)
			if !ok {
				normalized.WriteString(s[i : i+2])
				i += 2
				continue
			}
			if v >= 0x200000 {
				return "", fmt.Errorf("invalid Unicode escape %q", s[i:next])
			}
			if v > 0x10ffff {
				fmt.Fprintf(&normalized, `\\%c%08x`, s[i+1], v)
				i = next
				continue
			}
			if v >= 0xd800 && v <= 0xdbff {
				low, pairEnd, pairKind, pairOK := parseProtoUnicodeEscape(s, next, end)
				if pairOK && pairKind == 'u' && low >= 0xdc00 && low <= 0xdfff {
					normalized.WriteString(s[i:pairEnd])
					i = pairEnd
					continue
				}
			}
			if v >= 0xd800 && v <= 0xdfff {
				appendProtoCodePoint(&normalized, v)
				i = next
				continue
			}
			normalized.WriteString(s[i:next])
			i = next
		default:
			normalized.WriteString(s[i : i+2])
			i += 2
		}
	}
	normalized.WriteByte(s[end])
	return text.UnmarshalString(normalized.String())
}

func parseProtoUnicodeEscape(s string, start, end int) (value uint64, next int, kind byte, ok bool) {
	if start+2 > end || s[start] != '\\' {
		return 0, start, 0, false
	}
	kind = s[start+1]
	digits := 0
	switch kind {
	case 'u':
		digits = 4
	case 'U':
		digits = 8
	default:
		return 0, start, 0, false
	}
	next = start + 2 + digits
	if next > end {
		return 0, start, 0, false
	}
	v, err := strconv.ParseUint(s[start+2:next], 16, 32)
	if err != nil {
		return 0, start, 0, false
	}
	return v, next, kind, true
}

func appendProtoCodePoint(out *strings.Builder, v uint64) {
	for _, b := range []byte{
		0xe0 | byte(v>>12),
		0x80 | byte(v>>6)&0x3f,
		0x80 | byte(v)&0x3f,
	} {
		fmt.Fprintf(out, "\\%03o", b)
	}
}

func isHexDigit(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

func (x *protoLoader) parseRanges(
	rr []*protoparser.RangeDecl,
	minValue, maxValue int32,
	existing reservedRanges,
	kind string,
	allowNegative bool,
) reservedRanges {
	ranges := make([]reservedRange, 0, len(rr))
	for _, r := range rr {
		if !allowNegative && (strings.HasPrefix(r.Start, "-") || strings.HasPrefix(r.End, "-")) {
			x.Fatalf("%s - %s must not use negative numbers", r.Pos, kind)
		}
		from, ok := text.UnmarshalI32(r.Start)
		if !ok {
			x.Fatalf("%s - invalid integer %q", r.StartPos, r.Start)
		}
		v := reservedRange{from: from}
		if !r.IsRange {
			v.to = v.from
		} else if r.IsMax {
			v.to = maxValue
		} else {
			to, ok := text.UnmarshalI32(r.End)
			if !ok {
				x.Fatalf("%s - invalid integer %q", r.EndPos, r.End)
			}
			v.to = to
		}
		if v.from < minValue || v.from > maxValue || v.to < minValue || v.to > maxValue {
			x.Fatalf("%s - %s [%d, %d] is outside [%d, %d]",
				r.Pos, kind, v.from, v.to, minValue, maxValue)
		}
		if v.to < v.from {
			x.Fatalf("%s - %s end %d is less than start %d", r.Pos, kind, v.to, v.from)
		}
		checkOverlap := func(previous reservedRange) {
			if v.overlaps(previous) {
				x.Fatalf("%s - %s [%d, %d] overlaps [%d, %d]",
					r.Pos, kind, v.from, v.to, previous.from, previous.to)
			}
		}
		for _, previous := range existing {
			checkOverlap(previous)
		}
		for _, previous := range ranges {
			checkOverlap(previous)
		}
		ranges = append(ranges, v)
	}
	return ranges
}
