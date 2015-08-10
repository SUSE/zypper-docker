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
	"fmt"

	"github.com/codegangsta/cli"
)

// Returns the version string
func version() string {
	const (
		major = 0
		minor = 1
	)
	return fmt.Sprintf("%v.%v", major, minor)
}

// It returns an application with all the flags and subcommands already in
// place.
func newApp() *cli.App {
	app := cli.NewApp()

	app.Name = "zypper-docker"
	app.Usage = "Patching Docker images safely"
	app.Version = version()

	app.CommandNotFound = func(context *cli.Context, cmd string) {
		fmt.Printf("Incorrect usage: command '%v' does not exist.\n\n", cmd)
		cli.ShowAppHelp(context)
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "n, non-interactive",
			Usage: "Switches to non-interactive mode",
		},
		cli.BoolFlag{
			Name:  "no-gpg-checks",
			Usage: "Ignore GPG check failures and continue",
		},
		cli.BoolFlag{
			Name:  "gpg-auto-import-keys",
			Usage: "If new repository signing key is found, do not ask what to do; trust and import it automatically",
		},
		cli.BoolFlag{
			Name:  "f, force",
			Usage: "Ignore all the local caches",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "images",
			Usage:  "List all the images based on either OpenSUSE or SLES",
			Action: imagesCmd,
		},
		{
			Name:    "list-updates",
			Aliases: []string{"lu"},
			Usage:   "List all the available updates",
			Action:  listUpdatesCmd,
		},
		{
			Name:    "update",
			Aliases: []string{"up"},
			Usage:   "Install the available updates",
			Action:  updateCmd,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "l, auto-agree-with-licenses",
					Usage: "Automatically say yes to third party license confirmation prompt. By using this option, you choose to agree with licenses of all third-party software this command will install.",
				},
				cli.BoolFlag{
					Name:  "no-recommends",
					Usage: "By default, zypper installs also packages recommended by the requested ones. This option causes the recommended packages to be ignored and only the required ones to be installed.",
				},
				cli.StringFlag{
					Name:   "author",
					EnvVar: "USERNAME",
					Usage:  "Commit author to associate with the new layer (e.g., \"John Doe <john.doe@example.com>\")",
				},
				cli.StringFlag{
					Name:  "message",
					Value: "[zypper-docker] update",
					Usage: "Commit message to associated with the new layer",
				},
			},
		},
		{
			Name:    "list-patches",
			Aliases: []string{"lp"},
			Usage:   "List all the available patches",
			Action:  listPatchesCmd,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "b, bugzilla",
					Value: "",
					Usage: "List available needed patches for all Bugzilla issues, or issues whose number matches the given string.",
				},
				cli.StringFlag{
					Name:  "cve",
					Value: "",
					Usage: "List available needed patches for all CVE issues, or issues whose number matches the given string.",
				},
				cli.StringFlag{
					Name:  "date",
					Value: "",
					Usage: "List patches issued up to, but not including, the specified date (YYYY-MM-DD).",
				},
				cli.StringFlag{
					Name:  "issues",
					Value: "",
					Usage: "Look for issues whose number, summary, or description matches the specified string.",
				},
				cli.StringFlag{
					Name:  "g, category",
					Value: "",
					Usage: "List only patches with this category.",
				},
			},
		},
		{
			Name:   "patch",
			Usage:  "Install the available patches",
			Action: patchCmd,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "b, bugzilla",
					Value: "",
					Usage: "List available needed patches for all Bugzilla issues, or issues whose number matches the given string.",
				},
				cli.StringFlag{
					Name:  "cve",
					Value: "",
					Usage: "List available needed patches for all CVE issues, or issues whose number matches the given string.",
				},
				cli.StringFlag{
					Name:  "date",
					Value: "",
					Usage: "List patches issued up to, but not including, the specified date (YYYY-MM-DD).",
				},
				cli.StringFlag{
					Name:  "issues",
					Value: "",
					Usage: "Look for issues whose number, summary, or description matches the specified string.",
				},
				cli.BoolFlag{
					Name:  "l, auto-agree-with-licenses",
					Usage: "Automatically say yes to third party license confirmation prompt. By using this option, you choose to agree with licenses of all third-party software this command will install.",
				},
				cli.BoolFlag{
					Name:  "no-recommends",
					Usage: "By default, zypper installs also packages recommended by the requested ones. This option causes the recommended packages to be ignored and only the required ones to be installed.",
				},
				cli.StringFlag{
					Name:   "author",
					EnvVar: "USERNAME",
					Usage:  "Commit author to associate with the new layer (e.g., \"John Doe <john.doe@example.com>\")",
				},
				cli.StringFlag{
					Name:  "message",
					Value: "[zypper-docker] patch",
					Usage: "Commit message to associated with the new layer",
				},
			},
		},
		{
			Name:    "patch-check",
			Aliases: []string{"pchk"},
			Usage:   "Check for patches (to do)",
			Action:  patchCheckCmd,
		},
		{
			Name:   "ps",
			Usage:  "List all the containers that are outdated",
			Action: psCmd,
		},
	}
	return app
}
