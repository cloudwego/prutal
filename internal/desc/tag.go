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
)

var errGroupNotSupported = errors.New("group encoding not supported")

func (p *FieldDesc) parseStructTag(tag string) error {
	ss := strings.Split(tag, ",")
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		switch {
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
			p.ID = int32(n)
		}
	}
	if p.TagType == 0 {
		return errors.New("unknown tag type")
	}
	p.WireType = wireTypes[p.TagType]
	return nil
}

func parseKVTag(tag string) (TagType, error) {
	ss := strings.Split(tag, ",")
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		switch { // ignore field num 1 for k, 2 for v, and opt, only need type
		case s == "varint":
			return TypeVarint, nil
		case s == "zigzag32":
			return TypeZigZag32, nil
		case s == "zigzag64":
			return TypeZigZag64, nil
		case s == "fixed32":
			return TypeFixed32, nil
		case s == "fixed64":
			return TypeFixed64, nil
		case s == "bytes":
			return TypeBytes, nil
		case s == "group":
			return 0, errGroupNotSupported
		}
	}
	return TypeUnknown, fmt.Errorf("failed to parse tag type: %q", tag)
}
