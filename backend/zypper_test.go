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

import "testing"

/*
func TestFormatZypperCommand(t *testing.T) {
	cmd := formatZypperCommand("ref", "up")
	if cmd != "zypper ref && zypper up" {
		t.Fatalf("Wrong command '%v', expected 'zypper ref && zypper up'", cmd)
	}

	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		CLIContext = nil
	}()
	os.Args = []string{"exe", "--add-host", "host:ip", "test"}

	app := newApp()
	app.Commands = []cli.Command{{Name: "test", Action: getCmd("test", func(*cli.Context) {})}}
	capture.All(func() { app.RunAndExitOnError() })

	cmd = formatZypperCommand("ref", "up")
	expected := "zypper --non-interactive ref && zypper --non-interactive up"
	if cmd != expected {
		t.Fatalf("Wrong command '%v', expected '%v'", cmd, expected)
	}
}
*/

func TestIsZypperExitCodeSevere(t *testing.T) {
	notSevereExitCodes := []int{
		zypperExitOK,
		zypperExitInfRebootNeeded,
		zypperExitInfUpdateNeeded,
		zypperExitInfSecUpdateNeeded,
		zypperExitInfRestartNeeded,
		zypperExitOnSignal,
	}

	for _, code := range notSevereExitCodes {
		if isZypperExitCodeSevere(code) {
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
		if !isZypperExitCodeSevere(code) {
			t.Fatalf("Exit code %v should be considered a severe error", code)
		}
	}
}
