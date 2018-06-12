// Copyright (c) 2018 SUSE LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This is a thin wrapper on top of zypper that allows patching docker images
// in a safe way.
package main

import "os"

var exitWithCode func(code int)
var killChannel chan bool

func main() {
	listenSignals()

	exitWithCode = func(code int) {
		os.Exit(code)
	}

	os.Args = fixArgsForZypper(os.Args)
	app := newApp()
	app.RunAndExitOnError()

	// TODO: add tests to check for correctly passing exit codes
	os.Exit(int(zypperExitCode))
}
