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
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func int32p(v int32) *int32 { return &v }

// TestEdition2023PresenceEncoding runs the code prutalgen generates for every way an
// edition 2023 field tracks presence through the encoder, and expects what
// protobuf-go writes for the same values.
func TestEdition2023PresenceEncoding(t *testing.T) {
	for _, c := range []struct {
		name string
		m    *TestEdition2023Presence
		want []byte
	}{
		{"unset", &TestEdition2023Presence{}, nil},
		{"explicit empty bytes", &TestEdition2023Presence{ByExplicit: []byte{}}, []byte{0x0a, 0x00}},
		{"implicit empty bytes", &TestEdition2023Presence{ByImplicit: []byte{}}, nil},
		{"implicit set bytes", &TestEdition2023Presence{ByImplicit: []byte{1}}, []byte{0x12, 0x01, 0x01}},
		{"required empty bytes", &TestEdition2023Presence{ByRequired: []byte{}}, []byte{0x1a, 0x00}},
		{"aggregate implicit empty bytes", &TestEdition2023Presence{ByAggregate: []byte{}}, nil},
		{"explicit zero int32", &TestEdition2023Presence{IExplicit: int32p(0)}, []byte{0x28, 0x00}},
		{"implicit zero int32", &TestEdition2023Presence{IImplicit: 0}, nil},
		{"required zero int32", &TestEdition2023Presence{IRequired: int32p(0)}, []byte{0x38, 0x00}},
		{"implicit empty string", &TestEdition2023Presence{SImplicit: ""}, nil},
		{"oneof empty bytes", &TestEdition2023Presence{K: &TestEdition2023Presence_ByOneof{ByOneof: []byte{}}}, []byte{0x4a, 0x00}},
		{"list empty bytes", &TestEdition2023Presence{ByList: [][]byte{{}}}, []byte{0x52, 0x00}},
	} {
		t.Run(c.name, func(t *testing.T) {
			b, err := MarshalAppend(nil, c.m)
			assert.NoError(t, err)
			assert.BytesEqual(t, c.want, b)
			n, err := Size(c.m)
			assert.NoError(t, err)
			assert.Equal(t, len(c.want), n)
		})
	}
}
