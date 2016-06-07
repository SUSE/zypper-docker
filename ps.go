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
	"github.com/codegangsta/cli"
)

// zypper-docker ps
func psCmd(ctx *cli.Context) {
	state, err := backend.ListContainers(true)

	if err != nil {
		logAndFatalf("error: %v\n", err)
		return
	} else if state == nil {
		fmt.Println("There are no running containers to analyze.")
		return
	}

	if len(state.updated) > 0 {
		fmt.Println("Running containers whose images have been updated:")
		for _, container := range matches {
			fmt.Printf("  - %s [%s]\n", container.ID, container.Image)
		}
		fmt.Println("It is recommended to stop the container and start a new instance based on the new image created with zypper-docker")
	}

	if len(state.ignored) > 0 {
		if len(state.updated) > 0 {
			fmt.Printf("\n")
		}
		fmt.Println("The following containers have been ignored:")

		for _, cs := range state.ignored {
			fmt.Printf("  - %s [%s] (reason: %s)\n", cs.container.ID, cs.container.Image, cs.message)
		}
	}

	if len(state.unknown) > 0 {
		if len(state.updated) > 0 || len(state.ignored) > 0 {
			fmt.Printf("\n")
		}
		fmt.Println("The following containers have an unknown state or could not be analyzed:")

		for _, cs := range state.unknown {
			fmt.Printf("  - %s [%s] (reason: %s)\n", cs.container.ID, cs.container.Image, cs.message)
		}

		fmt.Println("Use either the \"list-patches-container\" or the \"list-updates-container\" commands to inspect them.")
	}
}
