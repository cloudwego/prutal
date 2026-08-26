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

package desc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/prutal/internal/protowire"
	"github.com/cloudwego/prutal/internal/wire"
)

var errGroupNotSupported = errors.New("group encoding not supported")

// Field number limits defined by the protobuf spec,
// sourced from internal/protowire to keep a single authority in the repo.
const (
	// maxFieldNumber is the maximum valid protobuf field number (2^29 - 1).
	maxFieldNumber = uint64(protowire.MaxValidNumber)

	// Field numbers in this range are reserved by the protobuf implementation.
	minReservedFieldNumber = uint64(protowire.FirstReservedNumber)
	maxReservedFieldNumber = uint64(protowire.LastReservedNumber)
)

func (p *FieldDesc) parseStructTag(tag string) error {
	ss := strings.Split(tag, ",")
loop:
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		switch {
		case strings.HasPrefix(s, "def="):
			// proto2 default: protobuf-go serializes it verbatim as the last
			// tag element, so it may contain commas (e.g. def=a,1); everything
			// after def= belongs to the default value, stop scanning
			break loop
		case s == "opt":
			// not in use
		case s == "req":
			p.Required = true // proto2
		case s == "rep":
			p.Repeated = true
		case s == "varint":
			p.TagType = TypeVarint
		case s == "zigzag32":
			p.TagType = TypeZigZag32
		case s == "zigzag64":
			p.TagType = TypeZigZag64
		case s == "fixed32":
			p.TagType = TypeFixed32
		case s == "fixed64":
			p.TagType = TypeFixed64
		case s == "bytes":
			p.TagType = TypeBytes
		case s == "group":
			return errGroupNotSupported
		case s == "packed":
			p.Packed = true
		case strings.Trim(s, "1234567890") == "":
			n, err := strconv.ParseUint(s, 10, 32)
			if err != nil {
				return err
			}
			// validate the unsigned value: int32(n) silently wraps for
			// n >= 1<<31 and would defeat the checks below
			if n == 0 || n > maxFieldNumber ||
				(n >= minReservedFieldNumber && n <= maxReservedFieldNumber) {
				return fmt.Errorf("invalid field number %d", n)
			}
			if p.ID != 0 {
				// a 2nd all-digit segment would silently renumber the field
				return fmt.Errorf("duplicated field number in tag: %d and %d", p.ID, n)
			}
			p.ID = int32(n)
		}
	}
	if p.TagType == 0 {
		return errors.New("unknown tag type")
	}
	if p.ID == 0 {
		return errors.New("missing or invalid field number")
	}
	if p.Packed {
		p.WireTag = wire.EncodeTag(p.ID, wire.TypeBytes)
	} else {
		p.WireTag = wire.EncodeTag(p.ID, wireTypes[p.TagType])
	}
	return nil
}

func parseKVTag(tag string, wantFieldNumber uint64, allowEnum bool) (TagType, error) {
	ss := strings.Split(tag, ",")
	var typ TagType
	var fieldNumber uint64
	hasFieldNumber := false
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		switch {
		case s == "varint":
			typ = TypeVarint
		case s == "zigzag32":
			typ = TypeZigZag32
		case s == "zigzag64":
			typ = TypeZigZag64
		case s == "fixed32":
			typ = TypeFixed32
		case s == "fixed64":
			typ = TypeFixed64
		case s == "bytes":
			typ = TypeBytes
		case s == "group":
			return 0, errGroupNotSupported
		case strings.HasPrefix(s, "enum="):
			if !allowEnum {
				return 0, errors.New("enum is not a valid map key type")
			}
		case strings.Trim(s, "1234567890") == "":
			n, err := strconv.ParseUint(s, 10, 32)
			if err != nil {
				return 0, err
			}
			if hasFieldNumber {
				return 0, fmt.Errorf("duplicated map entry field number: %d and %d", fieldNumber, n)
			}
			fieldNumber = n
			hasFieldNumber = true
		}
	}
	if typ == TypeUnknown {
		return TypeUnknown, fmt.Errorf("failed to parse tag type: %q", tag)
	}
	if fieldNumber != wantFieldNumber {
		return TypeUnknown, fmt.Errorf("invalid map entry field number %d, want %d", fieldNumber, wantFieldNumber)
	}
	return typ, nil
}
