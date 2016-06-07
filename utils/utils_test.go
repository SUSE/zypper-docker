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

package utils

import "testing"

func TestJoinAsArray(t *testing.T) {
	str := JoinAsArray([]string{}, false)
	if str != "[]" {
		t.Fatalf("Expected '[]', got: %s", str)
	}
	str = JoinAsArray([]string{}, true)
	if str != "" {
		t.Fatalf("Expected '', got: %s", str)
	}
	str = JoinAsArray([]string{"one"}, false)
	if str != "[\"one\"]" {
		t.Fatalf("Expected '[\"one\"]', got: %s", str)
	}
	str = JoinAsArray([]string{"one", "two"}, false)
	if str != "[\"one\", \"two\"]" {
		t.Fatalf("Expected '[\"one\", \"two\"]', got: %s", str)
	}
}

func TestRemoveDuplicates(t *testing.T) {
	expected := []string{"this", "string", "contains", "duplicates"}
	got := RemoveDuplicates([]string{"this", "string", "contains", "contains", "duplicates"})
	if len(expected) != len(got) {
		t.Fatalf("Expected %v, got %v", expected, got)
	}
	for i := 0; i < len(got); i++ {
		if expected[i] != got[i] {
			t.Fatalf("Expected %v, got %v", expected, got)
		}
	}

	got = RemoveDuplicates([]string{})
	if len(got) > 0 {
		t.Fatalf("Expected empty array, got %v", got)
	}
}
