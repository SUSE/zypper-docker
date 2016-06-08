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

package drivers

import (
	"errors"
	"testing"
)

type ownError struct {
}

func (ownError) Error() string {
	return "wat?!"
}

func TestIsNotSupported(t *testing.T) {
	not := notSupportedError{name: "a"}
	if !IsNotSupported(not) {
		t.Fatalf("Should've been not supported")
	}
	msg := "action not supported by 'a'"
	if not.Error() != msg {
		t.Fatalf("Got: %v; expecting: %v", not.Error(), msg)
	}

	if IsNotSupported(errors.New("some other")) {
		t.Fatalf("Should've been some other")
	}
	if IsNotSupported(ownError{}) {
		t.Fatalf("Should've been some other")
	}
}
