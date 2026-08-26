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
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestReservedRange(t *testing.T) {
	rr := reservedRanges{}
	rr = append(rr, reservedRange{1, 2}, reservedRange{5, 6})
	assert.False(t, rr.In(0))
	assert.True(t, rr.In(1))
	assert.True(t, rr.In(2))
	assert.False(t, rr.In(3))
	assert.False(t, rr.In(4))
	assert.True(t, rr.In(5))
	assert.True(t, rr.In(6))
	assert.False(t, rr.In(7))
	assert.False(t, rr.In(19000))
	assert.False(t, rr.In(19999))
}

func TestLoader_Reserved(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "example.com/testmessage";
message M {
  string f = 1;

	reserved 3,5;
	reserved 7 to 10;
	reserved 100 to max;

  enum e {
	reserved 30,50;
	reserved 70 to 100;
	reserved 1000 to max;
	ZERO = 0;
  }
}
`)

	m := p.Messages[0]
	type testcase struct {
		f int32
		v bool
	}

	{ // (*Message) IsReservedField
		cases := []testcase{
			{2, false},
			{3, true},
			{4, false},
			{5, true},
			{6, false},
			{7, true},
			{8, true},
			{10, true},
			{11, false},
			{99, false},
			{100, true},
			{101, true},
			{19000, true},
			{19999, true},
			{1000000, true},
		}
		for _, c := range cases {
			assert.Equal(t, c.v, m.IsReservedField(c.f), c.f)
		}
	}

	{ // (*Enum) IsReservedField
		e := m.Enums[0]
		cases := []testcase{
			{20, false},
			{30, true},
			{40, false},
			{50, true},
			{60, false},
			{70, true},
			{80, true},
			{100, true},
			{110, false},
			{990, false},
			{1000, true},
			{1001, true},
			{10000000, true},
		}
		for _, c := range cases {
			assert.Equal(t, c.v, e.IsReservedField(c.f), c.f)
		}
	}

}

func TestLoaderReservedContract(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "example.com/reserved";
message M {
  reserved "not-a-field";
  reserved "fo" "o", "fo";
  reserved 1 to 3, 4 to 5;
  reserved 18000 to 20000;
  reserved 536870911 to max;
}
enum E {
  reserved "ba" 'r';
  reserved -2147483648 to -2;
  reserved 2147483647 to max;
  MINUS_ONE = -1;
  ZERO = 0;
  V19000 = 19000;
  V19999 = 19999;
}
`)

	m := p.Messages[0]
	assert.True(t, m.reservedNames.Has("foo"))
	assert.True(t, m.reservedNames.Has("fo"))
	assert.True(t, m.IsReservedField(19000))
	assert.True(t, m.IsReservedField(19999))
	assert.True(t, m.IsReservedField(536870911))

	e := p.Enums[0]
	assert.True(t, e.reservedNames.Has("bar"))
	assert.True(t, e.IsReservedField(-2147483648))
	assert.True(t, e.IsReservedField(-2))
	assert.False(t, e.IsReservedField(-1))
	assert.False(t, e.IsReservedField(19000))
	assert.False(t, e.IsReservedField(19999))
	assert.True(t, e.IsReservedField(2147483647))
}

func TestLoaderEditionReservedNames(t *testing.T) {
	p := loadTestProto(t, `
edition = "2023";
option go_package = "example.com/reserved";
message M {
  reserved old, legacy, required, inf, nan;
  int32 live = 1;
}
enum E {
  reserved OLD, LEGACY, required, inf, nan;
  ZERO = 0;
}
`)
	assert.True(t, p.Messages[0].reservedNames.Has("old"))
	assert.True(t, p.Messages[0].reservedNames.Has("required"))
	assert.True(t, p.Messages[0].reservedNames.Has("inf"))
	assert.True(t, p.Messages[0].reservedNames.Has("nan"))
	assert.True(t, p.Enums[0].reservedNames.Has("OLD"))
	assert.True(t, p.Enums[0].reservedNames.Has("required"))
	assert.True(t, p.Enums[0].reservedNames.Has("inf"))
	assert.True(t, p.Enums[0].reservedNames.Has("nan"))

	expectProtoError(t, `
edition = "2023";
option go_package = "example.com/reserved";
message M { reserved old; int32 old = 1; }
`, `"old" uses reserved name`)
	expectProtoError(t, `
edition = "2023";
option go_package = "example.com/reserved";
enum E { reserved OLD; OLD = 0; }
`, `"OLD" uses reserved name`)
	expectProtoError(t, `
edition = "2023";
option go_package = "example.com/reserved";
message M { reserved "old"; }
`, "reserved names must be identifiers in editions")
}

