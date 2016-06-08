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
	"strings"

	"github.com/SUSE/zypper-docker/backend"
	"github.com/SUSE/zypper-docker/backend/drivers"
	"github.com/SUSE/zypper-docker/logger"
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
	return func(ctx *cli.Context) {
		// Initialize the logger and the drivers for this command.
		logger.Initialize(name, ctx.GlobalBool("debug"))
		drivers.Initialize(ctx)

		// And finally call the given command.
		f(ctx)
	}
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
// TODO: move
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

// commandFunc represents a function that accepts an image ID. This is used in
// the `commandInContainer` function.
type commandFunc func(string)

// commandInContainer executes the given commandFunc for the image in which the
// given container is based on. The container ID is extracted from the first
// argument as given in ctx.
func commandInContainer(f commandFunc, ctx *cli.Context) {
	containerID := ctx.Args().First()

	if container, err := backend.CheckContainer(containerID); err != nil {
		logger.Fatalf("%v", err)
	} else {
		f(container.Image)
	}
}

// updatePatchCmd executes an update/patch command depending on the argument
// zypperCmd.
func updatePatchCmd(cmd backend.UpdateKind, ctx *cli.Context) {
	if len(ctx.Args()) != 2 {
		logger.Fatalf("Wrong invocation: expected 2 arguments, %d given", len(ctx.Args()))
		return
	}

	img, dst := ctx.Args()[0], ctx.Args()[1]
	comment, author := ctx.String("message"), ctx.String("author")

	if repo, tag, err := backend.PerformUpdate(cmd, img, dst, comment, author); err != nil {
		logger.Fatalf("Could not update image: %v", err)
	} else {
		logger.Printf("%s:%s successfully created", repo, tag)
	}
}

// supportsSeverityFlag checks whether or not zypper's `list-patches` command
// supports the `--severity` flag in the specified image.
func supportsSeverityFlag(image string) (bool, error) {
	/*
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
	*/
	return false, nil
}
