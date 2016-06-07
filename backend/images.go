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

import (
	"fmt"
	"os"

	"github.com/docker/docker/api/client/formatter"
	"github.com/docker/engine-api/types"
)

// FetchImages fetches all the info from images. Set `force` to true if the
// cache has to be re-updated before performing the operation. Returns the
// fetched images on success, otherwise it returns the proper error.
func FetchImages(force bool) ([]types.Image, error) {
	if force {
		getCacheFile().reset()
	}

	client := GetDockerClient()
	return client.ImageList(types.ImageListOptions{All: false})
}

// PrintImages prints the given images as expected.
func PrintImages(images []types.Image) {
	suseImages := make([]types.Image, 0, len(images))
	cache := getCacheFile()
	counter := 0

	for _, img := range images {
		select {
		case <-KillChannel:
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

// ImageExists looks for a docker image defined by repo:tag and it returns true
// if the image already exists.
func ImageExists(repo, tag string) (bool, error) {
	client := GetDockerClient()

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
