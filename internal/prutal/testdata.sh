#!/bin/bash
set -e
unset GOARCH # fix go install

cd ../../prutalgen/
go install
cd -

add_license_header() (
	generated_file="$1"
	licensed_file=$(mktemp "${generated_file}.XXXXXX")
	trap 'rm -f -- "$licensed_file"' EXIT

	{
		cat <<'EOF'
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

EOF
		cat "$generated_file"
	} > "$licensed_file"

	chmod 0644 "$licensed_file"
	mv "$licensed_file" "$generated_file"
)

prutalgen --proto_path=. --go_out=../ ./testdata.proto
add_license_header testdata.pb.go

prutalgen --proto_path=. --go_out=../ ./testdata_edition2023.proto
add_license_header testdata_edition2023.pb.go
