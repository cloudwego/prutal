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

package prutal

import (
	"reflect"
	"sync"
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
	"github.com/cloudwego/prutal/internal/wire"
)

type mergeSub struct {
	V    int64   `protobuf:"varint,1,opt"`
	List []int64 `protobuf:"varint,2,rep"`

	unknownFields []byte
}

type mergeMsg struct {
	Sub  *mergeSub   `protobuf:"bytes,1,opt"`
	Subs []*mergeSub `protobuf:"bytes,2,rep"`
}

// protobuf defines concatenating two payloads as merging them, so a singular
// message field occurring twice must merge rather than be replaced, keeping
// the unknown fields of both occurrences.
func TestUnmarshalMergesSingularMessageField(t *testing.T) {
	b := (&wire.Builder{}).
		AppendBytesField(1, (&wire.Builder{}).AppendVarintField(1, 7).AppendVarintField(8, 1).Bytes()).
		AppendBytesField(1, (&wire.Builder{}).AppendVarintField(2, 9).AppendVarintField(9, 2).Bytes()).
		Bytes()

	v := &mergeMsg{}
	assert.NoError(t, Unmarshal(b, v))
	assert.Equal(t, int64(7), v.Sub.V)
	assert.SliceEqual(t, []int64{9}, v.Sub.List)
	// the unknown fields of both occurrences must survive the merge
	want := (&wire.Builder{}).AppendVarintField(8, 1).AppendVarintField(9, 2).Bytes()
	assert.BytesEqual(t, want, v.Sub.unknownFields)
}

// A repeated message field still appends: each occurrence is its own element.
func TestUnmarshalRepeatedMessageFieldAppends(t *testing.T) {
	b := (&wire.Builder{}).
		AppendBytesField(2, (&wire.Builder{}).AppendVarintField(1, 7).Bytes()).
		AppendBytesField(2, (&wire.Builder{}).AppendVarintField(1, 9).Bytes()).
		Bytes()

	v := &mergeMsg{}
	assert.NoError(t, Unmarshal(b, v))
	assert.Equal(t, 2, len(v.Subs))
	assert.Equal(t, int64(7), v.Subs[0].V)
	assert.Equal(t, int64(9), v.Subs[1].V)
}

