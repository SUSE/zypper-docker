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
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/SUSE/zypper-docker/logger"
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
func (*Zypper) ListSecurityUpdates(machine bool) (string, error) {
	if machine {
		return formatZypperCommand("-q ref", "-q -t -x lp"), nil
	}
	return formatZypperCommand("-q ref", "lp"), nil
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

// NeedsCLI returns true.
func (*Zypper) NeedsCLI() bool {
	return true
}

// SeverityCommand TODO
func (*Zypper) SeverityCommand() string {
	return "zypper lp --severity"
}

// SeveritySupported TODO
func (*Zypper) SeveritySupported(output string) bool {
	if strings.Contains(output, "Missing argument for --severity") {
		return true
	}
	if strings.Contains(output, "Unknown option '--severity'") {
		return false
	}
	return false
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

type Updates struct {
	Updates  int
	Security int
	Error    string

	List []UpdateInfo
}

// UpdateInfo TODO
type UpdateInfo struct {
	Name        string
	IsSecurity  bool
	Severity    string
	Kind        string
	Summary     string
	Description string
}

// ParseUpdateOutput TODO
func (*Zypper) ParseUpdateOutput(output []byte) Updates {
	var up Updates
	var t xml.Token
	var err error

	d := xml.NewDecoder(bytes.NewReader(output))
	for {
		t, err = d.RawToken()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}

		if e, ok := t.(xml.StartElement); ok && strings.EqualFold(e.Name.Local, "update") {
			update := UpdateInfo{
				Name:     attrValue(e.Attr, "name"),
				Severity: attrValue(e.Attr, "severity"),
				Kind:     attrValue(e.Attr, "kind"),
			}
			if update.Severity == "security" {
				update.IsSecurity = true
				up.Updates++
			} else {
				update.IsSecurity = false
				up.Security++
			}

			if update.Summary, err = nextElementValue(d); err != nil {
				logger.Printf("%v", err)
				continue
			}
			if update.Description, err = nextElementValue(d); err != nil {
				logger.Printf("%v", err)
				continue
			}
			up.List = append(up.List, update)
		}
	}

	if err != nil {
		up.Error = err.Error()
	}
	return up
}

func nextElementValue(d *xml.Decoder) (string, error) {
	// Skipping a blank token (which mixes eol, whitespace, etc., so we are
	// safe).
	if _, err := d.Token(); err != nil {
		return "", err
	}

	// After skipping a blank token, now we are sure that the next element is a
	// start element, which will contain the desired value.

	token, err := d.Token()
	if err != nil {
		return "", err
	}

	start, ok := token.(xml.StartElement)
	if !ok {
		return "", errors.New("not a start element")
	}

	var info string
	err = d.DecodeElement(&info, &start)
	return info, err
}

// attrValue returns the attribute value for the case-insensitive key
// `name', or the empty string if nothing is found.
func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}
