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
	"strings"

	"github.com/SUSE/zypper-docker/backend/drivers"
	"github.com/SUSE/zypper-docker/logger"
	"github.com/SUSE/zypper-docker/utils"
	"github.com/docker/distribution/reference"
	"github.com/docker/engine-api/types"
)

// getImageID fetches the image ID for the given image.
func getImageID(name string) (string, error) {
	client := getDockerClient()

	repo, tag, err := parseImageName(name)
	if err != nil {
		return "", err
	}
	if tag == "latest" && !strings.Contains(name, tag) {
		name = name + ":" + tag
	}

	images, err := client.ImageList(types.ImageListOptions{MatchName: repo, All: false})
	if err != nil {
		return "", err
	}

	if len(images) == 0 {
		return "", fmt.Errorf("cannot find image %s", name)
	}
	for _, image := range images {
		if utils.ArrayIncludeString(image.RepoTags, name) {
			return image.ID, nil
		}
	}

	return "", fmt.Errorf("cannot find image %s", name)
}

// Given a Docker image name it returns the repository and the tag composing it
// Returns the repository and the tag strings.
// Examples:
//   * suse/sles11sp3:1.0.0 -> repo is suse/sles11sp3, tag is 1.0.0
//   * suse/sles11sp3 -> repo is suse/sles11sp3, tag is latest
func parseImageName(name string) (string, string, error) {
	// TODO (mssola): The reference package has the Parse function that does
	// what we want. However, the returned object does not contain the tag
	// always. This leads into a grammar conflict from a client point of view.
	// For this reason, instead of using reference.Parse we use the regexpes
	// provided by the reference package (that Parse is using anyways).

	matches := reference.ReferenceRegexp.FindStringSubmatch(name)
	if matches == nil {
		return "", "",
			fmt.Errorf("could not parse '%s': %v", name, reference.ErrReferenceInvalidFormat)
	}
	if matches[1] == "" {
		return "", "", reference.ErrNameEmpty
	}
	if len(matches[1]) > reference.NameTotalLengthMax {
		return "", "", fmt.Errorf("could not parse '%s': %v", name, reference.ErrNameTooLong)
	}
	if matches[2] == "" {
		matches[2] = "latest"
	}
	return matches[1], matches[2], nil
}

// Exists with error if the image identified by repo and tag already exists
// Returns an error when the image already exists or something went wrong.
func preventImageOverwrite(repo, tag string) error {
	imageExists, err := ImageExists(repo, tag)

	if err != nil {
		return fmt.Errorf("cannot proceed safely: %v", err)
	}
	if imageExists {
		return fmt.Errorf("cannot overwrite an existing image. Please use a different repository/tag")
	}
	return nil
}

// humanizeCommandError tries to print an explicit and useful message for a
// failing command inside of a docker container (based on the given image).
func humanizeCommandError(cmd, image string, err error) {
	var reason string

	switch err.(type) {
	case dockerError:
		de := err.(dockerError)
		if de.exitCode == 127 {
			reason = "command not found"
		} else {
			reason = err.Error()
		}
	default:
		reason = err.Error()
	}

	msg := "could not execute command '%s' successfully in image '%s': %v"
	logger.Printf(msg, cmd, image, reason)
}

// Looks for the specified command inside of a Docker image.
// The given image string is just the ID of said image.
// It returns true if the command was successful, false otherwise.
func checkCommandInImage(img, cmd string) bool {
	containerID, err := startContainer(img, []string{cmd}, false, nil)

	defer removeContainer(containerID)

	if err != nil {
		humanizeCommandError(cmd, img, err)
		return false
	}
	return true
}

// Spawns a container from the specified image, runs the specified command inside
// of it and commits the results to a new image.
// The name of the new image is specified via target_repo and target_tag.
// The container is always deleted.
// If something goes wrong an error message is returned.
// Returns the ID of the new image on success.
func runCommandAndCommitToImage(img, targetRepo, targetTag, cmd, comment, author string) (string, error) {
	containerID, err := runCommandInContainer(img, []string{cmd}, os.Stdout)
	if err != nil {
		switch err.(type) {
		case dockerError:
			de := err.(dockerError)
			severe, err := drivers.Current().IsExitCodeSevere(de.exitCode)
			if severe && err == nil {
				return "", err
			}
		default:
			return "", err
		}
	}

	imageID, err := commitContainerToImage(img, containerID, targetRepo, targetTag, comment, author)

	// always remove the container
	removeContainer(containerID)

	return imageID, err
}
