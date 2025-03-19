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

	s, err := parseStruct(rt)
	if err != nil {
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

	T *Type
}

func (f *FieldDesc) String() string {
	return fmt.Sprintf("ID:%d Offset:%d Repeated:%v Packed:%v TagType:%v T:%v",
		f.ID, f.Offset, f.Repeated, f.Packed, f.TagType, f.T)
}

func (f *FieldDesc) IsOneof() bool {
	return f.OneofType != nil
}

func (f *FieldDesc) parse(sf reflect.StructField) (err error) {
	tag := sf.Tag.Get("protobuf")
	if tag == "" {
		panic("not protobuf field")
	}
	if err = f.parseStructTag(tag); err != nil {
		return
	}
	f.T, err = parseType(sf.Type)
	if err != nil {
		return
	}
	f.IsList = f.Repeated && f.T.Kind != reflect.Map
	f.IsMap = f.T.Kind == reflect.Map

	if f.IsMap {
		f.KeyType, err = parseKVTag(sf.Tag.Get("protobuf_key"))
		if err != nil {
			return
		}
		f.ValType, err = parseKVTag(sf.Tag.Get("protobuf_val"))
		if err != nil {
			return
		}
		f.KeyWireType = wireTypes[f.KeyType]
		f.ValWireType = wireTypes[f.ValType]
	}
	if err = f.checkTypeMatch(); err != nil {
		return
	}
	t := f.T
	f.AppendFunc = getAppendFunc(f.TagType, t.RealKind(), f.Packed)
	if f.T.Kind == reflect.Map {
		f.KeyAppendFunc = getAppendFunc(f.KeyType, t.K.RealKind(), false)
		f.ValAppendFunc = getAppendFunc(f.ValType, t.V.RealKind(), false)
	}
	if f.IsList {
		f.AppendRepeated = getAppendListFunc(f.TagType, t.RealKind())
	}
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
}