func TestUnmarshalResetsTarget(t *testing.T) {
	t.Run("rich struct including unknownFields and oneof", func(t *testing.T) {
		type resetAll struct {
			F1            int32           `protobuf:"varint,1,opt"`
			FOther        string          `protobuf:"bytes,2,opt"`
			PtrInt        *int32          `protobuf:"varint,3,opt"`
			FInt32        []int32         `protobuf:"varint,4,rep"`
			MapInt32Int32 map[int32]int32 `protobuf:"bytes,5,rep" protobuf_key:"varint,1,opt" protobuf_val:"varint,2,opt"`
			Nested        *mergeSub       `protobuf:"bytes,6,opt"`
			unknownFields []byte
		}

		b1 := (&wire.Builder{}).
			AppendStringField(2, "new").
			AppendVarintField(999, 1).
			Bytes()
		b2 := (&wire.Builder{}).
			AppendVarintField(1, 22).
			Bytes()

		opt := int32(99)
		p := &resetAll{
			F1:            1,
			FOther:        "stale",
			PtrInt:        &opt,
			FInt32:        []int32{1, 2},
			MapInt32Int32: map[int32]int32{1: 2},
			Nested:        &mergeSub{V: 88},
			unknownFields: []byte{0x08, 0x01},
		}
		assert.NoError(t, Unmarshal(b1, p))
		assert.Equal(t, int32(0), p.F1)
		assert.Equal(t, "new", p.FOther)
		assert.True(t, p.PtrInt == nil)
		assert.True(t, p.FInt32 == nil)
		assert.True(t, p.MapInt32Int32 == nil)
		assert.True(t, p.Nested == nil)
		assert.BytesEqual(t, (&wire.Builder{}).AppendVarintField(999, 1).Bytes(), p.unknownFields)

		// Second Unmarshal clears prior string and unknownFields.
		assert.NoError(t, Unmarshal(b2, p))
		assert.Equal(t, int32(22), p.F1)
		assert.Equal(t, "", p.FOther)
		assert.True(t, len(p.unknownFields) == 0)

		om := &TestOneofMessage{
			OneOfFieldA: &TestOneofMessage_Field1{Field1: true},
			OneOfFieldB: &TestOneofMessage_Field3{Field3: 7},
			OneOfFieldC: &TestOneofMessage_Field5{Field5: &TestofNestedMessage{Field1: true}},
		}
		assert.NoError(t, Unmarshal((&wire.Builder{}).AppendVarintField(2, 55).Bytes(), om))
		assert.Equal(t, int64(55), om.OneOfFieldA.(*TestOneofMessage_Field2).Field2)
		assert.True(t, om.OneOfFieldB == nil)
		assert.True(t, om.OneOfFieldC == nil)
	})

	t.Run("pointer unknownFields scalar zero-size and dynamic struct", func(t *testing.T) {
		type unknownPtr struct {
			V             int32 `protobuf:"varint,1,opt"`
			unknownFields *[]byte
		}
		uf := []byte{0x08, 0x01}
		up := &unknownPtr{V: 9, unknownFields: &uf}
		assert.NoError(t, Unmarshal(nil, up))
		assert.Equal(t, int32(0), up.V)
		assert.True(t, up.unknownFields == nil)

		type scalarOnly struct {
			A int32 `protobuf:"varint,1,opt"`
			B int64 `protobuf:"varint,2,opt"`
		}
		so := &scalarOnly{A: 7, B: 8}
		assert.NoError(t, Unmarshal((&wire.Builder{}).AppendVarintField(1, 3).Bytes(), so))
		assert.Equal(t, int32(3), so.A)
		assert.Equal(t, int64(0), so.B)

		type empty struct{}
		assert.NoError(t, Unmarshal(nil, &empty{}))

		dt := reflect.StructOf([]reflect.StructField{
			{Name: "A", Type: reflect.TypeOf(int32(0)), Tag: `protobuf:"varint,1,opt"`},
			{Name: "S", Type: reflect.TypeOf(""), Tag: `protobuf:"bytes,2,opt"`},
		})
		dv := reflect.New(dt)
		dv.Elem().Field(0).SetInt(99)
		dv.Elem().Field(1).SetString("stale")
		assert.NoError(t, Unmarshal((&wire.Builder{}).AppendVarintField(1, 5).Bytes(), dv.Interface()))
		assert.Equal(t, int64(5), dv.Elem().Field(0).Int())
		assert.Equal(t, "", dv.Elem().Field(1).String())
	})

	t.Run("invalid descriptor does not reset target", func(t *testing.T) {
		type groupBad struct {
			V int32 `protobuf:"group,1,opt"`
		}
		v := &groupBad{V: 42}
		if err := Unmarshal(nil, v); err == nil {
			t.Fatal("expected error for unsupported group")
		}
		assert.Equal(t, int32(42), v.V)
	})

	t.Run("concurrent unmarshal reset", func(t *testing.T) {
		type concurrentMsg struct {
			A int32  `protobuf:"varint,1,opt"`
			S string `protobuf:"bytes,2,opt"`
		}
		b := (&wire.Builder{}).AppendVarintField(1, 7).Bytes()
		var wg sync.WaitGroup
		for g := 0; g < 8; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 100; i++ {
					m := &concurrentMsg{A: 1, S: "stale"}
					if err := Unmarshal(b, m); err != nil {
						t.Errorf("Unmarshal failed: %v", err)
						return
					}
					if m.A != 7 || m.S != "" {
						t.Errorf("unexpected decoded struct: %+v", m)
						return
					}
				}
			}()
		}
		wg.Wait()
	})
}

type mergeOneofParent struct {
	Kind isMergeOneof `protobuf_oneof:"kind"`
}

type isMergeOneof interface {
	isMergeOneof()
}

type mergeOneofSub struct {
	Sub *mergeSub `protobuf:"bytes,1,opt"`
}

func (*mergeOneofSub) isMergeOneof() {}

type mergeOneofVal struct {
	Val mergeSub `protobuf:"bytes,2,opt"`
}

func (*mergeOneofVal) isMergeOneof() {}

type mergeOneofInt struct {
	Num int32 `protobuf:"varint,3,opt"`
}

func (*mergeOneofInt) isMergeOneof() {}

func (*mergeOneofParent) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*mergeOneofSub)(nil),
		(*mergeOneofVal)(nil),
		(*mergeOneofInt)(nil),
	}
}