func TestLoaderReservedStringEscapes(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "example.com/reserved";
message M {
  reserved "hex\x6", "upper\X42", "oct\7", "octtwo\77", "octwrap\400";
  reserved "unicode\u0062", "wide\U00000063", "pair\uD83D\uDE00", "question\?";
  reserved "surrogate\uD800", "above\U00110000", "legacymax\U001FFFFF";
  reserved "cross\uD83D" "\uDE00";
}
`)
	names := p.Messages[0].reservedNames
	assert.True(t, names.Has("upperB"))
	assert.True(t, names.Has("octwrap\x00"))
	assert.True(t, names.Has("unicodeb"))
	assert.True(t, names.Has("widec"))
	assert.True(t, names.Has("pair😀"))
	assert.True(t, names.Has("question?"))
	assert.True(t, names.Has("surrogate"+string([]byte{0xed, 0xa0, 0x80})))
	assert.True(t, names.Has(`above\U00110000`))
	assert.True(t, names.Has(`legacymax\U001fffff`))
	assert.True(t, names.Has("cross"+string([]byte{0xed, 0xa0, 0xbd, 0xed, 0xb8, 0x80})))
}

func TestLoaderExtensionRanges(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "example.com/reserved";
message M { extensions 10 to 20, 100 to max; }
`)
	ranges := p.Messages[0].extensionRanges
	assert.True(t, ranges.In(10))
	assert.True(t, ranges.In(20))
	assert.True(t, ranges.In(536870911))
}

func TestLoaderReservedIntegerLiterals(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "example.com/reserved";
message M {
  reserved 010, 0x10;
}
enum E {
  reserved -0x10, 010;
  ZERO = 0;
}
`)

	m := p.Messages[0]
	assert.True(t, m.IsReservedField(8))
	assert.False(t, m.IsReservedField(10))
	assert.True(t, m.IsReservedField(16))

	e := p.Enums[0]
	assert.True(t, e.IsReservedField(-16))
	assert.True(t, e.IsReservedField(8))
}

func TestLoaderReservedRejectsInvalidDeclarations(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload string
		want    string
	}{
		{
			"duplicate message name",
			`message M { reserved "old", "old"; }`,
			"reserved name \"old\" is duplicated",
		},
		{
			"escaped duplicate enum name",
			`enum E { reserved "foo"; reserved "f\157o"; ZERO = 0; }`,
			"reserved name \"foo\" is duplicated",
		},
		{
			"concatenated duplicate message name",
			`message M { reserved "fo" "o", "foo"; }`,
			"reserved name \"foo\" is duplicated",
		},
		{
			"concatenated reserved message name",
			`message M { reserved "fo" "o"; int32 foo = 1; }`,
			`"foo" uses reserved name`,
		},
		{
			"concatenated reserved enum name",
			`enum E { reserved "FO" "O"; FOO = 0; }`,
			`"FOO" uses reserved name`,
		},
		{
			"legacy identifier reserved name",
			`enum E { reserved OLD; ZERO = 0; }`,
			"reserved names must be string literals before editions",
		},
		{
			"overlapping message ranges",
			`message M { reserved 1 to 3; reserved 3 to 5; }`,
			"overlaps",
		},
		{
			"overlapping enum ranges",
			`enum E { reserved -2 to 0; reserved 0 to 2; ZERO = 3; }`,
			"overlaps",
		},
		{
			"message range starts at zero",
			`message M { reserved 0; }`,
			"outside [1, 536870911]",
		},
		{
			"negative message range",
			`message M { reserved -2 to -1; }`,
			"outside [1, 536870911]",
		},
		{
			"message range above maximum",
			`message M { reserved 536870912 to max; }`,
			"outside [1, 536870911]",
		},
		{
			"explicit enum reserved value",
			`enum E { reserved 19000; V = 19000; }`,
			"19000 is reserved",
		},
		{
			"negative extension range",
			`message M { extensions -2 to -1; }`,
			"extension range must not use negative numbers",
		},
		{
			"Unicode escape above source limit",
			`message M { reserved "x\U00200000"; }`,
			"invalid Unicode escape",
		},
		{
			"canonical duplicate legacy Unicode escape",
			`message M { reserved "x\U001FFFFF", "x\U001fffff"; }`,
			`reserved name "x\\U001fffff" is duplicated`,
		},
		{
			"adjacent range keyword",
			`message M { reserved 2to 3; }`,
			"space is required between a number",
		},
		{
			"overlapping extension ranges",
			`message M { extensions 10 to 20; extensions 20 to 30; }`,
			"extension range [20, 30] overlaps [10, 20]",
		},
		{
			"reserved and extension range overlap",
			`message M { reserved 10 to 20; extensions 15 to 25; }`,
			"overlaps extension range",
		},
		{
			"field in extension range",
			`message M { extensions 10 to 20; optional int32 f = 15; }`,
			"is in an extension range",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expectProtoError(t, `option go_package = "example.com/reserved";`+tc.payload, tc.want)
		})
	}

	expectProtoError(t, `
syntax = "proto3";
option go_package = "example.com/reserved";
message M { extensions 10 to 20; }
`, "extension ranges are not allowed in proto3")
}
