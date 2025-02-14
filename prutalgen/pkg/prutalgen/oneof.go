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
	"github.com/cloudwego/prutal/prutalgen/internal/parser"
	"github.com/cloudwego/prutal/prutalgen/internal/protobuf/strs"
)

type Oneof struct {
	Name string

	Options Options

	Msg    *Message
	Fields []*Field
}

func (o *Oneof) FieldName() string {
	return strs.GoCamelCase(o.Name)
}

func (o *Oneof) FieldType() string {
	return "is" + o.Msg.GoName + "_" + strs.GoCamelCase(o.Name)
}

func (x *protoLoader) EnterOneof(c *parser.OneofContext) {
	o := &Oneof{}
	o.Name = c.OneofName().GetText()
	o.Msg = x.currentMsg()
	o.Msg.Oneofs = append(o.Msg.Oneofs, o)
}
