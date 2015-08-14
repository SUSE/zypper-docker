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

// zypper-docker list-patches-container [flags] <container>
func listPatchesContainerCmd(ctx *cli.Context) {
	containerId := ctx.Args().First()
	container, err := checkContainerRunning(containerId)
	if err != nil {
		log.Println(err)
		exitWithCode(1)
	}

	listPatches(container.Image, ctx)
}

// zypper-docker list-patches [flags] <image>
func listPatchesCmd(ctx *cli.Context) {
	// It's safe to ignore the returned error because we set to false the
	// `getError` parameter of this function.
	listPatches(ctx.Args().First(), ctx)
}

// zypper-docker list-patches [flags] <image>
func listPatches(image string, ctx *cli.Context) {
	// It's safe to ignore the returned error because we set to false the
	// `getError` parameter of this function.
	_ = runStreamedCommand(
		image,
		cmdWithFlags("lp", ctx, []string{}, []string{}), false)
}

// zypper-docker patch [flags] image
func patchCmd(ctx *cli.Context) {
	if len(ctx.Args()) != 2 {
		log.Println("Wrong invocation")
		exitWithCode(1)
		return
	}

	img := ctx.Args()[0]
	repo, tag := parseImageName(ctx.Args()[1])
	if err := preventImageOverwrite(repo, tag); err != nil {
		log.Println(err)
		exitWithCode(1)
	}

	comment := ctx.String("message")
	author := ctx.String("author")

	boolFlags := []string{"l", "auto-agree-with-licenses", "no-recommends",
		"replacefiles"}
	toIgnore := []string{"author", "message"}

	cmd := fmt.Sprintf(
		"zypper ref && zypper -n %v",
		cmdWithFlags("patch", ctx, boolFlags, toIgnore))
	err := runCommandAndCommitToImage(
		img,
		repo,
		tag,
		cmd,
		comment,
		author)
	if err != nil {
		log.Println(err)
		exitWithCode(1)
	}

	log.Printf("%s:%s successfully created", repo, tag)
}
