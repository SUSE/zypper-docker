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
	"os"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/api/client/formatter"
	"github.com/docker/engine-api/types"
)

// Print all the images based on SUSE. It will print in a format that is as
// close to the `docker` command as possible.
func printImages(images []types.Image) {
	suseImages := make([]types.Image, 0, len(images))
	cache := getCacheFile()
	counter := 0

	for _, img := range images {
		select {
		case <-killChannel:
			return
		default:
			fmt.Printf("Inspecting image %d/%d\r", (counter + 1), len(images))
			if cache.isSUSE(img.ID) {
				suseImages = append(suseImages, img)
			}
		}
		counter++
	}

	imagesCtx := formatter.ImageContext{
		Context: formatter.Context{
			Output: os.Stdout,
			Format: "table",
			Quiet:  false,
			Trunc:  true,
		},
		Digest: false,
		Images: suseImages,
	}

	imagesCtx.Write()
	cache.flush()
}

// The images command prints all the images that are based on SUSE.
func imagesCmd(ctx *cli.Context) {
	client := getDockerClient()

	// On "force", just cleanup the cache.
	if ctx.GlobalBool("force") {
		cd := getCacheFile()
		cd.reset()
	}

	if imgs, err := client.ImageList(types.ImageListOptions{All: false}); err != nil {
		logAndFatalf("Cannot proceed safely: %v.", err)
	} else {
		printImages(imgs)
		exitWithCode(0)
	}
}

// Looks for a docker image defined by repo:tag
// Returns true if the image already exists, false otherwise
func checkImageExists(repo, tag string) (bool, error) {
	client := getDockerClient()

	images, err := client.ImageList(types.ImageListOptions{
		MatchName: repo,
		All:       false,
	})
	if err != nil {
		return false, err
	}
	if len(images) == 0 {
		return false, nil
	}

	ref := fmt.Sprintf("%s:%s", repo, tag)
	for _, image := range images {
		for _, t := range image.RepoTags {
			if ref == t {
				return true, nil
			}
		}
	}
	return false, nil
}
