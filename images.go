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
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/docker/pkg/units"
	"github.com/samalba/dockerclient"
)

// Returns whether the given ID matches an image that is based on SUSE.
func isSUSE(id string) bool {
	// TODO: (mssola) cache it ?
	return runCommandInContainer(id, []string{"zypper"})
}

// Returns a string that contains a description of how much has passed since
// the given timestamp until now.
func timeAgo(ts int64) string {
	created, now := time.Unix(ts, 0), time.Now().UTC()
	return units.HumanDuration(now.Sub(created))
}

// Print all the images based on SUSE. It will print in a format that is as
// close to the `docker` command as possible.
func printImages(imgs []*dockerclient.Image) {
	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	fmt.Fprintf(w, "REPOSITORY\tTAG\tIMAGE ID\tCREATED\tVIRTUAL SIZE\n")

	for _, img := range imgs {
		if isSUSE(img.Id) {
			if len(img.RepoTags) < 1 {
				continue
			}

			id := stringid.TruncateID(img.Id)
			repoTag := strings.SplitN(img.RepoTags[0], ":", 2)
			size := units.HumanSize(float64(img.VirtualSize))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s ago\t%s\n", repoTag[0], repoTag[1], id,
				timeAgo(img.Created), size)
		}
	}
	w.Flush()
}

// The images command prints all the images that are based on SUSE.
func imagesCmd(ctx *cli.Context) {
	client := getDockerClient()

	if imgs, err := client.ListImages(false); err != nil {
		log.Printf("%v\n", err)
	} else {
		printImages(imgs)
	}
}
