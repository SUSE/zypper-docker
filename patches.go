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

// zypper-docker list-patches [flags] <image>
func listPatchesCmd(ctx *cli.Context) {
	log.SetPrefix("[list-patches] ")
	// It's safe to ignore the returned error because we set to false the
	// `getError` parameter of this function.
	listPatches(ctx.Args().First(), ctx)
}

// zypper-docker list-patches-container [flags] <container>
func listPatchesContainerCmd(ctx *cli.Context) {
	log.SetPrefix("[list-patches-container] ")
	containerId := ctx.Args().First()
	if container, err := checkContainerRunning(containerId); err != nil {
		logAndFatalf("%v.\n", err)
	} else {
		listPatches(container.Image, ctx)
	}
}

// listParches calls the `zypper lp` command for the given image and the given
// arguments.
func listPatches(image string, ctx *cli.Context) {
	// It's safe to ignore the returned error because we set to false the
	// `getError` parameter of this function.
	_ = runStreamedCommand(
		image,
		cmdWithFlags("lp", ctx, []string{}, []string{}), false)
}

// zypper-docker patch [flags] image
func patchCmd(ctx *cli.Context) {
	log.SetPrefix("[patch] ")
	if len(ctx.Args()) != 2 {
		logAndFatalf("Wrong invocation: expected 2 arguments, %d given.\n", len(ctx.Args()))
		return
	}

	img := ctx.Args()[0]
	repo, tag, err := parseImageName(ctx.Args()[1])
	if err != nil {
		logAndFatalf("%v\n", err)
		return
	}
	if err = preventImageOverwrite(repo, tag); err != nil {
		logAndFatalf("%v\n", err)
		return
	}

	comment := ctx.String("message")
	author := ctx.String("author")

	boolFlags := []string{"l", "auto-agree-with-licenses", "no-recommends",
		"replacefiles"}
	toIgnore := []string{"author", "message"}

	cmd := fmt.Sprintf(
		"zypper ref && zypper -n %v",
		cmdWithFlags("patch", ctx, boolFlags, toIgnore))
	newImgId, err := runCommandAndCommitToImage(
		img,
		repo,
		tag,
		cmd,
		comment,
		author)
	if err != nil {
		logAndFatalf("Could not commit to the new image: %v.\n", err)
		return
	}

	logAndPrintf("%s:%s successfully created", repo, tag)

	cache := getCacheFile()
	if err := cache.updateCacheAfterUpdate(img, newImgId); err != nil {
		log.Println("Cannot add image details to zypper-docker cache")
		log.Println("This will break the \"zypper-docker ps\" feature")
		log.Println(err)
		exitWithCode(1)
	}
}