// TestUnmarshalMergesOneof verifies that same-case oneof members reuse existing
// wrappers on duplicate wire occurrences and UnmarshalMerge (matching
// protobuf-go), while switching cases or calling Unmarshal resets cleanly
// without mutating saved wrappers.
func TestUnmarshalMergesOneof(t *testing.T) {
	sub1 := (&wire.Builder{}).AppendVarintField(1, 7).AppendVarintField(8, 1).Bytes()
	sub2 := (&wire.Builder{}).AppendVarintField(2, 9).AppendVarintField(9, 2).Bytes()
	wantUnknown := (&wire.Builder{}).AppendVarintField(8, 1).AppendVarintField(9, 2).Bytes()

	t.Run("duplicate wire occurrence merges pointer message oneof", func(t *testing.T) {
		b := (&wire.Builder{}).
			AppendBytesField(1, sub1).
			AppendBytesField(1, sub2).
			Bytes()
		v := &mergeOneofParent{}
		assert.NoError(t, Unmarshal(b, v))
		got := v.Kind.(*mergeOneofSub).Sub
		assert.Equal(t, int64(7), got.V)
		assert.SliceEqual(t, []int64{9}, got.List)
		assert.BytesEqual(t, wantUnknown, got.unknownFields)
	})

	t.Run("UnmarshalMerge reuses existing pointer message oneof wrapper", func(t *testing.T) {
		existingSub := &mergeSub{
			V:             7,
			List:          []int64{3},
			unknownFields: (&wire.Builder{}).AppendVarintField(8, 1).Bytes(),
		}
		v := &mergeOneofParent{Kind: &mergeOneofSub{Sub: existingSub}}
		patch := (&wire.Builder{}).AppendBytesField(1, sub2).Bytes()

		assert.NoError(t, UnmarshalMerge(patch, v))
		gotWrapper := v.Kind.(*mergeOneofSub)
		assert.True(t, gotWrapper.Sub == existingSub)
		assert.Equal(t, int64(7), existingSub.V)
		assert.SliceEqual(t, []int64{3, 9}, existingSub.List)
		assert.BytesEqual(t, wantUnknown, existingSub.unknownFields)
	})

	t.Run("non-pointer message oneof merges on duplicate wire and UnmarshalMerge", func(t *testing.T) {
		b := (&wire.Builder{}).
			AppendBytesField(2, sub1).
			AppendBytesField(2, sub2).
			Bytes()
		v := &mergeOneofParent{}
		assert.NoError(t, Unmarshal(b, v))
		got := &v.Kind.(*mergeOneofVal).Val
		assert.Equal(t, int64(7), got.V)
		assert.SliceEqual(t, []int64{9}, got.List)
		assert.BytesEqual(t, wantUnknown, got.unknownFields)

		v2 := &mergeOneofParent{Kind: &mergeOneofVal{Val: mergeSub{V: 7, List: []int64{3}}}}
		assert.NoError(t, UnmarshalMerge((&wire.Builder{}).AppendBytesField(2, sub2).Bytes(), v2))
		got2 := &v2.Kind.(*mergeOneofVal).Val
		assert.Equal(t, int64(7), got2.V)
		assert.SliceEqual(t, []int64{3, 9}, got2.List)
	})

	t.Run("typed-nil wrapper allocates fresh wrapper", func(t *testing.T) {
		v := &mergeOneofParent{Kind: (*mergeOneofSub)(nil)}
		assert.NoError(t, UnmarshalMerge((&wire.Builder{}).AppendBytesField(1, sub1).Bytes(), v))
		got := v.Kind.(*mergeOneofSub)
		assert.True(t, got != nil && got.Sub != nil)
		assert.Equal(t, int64(7), got.Sub.V)
	})

	t.Run("wrapper with nil inner pointer allocates sub-message", func(t *testing.T) {
		origWrapper := &mergeOneofSub{Sub: nil}
		v := &mergeOneofParent{Kind: origWrapper}
		assert.NoError(t, UnmarshalMerge((&wire.Builder{}).AppendBytesField(1, sub1).Bytes(), v))
		got := v.Kind.(*mergeOneofSub)
		assert.True(t, got == origWrapper)
		assert.True(t, got.Sub != nil)
		assert.Equal(t, int64(7), got.Sub.V)
	})

	t.Run("switching oneof case and scalar oneof reuse", func(t *testing.T) {
		savedMsg := &mergeOneofSub{Sub: &mergeSub{V: 7}}
		v := &mergeOneofParent{Kind: savedMsg}

		// Message -> scalar switches wrapper without mutating savedMsg.
		assert.NoError(t, UnmarshalMerge((&wire.Builder{}).AppendVarintField(3, 11).Bytes(), v))
		assert.Equal(t, int32(11), v.Kind.(*mergeOneofInt).Num)
		assert.Equal(t, int64(7), savedMsg.Sub.V)

		// Scalar -> same scalar reuses existing wrapper (matching protobuf-go).
		savedInt := v.Kind.(*mergeOneofInt)
		assert.NoError(t, UnmarshalMerge((&wire.Builder{}).AppendVarintField(3, 22).Bytes(), v))
		assert.True(t, v.Kind.(*mergeOneofInt) == savedInt)
		assert.Equal(t, int32(22), savedInt.Num)

		// Typed-nil scalar wrapper allocates a fresh wrapper.
		vTypedNil := &mergeOneofParent{Kind: (*mergeOneofInt)(nil)}
		assert.NoError(t, UnmarshalMerge((&wire.Builder{}).AppendVarintField(3, 33).Bytes(), vTypedNil))
		assert.Equal(t, int32(33), vTypedNil.Kind.(*mergeOneofInt).Num)

		// Scalar -> message allocates a fresh message wrapper without mutating savedInt or savedMsg.
		assert.NoError(t, UnmarshalMerge((&wire.Builder{}).AppendBytesField(1, sub2).Bytes(), v))
		assert.SliceEqual(t, []int64{9}, v.Kind.(*mergeOneofSub).Sub.List)
		assert.Equal(t, int32(22), savedInt.Num)
		assert.Equal(t, int64(7), savedMsg.Sub.V)
	})

	t.Run("Unmarshal resets same-case oneof without merging or mutating saved wrapper", func(t *testing.T) {
		savedMsg := &mergeOneofSub{Sub: &mergeSub{V: 7, List: []int64{3}}}
		v := &mergeOneofParent{Kind: savedMsg}
		assert.NoError(t, Unmarshal((&wire.Builder{}).AppendBytesField(1, sub2).Bytes(), v))
		gotMsg := v.Kind.(*mergeOneofSub)
		assert.True(t, gotMsg != savedMsg)
		assert.True(t, gotMsg.Sub != savedMsg.Sub)
		assert.Equal(t, int64(0), gotMsg.Sub.V)
		assert.SliceEqual(t, []int64{9}, gotMsg.Sub.List)
		assert.Equal(t, int64(7), savedMsg.Sub.V)
		assert.SliceEqual(t, []int64{3}, savedMsg.Sub.List)

		savedInt := &mergeOneofInt{Num: 11}
		v.Kind = savedInt
		assert.NoError(t, Unmarshal((&wire.Builder{}).AppendVarintField(3, 22).Bytes(), v))
		gotInt := v.Kind.(*mergeOneofInt)
		assert.True(t, gotInt != savedInt)
		assert.Equal(t, int32(22), gotInt.Num)
		assert.Equal(t, int32(11), savedInt.Num)
	})
}

