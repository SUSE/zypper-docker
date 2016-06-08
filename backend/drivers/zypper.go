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

// Zypper implements the Driver interface for the zypper tool.
type Zypper struct{}

// GeneralUpdate TODO
func (*Zypper) GeneralUpdate() (string, error) {
	return zypperUpdate("up")
}

// SecurityUpdate TODO
func (*Zypper) SecurityUpdate() (string, error) {
	return zypperUpdate("patch")
}

// ListGeneralUpdates TODO
func (*Zypper) ListGeneralUpdates() (string, error) {
	return formatZypperCommand("ref", "lu"), nil
}

// ListSecurityUpdates TODO
func (*Zypper) ListSecurityUpdates() (string, error) {
	return formatZypperCommand("ref", "lp"), nil
}

// CheckPatches TODO
func (*Zypper) CheckPatches() (string, error) {
	return formatZypperCommand("ref", "pchk"), nil
}

// IsExitCodeSevere TODO
func (*Zypper) IsExitCodeSevere(code int) (bool, error) {
	switch code {
	case zypperExitOK:
	case zypperExitInfRebootNeeded:
	case zypperExitInfUpdateNeeded:
	case zypperExitInfSecUpdateNeeded:
	case zypperExitInfRestartNeeded:
	case zypperExitOnSignal:
	default:
		return true, nil
	}
	return false, nil
}

func (*Zypper) needsCLI() bool {
	return true
}

func zypperUpdate(subcommand string) (string, error) {
	boolFlags := []string{"l", "auto-agree-with-licenses", "no-recommends",
		"replacefiles"}
	toIgnore := []string{"author", "message"}

	cmd := formatZypperCommand("ref", fmt.Sprintf("-n %v", subcommand), "clean -a")
	return cmdWithFlags(cmd, cliContext, boolFlags, toIgnore), nil
}

// TODO
var specialFlags = []string{
	"--bugzilla",
	"--cve",
	"--issues",
}

// Returns a string containing the global flags being used.
func globalFlags() string {
	if cliContext == nil {
		return ""
	}

	res := "--non-interactive "
	flags := []string{"no-gpg-checks", "gpg-auto-import-keys"}

	for _, v := range flags {
		if cliContext.GlobalBool(v) {
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
