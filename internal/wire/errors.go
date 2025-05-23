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

package wire

import "fmt"

func newFieldNumErr(got, expect int32) error {
	return fmt.Errorf("field num not match: got %d expect %d", got, expect)
}

func newTypeNotMatchErr(got, expect Type) error {
	return fmt.Errorf("wire type not match: got %s expect %s", got, expect)
}
