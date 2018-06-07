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

import "github.com/codegangsta/cli"

// zypper-docker patch-check [flags] <image>
func patchCheckCmd(ctx *cli.Context) {
	imageID := ctx.Args().First()
	err := patchCheck(imageID, ctx)
	exitOnError(imageID, "zypper pchk", err)
}

// zypper-docker patch-check-container [flags] <image>
func patchCheckContainerCmd(ctx *cli.Context) {
	imageID, err := commandInContainer(patchCheck, ctx)
	exitOnError(imageID, "zypper pchk", err)
}

// patchCheck calls the `zypper pchk` command for the given image and the given
// arguments.
func patchCheck(image string, ctx *cli.Context) error {
	err := runStreamedCommand(
		image,
		cmdWithFlags("pchk", ctx, []string{}, []string{"base"}), true)
	return err
}
