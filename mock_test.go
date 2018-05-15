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
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
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

func (mc *mockClient) ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	if mc.listFail {
		return nil, errors.New("List Failed")
	}
	if mc.listEmpty {
		return nil, nil
	}

	// Let's return some more or less realistic images.
	if mc.listReturnOneImage {
		return []types.ImageSummary{
			types.ImageSummary{
				ID:          "2",
				ParentID:    "0",       // Not used
				Size:        254515796, // 254.5 MB
				VirtualSize: 254515796,
				RepoTags:    []string{"opensuse:13.2"},
				Created:     time.Now().UnixNano(),
			},
		}, nil
	}

	return []types.ImageSummary{
		types.ImageSummary{
			ID:          "1",
			ParentID:    "0",       // Not used
			Size:        254515796, // 254.5 MB
			VirtualSize: 254515796,
			RepoTags:    []string{"opensuse:latest", "opensuse:tag"},
			Created:     time.Now().UnixNano(),
		},
		types.ImageSummary{
			ID:          "2",
			ParentID:    "0",       // Not used
			Size:        254515796, // 254.5 MB
			VirtualSize: 254515796,
			RepoTags:    []string{"opensuse:13.2"},
			Created:     time.Now().UnixNano(),
		},
		types.ImageSummary{
			ID:          "3",
			ParentID:    "0",       // Not used
			Size:        254515796, // 254.5 MB
			VirtualSize: 254515796,
			RepoTags:    []string{"ubuntu:latest"},
			Created:     time.Now().UnixNano(),
		},
		types.ImageSummary{
			ID:          "4",
			ParentID:    "0",       // Not used
			Size:        254515796, // 254.5 MB
			VirtualSize: 254515796,
			RepoTags:    []string{}, // Invalid image
			Created:     time.Now().UnixNano(),
		},
		types.ImageSummary{
			ID:          "5",
			ParentID:    "0",       // Not used
			Size:        254515796, // 254.5 MB
			VirtualSize: 254515796,
			RepoTags:    []string{"busybox:latest"}, // Invalid image
			Created:     time.Now().UnixNano(),
		},
	}, nil
}

func (mc *mockClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error) {
	var (
		warnings []string
		name     string
	)

	if mc.createFail {
		return container.ContainerCreateCreatedBody{}, errors.New("Create failed")
	}
	if mc.createWarnings {
		warnings = []string{"warning"}
	}

	// workaround: TestHasZypperMinVersion{Fail,Success} breaks TestSetupLoggerHome
	// due to this string being printed to the log file.
	if !mc.suppressLog {
		name = fmt.Sprintf("zypper-docker-private-%s", config.Image)
	}

	mc.lastCmd = config.Cmd

	return container.ContainerCreateCreatedBody{ID: name, Warnings: warnings}, nil
}

func (mc *mockClient) ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error {
	if mc.startFail {
		return errors.New("Start failed")
	}
	if containerID == "zypper-docker-private-3" {
		// Ubuntu doesn't have zypper: fail.
		return errors.New("Start failed")
	}
	return nil
}

func (mc *mockClient) ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error {
	if mc.removeFail {
		return errors.New("Remove failed")
	}
	if !mc.suppressLog {
		log.Printf("Removed container %v", containerID)
	}
	return nil
}

func (mc *mockClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error) {
	resultC := make(chan container.ContainerWaitOKBody)
	errC := make(chan error)

	go func() {
		time.Sleep(mc.waitSleep)
		if mc.waitFail {
			errC <- errors.New("Wait failed")
			resultC <- container.ContainerWaitOKBody{StatusCode: -1}
			return
		}
		if mc.commandFail {
			// If commandExit was not specified, just exit with 1.
			if mc.commandExit == 0 {
				resultC <- container.ContainerWaitOKBody{StatusCode: 1}
				errC <- nil
				return
			}
			resultC <- container.ContainerWaitOKBody{StatusCode: (int64)(mc.commandExit)}
			errC <- nil
			return
		}
		resultC <- container.ContainerWaitOKBody{StatusCode: 0}
		errC <- nil
	}()
	return resultC, errC
}

func (mc *mockClient) ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
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

func (mc *mockClient) ContainerKill(ctx context.Context, containerID, signal string) error {
	if mc.killFail {
		return fmt.Errorf("Fake failure while killing container")
	}
	return nil
}

func (mc *mockClient) ContainerCommit(ctx context.Context, container string, options types.ContainerCommitOptions) (types.IDResponse, error) {
	if mc.commitFail {
		return types.IDResponse{ID: ""}, fmt.Errorf("Fake failure while committing container")
	}
	return types.IDResponse{ID: "fake image ID"}, nil
}

func (mc *mockClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
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

func (mc *mockClient) ContainerResize(ctx context.Context, containerID string, options types.ResizeOptions) error {
	// Do nothing
	return nil
}

func (mc *mockClient) ImageInspectWithRaw(ctx context.Context, imageID string) (types.ImageInspect, []byte, error) {
	if mc.inspectFail {
		return types.ImageInspect{}, []byte{}, errors.New("inspect fail")
	}
	return types.ImageInspect{Config: &container.Config{Image: "1"}}, []byte{}, nil
}

func (mc *mockClient) ImageRemove(ctx context.Context, image string, options types.ImageRemoveOptions) ([]types.ImageDeleteResponseItem, error) {
	if mc.removeFail {
		return []types.ImageDeleteResponseItem{}, errors.New("remove fail")
	}
	return nil, nil
}

func (mc *mockClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	if mc.inspectFail {
		return types.ContainerJSON{}, errors.New("inspect fail")
	}
	return types.ContainerJSON{Config: &container.Config{Image: "1"}}, nil
}
