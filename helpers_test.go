// Copyright (c) 2015 SUSE LLC. All rights reserved.
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

import (
	"strings"
	"testing"
)

func TestParseImageName(t *testing.T) {
	// map with name as value and a string list with two enteries (repo and tag)
	// as value
	data := make(map[string][]string)
	data["opensuse:13.2"] = []string{"opensuse", "13.2"}
	data["opensuse"] = []string{"opensuse", "latest"}

	for name, expected := range data {
		repo, tag := parseImageName(name)
		if repo != expected[0] {
			t.Fatalf("repository %s is different from the expected %s", repo, expected[0])
		}
		if tag != expected[1] {
			t.Fatalf("tag %s is different from the expected %s", tag, expected[1])
		}
	}
}

func TestPreventImageOverwriteImageCheckImageFailure(t *testing.T) {
	dockerClient = &mockClient{listFail: true}

	err := preventImageOverwrite("opensuse", "13.2")

	if err == nil {
		t.Fatalf("Expected error")
	}
	if !strings.Contains(err.Error(), "List Failed") {
		t.Fatal("Wrong error message")
	}
}

func TestPreventImageOverwriteImageExists(t *testing.T) {
	dockerClient = &mockClient{}

	err := preventImageOverwrite("opensuse", "13.2")

	if err == nil {
		t.Fatalf("Expected error")
	}
	if !strings.Contains(err.Error(), "Cannot overwrite an existing image.") {
		t.Fatal("Wrong error message")
	}
}