func (t *Type) RealKind() reflect.Kind {
	if t.IsPointer || t.IsSlice {
		return t.V.RealKind()
	}
	return t.Kind
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

var cachedTypes = map[reflect.Type]*Type{}

func parseType(rt reflect.Type) (ret *Type, err error) {
	if t := cachedTypes[rt]; t != nil {
		return t, nil
	}
	ret = &Type{}

	cachedTypes[rt] = ret // fix cyclic refs

	ret.T = rt
	ret.Kind = rt.Kind()
	ret.Size = rt.Size()
	ret.Align = rt.Align()

	if rt == bytesType { // special case
		ret.Kind = KindBytes
	}

	switch ret.Kind {
	case reflect.Ptr, reflect.Slice, KindBytes, reflect.String,
		reflect.Map, reflect.Struct:
		// for these types, we can't use span mem allocator
		// coz then may contain pointer
		ret.MallocAbiType = hack.ReflectTypePtr(ret.T)
	}

	ret.IsPointer = ret.Kind == reflect.Pointer
	ret.IsSlice = ret.Kind == reflect.Slice

	ret.SliceLike = ret.Kind == reflect.Slice ||
		ret.Kind == KindBytes ||
		ret.Kind == reflect.String

	switch rt.Kind() {
	case reflect.Map:
		ret.K, err = parseType(rt.Key())
		if err != nil {
			break
		}
		ret.V, err = parseType(rt.Elem())
		if err != nil {
			break
		}
		ret.MapTmpVarsPool.New = func() interface{} {
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
		ret.S, err = parseStruct(rt)
	case reflect.Slice:
		ret.V, err = parseType(rt.Elem())
	case reflect.Pointer:
		ret.V, err = parseType(rt.Elem())
		if err == nil && ret.V.IsPointer {
			err = errors.New("multilevel pointer")
		}
	default:
	}
	if err != nil {
		delete(cachedTypes, rt)
		return nil, err
	}
	return ret, nil
}

var cachedStructs = map[reflect.Type]*StructDesc{}

func parseStruct(rt reflect.Type) (s *StructDesc, err error) {
	if s = cachedStructs[rt]; s != nil {
		return s, nil
	}

	s = &StructDesc{}
	cachedStructs[rt] = s // fix cyclic refs
	defer func() {
		if err != nil {
			delete(cachedStructs, rt)
		}
	}()

	var oneofs []reflect.StructField
	var fields []FieldDesc
	for i, n := 0, rt.NumField(); i < n; i++ {
		sf := rt.Field(i)
		tag := sf.Tag.Get("protobuf")
		if tag == "" {
			if sf.Tag.Get("protobuf_oneof") != "" {
				oneofs = append(oneofs, sf)
			}
			continue
		}
		f := FieldDesc{Name: sf.Name, Offset: sf.Offset}
		if err = f.parse(sf); err != nil {
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
				f := FieldDesc{Name: o.Name, Offset: o.Offset, OneofType: reflect.PointerTo(rt)}
				f.IfaceTab = hack.IfaceTab(o.Type, rt)
				if err = f.parse(rt.Field(0)); err != nil {
					return nil, fmt.Errorf("parse field %q oneof %q err: %w", o.Name, rt.String(), err)
				}
				fields = append(fields, f)
			}
		}
	}

	// reduce in-use objects
	ff := make([]FieldDesc, len(fields))
	copy(ff, fields)
	s.Fields = make([]*FieldDesc, len(ff))
	for i := range ff {
		s.Fields[i] = &ff[i]
	}

	// sort by ID
	sort.Slice(s.Fields, func(i, j int) bool {
		return s.Fields[i].ID < s.Fields[j].ID
	})

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

func getAppendFunc(t TagType, k reflect.Kind, packed bool) func(b []byte, p unsafe.Pointer) []byte {
	if packed {
		switch t {
		case TypeVarint:
			switch k {
			case reflect.Int32, reflect.Uint32:
				return wire.UnsafeAppendPackedVarintU32
			case reflect.Int64, reflect.Uint64:
				return wire.UnsafeAppendPackedVarintU64
			case reflect.Bool:
				return wire.UnsafeAppendPackedBool
			}
		case TypeZigZag32:
			return wire.UnsafeAppendPackedZigZag32
		case TypeZigZag64:
			return wire.UnsafeAppendPackedZigZag64
		case TypeFixed32:
			return wire.UnsafeAppendPackedFixed32
		case TypeFixed64:
			return wire.UnsafeAppendPackedFixed64
		case TypeBytes:
			panic("packed on bytes field")
		default:
			panic(fmt.Sprintf("unknown tag type: %q", t))
		}
	}
	switch t {
	case TypeVarint:
		switch k {
		case reflect.Int32, reflect.Uint32:
			return wire.UnsafeAppendVarintU32
		case reflect.Int64, reflect.Uint64:
			return wire.UnsafeAppendVarintU64
		case reflect.Bool:
			return wire.UnsafeAppendBool
		}
	case TypeZigZag32:
		return wire.UnsafeAppendZigZag32
	case TypeZigZag64:
		return wire.UnsafeAppendZigZag64
	case TypeFixed32:
		return wire.UnsafeAppendFixed32
	case TypeFixed64:
		return wire.UnsafeAppendFixed64
	case TypeBytes:
		switch k {
		case reflect.String: // string
			return wire.UnsafeAppendString
		case KindBytes: // []byte
			return wire.UnsafeAppendBytes
		}
	default:
		panic(fmt.Sprintf("unknown tag type: %q", t))
	}
	return nil
}

func getAppendListFunc(t TagType, k reflect.Kind) func(b []byte, id int32, p unsafe.Pointer) []byte {
	switch t {
	case TypeVarint:
		switch k {
		case reflect.Int64, reflect.Uint64:
			return wire.UnsafeAppendVarintU64List
		case reflect.Int32, reflect.Uint32:
			return wire.UnsafeAppendVarintU32List
		case reflect.Bool:
			return wire.UnsafeAppendBoolList
		}
	case TypeZigZag32:
		return wire.UnsafeAppendZigZag32List
	case TypeZigZag64:
		return wire.UnsafeAppendZigZag64List
	case TypeFixed32:
		return wire.UnsafeAppendFixed32List
	case TypeFixed64:
		return wire.UnsafeAppendFixed64List
	case TypeBytes:
		switch k {
		case reflect.String: // string
			return wire.UnsafeAppendStringList
		case KindBytes: // []byte
			return wire.UnsafeAppendBytesList
		}
	}
	return nil
}
