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
