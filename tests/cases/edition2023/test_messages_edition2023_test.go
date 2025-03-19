package edition2023

import (
	"testing"

	"github.com/cloudwego/prutal"
	"github.com/cloudwego/prutal/internal/testutils/assert"
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"google.golang.org/protobuf/proto"
)

func TestEdition2023(t *testing.T) {
	p := &TestAllTypesEdition2023{}
	err := faker.FakeData(p, options.WithIgnoreInterface(true), options.WithRandomMapAndSliceMaxSize(33))
	assert.NoError(t, err)

	bs, err := proto.Marshal(p)
	assert.NoError(t, err)

	p0 := &TestAllTypesEdition2023{}
	err = prutal.Unmarshal(bs, p0)
	assert.NoError(t, err)

	assert.DeepEqual(t, p, p0)

	bs, err = prutal.Marshal(p)
	assert.NoError(t, err)

	p0 = &TestAllTypesEdition2023{}
	err = proto.Unmarshal(bs, p0)
	assert.NoError(t, err)

	assert.DeepEqual(t, p, p0)
}
