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
	"log"

	"github.com/codegangsta/cli"
)

// zypper-docker patch-check [flags] <image>
func patchCheckCmd(ctx *cli.Context) {
	err := runStreamedCommand(ctx.Args().First(), "pchk", true)
	if err == nil {
		return
	}

	switch err.(type) {
	case dockerError:
		// According to zypper's documentation:
		// 	100 - There are patches available for installation.
		// 	101 - There are security patches available for installation.
		// Therefore, if the returned exit code is one of the specified above,
		// then we do nothing.
		de := err.(dockerError)
		if de.ExitCode == 100 || de.ExitCode == 101 {
			return
		}
	}
	log.Printf("Error: %v\n", err)
	exitWithCode(1)
}
