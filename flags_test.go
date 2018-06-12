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

package main

import "testing"

func TestVersion(t *testing.T) {
	if version() != "2.0.0" {
		t.Fatal("Wrong version")
	}
}

func TestNewApp(t *testing.T) {
	app := newApp()

	if len(app.Flags) != 5 {
		t.Fatal("Wrong number of global flags")
	}
	if len(app.Commands) != 10 {
		t.Fatal("Wrong number of subcommands")
	}
}
