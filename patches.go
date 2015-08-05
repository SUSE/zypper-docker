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
	"os"
	"strings"

	"github.com/codegangsta/cli"
)

// It appends the set flags with the given command.
func cmdWithFlags(cmd string, ctx *cli.Context) string {
	for _, name := range ctx.FlagNames() {
		if value := ctx.String(name); value != "" {
			var dash string
			if len(name) == 1 {
				dash = "-"
			} else {
				dash = "--"
			}

			cmd += fmt.Sprintf(" %v%s %s", dash, name, value)
		}
	}
	return cmd
}

// zypper-docker list-patches [flags] <image>
func listPatchesCmd(ctx *cli.Context) {
	// It's safe to ignore the returned error because we set to false the
	// `getError` parameter of this function.
	_ = runStreamedCommand(ctx.Args().First(), cmdWithFlags("lp", ctx), false)
}

// zypper-docker patch [flags] image
func patchCmd(ctx *cli.Context) {
	if len(ctx.Args()) != 2 {
		log.Println("Wrong invocation")
		exitWithCode(1)
		return
	}

	img := ctx.Args()[0]
	target := strings.SplitN(ctx.Args()[1], ":", 2)
	var repo, tag string
	repo = target[0]
	if len(target) != 2 {
		tag = "latest"
	} else {
		tag = target[1]
	}

	imageExists, err := checkImageExists(repo, tag)
	if err != nil {
		log.Println(err)
		exitWithCode(1)
		return
	}
	if imageExists {
		log.Println("Cannot overwrite an existing image. Please use a different repository/tag.")
		exitWithCode(1)
		return
	}

	cmd := fmt.Sprintf("zypper ref && zypper -n %v", cmdWithFlags("patch", ctx))
	id, err := runCommandInContainer(img, []string{"/bin/sh", "-c", cmd}, true)
	if err != nil {
		log.Println(err)
		exitWithCode(1)
	}

	comment := "[zypper-docker] apply patches"
	author := os.Getenv("USER")
	err = commitContainerToImage(id, repo, tag, comment, author)

	// always remove the container
	removeContainer(id)

	if err != nil {
		log.Println(err)
		exitWithCode(1)
	}
	log.Printf("%s:%s successfully created", repo, tag)
}
