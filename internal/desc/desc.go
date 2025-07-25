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
	"reflect"
	"sort"
	"strings"
	"sync"
	"unsafe"

	"github.com/cloudwego/prutal/internal/hack"
	"github.com/cloudwego/prutal/internal/wire"
)

var cache = newMapStructDesc()

var (
	parsemu      sync.Mutex
	errNotStruct = errors.New("input not struct")
)

func CacheGet(rv reflect.Value) *StructDesc {
	typ := hack.ReflectValueTypePtr(rv)
	return cache.Get(typ)
}

// GetOrParse ...
func GetOrParse(rv reflect.Value) (*StructDesc, error) {
	typ := hack.ReflectValueTypePtr(rv)
	ret := cache.Get(typ)
	if ret != nil {
		return ret, nil
	}
	rt := rv.Type()
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return nil, errNotStruct
	}

	parsemu.Lock()
	defer parsemu.Unlock()

	ret = cache.Get(typ)
	if ret != nil {
		return ret, nil
	}

	t, err := parseType(rt)
	if err != nil {
		return nil, err
	}
	s := t.S
	if s == nil {
		panic("t.S == nil")
	}
	if err := s.FinalizeFields(); err != nil {
		return nil, err
	}
	cache.Set(typ, s)
	return s, nil
}

const maxDirectFieldMapID = 1000

type StructDesc struct {
	Fields []*FieldDesc // sorted by ID

	// []byte or *[]byte
	// for *[]byte, see: https://go-review.googlesource.com/c/protobuf/+/244937
	HasUnknownFields     bool
	UnknownFieldsPointer bool
	UnknownFieldsOffset  uintptr

	// for GetField
	mFields0 []*FieldDesc         // direct ID map
	mFields1 map[int32]*FieldDesc // slow hash map

	Empty unsafe.Pointer // point to a zero struct for encoding list or map

	finalized bool // for FinalizeFields
}

func (p *StructDesc) GetField(v int32) *FieldDesc {
	if v < int32(len(p.mFields0)) {
		return p.mFields0[v]
	}
	return p.mFields1[v]
}

func (p *StructDesc) String() string {
	var buf strings.Builder
	buf.WriteString("StructDesc {\n")
	buf.WriteString("Fields:\n")
	for _, f := range p.Fields {
		fmt.Fprintf(&buf, " %v\n", f)
	}
	buf.WriteString("}\n")
	return buf.String()
}

func (p *StructDesc) FinalizeFields() error {
	if p.finalized {
		return nil
	}
	p.finalized = true
	for _, f := range p.Fields {
		if err := f.finalizeField(); err != nil {
			p.finalized = false
			return err
		}
		for _, t := range []*Type{f.T, f.T.K, f.T.V} {
			if t == nil {
				continue
			}
			if err := t.finalizeType(); err != nil {
				p.finalized = false
				return err
			}
		}
	}
	return nil
}

var wireTypes = []wire.Type{
	TypeVarint:   wire.TypeVarint,
	TypeZigZag32: wire.TypeVarint,
	TypeZigZag64: wire.TypeVarint,
	TypeFixed32:  wire.TypeFixed32,
	TypeFixed64:  wire.TypeFixed64,
	TypeBytes:    wire.TypeBytes,
}

type FieldDesc struct {
	ID       int32
	Name     string
	Offset   uintptr
	Tag      reflect.StructTag
	Required bool
	Repeated bool
	Packed   bool
	IsList   bool
	IsMap    bool

	TagType TagType
	WireTag uint64 //  wire.EncodeTag(f.ID, wireType)

	// only for oneof types
	// Kind==reflect.Pointer, coz we always use pointer for checking
	OneofType reflect.Type
	IfaceTab  uintptr // from OneofFieldIfaceTab

	// only for scalar types (including packed scalar types)
	AppendFunc wire.AppendFunc

	// only for list or map scalar types
	AppendRepeated wire.AppendRepeatedFunc

	// only for map type
	KeyType TagType
	ValType TagType

	KeyWireType wire.Type
	ValWireType wire.Type

	KeyAppendFunc wire.AppendFunc
	ValAppendFunc wire.AppendFunc

	// only for packed types, and some map types
	DecodeFunc func(b []byte, p unsafe.Pointer) error

	T *Type
}

func (f *FieldDesc) String() string {
	return fmt.Sprintf("ID:%d Name:%s Offset:%d Repeated:%v Packed:%v TagType:%v T:%v",
		f.ID, f.Name, f.Offset, f.Repeated, f.Packed, f.TagType, f.T)
}

func (f *FieldDesc) IsOneof() bool {
	return f.OneofType != nil
}

func (f *FieldDesc) parse(rt reflect.Type) (err error) {
	tag := f.Tag.Get("protobuf")
	if tag == "" {
		panic("not protobuf field")
	}
	if err = f.parseStructTag(tag); err != nil {
		return
	}
	f.T, err = parseType(rt)
	return
}

