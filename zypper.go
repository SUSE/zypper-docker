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
	ZYPPER_EXIT_OK                    = 0
	ZYPPER_EXIT_ERR_BUG               = 1
	ZYPPER_EXIT_ERR_SYNTAX            = 2
	ZYPPER_EXIT_ERR_INVALID_ARGS      = 3
	ZYPPER_EXIT_ERR_ZYPP              = 4
	ZYPPER_EXIT_ERR_PRIVILEGES        = 5
	ZYPPER_EXIT_NO_REPOS              = 6
	ZYPPER_EXIT_ZYPP_LOCKED           = 7
	ZYPPER_EXIT_ERR_COMMIT            = 8
	ZYPPER_EXIT_INF_UPDATE_NEEDED     = 100
	ZYPPER_EXIT_INF_SEC_UPDATE_NEEDED = 101
	ZYPPER_EXIT_INF_REBOOT_NEEDED     = 102
	ZYPPER_EXIT_INF_RESTART_NEEDED    = 103
	ZYPPER_EXIT_INF_CAP_NOT_FOUND     = 104
	ZYPPER_EXIT_ON_SIGNAL             = 105
)

// Given zypper's exit code returns true if the error is
// a severe one. False otherwise. Severe errors will cause
// zypper-docker to exit with error.
func isZypperExitCodeSevere(errCode int) bool {
	switch errCode {
	case ZYPPER_EXIT_OK:
		return false
	case ZYPPER_EXIT_INF_REBOOT_NEEDED:
		return false
	case ZYPPER_EXIT_INF_UPDATE_NEEDED:
		return false
	case ZYPPER_EXIT_INF_SEC_UPDATE_NEEDED:
		return false
	case ZYPPER_EXIT_INF_RESTART_NEEDED:
		return false
	case ZYPPER_EXIT_ON_SIGNAL:
		return false
	default:
		return true
	}
}
