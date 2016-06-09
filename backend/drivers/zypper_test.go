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
	"flag"
	"os"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/mssola/capture"
)

func mockApp() *cli.App {
	app := cli.NewApp()
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
			Usage: "Show all the logged messages on stdout (ignored when using the serve command)",
		},
		cli.StringSliceFlag{
			Name:  "add-host",
			Usage: "Add a custom host-to-IP mapping (host:ip)",
		},
	}
	app.Commands = []cli.Command{{Name: "test", Action: func(ctx *cli.Context) {
		Initialize(ctx)
	}}}
	return app
}

func TestFormatZypperCommand(t *testing.T) {
	cmd := formatZypperCommand("ref", "up")
	if cmd != "zypper ref && zypper up" {
		t.Fatalf("Wrong command '%v', expected 'zypper ref && zypper up'", cmd)
	}

	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		cliContext = nil
	}()
	os.Args = []string{"exe", "--add-host", "host:ip", "test"}

	app := mockApp()
	capture.All(func() { app.RunAndExitOnError() })

	cmd = formatZypperCommand("ref", "up")
	expected := "zypper --non-interactive ref && zypper --non-interactive up"
	if cmd != expected {
		t.Fatalf("Wrong command '%v', expected '%v'", cmd, expected)
	}
}

func TestIsZypperExitCodeSevere(t *testing.T) {
	notSevereExitCodes := []int{
		zypperExitOK,
		zypperExitInfRebootNeeded,
		zypperExitInfUpdateNeeded,
		zypperExitInfSecUpdateNeeded,
		zypperExitInfRestartNeeded,
		zypperExitOnSignal,
	}

	z := &Zypper{}
	for _, code := range notSevereExitCodes {
		if is, _ := z.IsExitCodeSevere(code); is {
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
		zypperExitIndCapNotFound,
		127,
	}

	for _, code := range severeExitCodes {
		if is, _ := z.IsExitCodeSevere(code); !is {
			t.Fatalf("Exit code %v should be considered a severe error", code)
		}
	}
}

func TestCmdWithFlags(t *testing.T) {
	cmd := cli.Command{
		Name:  "lp",
		Usage: "List all the images based on either OpenSUSE or SLES",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "bugzilla",
				Value: "",
				Usage: "List available needed patches for all Bugzilla issues, or issues whose number matches the given string.",
			},
			cli.StringFlag{
				Name:  "cve",
				Value: "",
				Usage: "List available needed patches for all CVE issues, or issues whose number matches the given string.",
			},
			cli.StringFlag{
				Name:  "issues",
				Value: "",
				Usage: "doc",
			},
			cli.StringFlag{
				Name:  "to-ignore",
				Value: "",
				Usage: "Should not be forwarded.",
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
				Name:  "explode",
				Usage: "Boom",
			},
		},
	}

	set := flag.NewFlagSet("test", 0)
	set.String("bugzilla", "bugzilla_value", "doc")
	set.String("cve", "cve_value", "doc")
	set.String("to-ignore", "to_ignore_value", "doc")
	set.String("issues", "", "doc")
	set.Bool("l", true, "doc")
	set.Bool("no-recommends", true, "doc")
	err := set.Parse([]string{
		"--bugzilla", "bugzilla_value",
		"--cve", "cve_value",
		"--to-ignore", "to_ignore_value",
		"--issues", "",
		"-l",
		"--no-recommends",
	})
	if err != nil {
		t.Fatal("cannot parse flags")
	}

	ctx := cli.NewContext(nil, set, nil)
	ctx.Command = cmd

	boolFlags := []string{"l", "auto-agree-with-licenses", "no-recommends"}
	toIgnore := []string{"to-ignore"}
	actual := cmdWithFlags("cmd", ctx, boolFlags, toIgnore)
	expected := "cmd --bugzilla=bugzilla_value --cve=cve_value --issues  -l --no-recommends"

	if expected != actual {
		t.Fatal("Wrong command")
	}
}

func TestHostConfig(t *testing.T) {
	hc := GetHostConfig()
	if len(hc.ExtraHosts) != 0 {
		t.Fatalf("Wrong number of extra hosts: %v; Expected: 1", len(hc.ExtraHosts))
	}

	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		cliContext = nil
	}()
	os.Args = []string{"exe", "--add-host", "host:ip", "test"}

	app := mockApp()
	capture.All(func() { app.RunAndExitOnError() })

	hc = GetHostConfig()
	if len(hc.ExtraHosts) != 1 {
		t.Fatalf("Wrong number of extra hosts: %v; Expected: 1", len(hc.ExtraHosts))
	}
	if hc.ExtraHosts[0] != "host:ip" {
		t.Fatalf("Did not expect %v", hc.ExtraHosts[0])
	}
}
