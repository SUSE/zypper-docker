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

// PATCH

func TestPatchCommand(t *testing.T) {
	cases := testCases{
		{"Wrong number of arguments", &mockClient{}, 1, []string{}, true, "Wrong invocation: expected 2 arguments, 0 given.", ""},
		{"Wrong format of image name", &mockClient{}, 1, []string{"ori", "dollar$$"}, true, "Could not parse 'dollar$$': invalid reference format", ""},
		{"List Command fails", &mockClient{listFail: true}, 1, []string{"ori", "opensuse:13.2"}, true, "Cannot proceed safely: List Failed.", ""},
		{"Overwrite detected", &mockClient{}, 1, []string{"ori", "opensuse:13.2"}, true, "Cannot overwrite an existing image. Please use a different repository/tag.", ""},
		{"Start fail on commit", &mockClient{startFail: true}, 1, []string{"ori", "new:1.0.0"}, true, "Could not commit to the new image: Start failed.", ""},
		{"Cannot update cache", &mockClient{}, 1, []string{"ori", "new:1.0.0"}, false, "Cannot add image details to zypper-docker cache", ""},
		{"Cannot inspect", &mockClient{inspectFail: true}, 1, []string{"opensuse:13.2", "new:1.0.0"}, true, "could not inspect image 'opensuse:13.2': inspect fail!", ""},
		{"Patch success", &mockClient{listReturnOneImage: true}, 0, []string{"opensuse:13.2", "new:1.0.0"}, true, "new:1.0.0 successfully created", ""},
	}
	cases.run(t, patchCmd, "zypper -n patch", "")
}

// LIST PATCHES

func TestListPatchesCommand(t *testing.T) {
	cases := testCases{
		{"No image specified", &mockClient{}, 1, []string{}, true, "no image name specified", ""},
		{"Command fail", &mockClient{commandFail: true}, 1, []string{"opensuse:13.2"}, false, "Error: Command exited with status 1", ""},
		{"List patches", &mockClient{}, 0, []string{"opensuse:13.2"}, false, "Removed container zypper-docker-private-opensuse:13.2", "streaming buffer initialized"},
	}
	cases.run(t, listPatchesCmd, "zypper lp", "")
}

// LIST PATCHES CONTAINER

func TestListPatchesContainerCommand(t *testing.T) {
	cases := testCases{
		{"List fails on list patch container", &mockClient{listFail: true}, 1, []string{"opensuse:13.2"}, true, "Error while fetching running containers: Fake failure while listing containers", ""},
		{"Patches container successfully", &mockClient{}, 0, []string{"suse"}, false, "Removed container zypper-docker-private-opensuse:13.2", "streaming buffer initialized"},
	}
	cases.run(t, listPatchesContainerCmd, "zypper lp", "")
}
