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

package edition2023

import (
	"testing"

	"github.com/cloudwego/prutal"
	"github.com/cloudwego/prutal/internal/testutils/assert"
	"github.com/cloudwego/prutal/tests/cases/edition2023/prutalout"
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

	// Size test: prutal.Size must match len(prutal.Marshal)
	bs, err = prutal.Marshal(p)
	assert.NoError(t, err)
	sz, err := prutal.Size(p)
	assert.NoError(t, err)
	assert.Equal(t, len(bs), sz)
}

func TestEdition2023PrutalgenInterop(t *testing.T) {
	oneofs := []isTestAllTypesEdition2023_OneofField{
		&TestAllTypesEdition2023_OneofUint32{OneofUint32: 42},
		&TestAllTypesEdition2023_OneofNestedMessage{OneofNestedMessage: &TestAllTypesEdition2023_NestedMessage{A: proto.Int32(7)}},
		&TestAllTypesEdition2023_OneofString{OneofString: "hello"},
		&TestAllTypesEdition2023_OneofBytes{OneofBytes: []byte("world")},
		&TestAllTypesEdition2023_OneofBool{OneofBool: true},
		&TestAllTypesEdition2023_OneofEnum{OneofEnum: TestAllTypesEdition2023_NEG},
	}

	for _, of := range oneofs {
		p := &TestAllTypesEdition2023{}
		err := faker.FakeData(p, options.WithIgnoreInterface(true), options.WithRandomMapAndSliceMaxSize(16))
		assert.NoError(t, err)
		p.OneofField = of

		protoBytes, err := proto.Marshal(p)
		assert.NoError(t, err)

		// Decode protobuf-go wire bytes into prutalgen-generated struct.
		gen := &prutalout.TestAllTypesEdition2023{}
		assert.NoError(t, prutal.Unmarshal(protoBytes, gen))

		// Re-encode from prutalgen-generated struct and verify Size + round-trip.
		prutalBytes, err := prutal.Marshal(gen)
		assert.NoError(t, err)
		sz, err := prutal.Size(gen)
		assert.NoError(t, err)
		assert.Equal(t, len(prutalBytes), sz)

		roundTrip := &TestAllTypesEdition2023{}
		assert.NoError(t, proto.Unmarshal(prutalBytes, roundTrip))
		assert.True(t, proto.Equal(p, roundTrip))
	}
}

func TestEdition2023FeaturesPrutalgenInterop(t *testing.T) {
	zeroI32 := int32(0)
	reqI32 := int32(9)
	emptyStr := ""
	openZero := OpenEnum_OPEN_UNSPECIFIED
	closedVal := ClosedEnum_CLOSED_B

	ref := &FeaturesInterop{
		IExplicit:          &zeroI32,
		IImplicit:          0,
		IRequired:          &reqI32,
		ByExplicit:         []byte{},
		ByImplicit:         []byte{},
		ByRequired:         []byte{},
		SExplicit:          &emptyStr,
		SImplicit:          "",
		PackedI32:          []int32{-1, 0, 128},
		ExpandedI32:        []int32{-2, 0, 256},
		PackedOpenEnum:     []OpenEnum{OpenEnum_OPEN_UNSPECIFIED, OpenEnum_OPEN_B},
		ExpandedClosedEnum: []ClosedEnum{ClosedEnum_CLOSED_A, ClosedEnum_CLOSED_B},
		OpenExplicit:       &openZero,
		OpenImplicit:       OpenEnum_OPEN_A,
		ClosedExplicit:     &closedVal,
		Child: &FeaturesInterop{
			IExplicit:  proto.Int32(11),
			IRequired:  proto.Int32(12),
			ByExplicit: []byte{1, 2},
			ByImplicit: []byte{3},
			ByRequired: []byte{4},
		},
		Pick: &FeaturesInterop_PickMsg{
			PickMsg: &FeaturesInterop{
				IImplicit:  77,
				IRequired:  proto.Int32(78),
				ByRequired: []byte{5},
			},
		},
	}

	refBytes, err := proto.Marshal(ref)
	assert.NoError(t, err)

	// Decode into prutalgen-generated struct and verify getters/presence.
	gen := &prutalout.FeaturesInterop{}
	assert.NoError(t, prutal.Unmarshal(refBytes, gen))
	assert.True(t, gen.IExplicit != nil && gen.GetIExplicit() == 0)
	assert.Equal(t, int32(0), gen.GetIImplicit())
	assert.True(t, gen.IRequired != nil && gen.GetIRequired() == 9)
	assert.True(t, gen.ByExplicit != nil && len(gen.ByExplicit) == 0)
	assert.True(t, gen.ByImplicit == nil)
	assert.True(t, gen.ByRequired != nil && len(gen.ByRequired) == 0)
	assert.True(t, gen.SExplicit != nil && gen.GetSExplicit() == "")

	// Explicitly set empty non-nil bytes on implicit field: prutalgen's
	// "implicit" tag marker must omit it on the wire, matching proto.Marshal.
	gen.ByImplicit = []byte{}

	genBytes, err := prutal.Marshal(gen)
	assert.NoError(t, err)
	assert.BytesEqual(t, refBytes, genBytes)

	sz, err := prutal.Size(gen)
	assert.NoError(t, err)
	assert.Equal(t, len(refBytes), sz)
	assert.Equal(t, proto.Size(ref), sz)

	roundTrip := &FeaturesInterop{}
	assert.NoError(t, proto.Unmarshal(genBytes, roundTrip))
	assert.True(t, proto.Equal(ref, roundTrip))
}
