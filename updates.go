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
	"github.com/codegangsta/cli"
	"log"
)

func listUpdatesCmd(ctx *cli.Context) {
	imageName := ctx.Args().First()

	if imageName == "" {
		log.Printf("Error: no image name specified\n")
		exitWithCode(1)
	}

	id, err := runCommandInContainer(
		imageName,
		[]string{"/bin/sh", "-c", "zypper ref && zypper lu"},
		true)
	removeContainer(id)

	if err != nil {
		log.Printf("Error: %s\n", err)
		exitWithCode(1)
	}
}
