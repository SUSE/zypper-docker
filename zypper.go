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

const (
	zypperExitOK                 = 0
	zypperExitErrBug             = 1
	zypperExitErrSyntax          = 2
	zypperExitErrInvalidArgs     = 3
	zypperExitErrZyp             = 4
	zypperExitErrPrivileges      = 5
	zypperExitNoRepos            = 6
	zypperExitZyppLocked         = 7
	zypperExitErrCommit          = 8
	zypperExitInfUpdateNeeded    = 100
	zypperExitInfSecUpdateNeeded = 101
	zypperExitInfRebootNeeded    = 102
	zypperExitInfRestartNeeded   = 103
	zypperExitInfCapNotFound     = 104
	zypperExitOnSignal           = 105
	zypperExitInfReposSkipped    = 106
)

// isZypperExitCodeSevere returns true if errCode is a severe zypper error
// code, and will cause zypper-docker to exit with error.
func isZypperExitCodeSevere(errCode int) bool {
	// Codes below 100 denote an error, codes above 100 provide a specific
	// information, 0 represents a normal successful run.
	if errCode > 0 && errCode < 100 {
		return true
	}
	return false
}
