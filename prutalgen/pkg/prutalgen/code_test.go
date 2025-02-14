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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestGenCode(t *testing.T) {
	g := NewGoCodeGen()
	g.Getter = true
	p := &Proto{ProtoFile: filepath.Join(t.TempDir(), "test.proto"), l: testLogger{t}}
	p.GoPackage = "test"

	p.Enums = []*Enum{{
		GoName: "Enum1",
		Fields: []*EnumField{
			{GoName: "Enum1_Zero", Value: 0},
			{GoName: "Enum1_N", Value: 10001},
		},
		Proto: p,
	}}

	m := &Message{
		GoName: "Message1",
		Fields: []*Field{
			{GoName: "Field1", Type: &Type{Name: "string"}, FieldNumber: 20001},
			{GoName: "Field2", Type: &Type{Name: "uint64"}, Repeated: true, FieldNumber: 20002},
			{GoName: "Field3", Type: &Type{Name: "string"}, Optional: true, FieldNumber: 20003},
			{GoName: "Field4", Key: &Type{Name: "string"}, Type: &Type{Name: "string"}, FieldNumber: 20004},
		},
		Proto: p,
	}
	p.Messages = []*Message{m}

	m.Oneofs = []*Oneof{{
		Name: "Oneof1",
		Fields: []*Field{
			{GoName: "OneofA", Type: &Type{Name: "string"}, FieldNumber: 20101},
			{GoName: "OneofB", Type: &Type{Name: "string"}, FieldNumber: 20102},
		},
		Msg: m,
	}}
	oneof := m.Oneofs[0]
	oneof.Fields[0].Oneof = oneof
	oneof.Fields[1].Oneof = oneof
	m.Fields = append(m.Fields, oneof.Fields...)
	for _, f := range m.Fields {
		f.Msg = m
	}

	m.Enums = []*Enum{{
		GoName: "NestedEnum",
		Fields: []*EnumField{
			{GoName: "NestedEnum_Zero", Value: 0},
			{GoName: "NestedEnum_N", Value: 30001},
		},
		Proto: p,
	}}

	m.Messages = []*Message{{
		GoName: "NestedMsg",
		Fields: []*Field{
			{GoName: "NField1", Type: &Type{Name: "string"}, FieldNumber: 40001},
		},
		Proto: p,
	}}
	for _, f := range m.Messages[0].Fields {
		f.Msg = m.Messages[0]
	}

	_ = g.Gen(p, GenBySourceRelative, "")

	outfn := filepath.Join(filepath.Dir(p.ProtoFile), "test.pb.go")
	b, err := os.ReadFile(outfn)
	assert.NoError(t, err)

	src := string(b)
	lines := strings.Split(src, "\n")
	assertLine := func(s string) {
		t.Helper()
		for _, l := range lines {
			if strings.Contains(l, s) {
				return
			}
		}
		t.Fatal("not match", s)
	}
	t.Log(src)
	assertLine("Enum1 = 0")
	assertLine("Enum1 = 10001")
	assertLine("NestedEnum = 0")
	assertLine("NestedEnum = 30001")
	assertLine("Field1 string")
	assertLine("Field2 []uint64")
	assertLine("Field3 *string")
	assertLine("Field4 map[string]string")
	assertLine("OneofA string")
	assertLine("OneofB string")
	assertLine("NField1 string")
}
