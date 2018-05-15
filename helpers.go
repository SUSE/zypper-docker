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
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/go-units"
)

var specialFlags = []string{
	"--bugzilla",
	"--cve",
	"--issues",
}

// Decorate the given command so it adds some extra information to it before
// executing it.
func getCmd(name string, f func(ctx *cli.Context)) func(*cli.Context) {
	return func(ctx *cli.Context) {
		log.SetPrefix("[" + name + "] ")
		setupLogger(ctx)
		currentContext = ctx
		f(ctx)
	}
}

// Returns a string containing the global flags being used.
func globalFlags() string {
	if currentContext == nil {
		return ""
	}

	res := "--non-interactive "
	flags := []string{"no-gpg-checks", "gpg-auto-import-keys"}

	for _, v := range flags {
		if currentContext.GlobalBool(v) {
			res = res + "--" + v + " "
		}
	}
	return res
}

// Concatenate the given zypper commands, while adding the global flags
// currently in place.
func formatZypperCommand(cmds ...string) string {
	flags := globalFlags()

	for k, v := range cmds {
		cmds[k] = "zypper " + flags + v
	}
	return strings.Join(cmds, " && ")
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
	// TODO (mssola): The reference package has the Parse function that does
	// what we want. However, the returned object does not contain the tag
	// always. This leads into a grammar conflict from a client point of view.
	// For this reason, instead of using reference.Parse we use the regexpes
	// provided by the reference package (that Parse is using anyways).

	matches := reference.ReferenceRegexp.FindStringSubmatch(name)
	if matches == nil {
		return "", "",
			fmt.Errorf("Could not parse '%s': %v", name, reference.ErrReferenceInvalidFormat)
	}
	if matches[1] == "" {
		return "", "", reference.ErrNameEmpty
	}
	if len(matches[1]) > reference.NameTotalLengthMax {
		return "", "", fmt.Errorf("Could not parse '%s': %v", name, reference.ErrNameTooLong)
	}
	if matches[2] == "" {
		matches[2] = "latest"
	}
	return matches[1], matches[2], nil
}

// Exists with error if the image identified by repo and tag already exists
// Returns an error when the image already exists or something went wrong.
func preventImageOverwrite(repo, tag string) error {
	imageExists, err := checkImageExists(repo, tag)

	if err != nil {
		return fmt.Errorf("Cannot proceed safely: %v", err)
	}
	if imageExists {
		return fmt.Errorf("Cannot overwrite an existing image. Please use a different repository/tag")
	}
	return nil
}

func getImageID(name string) (string, error) {
	client := getDockerClient()

	repo, tag, err := parseImageName(name)
	if err != nil {
		return "", err
	}
	if tag == "latest" && !strings.Contains(name, tag) {
		name = name + ":" + tag
	}

	images, err := client.ImageList(context.Background(), types.ImageListOptions{All: false, Filters: filters.NewArgs(filters.Arg("reference", repo))})
	if err != nil {
		return "", err
	}

	if len(images) == 0 {
		return "", fmt.Errorf("Cannot find image %s", name)
	}
	for _, image := range images {
		if arrayIncludeString(image.RepoTags, name) {
			return image.ID, nil
		}
	}

	return "", fmt.Errorf("Cannot find image %s", name)
}

// commandFunc represents a function that accepts an image ID and the CLI
// context. This is used in the commandInContainer function.
type commandFunc func(string, *cli.Context)

// commandInContainer extracts the containerID from the given ctx. The function
// first checks whether the container exists. If the container is running the
// given commandFunc is executed on the image the container is based on. If the
// --force flag is set the container is committed to a new image instead. In
// case the container is not running it is committed to an image first.
// Afterward commandFunc is executed on the new image.
func commandInContainer(f commandFunc, ctx *cli.Context) {
	containerID := ctx.Args().First()
	// check if the container exists
	if !checkContainerExists(containerID) {
		logAndFatalf("Container %s does not exist.\n", containerID)
	}
	// check whether the container is running
	if container, err := checkContainerRunning(containerID); err != nil {
		logAndPrintf("Checking stopped container %s ...\n", containerID)
		// execute commitAndExecute on the stopped container
		err = commitAndExecute(f, ctx, containerID)
		if err != nil {
			logAndFatalf("%v.\n", err)
		}
	} else {
		// if the force flag is set commitAndExecute is executed on the running container
		if ctx.GlobalBool("force") {
			logAndPrintf("WARNING: Force flag used. Manually installed packages will be analyzed as well.\n")
			err = commitAndExecute(f, ctx, containerID)
			if err != nil {
				logAndFatalf("%v.\n", err)
			}
		} else {
			// directly call commandFunc on the image the container is based on
			logAndPrintf("WARNING: Only the source image from this container will be inspected. Manually installed packages won't be taken into account.\n")
			f(container.Image, ctx)
		}
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

	cmd := formatZypperCommand("ref", fmt.Sprintf("-n %v", zypperCmd), "clean -a")
	cmd = cmdWithFlags(cmd, ctx, boolFlags, toIgnore)
	newImgID, err := runCommandAndCommitToImage(
		img,
		repo,
		tag,
		cmd,
		comment,
		author)
	if err != nil {
		logAndFatalf("Could not commit to the new image: %v\n", err)
		return
	}

	logAndPrintf("%s:%s successfully created\n", repo, tag)

	cache := getCacheFile()
	if err := cache.updateCacheAfterUpdate(img, newImgID); err != nil {
		log.Println("Cannot add image details to zypper-docker cache")
		log.Println("This will break the \"zypper-docker ps\" feature")
		log.Println(err)
	}
}

// joinAsArray joins the given array of commands so it's compatible to what is
// expected from a dockerfile syntax.
func joinAsArray(cmds []string, emptyArray bool) string {
	if emptyArray && len(cmds) == 0 {
		return ""
	}

	str := "["
	for i, v := range cmds {
		str += "\"" + v + "\""
		if i < len(cmds)-1 {
			str += ", "
		}
	}
	return str + "]"
}

// supportsSeverityFlag checks whether or not zypper's `list-patches` command
// supports the `--severity` flag in the specified image.
func supportsSeverityFlag(image string) (bool, error) {
	buf := bytes.NewBuffer([]byte{})
	id, err := runCommandInContainer(image, []string{"zypper lp --severity"}, buf)
	defer removeContainer(id)

	if strings.Contains(buf.String(), "Missing argument for --severity") {
		return true, nil
	}
	if strings.Contains(buf.String(), "Unknown option '--severity'") {
		return false, nil
	}
	return false, err
}

// removeDuplicates removes duplicate entries from an array of strings. Should
// the resulting array be empty, it does not return nil but an empty array.
func removeDuplicates(elements []string) []string {
	seen := make(map[string]bool)
	var res []string

	for _, v := range elements {
		if seen[v] {
			continue
		} else {
			seen[v] = true
			res = append(res, v)
		}
	}

	// make sure not to return nil
	if res == nil {
		return []string{}
	}

	return res
}

// format and print given images to match `docker images` output
func formatAndPrint(images []types.ImageSummary) {
	writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	fmt.Fprintln(writer, "REPOSITORY\tTAG\tIMAGE ID\tCREATED\tSIZE")

	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			repo := strings.Split(repoTag, ":")[0]
			tag := strings.Split(repoTag, ":")[1]
			truncID := stringid.TruncateID(img.ID)
			createdSince := units.HumanDuration(time.Now().UTC().Sub(time.Unix(img.Created, 0))) + " ago"
			hSize := units.HumanSize(float64(img.Size))

			fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n", repo, tag, truncID, createdSince, hSize)
		}
	}
	writer.Flush()
}
