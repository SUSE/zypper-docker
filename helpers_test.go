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
	"flag"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
)

func TestParseImageName(t *testing.T) {
	// map with name as value and a string list with two enteries (repo and tag)
	// as value
	data := make(map[string][]string)
	data["opensuse:13.2"] = []string{"opensuse", "13.2"}
	data["opensuse"] = []string{"opensuse", "latest"}

	for name, expected := range data {
		repo, tag := parseImageName(name)
		if repo != expected[0] {
			t.Fatalf("repository %s is different from the expected %s", repo, expected[0])
		}
		if tag != expected[1] {
			t.Fatalf("tag %s is different from the expected %s", tag, expected[1])
		}
	}
}

func TestPreventImageOverwriteImageCheckImageFailure(t *testing.T) {
	dockerClient = &mockClient{listFail: true}

	err := preventImageOverwrite("opensuse", "13.2")

	if err == nil {
		t.Fatalf("Expected error")
	}
	if !strings.Contains(err.Error(), "List Failed") {
		t.Fatal("Wrong error message")
	}
}

func TestPreventImageOverwriteImageExists(t *testing.T) {
	dockerClient = &mockClient{}

	err := preventImageOverwrite("opensuse", "13.2")

	if err == nil {
		t.Fatalf("Expected error")
	}
	if !strings.Contains(err.Error(), "Cannot overwrite an existing image.") {
		t.Fatal("Wrong error message")
	}
}

func TestCmdWithFlags(t *testing.T) {
	cmd := cli.Command{
		Name:  "lp",
		Usage: "List all the images based on either OpenSUSE or SLES",
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
				Usage: "Install patches issued up to, but not including, the specified date (YYYY-MM-DD).",
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
	set.String("b", "bugzilla_value", "doc")
	set.String("cve", "cve_value", "doc")
	set.String("to-ignore", "to_ignore_value", "doc")
	set.String("date", "", "doc")
	set.Bool("l", true, "doc")
	set.Bool("no-recommends", true, "doc")
	err := set.Parse([]string{
		"-b", "bugzilla_value",
		"--cve", "cve_value",
		"--to-ignore", "to_ignore_value",
		"--date", "",
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
	expected := "cmd -b bugzilla_value --cve cve_value --date  -l --no-recommends"

	if expected != actual {
		t.Fatal("Wrong command")
	}
}

func TestSanitizeStringSpecialFlagUsedAsBool(t *testing.T) {
	input := []string{"zypper-docker", "lp", "--bugzilla", "image"}
	expected := []string{"zypper-docker", "lp", "--bugzilla", "", "image"}

	actual := fixArgsForZypper(input)
	if err := compareStringSlices(actual, expected); err != nil {
		t.Fatalf("Wrong sanitization %v", err)
	}
}

func TestSanitizeStringSpecialFlagUsedAsStringWithEmptyValue(t *testing.T) {
	// this can be achieved by calling zypper-docker lp --bugzilla "" image
	input := []string{"zypper-docker", "lp", "--bugzilla", "", "image"}
	expected := []string{"zypper-docker", "lp", "--bugzilla", "", "image"}

	actual := fixArgsForZypper(input)
	if err := compareStringSlices(actual, expected); err != nil {
		t.Fatalf("Wrong sanitization %v", err)
	}
}

func TestSanitizeStringSpecialFlagUsedAsString(t *testing.T) {
	input := []string{"zypper-docker", "lp", "--bugzilla=bnc123", "image"}
	expected := []string{"zypper-docker", "lp", "--bugzilla", "bnc123", "image"}

	actual := fixArgsForZypper(input)
	if err := compareStringSlices(actual, expected); err != nil {
		t.Fatalf("Wrong sanitization %v", err)
	}
}
