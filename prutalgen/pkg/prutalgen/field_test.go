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

func TestField(t *testing.T) {
	f := &Field{}

	// OptionGenNoPointer
	f.Options = nil
	assert.False(t, f.OptionGenNoPointer())

	f.Options = Options{{Name: optionNoPointer, Value: "true"}}
	assert.True(t, f.OptionGenNoPointer())

	f.Options = Options{{Name: gogoproto_nullable, Value: "false"}}
	assert.True(t, f.OptionGenNoPointer())

	reset := func(f *Field) {
		f.Options = nil
		f.Key, f.Type = nil, nil
		f.Repeated, f.Optional, f.Required = false, false, false
		f.Msg = &Message{Proto: &Proto{}}
	}

	// GoTypeName:map
	reset(f)
	f.Key = &Type{Name: "float"}
	f.Type = &Type{Name: "double"}
	assert.Equal(t, "map[float32]float64", f.GoTypeName())

	// GoTypeName:slice
	reset(f)
	f.Type = &Type{Name: "double"}
	assert.Equal(t, "float64", f.GoTypeName())
	f.Repeated = true
	assert.Equal(t, "[]float64", f.GoTypeName())

	// GoTypeName:optional or proto2
	reset(f)
	f.Type = &Type{Name: "double"}
	assert.Equal(t, "float64", f.GoTypeName())
	f.Optional = true
	assert.Equal(t, "*float64", f.GoTypeName())
	f.Optional = false
	assert.Equal(t, "float64", f.GoTypeName())
	f.Msg = &Message{Proto: &Proto{Edition: editionProto2}}
	assert.Equal(t, "*float64", f.GoTypeName())

	// GoTypeName: bytes
	reset(f)
	f.Optional = true
	f.Type = &Type{Name: "bytes"} // no pointer even f.Optional=true
	assert.Equal(t, "[]byte", f.GoTypeName())

	// GoTypeName:message
	reset(f)
	nopointerOpt := Options{{Name: optionNoPointer, Value: "true"}}
	f.Type = &Type{Name: "Msg", typ: &Message{GoName: "Msg"}}
	assert.Equal(t, "*Msg", f.GoTypeName())
	f.Options = nopointerOpt
	assert.Equal(t, "Msg", f.GoTypeName())
	f.Options = nil
	f.Repeated = true
	assert.Equal(t, "[]*Msg", f.GoTypeName())
	f.Options = nopointerOpt
	assert.Equal(t, "[]Msg", f.GoTypeName())

	// GoZero: map
	f.Key = &Type{Name: "float"}
	f.Type = &Type{Name: "double"}
	assert.Equal(t, "nil", f.GoZero())
	f.Key = nil

	// GoZero: slice
	f.Repeated = true
	f.Type = &Type{Name: "double"}
	assert.Equal(t, "nil", f.GoZero())
	f.Repeated = false

	// GoZero: scalars
	f.Type = &Type{Name: "double"}
	assert.Equal(t, "0", f.GoZero())
	f.Type = &Type{Name: "bool"}
	assert.Equal(t, "false", f.GoZero())
	f.Type = &Type{Name: "string"}
	assert.Equal(t, `""`, f.GoZero())
	f.Type = &Type{Name: "bytes"}
	assert.Equal(t, "nil", f.GoZero())

	// GoZero: enum
	e := &Enum{
		Name:   "my_enum",
		GoName: "MyEnum",
		Fields: []*EnumField{{Name: "Zero", GoName: "MyEnum_Zero"}},
	}
	f.Type = &Type{Name: "my_enum",
		typ: e,
		p:   &Proto{GoImport: "base"},
	}
	assert.Equal(t, "base.MyEnum_Zero", f.GoZero())
	f.Type.p = nil
	assert.Equal(t, "MyEnum_Zero", f.GoZero())
	e.Fields = nil
	assert.Equal(t, "0", f.GoZero())

	// GOZero: struct
	f.Options = nil
	f.Type = &Type{Name: "my_struct",
		typ: &Message{GoName: "MyStruct"},
	}
	assert.Equal(t, "nil", f.GoZero())
	f.Options = nopointerOpt
	assert.Equal(t, "MyStruct{}", f.GoZero())

}

func TestLoader_Field(t *testing.T) {
	p := loadTestProto(t, `
option go_package = "testfield";
message M {
required string email = 1;
repeated string names = 2;
optional string tel = 3;
map<string,string> extra = 4;

message N {
}
N m = 5;

enum E {
	V = 0;
}

E e = 6 [(prutal.test) = true];
}`,
	)

	m := p.Messages[0]
	assert.Equal(t, 6, len(m.Fields))

	f1 := m.Fields[0]
	t.Log(f1.String())
	assert.Equal(t, int32(1), f1.FieldNumber)
	assert.True(t, f1.Required)

	f2 := m.Fields[1]
	t.Log(f2.String())
	assert.Equal(t, int32(2), f2.FieldNumber)
	assert.True(t, f2.Repeated)

	f3 := m.Fields[2]
	t.Log(f3.String())
	assert.Equal(t, int32(3), f3.FieldNumber)
	assert.True(t, f3.Optional)

	f4 := m.Fields[3]
	t.Log(f4.String())
	assert.Equal(t, int32(4), f4.FieldNumber)
	assert.True(t, f4.IsMap())

	f5 := m.Fields[4]
	t.Log(f5.String())
	assert.Equal(t, int32(5), f5.FieldNumber)
	assert.True(t, f5.IsMessage())

	f6 := m.Fields[5]
	t.Log(f6.String())
	assert.Equal(t, int32(6), f6.FieldNumber)
	assert.True(t, f6.IsEnum())
	o, _ := f6.Options.Get("(prutal.test)")
	assert.Equal(t, "true", o)
}
