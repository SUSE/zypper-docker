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

// zypper-docker list-updates [flags] <image>
func listUpdatesCmd(ctx *cli.Context) {
	imageID := ctx.Args().First()
	err := listUpdates(imageID, ctx)
	exitOnError(imageID, "zypper lu", err)
}

// zypper-docker list-updates-container [flags] <container>
func listUpdatesContainerCmd(ctx *cli.Context) {
	imageID, err := commandInContainer(listUpdates, ctx)
	exitOnError(imageID, "zypper lu", err)
}

// listUpdates lists all the updates available for the given image with the
// given arguments.
func listUpdates(image string, ctx *cli.Context) error {
	err := runStreamedCommand(image, "lu", true)
	return err
}

// zypper-docker update [flags] image new-image
func updateCmd(ctx *cli.Context) {
	updatePatchCmd("up", ctx)
}
