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

	"github.com/codegangsta/cli"
)

// runStreamedCommand is a convenient wrapper of the `runCommandInContainer`
// for functions that just need to run a command on streaming without the
// burden of removing the resulting container, etc.
//
// The image has to be provided, otherwise this function will exit with 1 as
// the status code and it will log that no image was provided. The given
// command will be executed as "zypper ref && zypper <command>".
//
// If getError is set to false, then this function will always return nil.
// Otherwise, it will return the error as given by the `runCommandInContainer`
// function.
func runStreamedCommand(img, cmd string, getError bool) error {
	if img == "" {
		log.Println("Error: no image name specified.")
		exitWithCode(1)
		return nil
	}

	cmd = fmt.Sprintf("zypper ref && zypper %v", cmd)
	id, err := runCommandInContainer(img, []string{"/bin/sh", "-c", cmd}, true)
	removeContainer(id)

	if getError {
		return err
	}
	if err != nil {
		log.Printf("Error: %s\n", err)
		exitWithCode(1)
	}
	return nil
}

func listUpdatesCmd(ctx *cli.Context) {
	runStreamedCommand(ctx.Args().First(), "lu", false)
}
