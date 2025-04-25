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

package prutalgen

import (
	"testing"

	"github.com/cloudwego/prutal/internal/testutils/assert"
)

func TestDirectives_ParseAndIsSet(t *testing.T) {
	// Test positive directive
	var d Directives
	d.Parse("//prutalgen:pointer")
	val, ok := d.IsSet("pointer")
	assert.Equal(t, true, val)
	assert.Equal(t, true, ok)

	// Test negative directive
	d.Parse("//prutalgen:no_pointer")
	val, ok = d.IsSet("pointer")
	assert.Equal(t, false, val)
	assert.Equal(t, true, ok)

	// Test first one wins - negative then positive
	d.Parse("//prutalgen:no_pointer", "//prutalgen:pointer")
	val, ok = d.IsSet("pointer")
	assert.Equal(t, false, val)
	assert.Equal(t, true, ok)

	// Test first one wins - positive then negative
	d.Parse("//prutalgen:pointer", "//prutalgen:no_pointer")
	val, ok = d.IsSet("pointer")
	assert.Equal(t, true, val)
	assert.Equal(t, true, ok)

	// Test unknown directive
	d.Parse("//prutalgen:pointer")
	val, ok = d.IsSet("unknown")
	assert.Equal(t, false, val)
	assert.Equal(t, false, ok)

	// Test empty directives
	d.Parse()
	val, ok = d.IsSet("pointer")
	assert.Equal(t, false, val)
	assert.Equal(t, false, ok)
}

func TestDirectives_Has(t *testing.T) {
	// Test has directive
	var d Directives
	d.Parse("//prutalgen:pointer", "//prutalgen:unknown_fields")
	assert.Equal(t, true, d.Has("pointer"))

	// Test does not have directive
	d.Parse("//prutalgen:unknown_fields")
	assert.Equal(t, false, d.Has("pointer"))

	// Test empty directives
	d.Parse()
	assert.Equal(t, false, d.Has("pointer"))
}

func TestDirectives_Parse(t *testing.T) {
	// Test single directive
	var d Directives
	d.Parse("//prutalgen:pointer")
	assert.Equal(t, 1, len(d))
	assert.Equal(t, "pointer", d[0])

	// Test multiple directives
	d.Parse("//prutalgen:pointer", "//prutalgen:unknown_fields")
	assert.Equal(t, 2, len(d))
	assert.Equal(t, "pointer", d[0])
	assert.Equal(t, "unknown_fields", d[1])

	// Test empty input
	d.Parse()
	assert.Equal(t, 0, len(d))

	// Test directive with comma
	d.Parse("//prutalgen:pointer,unknown_fields")
	assert.Equal(t, 2, len(d))
	assert.Equal(t, "pointer", d[0])
	assert.Equal(t, "unknown_fields", d[1])

	// Test directive with spaces
	d.Parse("//prutalgen: pointer ", "//prutalgen: unknown_fields ")
	assert.Equal(t, 2, len(d))
	assert.Equal(t, "pointer", d[0])
	assert.Equal(t, "unknown_fields", d[1])
}

func TestDirectives_Reset(t *testing.T) {
	d := Directives{"pointer", "no_unknown_fields"}
	d.Reset()
	assert.Equal(t, 0, len(d))
}