func TestUnmarshalMerge(t *testing.T) {
	unkInit := (&wire.Builder{}).AppendVarintField(98, 1).Bytes()
	unk2 := (&wire.Builder{}).AppendVarintField(99, 2).Bytes()
	b2 := (&wire.Builder{}).
		AppendVarintField(2, 9).
		AppendVarintField(9, 2).
		Bytes()

	t.Run("merges singular message slice map and unknownFields", func(t *testing.T) {
		type mergeAll struct {
			A             int32           `protobuf:"varint,1,opt"`
			S             string          `protobuf:"bytes,2,opt"`
			List          []int32         `protobuf:"varint,3,rep"`
			Map           map[int32]int32 `protobuf:"bytes,4,rep" protobuf_key:"varint,1,opt" protobuf_val:"varint,2,opt"`
			Sub           *mergeSub       `protobuf:"bytes,5,opt"`
			unknownFields []byte
		}

		p := &mergeAll{
			A:             10,
			S:             "keep",
			List:          []int32{1},
			Map:           map[int32]int32{1: 100, 2: 200},
			Sub:           &mergeSub{V: 7, List: []int64{3}},
			unknownFields: append([]byte(nil), unkInit...),
		}

		patch := (&wire.Builder{}).
			AppendStringField(2, "updated").
			AppendVarintField(3, 2).
			AppendBytesField(4, (&wire.Builder{}).AppendVarintField(1, 1).AppendVarintField(2, 101).Bytes()).
			AppendBytesField(4, (&wire.Builder{}).AppendVarintField(1, 3).AppendVarintField(2, 300).Bytes()).
			AppendBytesField(5, b2).
			Bytes()
		patch = append(patch, unk2...)

		assert.NoError(t, UnmarshalMerge(patch, p))
		assert.Equal(t, int32(10), p.A)
		assert.Equal(t, "updated", p.S)
		assert.SliceEqual(t, []int32{1, 2}, p.List)
		assert.Equal(t, int32(101), p.Map[1])
		assert.Equal(t, int32(200), p.Map[2])
		assert.Equal(t, int32(300), p.Map[3])
		assert.Equal(t, int64(7), p.Sub.V)
		assert.SliceEqual(t, []int64{3, 9}, p.Sub.List)
		assert.BytesEqual(t, append(append([]byte(nil), unkInit...), unk2...), p.unknownFields)
	})

	t.Run("merges pointer unknownFields", func(t *testing.T) {
		type mergePtrUnknown struct {
			unknownFields *[]byte
		}
		uf := append([]byte(nil), unkInit...)
		p := &mergePtrUnknown{unknownFields: &uf}
		assert.NoError(t, UnmarshalMerge(unk2, p))
		assert.BytesEqual(t, append(append([]byte(nil), unkInit...), unk2...), *p.unknownFields)
	})
}
