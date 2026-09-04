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

var cache = &mapStructDesc{}

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

	parsed := false
	startParseScope()
	defer finishParseScope(&parsed)

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
	parsed = true
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
			return fmt.Errorf("finalize field %q err: %w", f.Name, err)
		}
		for _, t := range []*Type{f.T, f.T.K, f.T.V} {
			if t == nil {
				continue
			}
			if err := t.finalizeType(); err != nil {
				p.finalized = false
				return fmt.Errorf("finalize field %q err: %w", f.Name, err)
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

// ZeroKind tells the encoder and sizer how to test a field for its zero
// value, so that the hot loop needs no type dispatch of its own. It is
// resolved once, when the field is finalized.
type ZeroKind uint8

const (
	// ZeroKindNone marks a field that is never skipped: a set oneof member is
	// serialized even when it holds a zero value, otherwise the chosen case
	// would be lost on round-trip. Types without a cheap test end up here too.
	ZeroKindNone ZeroKind = iota
	ZeroKindU64           // int64, uint64, float64, or a pointer or map on 64-bit
	ZeroKindU32           // int32, uint32, float32, or a pointer or map on 32-bit
	ZeroKindU8            // bool
	ZeroKindLen           // string, []byte or slice: zero when the header Len is 0
)

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
	ZeroKind ZeroKind

	TagType TagType
	WireTag uint64 //  wire.EncodeTag(f.ID, wireType)

	// only for oneof types
	// Kind==reflect.Pointer, coz we always use pointer for checking
	OneofType reflect.Type
	IfaceTab  uintptr // from OneofFieldIfaceTab

	WireTagSize int // pre-computed size of WireTag varint

	// only for scalar types (including packed scalar types)
	AppendFunc wire.AppendFunc
	SizeFunc   wire.SizeFunc

	// only for list or map scalar types
	AppendRepeated wire.AppendRepeatedFunc

	// only for map type
	KeyType TagType
	ValType TagType

	KeyWireType wire.Type
	ValWireType wire.Type

	KeyAppendFunc wire.AppendFunc
	ValAppendFunc wire.AppendFunc

	KeySizeFunc wire.SizeFunc
	ValSizeFunc wire.SizeFunc

	SizeMapFunc wire.SizeMapFunc

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
		// only reachable via oneof wrapper fields;
		// parseStruct skips regular fields without the tag
		return errors.New("missing protobuf tag")
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
		f.KeyType, err = parseKVTag(f.Tag.Get("protobuf_key"), 1, false)
		if err != nil {
			return
		}
		f.ValType, err = parseKVTag(f.Tag.Get("protobuf_val"), 2, true)
		if err != nil {
			return
		}
		f.KeyWireType = wireTypes[f.KeyType]
		f.ValWireType = wireTypes[f.ValType]
	}
	if err = f.checkTypeMatch(); err != nil {
		return
	}
	f.ZeroKind = zeroKindOf(f)
	f.WireTagSize = wire.SizeVarint(f.WireTag)
	f.AppendFunc = getAppendFunc(f.TagType, t.RealKind(), f.Packed)
	f.SizeFunc = getSizeFunc(f.TagType, t.RealKind())
	if f.T.Kind == reflect.Map {
		f.KeyAppendFunc = getAppendFunc(f.KeyType, t.K.RealKind(), false)
		f.ValAppendFunc = getAppendFunc(f.ValType, t.V.RealKind(), false)
		f.KeySizeFunc = getSizeFunc(f.KeyType, t.K.RealKind())
		f.ValSizeFunc = getSizeFunc(f.ValType, t.V.RealKind())
		f.SizeMapFunc = getSizeMapFunc(f)
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

// zeroKindOf resolves how the encoder and sizer test f for its zero value.
func zeroKindOf(f *FieldDesc) ZeroKind {
	t := f.T
	switch {
	case f.IsOneof():
		return ZeroKindNone
	case t.SliceLike:
		// checked before the sizes: on 386 a string header is 8 bytes and
		// "" has a non-nil data pointer, so a size-based test would treat
		// an empty string as non-zero
		return ZeroKindLen
	case t.Size == 8:
		return ZeroKindU64
	case t.Size == 4:
		return ZeroKindU32
	case t.Size == 1:
		return ZeroKindU8
	}
	return ZeroKindNone
}

func (f *FieldDesc) checkTypeMatch() error {
	t := f.T
	if f.IsOneof() {
		// oneof fields are never repeated or map in protobuf; the decoder
		// also allocates a fresh wrapper per wire record, so a slice or map
		// here would silently keep only the last record on decode
		if f.Repeated || t.IsSlice || t.Kind == reflect.Map {
			return fmt.Errorf("unsupported repeated or map field in oneof %s", t.T)
		}
		if t.IsPointer && t.V.Kind != reflect.Struct {
			return fmt.Errorf("unsupported oneof pointer to non-message type %s", t.T)
		}
	}

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
		if t.IsSlice {
			// The element type (after at most one pointer indirection) must be a
			// scalar/struct, not another slice or map. Otherwise RealKind would
			// flatten e.g. [][]int32 to int32 and pass validation, and the decoder
			// would write a scalar slice header into a slice-of-slice element,
			// causing type confusion.
			el := t.V
			if el.IsPointer {
				el = el.V
				// pointer elements are only valid for message (struct) types:
				// scalar pointer elements would be routed to the flat-stride
				// list coders, corrupting the output
				if el.Kind != reflect.Struct {
					return fmt.Errorf("unsupported repeated pointer to non-message type %s", t.T)
				}
			}
			if el.IsSlice || el.Kind == reflect.Map {
				return fmt.Errorf("unsupported nested repeated field type %s", t.T)
			}
		}
	}

	if !f.Repeated {
		// converse of the check above: a slice or map (but not []byte, which is
		// the scalar bytes type) without `rep` would take an unsupported encode
		// path and read the container header as a value, producing corrupt output
		et := t
		if et.IsPointer {
			et = et.V
		}
		if et.IsSlice || et.Kind == reflect.Map {
			return errors.New("slice or map field must be repeated")
		}
	}

	if err := IsFieldTypeMatchReflectKind(f.TagType, t.RealKind()); err != nil {
		return err
	}
	if t.Kind == reflect.Map {
		// non-repeated maps were already rejected above
		// pointer map values are only valid for message (struct) types:
		// scalar pointer values would be read as flat values by the value
		// coder, leaking pointer bits (same class as the slice check above)
		if t.V.IsPointer && t.V.V.Kind != reflect.Struct {
			return fmt.Errorf("unsupported map value pointer to non-message type %s", t.T)
		}
		// slice/map values would flatten to a scalar RealKind, pass the
		// match below and then be incorrectly encoded by a flat value coder
		if t.V.IsSlice || t.V.Kind == reflect.Map {
			return fmt.Errorf("unsupported map value type %s", t.T)
		}
		// pointer keys flatten to a scalar RealKind and would leak the key
		// pointer bits on encode
		if t.K.IsPointer {
			return fmt.Errorf("unsupported map key pointer type %s", t.T)
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

var (
	// Entries inserted by the GetOrParse call currently holding parsemu. They
	// are rolled back as a unit if that call fails or panics.
	parseScopeTypes   map[reflect.Type]struct{}
	parseScopeStructs map[reflect.Type]struct{}
)

func startParseScope() {
	parseScopeTypes = make(map[reflect.Type]struct{})
	parseScopeStructs = make(map[reflect.Type]struct{})
}

func finishParseScope(parsed *bool) {
	if !*parsed {
		for rt := range parseScopeTypes {
			delete(cachedTypes, rt)
		}
		for rt := range parseScopeStructs {
			delete(cachedStructs, rt)
		}
	}
	parseScopeTypes = nil
	parseScopeStructs = nil
}

func parseStruct(rt reflect.Type) (s *StructDesc, err error) {
	if s = cachedStructs[rt]; s != nil {
		return s, nil
	}
	s = &StructDesc{Empty: reflect.New(rt).UnsafePointer()}
	cachedStructs[rt] = s // fix cyclic refs
	if parseScopeStructs != nil {
		parseScopeStructs[rt] = struct{}{}
	}
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
		// Type.Implements below panics on a non-interface type.
		for _, o := range oneofs {
			if o.Type.Kind() != reflect.Interface {
				return nil, fmt.Errorf("parse field %q err: oneof field is not an interface: %s", o.Name, o.Type)
			}
			if o.Type.NumMethod() == 0 {
				return nil, fmt.Errorf("parse field %q err: oneof field is an empty interface", o.Name)
			}
		}

		wrappers := searchOneofWrappers(rt)
		matched := make([]bool, len(oneofs))
		for _, v := range wrappers {
			wt := reflect.TypeOf(v)
			if wt == nil || wt.Kind() != reflect.Pointer || wt.Elem().Kind() != reflect.Struct {
				return nil, fmt.Errorf("oneof wrapper %T is not a pointer to struct", v)
			}
			st := wt.Elem() // Pointer -> Struct
			match := -1
			for i, o := range oneofs {
				if !wt.Implements(o.Type) {
					continue
				}
				if match >= 0 {
					return nil, fmt.Errorf("oneof wrapper %s implements multiple oneof interfaces", wt)
				}
				match = i
			}
			if match < 0 {
				continue
			}
			o := oneofs[match]
			if st.NumField() != 1 { // The struct must contain exactly one field.
				return nil, fmt.Errorf("parse field %q oneof %q err: field number != 1", o.Name, st.String())
			}
			field := st.Field(0)
			f := &FieldDesc{Name: o.Name, Offset: o.Offset, Tag: field.Tag, OneofType: wt}
			f.IfaceTab = hack.IfaceTab(o.Type, st)
			if err = f.parse(field.Type); err != nil {
				return nil, fmt.Errorf("parse field %q oneof %q err: %w", o.Name, st.String(), err)
			}
			fields = append(fields, f)
			matched[match] = true
		}
		for i, ok := range matched {
			if !ok {
				return nil, fmt.Errorf("parse field %q err: oneof has no wrappers", oneofs[i].Name)
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
			k = len(s.Fields) - i
			break
		}
		maxn = int(f.ID)
	}
	s.mFields0 = make([]*FieldDesc, maxn+1)
	s.mFields1 = make(map[int32]*FieldDesc, k)
	for i, f := range s.Fields {
		if f.ID <= maxDirectFieldMapID {
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
		// Must be a direct field. FieldByName also matches promoted fields of
		// embedded structs, whose Offset is relative to the embedded struct,
		// not rt; applying it to the outer base would corrupt memory.
		if len(f.Index) != 1 {
			continue
		}
		// Only []byte or *[]byte may be used as unknown-fields storage; any
		// other type would be misinterpreted as a slice header by the decoder.
		ft := f.Type
		isBytes := ft == bytesType
		isBytesPtr := ft.Kind() == reflect.Pointer && ft.Elem() == bytesType
		if !isBytes && !isBytesPtr {
			continue
		}
		s.HasUnknownFields = true
		s.UnknownFieldsPointer = isBytesPtr
		s.UnknownFieldsOffset = f.Offset
		break
	}

	return s, nil
}
