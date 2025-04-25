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
	assert.Equal(t, "MyMsg_EEnum", e.GoName)
	assert.Equal(t, 2, len(e.Fields))
	assert.Equal(t, "ENUM0", e.Fields[0].Name)
	assert.Equal(t, "MyMsg_ENUM0", e.Fields[0].GoName)
	assert.Equal(t, int32(0), e.Fields[0].Value)
	assert.Equal(t, "ENUM2", e.Fields[1].Name)
	assert.Equal(t, "MyMsg_ENUM2", e.Fields[1].GoName)
	assert.Equal(t, int32(2), e.Fields[1].Value)
}