func (f *FieldDesc) finalizeField() (err error) {
	t := f.T
	f.IsList = f.Repeated && t.Kind != reflect.Map
	f.IsMap = t.Kind == reflect.Map

	if f.IsMap {
		f.KeyType, err = parseKVTag(f.Tag.Get("protobuf_key"))
		if err != nil {
			return
		}
		f.ValType, err = parseKVTag(f.Tag.Get("protobuf_val"))
		if err != nil {
			return
		}
		f.KeyWireType = wireTypes[f.KeyType]
		f.ValWireType = wireTypes[f.ValType]
	}
	if err = f.checkTypeMatch(); err != nil {
		return
	}
	f.AppendFunc = getAppendFunc(f.TagType, t.RealKind(), f.Packed)
	if f.T.Kind == reflect.Map {
		f.KeyAppendFunc = getAppendFunc(f.KeyType, t.K.RealKind(), false)
		f.ValAppendFunc = getAppendFunc(f.ValType, t.V.RealKind(), false)
	}
	if f.IsList {
		f.AppendRepeated = getAppendListFunc(f)
	}
	if f.IsMap {
		f.AppendRepeated = getAppendMapFunc(f)
	}
	f.DecodeFunc = getDecodeFunc(f)
	return
}

func (f *FieldDesc) checkTypeMatch() error {
	t := f.T
	if f.Packed {
		if !f.Repeated {
			return errors.New("packed field is not repeated field")
		}
		if !t.IsSlice {
			return errors.New("packed field is not slice")
		}
		switch f.TagType {
		case TypeVarint, TypeZigZag32, TypeZigZag64, TypeFixed32, TypeFixed64:
		default:
			return errors.New("packed field only for scalar types except string or bytes")
		}
	}
	if f.Repeated {
		if !t.IsSlice && t.Kind != reflect.Map {
			return fmt.Errorf("repeated field is not slice or map")
		}
	}

	if err := IsFieldTypeMatchReflectKind(f.TagType, t.RealKind()); err != nil {
		return err
	}
	if t.Kind == reflect.Map {
		if !f.Repeated {
			return errors.New("must be repeated field for map")
		}
		if err := IsFieldKeyTypeMatchReflectKind(f.KeyType, t.K.RealKind()); err != nil {
			return err
		}
		if err := IsFieldTypeMatchReflectKind(f.ValType, t.V.RealKind()); err != nil {
			return err
		}
	}
	return nil
}

var cachedStructs = map[reflect.Type]*StructDesc{}

func parseStruct(rt reflect.Type) (s *StructDesc, err error) {
	if s = cachedStructs[rt]; s != nil {
		return s, nil
	}
	s = &StructDesc{Empty: reflect.New(rt).UnsafePointer()}
	cachedStructs[rt] = s // fix cyclic refs
	defer func() {
		if err != nil {
			delete(cachedStructs, rt)
		}
	}()

	var oneofs []reflect.StructField
	var fields []*FieldDesc
	for i, n := 0, rt.NumField(); i < n; i++ {
		sf := rt.Field(i)
		tag := sf.Tag.Get("protobuf")
		if tag == "" {
			if sf.Tag.Get("protobuf_oneof") != "" {
				oneofs = append(oneofs, sf)
			}
			continue
		}
		f := &FieldDesc{Name: sf.Name, Offset: sf.Offset, Tag: sf.Tag}
		if err = f.parse(sf.Type); err != nil {
			return nil, fmt.Errorf("parse field %q err: %w", sf.Name, err)
		}
		fields = append(fields, f)
	}

	if len(oneofs) > 0 {
		for _, v := range searchOneofWrappers(rt) {
			rt := reflect.TypeOf(v)
			for _, o := range oneofs {
				if !rt.Implements(o.Type) {
					continue
				}
				// Pointer -> Struct
				rt = rt.Elem()
				if rt.NumField() != 1 { // The struct must contains exactly one field
					return nil, fmt.Errorf("parse field %q oneof %q err: field number != 1", o.Name, rt.String())
				}
				field := rt.Field(0)
				f := &FieldDesc{Name: o.Name, Offset: o.Offset, Tag: field.Tag, OneofType: reflect.PointerTo(rt)}
				f.IfaceTab = hack.IfaceTab(o.Type, rt)
				if err = f.parse(field.Type); err != nil {
					return nil, fmt.Errorf("parse field %q oneof %q err: %w", o.Name, rt.String(), err)
				}
				fields = append(fields, f)
			}
		}
	}

	// sort by ID
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].ID < fields[j].ID
	})
	s.Fields = fields

	k := 0 // for s.mFields1
	maxn := 0
	for i, f := range s.Fields {
		if f.ID > maxDirectFieldMapID {
			k = len(s.Fields) - 1 - i
			break
		}
		maxn = int(f.ID)
	}
	s.mFields0 = make([]*FieldDesc, maxn+1)
	s.mFields1 = make(map[int32]*FieldDesc, k)
	for i, f := range s.Fields {
		if f.ID < maxDirectFieldMapID {
			s.mFields0[int(f.ID)] = f
		} else {
			s.mFields1[f.ID] = f
		}

		if i > 0 && f.ID == s.Fields[i-1].ID {
			return nil, fmt.Errorf("duplicated field number: %d for field %q and %q",
				f.ID, f.Name, s.Fields[i-1].Name)
		}
	}

	// unknownFields: latest version
	// XXX_unrecognized: old version protobuf
	for _, name := range []string{"unknownFields", "XXX_unrecognized"} {
		f, ok := rt.FieldByName(name)
		if !ok {
			continue
		}
		ft := f.Type
		if ft != bytesType && // not []byte nor *[]byte?
			ft.Kind() == reflect.Pointer && ft.Elem() != bytesType {
			continue
		}
		s.HasUnknownFields = true
		s.UnknownFieldsPointer = ft.Kind() == reflect.Pointer
		s.UnknownFieldsOffset = f.Offset
		break
	}

	return s, nil
}
