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

package backend

import (
	"fmt"
	"strings"

	"github.com/SUSE/zypper-docker/utils"
	"github.com/codegangsta/cli"
)

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
	zypperExitIndCapNotFound     = 104
	zypperExitOnSignal           = 105
)

// Returns a string containing the global flags being used.
func globalFlags() string {
	if CLIContext == nil {
		return ""
	}

	res := "--non-interactive "
	flags := []string{"no-gpg-checks", "gpg-auto-import-keys"}

	for _, v := range flags {
		if CLIContext.GlobalBool(v) {
			res = res + "--" + v + " "
		}
	}
	return res
}

// Concatenate the given zypper commands, while adding the global flags
// currently in place.
func formatZypperCommand(cmds ...string) string {
	flags := globalFlags()

	for k, v := range cmds {
		cmds[k] = "zypper " + flags + v
	}
	return strings.Join(cmds, " && ")
}

// Given zypper's exit code returns true if the error is
// a severe one. False otherwise. Severe errors will cause
// zypper-docker to exit with error.
func isZypperExitCodeSevere(errCode int) bool {
	switch errCode {
	case zypperExitOK:
		return false
	case zypperExitInfRebootNeeded:
		return false
	case zypperExitInfUpdateNeeded:
		return false
	case zypperExitInfSecUpdateNeeded:
		return false
	case zypperExitInfRestartNeeded:
		return false
	case zypperExitOnSignal:
		return false
	default:
		return true
	}
}

// It appends the set flags with the given command.
// `boolFlags` is a list of strings containing the names of the boolean
// command line options. These have to be handled in a slightly different
// way because zypper expects `--boolflag` instead of `--boolflag true`. Also
// boolean flags with a false value are ignored because zypper set all the
// undefined bool flags to false by default.
// `toIgnore` contains a list of flag names to not be passed to the final
//  command, this is useful to prevent zypper-docker only parameters to be
// forwarded to zypper (eg: `--author` or `--message`).
func cmdWithFlags(cmd string, ctx *cli.Context, boolFlags, toIgnore []string) string {
	for _, name := range ctx.FlagNames() {
		if utils.ArrayIncludeString(toIgnore, name) {
			continue
		}

		if value := ctx.String(name); ctx.IsSet(name) {
			var dash string
			if len(name) == 1 {
				dash = "-"
			} else {
				dash = "--"
			}

			if utils.ArrayIncludeString(boolFlags, name) {
				cmd += fmt.Sprintf(" %v%s", dash, name)
			} else {
				if utils.ArrayIncludeString(specialFlags, fmt.Sprintf("%v%s", dash, name)) && value != "" {
					cmd += fmt.Sprintf(" %v%s=%s", dash, name, value)
				} else {
					cmd += fmt.Sprintf(" %v%s %s", dash, name, value)
				}
			}
		}
	}

	return cmd
}
