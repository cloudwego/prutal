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
	assert.False(t, e.OptionGenNoPrefix())
	assert.True(t, e.OptionGenNameMapping())

	oo := Options{
		{Name: optionNoEnumPrefix, Value: ""},
		{Name: gogoproto_enum_prefix, Value: "false"},     // lower higher priority
		{Name: gogoproto_enum_prefix_all, Value: "false"}, // lower higher priority
	}
	e.Options = oo
	oo[0].Value = "true"
	assert.True(t, e.OptionGenNoPrefix())
	oo[0].Value = "false"
	assert.False(t, e.OptionGenNoPrefix())
	e.Options = oo[1:]
	assert.True(t, e.OptionGenNoPrefix())

	e.Options = nil
	e.Proto.Options = oo
	oo[0].Value = "true"
	assert.True(t, e.OptionGenNoPrefix())
	oo[0].Value = "false"
	assert.False(t, e.OptionGenNoPrefix())
	e.Proto.Options = oo[1:]
	assert.True(t, e.OptionGenNoPrefix())

	oo = Options{
		{Name: optionEnumNameMapping, Value: ""},
	}

	e.Options = oo
	oo[0].Value = "true"
	assert.True(t, e.OptionGenNameMapping())
	oo[0].Value = "false"
	assert.False(t, e.OptionGenNameMapping())
	e.Options = nil
	e.Proto.Options = oo
	oo[0].Value = "true"
	assert.True(t, e.OptionGenNameMapping())
	oo[0].Value = "false"
	assert.False(t, e.OptionGenNameMapping())
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
}

func TestLoader_Enum(t *testing.T) {
	f := loadTestProto(t, `
option go_package = "test";
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


enum myEnum1 {
	option (prutal.gen_no_enum_prefix) = true;
  A = 0;
  B = 1;
  C = 2 [(prutal.gen_no_enum_prefix) = false];
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
	assert.Equal(t, "A", e.Fields[0].Name) // prutal.gen_no_enum_prefix=true
	assert.Equal(t, "A", e.Fields[0].GoName)
	assert.Equal(t, int32(0), e.Fields[0].Value)
	assert.Equal(t, "B", e.Fields[1].Name) // prutal.gen_no_enum_prefix=true
	assert.Equal(t, "B", e.Fields[1].GoName)
	assert.Equal(t, int32(1), e.Fields[1].Value)
	assert.Equal(t, "C", e.Fields[2].Name) // prutal.gen_no_enum_prefix=true
	assert.Equal(t, "MyEnum1_C", e.Fields[2].GoName)
	assert.Equal(t, int32(2), e.Fields[2].Value)

	m := f.Messages[0]
	ee = m.Enums
	assert.Equal(t, 1, len(ee))
	e = ee[0]
	assert.Equal(t, "eEnum", e.Name)
	assert.Equal(t, "MyMsg_EEnum", e.GoName)
	assert.Equal(t, 2, len(e.Fields))
	assert.Equal(t, "ENUM0", e.Fields[0].Name)
	assert.Equal(t, "MyMsg_ENUM0", e.Fields[0].GoName)
	assert.Equal(t, int32(0), e.Fields[0].Value)
	assert.Equal(t, "ENUM2", e.Fields[1].Name)
	assert.Equal(t, "MyMsg_ENUM2", e.Fields[1].GoName)
	assert.Equal(t, int32(2), e.Fields[1].Value)
}
