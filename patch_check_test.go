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

import "testing"

// PATCH-CHECK

func TestPatchCheckCommand(t *testing.T) {
	cases := testCases{
		{"Image not specified", &mockClient{}, 1, []string{}, true, "Error: no image name specified.", ""},
		{"Invalid error", &mockClient{commandFail: true, commandExit: 2}, 1, []string{"opensuse:13.2"}, false,
			"Could not execute command 'zypper pchk' successfully in image 'opensuse:13.2': Command exited with status 2.",
			"streaming buffer initialized"},
		{"Supported non-zero exit", &mockClient{commandFail: true, commandExit: 100}, 100, []string{"opensuse:13.2"}, false,
			"Removed container zypper-docker-private-opensuse:13.2",
			"streaming buffer initialized"},
		{"Ok", &mockClient{}, 0, []string{"opensuse:13.2"}, false, "Removed container zypper-docker-private-opensuse:13.2",
			"streaming buffer initialized"},
	}
	cases.run(t, patchCheckCmd, "zypper pchk", "")
}

// PATCH-CHECK-CONTAINER

func TestPatchCheckContainerCommand(t *testing.T) {
	cases := testCases{
		{"List Command fails", &mockClient{listFail: true}, 1, []string{"opensuse:13.2"}, true,
			"Error while fetching running containers: Fake failure while listing containers.", ""},
		{"Ok", &mockClient{}, 0, []string{"suse"}, false, "Removed container zypper-docker-private-opensuse:13.2",
			"streaming buffer initialized"},
	}
	cases.run(t, patchCheckContainerTest, "zypper pchk", "")
}
