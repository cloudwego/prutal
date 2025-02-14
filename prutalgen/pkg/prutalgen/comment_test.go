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

func TestConsumeComment(t *testing.T) {

	f := loadTestProto(t, `option go_package = "testcomment";

// comment line1
message Hello {
// comment line2
string a = 1; // comment line3
// comment line4
string b = 2; /*
comment line5
line5+
line5++
*/
// comment line6
// line7
// line8
string c = 3;
// comment line9

/*
comment line10
line11
line12
*/
string d = 4;
}
	`)

	m := f.Messages[0]
	assert.Equal(t, "// comment line1", m.HeadComment)
	assert.Equal(t, "// comment line2", m.Fields[0].HeadComment)
	assert.Equal(t, "// comment line3", m.Fields[0].InlineComment)
	assert.Equal(t, "// comment line4", m.Fields[1].HeadComment)
	assert.Equal(t, "/*\ncomment line5\nline5+\nline5++\n*/", m.Fields[1].InlineComment)
	assert.Equal(t, "// comment line6\n// line7\n// line8", m.Fields[2].HeadComment)
	assert.Equal(t, "", m.Fields[2].InlineComment) // should not contain line9
	assert.Equal(t, "/*\ncomment line10\nline11\nline12\n*/", m.Fields[3].HeadComment)
}
