package others

import (
	"testing"

	"github.com/cloudwego/prutal"
	"github.com/cloudwego/prutal/internal/testutils/assert"
	"google.golang.org/protobuf/proto"
)

func TestNilElem(t *testing.T) {
	p0 := &TestNilElemMsg{V: 7}
	p := &TestNilElemMsg{V: 1,
		NestedList: []*TestNilElemMsg{nil, p0, nil},
		NestedMap:  map[int64]*TestNilElemMsg{9: nil},
	}

	// test Marshal by proto

	b, err := proto.Marshal(p)

	p1 := &TestNilElemMsg{}
	err = prutal.Unmarshal(b, p1)
	assert.NoError(t, err)
	assert.DeepEqual(t, p1.NestedList[0], &TestNilElemMsg{})
	assert.DeepEqual(t, p1.NestedMap[9], &TestNilElemMsg{})

	p2 := &TestNilElemMsg{}
	err = proto.Unmarshal(b, p2)
	assert.NoError(t, err)

	assert.DeepEqual(t, p1, p2)

	// test Marshal by prutal

	b, err = prutal.Marshal(p)
	p1 = &TestNilElemMsg{}
	err = prutal.Unmarshal(b, p1)
	assert.NoError(t, err)

	p2 = &TestNilElemMsg{}
	err = proto.Unmarshal(b, p2)
	assert.NoError(t, err)

	assert.DeepEqual(t, p1, p2)
}

func TestNegativeInt32(t *testing.T) {
	p := &TestNegativeInt32Msg{
		V:              -1,
		List:           []int32{-1, -100, -1000, 0, 1, 100},
		PackedList:     []int32{-1, -2, -3},
		MapInt32Int32:  map[int32]int32{-1: -2, -100: -200, 1: 2},
		MapInt32String: map[int32]string{-1: "neg", 0: "zero", 1: "pos"},
		MapStringInt32: map[string]int32{"neg": -1, "zero": 0, "pos": 1},
		Nested:         &TestNegativeInt32Msg{V: -999},
	}

	// proto.Marshal -> prutal.Unmarshal
	bs, err := proto.Marshal(p)
	assert.NoError(t, err)

	p1 := &TestNegativeInt32Msg{}
	err = prutal.Unmarshal(bs, p1)
	assert.NoError(t, err)
	assert.DeepEqual(t, p, p1)

	// prutal.Marshal -> proto.Unmarshal
	bs, err = prutal.Marshal(p)
	assert.NoError(t, err)

	p2 := &TestNegativeInt32Msg{}
	err = proto.Unmarshal(bs, p2)
	assert.NoError(t, err)
	assert.DeepEqual(t, p, p2)

	// prutal.Marshal must produce same size as proto.Marshal
	assert.Equal(t, proto.Size(p), len(bs))
}
