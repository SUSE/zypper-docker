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

// TODO: remove that stinking help subcommad (but keeping it on the options)

// It returns an application with all the flags and subcommands already in
// place.
func newApp() *cli.App {
	app := cli.NewApp()

	app.Name = "zypper-docker"
	app.Usage = "Patching Docker images safely"
	app.Version = version()

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "n, -non-interactive",
			Usage: "Switches to non-interactive mode",
		},
		cli.BoolFlag{
			Name:  "-no-gpg-checks",
			Usage: "Ignore GPG check failures and continue",
		},
		cli.BoolFlag{
			Name:  "-gpg-auto-import-keys",
			Usage: "If new repository signing key is found, do not ask what to do; trust and import it automatically",
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
			Name:    "list-patches",
			Aliases: []string{"lp"},
			Usage:   "List all the available patches",
			Action:  listPatchescmd,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "b, bugzilla",
					Value: "",
				},
				cli.StringFlag{
					Name:  "-cve",
					Value: "",
				},
				cli.StringFlag{
					Name:  "-date",
					Value: "",
				},
				cli.StringFlag{
					Name:  "-issues",
					Value: "",
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
				},
				cli.StringFlag{
					Name:  "-cve",
					Value: "",
				},
				cli.StringFlag{
					Name:  "-date",
					Value: "",
				},
				cli.StringFlag{
					Name:  "-issues",
					Value: "",
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
