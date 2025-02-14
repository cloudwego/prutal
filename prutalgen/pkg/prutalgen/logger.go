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
	"log"
	"os"
)

type LoggerIface interface {
	Fatalf(format string, v ...any) // must call Exit
	Printf(format string, v ...any)
}

var defaultLogger LoggerIface = log.New(os.Stderr, "", 0)

// LoggerFunc converts a func to LoggerIface
type LoggerFunc func(format string, v ...any)

func (f LoggerFunc) Printf(format string, v ...any) {
	f(format, v...)
}

func (f LoggerFunc) Fatalf(format string, v ...any) {
	f(format, v...)
	os.Exit(1)
}
