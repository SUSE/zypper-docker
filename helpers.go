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
	"regexp"
	"strings"

	"github.com/SUSE/dockerclient"
	"github.com/codegangsta/cli"
)

var specialFlags = []string{
	"--bugzilla",
	"--cve",
	"--issues",
}

// Decorate the given command so it adds some extra information to it before
// executing it.
func getCmd(name string, f func(ctx *cli.Context)) func(*cli.Context) {
	log.SetPrefix("[" + name + "] ")

	return func(ctx *cli.Context) {
		setupLogger(ctx)
		currentContext = ctx
		f(ctx)
	}
}

func arrayIncludeString(arr []string, s string) bool {
	for _, i := range arr {
		if i == s {
			return true
		}
	}
	return false
}

// It appends the set flags with the given command.
// `boolFlags` is a list of strings containing the names of the boolean
// command line options. These have to be handled in a slightly different
// way because zypper expects `--boolflag` instead of `--boolflag true`. Also
// boolean flags with a false value are ignored because zypper set all the
// undefined bool flags to false by default.
// `toIgnore` contains a list of flag names to not be passed to the final
//  command, this is useful to prevent zypper-docker only parameters to be
// forwarded to zypper (eg: `--author` or `--message`).
func cmdWithFlags(cmd string, ctx *cli.Context, boolFlags, toIgnore []string) string {
	for _, name := range ctx.FlagNames() {
		if arrayIncludeString(toIgnore, name) {
			continue
		}

		if value := ctx.String(name); ctx.IsSet(name) {
			var dash string
			if len(name) == 1 {
				dash = "-"
			} else {
				dash = "--"
			}

			if arrayIncludeString(boolFlags, name) {
				cmd += fmt.Sprintf(" %v%s", dash, name)
			} else {
				if arrayIncludeString(specialFlags, fmt.Sprintf("%v%s", dash, name)) && value != "" {
					cmd += fmt.Sprintf(" %v%s=%s", dash, name, value)
				} else {
					cmd += fmt.Sprintf(" %v%s %s", dash, name, value)
				}
			}
		}
	}

	return cmd
}

// This function clears a list of args (like the one provided by `os.Args`)
// to match with some special cases of zypper.
// For example:
//   zypper lp --bugzilla
// In the above case --buzilla acts as a boolean flag, while with:
//   zypper lp --bugzilla=123
// acts like a string flag.
// We have to differentiate between invocations with and without the "=".
// When the "=" is not found we have to artificially inject an empty string
// to avoid the next parameter to be considered the flag value.
func fixArgsForZypper(args []string) []string {
	sanitizedArgs := []string{}
	skip := false

	for pos, arg := range args {
		if skip {
			skip = false
			continue
		}

		special := false
		for _, specialFlag := range specialFlags {
			if specialFlag == arg {
				sanitizedArgs = append(sanitizedArgs, arg)
				sanitizedArgs = append(sanitizedArgs, "")
				special = true

				if len(args) > (pos+1) && args[pos+1] == "" {
					skip = true
				}
				break
			} else if strings.Contains(arg, specialFlag+"=") {
				argAndValue := strings.SplitN(arg, "=", 2)

				sanitizedArgs = append(sanitizedArgs, argAndValue[0])
				sanitizedArgs = append(sanitizedArgs, argAndValue[1])
				special = true
				break
			}
		}
		if !special {
			sanitizedArgs = append(sanitizedArgs, arg)
		}
	}

	return sanitizedArgs
}

// Given a Docker image name it returns the repository and the tag composing it
// Returns the repository and the tag strings.
// Examples:
//   * suse/sles11sp3:1.0.0 -> repo is suse/sles11sp3, tag is 1.0.0
//   * suse/sles11sp3 -> repo is suse/sles11sp3, tag is latest
func parseImageName(name string) (string, string, error) {
	var repo, tag string

	re := regexp.MustCompile("([a-z0-9\\._-]+(:[a-z0-9\\._-]+)?\\z)")
	match := re.FindAllString(name, -1)
	if match == nil || len(match) == 0 {
		return "", "", fmt.Errorf("Cannot parse image name")
	}
	repoAndTag := strings.SplitN(match[0], ":", 2)
	repo = repoAndTag[0]

	if len(repoAndTag) != 2 {
		tag = "latest"
	} else {
		tag = repoAndTag[1]
	}

	return repo, tag, nil
}

// Exists with error if the image identified by repo and tag already exists
// Returns an error when the image already exists or something went wrong.
func preventImageOverwrite(repo, tag string) error {
	imageExists, err := checkImageExists(repo, tag)
	if err != nil {
		return fmt.Errorf("Cannot proceed safely: %v.", err)
	}
	if imageExists {
		return fmt.Errorf("Cannot overwrite an existing image. Please use a different repository/tag.")
	}
	return nil
}

func getImageId(name string) (string, error) {
	client := getDockerClient()

	repo, tag, err := parseImageName(name)
	if err != nil {
		return "", err
	}
	if tag == "latest" && !strings.Contains(name, tag) {
		name = name + ":" + tag
	}

	images, err := client.ListImages(true, repo, &dockerclient.ListFilter{})
	if err != nil {
		return "", err
	}

	if len(images) == 0 {
		return "", fmt.Errorf("Cannot find image %s", name)
	}
	for _, image := range images {
		if arrayIncludeString(image.RepoTags, name) {
			return image.Id, nil
		}
	}

	return "", fmt.Errorf("Cannot find image %s", name)
}

// commandFunc represents a function that accepts an image ID and the CLI
// context. This is used in the commandInContainer function.
type commandFunc func(string, *cli.Context)

// commandInContainer executes the given commandFunc for the image in which the
// given container is based on. The container ID is extracted from the first
// argument as given in ctx.
func commandInContainer(f commandFunc, ctx *cli.Context) {
	containerID := ctx.Args().First()

	if container, err := checkContainerRunning(containerID); err != nil {
		logAndFatalf("%v.\n", err)
	} else {
		f(container.Image, ctx)
	}
}

// updatePatchCmd executes an update/patch command depending on the argument
// zypperCmd.
func updatePatchCmd(zypperCmd string, ctx *cli.Context) {
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
		cmdWithFlags(zypperCmd, ctx, boolFlags, toIgnore))
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
