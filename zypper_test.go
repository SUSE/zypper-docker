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
	"testing"
)

func TestIsZypperExitCodeSevere(t *testing.T) {
	notSevereExitCodes := []int{
		ZYPPER_EXIT_OK,
		ZYPPER_EXIT_INF_REBOOT_NEEDED,
		ZYPPER_EXIT_INF_REBOOT_NEEDED,
		ZYPPER_EXIT_INF_UPDATE_NEEDED,
		ZYPPER_EXIT_INF_SEC_UPDATE_NEEDED,
		ZYPPER_EXIT_INF_RESTART_NEEDED,
		ZYPPER_EXIT_ON_SIGNAL,
	}

	for _, code := range notSevereExitCodes {
		if isZypperExitCodeSevere(code) {
			t.Fatalf("Exit code %v should not be considered a severe error", code)
		}
	}

	severeExitCodes := []int{
		ZYPPER_EXIT_ERR_BUG,
		ZYPPER_EXIT_ERR_SYNTAX,
		ZYPPER_EXIT_ERR_INVALID_ARGS,
		ZYPPER_EXIT_ERR_ZYPP,
		ZYPPER_EXIT_ERR_PRIVILEGES,
		ZYPPER_EXIT_NO_REPOS,
		ZYPPER_EXIT_ZYPP_LOCKED,
		ZYPPER_EXIT_ERR_COMMIT,
		ZYPPER_EXIT_INF_CAP_NOT_FOUND,
		127,
	}

	for _, code := range severeExitCodes {
		if !isZypperExitCodeSevere(code) {
			t.Fatalf("Exit code %v should be considered a severe error", code)
		}
	}
}
