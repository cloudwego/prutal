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

func TestEnum(t *testing.T) {
	e := &Enum{Proto: &Proto{}}
	assert.False(t, e.genNoPrefix())
	assert.True(t, e.genMapping())

	// genNoPrefix, case: Enum Directives
	e.Directives = Directives{prutalNoEnumPrefix}
	e.Options = nil
	assert.True(t, e.genNoPrefix())
	e.Directives = nil
	assert.False(t, e.genNoPrefix())

	// genNoPrefix, case: gogoproto options
	oo := Options{
		{Name: gogoproto_enum_prefix, Value: "false"},     // lower higher priority
		{Name: gogoproto_enum_prefix_all, Value: "false"}, // lower higher priority
	}
	e.Directives = nil
	e.Options = oo
	assert.True(t, e.genNoPrefix())

	// genNoPrefix, case: Proto Directives
	e.Options = nil
	e.Proto.Directives = Directives{prutalNoEnumPrefix}
	assert.True(t, e.genNoPrefix())
	e.Proto.Directives = nil
	assert.False(t, e.genNoPrefix())

	// genNoPrefix, case: Proto gogoproto options
	e.Proto.Options = oo
	assert.True(t, e.genNoPrefix())
	e.Proto.Options = nil

	// genMapping, case: Enum Directives
	e.Directives = Directives{prutalNoEnumMapping}
	assert.False(t, e.genMapping())

	// genMapping, case: Proto Directives
	e.Directives = nil
	e.Proto.Directives = Directives{prutalNoEnumMapping}
	assert.False(t, e.genMapping())
	e.Proto.Directives = nil
	assert.True(t, e.genMapping())

}

func TestEnum_Verify(t *testing.T) {
	p := &Proto{Package: "test.enum.verify"}
	e := &Enum{
		Name:  "e",
		Proto: p,
		Fields: []*EnumField{
			{Name: "ev1", Value: 1},
		},
	}
	p.Enums = []*Enum{e}

	// reserved
	e.reserved = append(e.reserved, reservedRange{from: 1, to: 1})
	assert.ErrorContains(t, p.verify(), "1 is reserved")
	e.reserved = nil
	assert.NoError(t, p.verify())

	// duplicated
	e.Fields = append(e.Fields, e.Fields[0])
	assert.ErrorContains(t, p.verify(), "1 is duplicated")
	e.Fields = e.Fields[:1]
	assert.NoError(t, p.verify())

	// allow_alias
	e.Fields = []*EnumField{
		{Name: "A", Value: 0},
		{Name: "B", Value: 1},
		{Name: "C", Value: 1}, // alias of B
	}
	assert.ErrorContains(t, p.verify(), "1 is duplicated")
	e.Options = Options{{Name: option_allow_alias, Value: "true"}}
	assert.NoError(t, p.verify())
	e.Fields = e.Fields[:2]
	assert.ErrorContains(t, p.verify(), "allows aliases, but none were found")
	e.Options = nil

	// open enum validation
	p.Edition = editionProto3
	e.Fields = []*EnumField{{Name: "E_ONE", Value: 1}}
	assert.ErrorContains(t, p.verify(), "must have zero number for the first value")
	e.Fields = []*EnumField{
		{Name: "E_FOO", Value: 0},
		{Name: "FOO", Value: 1},
	}
	assert.ErrorContains(t, p.verify(), "name conflict")
	e.Fields = nil
	assert.ErrorContains(t, p.verify(), "must contain at least one value declaration")
}

func TestEnumResolveDoesNotRenameValuesForMappingHelpers(t *testing.T) {
	p := &Proto{Directives: Directives{prutalNoEnumMapping}}
	e := &Enum{
		Name:  "e",
		Proto: p,
		Fields: []*EnumField{
			{Name: "name"},
			{Name: "name_"},
		},
	}
	for _, f := range e.Fields {
		f.Enum = e
	}

	e.resolve()
	assert.Equal(t, "E_name", e.Fields[0].GoName)
	assert.Equal(t, "E_name_", e.Fields[1].GoName)
}

func TestEnumEditionOpenSemantics(t *testing.T) {
	p := &Proto{Edition: edition2023}
	e := &Enum{
		Name:   "E",
		Proto:  p,
		Fields: []*EnumField{{Name: "ONE", Value: 1}},
	}
	p.Enums = []*Enum{e}

	assert.ErrorContains(t, p.verify(), "must have zero number for the first value")

	p.Options = Options{{Name: f_enum_type, Value: "CLOSED"}}
	assert.NoError(t, p.verify())

	e.Options = Options{{Name: f_enum_type, Value: "OPEN"}}
	assert.ErrorContains(t, p.verify(), "must have zero number for the first value")

	p.Options = Options{{Name: f_enum_type, Value: "OPEN"}}
	e.Options = Options{{Name: f_enum_type, Value: "CLOSED"}}
	e.Fields = []*EnumField{
		{Name: "E_FOO", Value: 1},
		{Name: "FOO", Value: 2},
	}
	assert.NoError(t, p.verify())
	assert.False(t, verifyOption(f_enum_type, "UNKNOWN"))
}

