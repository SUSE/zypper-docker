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

import (
	"testing"
)

func TestIsZypperExitCodeSevere(t *testing.T) {
	notSevereExitCodes := []int{
		zypperExitOK,
		zypperExitInfRebootNeeded,
		zypperExitInfUpdateNeeded,
		zypperExitInfSecUpdateNeeded,
		zypperExitInfRestartNeeded,
		zypperExitInfCapNotFound,
		zypperExitOnSignal,
		zypperExitInfReposSkipped,
		666,
	}

	for _, code := range notSevereExitCodes {
		if isZypperExitCodeSevere(code) {
			t.Fatalf("Exit code %v should not be considered a severe error", code)
		}
	}

	severeExitCodes := []int{
		zypperExitErrBug,
		zypperExitErrSyntax,
		zypperExitErrInvalidArgs,
		zypperExitErrZyp,
		zypperExitErrPrivileges,
		zypperExitNoRepos,
		zypperExitZyppLocked,
		zypperExitErrCommit,
		42,
		99,
	}

	for _, code := range severeExitCodes {
		if !isZypperExitCodeSevere(code) {
			t.Fatalf("Exit code %v should be considered a severe error", code)
		}
	}
}
