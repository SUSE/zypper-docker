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
	"log"
	"os/user"

	"github.com/codegangsta/cli"
)

// A pointer to the current context.
var currentContext *cli.Context

// Returns the version string
func version() string {
	const (
		major = 1
		minor = 2
		patch = 0
	)
	return fmt.Sprintf("%v.%v.%v", major, minor, patch)
}

func defaultCommitAuthor() string {
	current, err := user.Current()
	if err != nil {
		log.Printf("Cannot determine current user: %s", err)
		return ""
	}

	if current.Name != "" {
		return current.Name
	}
	return current.Username
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
		cli.BoolFlag{
			Name:  "d, debug",
			Usage: "Show all the logged messages on stdout",
		},
		cli.StringSliceFlag{
			Name:  "add-host",
			Usage: "Add a custom host-to-IP mapping (host:ip)",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:      "images",
			Usage:     "List all the images based on either OpenSUSE or SLES",
			Action:    getCmd("images", imagesCmd),
			ArgsUsage: " ",
		},
		{
			Name:    "list-updates",
			Aliases: []string{"lu"},
			Usage:   "List all the available updates",
			Action:  getCmd("list-updates", listUpdatesCmd),
			UsageText: `zypper-docker list-updates <image>
   zypper-docker lu <image>

Where <image> is the name of the openSUSE/SUSE Linux Enterprise image to use.
If the tag has not been provided, then "latest" is the one that will be used.`,
		},
		{
			Name:    "list-updates-container",
			Aliases: []string{"luc"},
			Usage:   "List all the available updates for the given container",
			Action:  getCmd("list-updates-container", listUpdatesContainerCmd),
			UsageText: `zypper-docker list-updates-container <container-id>
   zypper-docker luc <container-id>

Where <container-id> is either the container ID or the name of the container
to be used.`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "base",
					Usage: "Analyze the base image of the container for updates.",
				},
			},
		},
		{
			Name:    "update",
			Aliases: []string{"up"},
			Usage:   "Install the available updates",
			Action:  getCmd("update", updateCmd),
			UsageText: `zypper-docker update [command options] <image> <new-image>
   zypper-docker up [command options] <image> <new-image>

Where <image> is the name of the openSUSE/SUSE Linux Enterprise image to
update. Since zypper-docker does not overwrite images, <new-image> is the name
of the image that will be created on this operation. This new image will be the
same as the old one plus the applied updates.

If the tag has not been provided on either <image> or <new-image>, then
"latest" is the one that will be used.`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "l, auto-agree-with-licenses",
					Usage: "Automatically say yes to third party license confirmation prompt. By using this option, you choose to agree with licenses of all third-party software this command will install.",
				},
				cli.BoolFlag{
					Name:  "no-recommends",
					Usage: "By default, zypper installs also packages recommended by the requested ones. This option causes the recommended packages to be ignored and only the required ones to be installed.",
				},
				cli.BoolFlag{
					Name:  "replacefiles",
					Usage: "Install the packages even if they replace files from other, already installed, packages. Default is to treat file conflicts as an error.",
				},
				cli.StringFlag{
					Name:  "author",
					Value: defaultCommitAuthor(),
					Usage: "Commit author to associate with the new layer (e.g., \"John Doe <john.doe@example.com>\")",
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
			Action:  getCmd("list-patches", listPatchesCmd),
			UsageText: `zypper-docker list-patches [command options] <image>
   zypper-docker lp [command options] <image>

Where <image> is the name of the openSUSE/SUSE Linux Enterprise image to use.
If the tag has not been provided, then "latest" is the one that will be used.`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "bugzilla",
					Value: "",
					Usage: "List available needed patches for all Bugzilla issues, or issues whose number matches the given string (--bugzilla=#).",
				},
				cli.StringFlag{
					Name:  "cve",
					Value: "",
					Usage: "List available needed patches for all CVE issues, or issues whose number matches the given string (--cve=#).",
				},
				cli.StringFlag{
					Name:  "date",
					Value: "",
					Usage: "List patches issued up to, but not including, the specified date (YYYY-MM-DD).",
				},
				cli.StringFlag{
					Name:  "issues",
					Value: "",
					Usage: "Look for issues whose number, summary, or description matches the specified string (--issue=string).",
				},
				cli.StringFlag{
					Name:  "g, category",
					Value: "",
					Usage: "List only patches with this category.",
				},
				cli.StringFlag{
					Name:  "severity",
					Value: "",
					Usage: "List only patches with this severity.",
				},
			},
		},
		{
			Name:    "list-patches-container",
			Aliases: []string{"lpc"},
			Usage:   "List all the available patches for the given container",
			Action:  getCmd("list-patches-container", listPatchesContainerCmd),
			UsageText: `zypper-docker list-patches-container [command options] <container-id>
   zypper-docker lpc [command options] <container-id>

Where <container-id> is either the container ID or the name of the container
to be used.`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "base",
					Usage: "Analyse the base image of the container for patches.",
				},
				cli.StringFlag{
					Name:  "b, bugzilla",
					Value: "",
					Usage: "List available needed patches for all Bugzilla issues, or issues whose number matches the given string (--bugzilla=#).",
				},
				cli.StringFlag{
					Name:  "cve",
					Value: "",
					Usage: "List available needed patches for all CVE issues, or issues whose number matches the given string (--cve=#).",
				},
				cli.StringFlag{
					Name:  "date",
					Value: "",
					Usage: "List patches issued up to, but not including, the specified date (YYYY-MM-DD).",
				},
				cli.StringFlag{
					Name:  "issues",
					Value: "",
					Usage: "Look for issues whose number, summary, or description matches the specified string (--issue=string).",
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
			Action: getCmd("patch", patchCmd),
			ArgsUsage: `<image> <new-image>

Where <image> is the name of the openSUSE/SUSE Linux Enterprise image to
patch. Since zypper-docker does not overwrite images, <new-image> is the name
of the image that will be created on this operation. This new image will be the
same as the old one plus the applied patches.

If the tag has not been provided on either <image> or <new-image>, then
"latest" is the one that will be used.`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "bugzilla",
					Value: "",
					Usage: "Install available needed patches for all Bugzilla issues, or issues whose number matches the given string (--bugzilla=#).",
				},
				cli.StringFlag{
					Name:  "cve",
					Value: "",
					Usage: "Install available needed patches for all CVE issues, or issues whose number matches the given string (--cve=#).",
				},
				cli.StringFlag{
					Name:  "date",
					Value: "",
					Usage: "Install patches issued up to, but not including, the specified date (YYYY-MM-DD).",
				},
				cli.StringFlag{
					Name:  "g, category",
					Value: "",
					Usage: "Install only patches with this category.",
				},
				cli.BoolFlag{
					Name:  "l, auto-agree-with-licenses",
					Usage: "Automatically say yes to third party license confirmation prompt. By using this option, you choose to agree with licenses of all third-party software this command will install.",
				},
				cli.BoolFlag{
					Name:  "no-recommends",
					Usage: "By default, zypper installs also packages recommended by the requested ones. This option causes the recommended packages to be ignored and only the required ones to be installed.",
				},
				cli.BoolFlag{
					Name:  "replacefiles",
					Usage: "Install the packages even if they replace files from other, already installed, packages. Default is to treat file conflicts as an error.",
				},
				cli.StringFlag{
					Name:  "author",
					Value: defaultCommitAuthor(),
					Usage: "Commit author to associate with the new layer (e.g., \"John Doe <john.doe@example.com>\")",
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
			Usage:   "Check for patches",
			Action:  getCmd("patch-check", patchCheckCmd),
			UsageText: `zypper-docker patch-check <image>
   zypper-docker pchk <image>

Where <image> is the name of the openSUSE/SUSE Linux Enterprise image to use.
If the tag has not been provided, then "latest" is the one that will be used.`,
		},
		{
			Name:    "patch-check-container",
			Aliases: []string{"pchkc"},
			Usage:   "Check for patches available for the given container",
			Action:  getCmd("patch-check-container", patchCheckContainerCmd),
			UsageText: `zypper-docker patch-check-container <container-id>
   zypper-docker pchkc <container-id>

Where <container-id> is either the container ID or the name of the container
to be used.`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "base",
					Usage: "Execute a patch-check on the base image of the container.",
				},
			},
		},
		{
			Name:      "ps",
			Usage:     "List all the containers that are outdated",
			Action:    getCmd("ps", psCmd),
			ArgsUsage: " ",
		},
	}
	return app
}
