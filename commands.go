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

	"github.com/SUSE/zypper-docker/backend"
	"github.com/SUSE/zypper-docker/logger"
	"github.com/SUSE/zypper-docker/utils"
	"github.com/codegangsta/cli"
)

// zypper-docker images
func imagesCmd(ctx *cli.Context) {
	if imgs, err := backend.FetchImages(ctx.GlobalBool("force")); err != nil {
		logger.Fatalf("Cannot proceed safely: %v", err)
	} else {
		backend.PrintImages(imgs)
	}
}

// zypper-docker list-updates [flags] <image>
func listUpdatesCmd(ctx *cli.Context) {
	listUpdates(ctx.Args().First(), ctx)
}

// zypper-docker list-updates-container [flags] <container>
func listUpdatesContainerCmd(ctx *cli.Context) {
	commandInContainer(listUpdates, ctx)
}

// listUpdates lists all the updates available for the given image with the
// given arguments.
func listUpdates(image string, ctx *cli.Context) {
	if err := backend.ListUpdates(backend.General, image); err != nil {
		logger.Fatalf("Failed to list updates: %v", err)
	}
}

// zypper-docker update [flags] image new-image
func updateCmd(ctx *cli.Context) {
	updatePatchCmd(backend.General, ctx)
}

// zypper-docker list-patches [flags] <image>
func listPatchesCmd(ctx *cli.Context) {
	// It's safe to ignore the returned error because we set to false the
	// `getError` parameter of this function.
	listPatches(ctx.Args().First(), ctx)
}

// zypper-docker list-patches-container [flags] <container>
func listPatchesContainerCmd(ctx *cli.Context) {
	commandInContainer(listPatches, ctx)
}

// listParches calls the `zypper lp` command for the given image and the given
// arguments.
func listPatches(image string, ctx *cli.Context) {
	if image == "" {
		logger.Fatalf("error: no image name specified")
		return
	}

	if severity := ctx.String("severity"); severity != "" {
		if err := backend.SeveritySupported(image); err != nil {
			logger.Fatalf("error: %v", err)
			return
		}
	}

	if err := backend.ListUpdates(backend.Security, image); err != nil {
		logger.Fatalf("Failed to list security updates: %v", err)
	}
}

// zypper-docker patch [flags] image
func patchCmd(ctx *cli.Context) {
	updatePatchCmd(backend.Security, ctx)
}

// zypper-docker patch-check [flags] <image>
func patchCheckCmd(ctx *cli.Context) {
	patchCheck(ctx.Args().First(), ctx)
}

// zypper-docker patch-check-container [flags] <image>
func patchCheckContainerCmd(ctx *cli.Context) {
	commandInContainer(patchCheck, ctx)
}

// patchCheck calls the `zypper pchk` command for the given image and the given
// arguments.
func patchCheck(image string, ctx *cli.Context) {
	updates, security, err := backend.HasPatches(image)
	if err != nil {
		logger.Fatalf("%v", err)
		return
	}

	// From zypper's documentation:
	// 	- 100: there are patches available for installation.
	// 	- 101: there are security patches available for installation.
	if updates {
		utils.ExitWithCode(100)
	} else if security {
		utils.ExitWithCode(101)
	}
	utils.ExitWithCode(1)
}

// zypper-docker ps
func psCmd(ctx *cli.Context) {
	state, err := backend.ListContainers(true)

	if err != nil {
		logger.Fatalf("error: %v", err)
		return
	} else if state == nil {
		fmt.Println("There are no running containers to analyze.")
		return
	}

	if len(state.Updated) > 0 {
		fmt.Println("Running containers whose images have been updated:")
		for _, container := range state.Updated {
			fmt.Printf("  - %s [%s]\n", container.ID, container.Image)
		}
		fmt.Println("It is recommended to stop the container and start a new instance based on the new image created with zypper-docker")
	}

	if len(state.Ignored) > 0 {
		if len(state.Updated) > 0 {
			fmt.Printf("\n")
		}
		fmt.Println("The following containers have been ignored:")

		for _, cs := range state.Ignored {
			fmt.Printf("  - %s [%s] (reason: %s)\n", cs.Container.ID, cs.Container.Image, cs.Message)
		}
	}

	if len(state.Unknown) > 0 {
		if len(state.Updated) > 0 || len(state.Ignored) > 0 {
			fmt.Printf("\n")
		}
		fmt.Println("The following containers have an unknown state or could not be analyzed:")

		for _, cs := range state.Unknown {
			fmt.Printf("  - %s [%s] (reason: %s)\n", cs.Container.ID, cs.Container.Image, cs.Message)
		}

		fmt.Println("Use either the \"list-patches-container\" or the \"list-updates-container\" commands to inspect them.")
	}
}