func TestLoaderEnumEditionAggregateSemantics(t *testing.T) {
	p := loadTestProto(t, `
edition = "2023";
option go_package = "example.com/enum";
option features = { enum_type: CLOSED };
enum FileClosed { FILE_ONE = 1; }
enum EnumOpen {
  option features = { enum_type: OPEN };
  ENUM_ZERO = 0;
}
`)

	assert.False(t, p.Enums[0].isOpen())
	assert.True(t, p.Enums[1].isOpen())

	p = loadTestProto(t, `
edition = "2023";
option go_package = "example.com/enum";
option features = { enum_type: OPEN };
enum EnumClosed {
  option features = { enum_type: CLOSED };
  ENUM_ONE = 1;
}
`)
	assert.False(t, p.Enums[0].isOpen())
}

func TestLoaderEnumMinimumValue(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "example.com/enum";
enum Decimal { DECIMAL_MIN = -2147483648; }
enum Hex { HEX_MIN = -0x80000000; }
`)

	assert.Equal(t, int32(-2147483648), p.Enums[0].Fields[0].Value)
	assert.Equal(t, int32(-2147483648), p.Enums[1].Fields[0].Value)
}

func TestLoader_EnumAllowAlias(t *testing.T) {
	f := loadTestProto(t, `
option go_package = "example.com/test";
enum Status {
  option allow_alias = true;
  UNKNOWN = 0;
  STARTED = 1;
  RUNNING = 1;
}
`)
	t.Log(f.String())

	ee := f.Enums
	assert.Equal(t, 1, len(ee))

	e := ee[0]
	assert.Equal(t, "Status", e.Name)
	assert.True(t, e.allowAlias())
	assert.Equal(t, 3, len(e.Fields))
	assert.Equal(t, "UNKNOWN", e.Fields[0].Name)
	assert.Equal(t, int32(0), e.Fields[0].Value)
	assert.Equal(t, "STARTED", e.Fields[1].Name)
	assert.Equal(t, int32(1), e.Fields[1].Value)
	assert.Equal(t, "RUNNING", e.Fields[2].Name)
	assert.Equal(t, int32(1), e.Fields[2].Value)
}

func TestLoader_Enum(t *testing.T) {
	f := loadTestProto(t, `
option go_package = "example.com/test";
enum myEnum0 {
  ENUM0 = 0;
  ENUM1 = 1;
}


message myMsg {
enum eEnum {
  ENUM0 = 0;
  ENUM2 = 2;
}
}


//prutalgen:no_enum_prefix
enum myEnum1 {
  A = 0;
  B = 1;
  C = 2;
}
`)
	t.Log(f.String())

	ee := f.Enums
	assert.Equal(t, 2, len(ee))

	e := ee[0]
	assert.Equal(t, "myEnum0", e.Name)
	assert.Equal(t, "MyEnum0", e.GoName)
	assert.Equal(t, 2, len(e.Fields))
	assert.Equal(t, "ENUM0", e.Fields[0].Name)
	assert.Equal(t, "MyEnum0_ENUM0", e.Fields[0].GoName)
	assert.Equal(t, int32(0), e.Fields[0].Value)
	assert.Equal(t, "ENUM1", e.Fields[1].Name)
	assert.Equal(t, "MyEnum0_ENUM1", e.Fields[1].GoName)
	assert.Equal(t, int32(1), e.Fields[1].Value)

	e = ee[1]
	assert.Equal(t, "myEnum1", e.Name)
	assert.Equal(t, "MyEnum1", e.GoName)
	assert.Equal(t, 3, len(e.Fields))
	assert.Equal(t, "A", e.Fields[0].Name)
	assert.Equal(t, "A", e.Fields[0].GoName)
	assert.Equal(t, int32(0), e.Fields[0].Value)
	assert.Equal(t, "B", e.Fields[1].Name)
	assert.Equal(t, "B", e.Fields[1].GoName)
	assert.Equal(t, int32(1), e.Fields[1].Value)
	assert.Equal(t, "C", e.Fields[2].Name)
	assert.Equal(t, "C", e.Fields[2].GoName)
	assert.Equal(t, int32(2), e.Fields[2].Value)

	m := f.Messages[0]
	ee = m.Enums
	assert.Equal(t, 1, len(ee))
	e = ee[0]
	assert.Equal(t, "eEnum", e.Name)
	assert.Equal(t, "MyMsgEEnum", e.GoName) // GoCamelCase("myMsg.eEnum") = "MyMsgEEnum"
	assert.Equal(t, 2, len(e.Fields))
	assert.Equal(t, "ENUM0", e.Fields[0].Name)
	assert.Equal(t, "MyMsg_ENUM0", e.Fields[0].GoName) // enum value uses message GoName prefix
	assert.Equal(t, int32(0), e.Fields[0].Value)
	assert.Equal(t, "ENUM2", e.Fields[1].Name)
	assert.Equal(t, "MyMsg_ENUM2", e.Fields[1].GoName)
	assert.Equal(t, int32(2), e.Fields[1].Value)
}
