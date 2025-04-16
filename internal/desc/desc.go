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

var bytesType = reflect.TypeOf([]byte{})

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
	Repeated bool
	Packed   bool
	TagType  TagType
	WireType wire.Type

	IsList bool
	IsMap  bool

	// only for oneof types
	// Kind==reflect.Pointer, coz we always use pointer for checking
	OneofType reflect.Type
	IfaceTab  uintptr // from OneofFieldIfaceTab

	// only for scalar types (including packed scalar types)
	AppendFunc func(b []byte, p unsafe.Pointer) []byte

	// only for list or map scalar types
	AppendRepeated func(b []byte, id int32, p unsafe.Pointer) []byte

	// only for map type
	KeyType TagType
	ValType TagType

	KeyWireType wire.Type
	ValWireType wire.Type

	KeyAppendFunc func(b []byte, p unsafe.Pointer) []byte
	ValAppendFunc func(b []byte, p unsafe.Pointer) []byte

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
		f.AppendRepeated = getAppendListFunc(f.TagType, t.RealKind())
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

const KindBytes reflect.Kind = 5000 // for []byte

type Type struct {
	T reflect.Type

	// true if t.Kind == reflect.Pointer
	IsPointer bool

	// true if t.Kind == reflect.Slice,
	// false for []byte which is considered to be scalar type
	IsSlice bool

	SliceLike bool // reflect.Slice, reflect.String, KindBytes

	// cache reflect.Type returns for performance
	Kind  reflect.Kind
	Size  uintptr
	Align int

	// for decoder
	MallocAbiType uintptr

	K *Type       // for map
	V *Type       // for pointer, slice or map
	S *StructDesc // struct

	// for map only
	MapTmpVarsPool sync.Pool // for decoder tmp vars

	finalized bool
}

func (t *Type) RealKind() reflect.Kind {
	if t.IsPointer || t.IsSlice {
		return t.V.RealKind()
	}
	return t.Kind
}

func (t *Type) finalizeType() error {
	if t.finalized {
		return nil
	}
	t.finalized = true
	if t.S != nil {
		if err := t.S.FinalizeFields(); err != nil {
			t.finalized = false
			return err
		}
	}
	if t.V != nil {
		if err := t.V.finalizeType(); err != nil {
			t.finalized = false
			return err
		}
	}
	return nil
}

// TmpMapVars contains key and value tmp vars used for updating associated map for a type
type TmpMapVars struct {
	m reflect.Value

	k  reflect.Value  // t.K.T
	kp unsafe.Pointer // *t.K.T

	v  reflect.Value  // t.V.T
	vp unsafe.Pointer // *t.V.T

	// zero value of v,
	// only used when non-pointer struct as map val
	// we need to zero the tmp var before using it
	z reflect.Value
}

func (p *TmpMapVars) MapWithPtr(x unsafe.Pointer) reflect.Value {
	return hack.ReflectValueWithPtr(p.m, x)
}

func (p *TmpMapVars) KeyPointer() unsafe.Pointer { return p.kp }
func (p *TmpMapVars) ValPointer() unsafe.Pointer { return p.vp }
func (p *TmpMapVars) Update(m reflect.Value)     { m.SetMapIndex(p.k, p.v) }
func (p *TmpMapVars) Reset() {
	if p.z.IsValid() {
		p.v.Set(p.z)
	}
}

func (t *Type) String() string {
	switch t.Kind {
	case reflect.Struct:
		return fmt.Sprintf("%+v", t.S)
	default:
		return fmt.Sprintf("%+v", t.T)
	}
}

var (
	cachedTypes = map[reflect.Type]*Type{}
)

func noopFinalizeField(_ *Type) error { return nil }

func parseType(rt reflect.Type) (t *Type, err error) {
	if t = cachedTypes[rt]; t != nil {
		return t, nil
	}

	t = &Type{}
	cachedTypes[rt] = t // reuse result and also fix cyclic refs

	t.T = rt
	t.Kind = rt.Kind()
	t.Size = rt.Size()
	t.Align = rt.Align()

	if rt == bytesType { // special case
		t.Kind = KindBytes
	}

	switch t.Kind {
	case reflect.Ptr, reflect.Slice, KindBytes, reflect.String,
		reflect.Map, reflect.Struct:
		// for these types, we can't use span mem allocator
		// coz then may contain pointer
		t.MallocAbiType = hack.ReflectTypePtr(t.T)
	}

	t.IsPointer = t.Kind == reflect.Pointer
	t.IsSlice = t.Kind == reflect.Slice

	t.SliceLike = t.Kind == reflect.Slice ||
		t.Kind == KindBytes ||
		t.Kind == reflect.String

	switch rt.Kind() {
	case reflect.Map:
		t.K, err = parseType(rt.Key())
		if err != nil {
			break
		}
		t.V, err = parseType(rt.Elem())
		if err != nil {
			break
		}
		t.MapTmpVarsPool.New = func() interface{} {
			m := &TmpMapVars{}
			m.m = reflect.New(rt).Elem()
			m.k = reflect.New(rt.Key())
			m.kp = m.k.UnsafePointer()
			m.k = m.k.Elem()
			m.v = reflect.New(rt.Elem())
			m.vp = m.v.UnsafePointer()
			m.v = m.v.Elem()
			if rt.Elem().Kind() == reflect.Struct {
				m.z = reflect.Zero(rt.Elem())
			}
			return m
		}
	case reflect.Struct:
		t.S, err = parseStruct(rt)
	case reflect.Slice:
		t.V, err = parseType(rt.Elem())
	case reflect.Pointer:
		t.V, err = parseType(rt.Elem())
		if err == nil && t.V.IsPointer {
			err = errors.New("multilevel pointer")
		}
	default:
	}
	if err != nil {
		delete(cachedTypes, rt)
		return nil, err
	}
	return t, nil
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
				if rt.NumField() != 1 { // The struct must contains extractly one field
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
