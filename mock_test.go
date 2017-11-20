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
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
)

type mockClient struct {
	createFail         bool
	createWarnings     bool
	removeFail         bool
	startFail          bool
	waitSleep          time.Duration
	waitFail           bool
	commandFail        bool
	commandExit        int
	listFail           bool
	listEmpty          bool
	listReturnOneImage bool
	logFail            bool
	lastCmd            []string
	killFail           bool
	commitFail         bool
	inspectFail        bool
	zypperBadVersion   bool
	zypperGoodVersion  bool
	suppressLog        bool
}

func (mc *mockClient) ImageList(options types.ImageListOptions) ([]types.Image, error) {
	if mc.listFail {
		return nil, errors.New("List Failed")
	}
	if mc.listEmpty {
		return nil, nil
	}

	// Let's return some more or less realistic images.
	if mc.listReturnOneImage {
		return []types.Image{
			types.Image{
				ID:          "2",
				ParentID:    "0",       // Not used
				Size:        254515796, // 254.5 MB
				VirtualSize: 254515796,
				RepoTags:    []string{"opensuse:13.2"},
				Created:     time.Now().UnixNano(),
			},
		}, nil
	}

	return []types.Image{
		types.Image{
			ID:          "1",
			ParentID:    "0",       // Not used
			Size:        254515796, // 254.5 MB
			VirtualSize: 254515796,
			RepoTags:    []string{"opensuse:latest", "opensuse:tag"},
			Created:     time.Now().UnixNano(),
		},
		types.Image{
			ID:          "2",
			ParentID:    "0",       // Not used
			Size:        254515796, // 254.5 MB
			VirtualSize: 254515796,
			RepoTags:    []string{"opensuse:13.2"},
			Created:     time.Now().UnixNano(),
		},
		types.Image{
			ID:          "3",
			ParentID:    "0",       // Not used
			Size:        254515796, // 254.5 MB
			VirtualSize: 254515796,
			RepoTags:    []string{"ubuntu:latest"},
			Created:     time.Now().UnixNano(),
		},
		types.Image{
			ID:          "4",
			ParentID:    "0",       // Not used
			Size:        254515796, // 254.5 MB
			VirtualSize: 254515796,
			RepoTags:    []string{}, // Invalid image
			Created:     time.Now().UnixNano(),
		},
		types.Image{
			ID:          "5",
			ParentID:    "0",       // Not used
			Size:        254515796, // 254.5 MB
			VirtualSize: 254515796,
			RepoTags:    []string{"busybox:latest"}, // Invalid image
			Created:     time.Now().UnixNano(),
		},
	}, nil
}

func (mc *mockClient) ContainerCreate(config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (types.ContainerCreateResponse, error) {
	var (
		warnings []string
		name     string
	)

	if mc.createFail {
		return types.ContainerCreateResponse{}, errors.New("Create failed")
	}
	if mc.createWarnings {
		warnings = []string{"warning"}
	}

	// workaround: TestHasZypperMinVersion{Fail,Success} breaks TestSetupLoggerHome
	// due to this string being printed to the log file.
	if !mc.suppressLog {
		name = fmt.Sprintf("zypper-docker-private-%s", config.Image)
	}

	mc.lastCmd = config.Cmd.Slice()

	return types.ContainerCreateResponse{ID: name, Warnings: warnings}, nil
}

func (mc *mockClient) ContainerStart(id string) error {
	if mc.startFail {
		return errors.New("Start failed")
	}
	if id == "zypper-docker-private-3" {
		// Ubuntu doesn't have zypper: fail.
		return errors.New("Start failed")
	}
	return nil
}

func (mc *mockClient) ContainerRemove(options types.ContainerRemoveOptions) error {
	if mc.removeFail {
		return errors.New("Remove failed")
	}
	if !mc.suppressLog {
		log.Printf("Removed container %v", options.ContainerID)
	}
	return nil
}

func (mc *mockClient) ContainerWait(containerID string) (int, error) {
	time.Sleep(mc.waitSleep)
	if mc.waitFail {
		return -1, errors.New("Wait failed")
	}
	if mc.commandFail {
		// If commandExit was not specified, just exit with 1.
		if mc.commandExit == 0 {
			mc.commandExit = 1
		}
		return mc.commandExit, nil
	}
	return 0, nil
}

func (mc *mockClient) ContainerLogs(options types.ContainerLogsOptions) (io.ReadCloser, error) {
	var err error

	if mc.logFail {
		return nil, fmt.Errorf("Fake log failure")
	}
	cb := &closingBuffer{bytes.NewBuffer([]byte{})}
	if mc.zypperBadVersion {
		_, err = cb.WriteString("Unknown option '--severity'\n")
	} else if mc.zypperGoodVersion {
		_, err = cb.WriteString("Missing argument for --severity\n")
	} else {
		_, err = cb.WriteString("streaming buffer initialized\n")
	}
	return cb, err
}

func (mc *mockClient) ContainerKill(id, signal string) error {
	if mc.killFail {
		return fmt.Errorf("Fake failure while killing container")
	}
	return nil
}

func (mc *mockClient) ContainerCommit(options types.ContainerCommitOptions) (types.ContainerCommitResponse, error) {
	if mc.commitFail {
		return types.ContainerCommitResponse{ID: ""}, fmt.Errorf("Fake failure while committing container")
	}
	return types.ContainerCommitResponse{ID: "fake image ID"}, nil
}

func (mc *mockClient) ContainerList(options types.ContainerListOptions) ([]types.Container, error) {
	if mc.listFail {
		return []types.Container{},
			fmt.Errorf("Fake failure while listing containers")
	}

	if mc.listEmpty {
		return []types.Container{}, nil
	}

	return []types.Container{
		types.Container{
			ID:    "35ae93c88cf8ab18da63bb2ad2dfd2399d745f292a344625fbb65892b7c25a01",
			Names: []string{"/suse"},
			Image: "opensuse:13.2",
		},
		types.Container{
			ID:    "2",
			Names: []string{"/not_suse"},
			Image: "busybox:latest",
		},
		types.Container{
			ID:    "3",
			Names: []string{"/ubuntu"},
			Image: "ubuntu:latest",
		},
		types.Container{
			ID:    "4",
			Names: []string{"/unknown_image"},
			Image: "foo",
		},
	}, nil
}

func (mc *mockClient) ContainerResize(options types.ResizeOptions) error {
	// Do nothing
	return nil
}

func (mc *mockClient) ImageInspectWithRaw(imageID string, getSize bool) (types.ImageInspect, []byte, error) {
	if mc.inspectFail {
		return types.ImageInspect{}, []byte{}, errors.New("inspect fail")
	}
	return types.ImageInspect{Config: &container.Config{Image: "1"}}, []byte{}, nil
}

func (mc *mockClient) ImageRemove(options types.ImageRemoveOptions) ([]types.ImageDelete, error) {
	if mc.removeFail {
		return []types.ImageDelete{}, errors.New("remove fail")
	}
	return nil, nil
}

func (mc *mockClient) ContainerInspect (containerID string) (types.ContainerJSON, error) {
	if mc.inspectFail {
		return types.ContainerJSON{}, errors.New("inspect fail")
	}
	return types.ContainerJSON{Config: &container.Config{Image: "1"}}, nil
}
