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

	"github.com/mssola/dockerclient"
)

var exitInvocations int

type mockClient struct {
	createFail  bool
	removeFail  bool
	startFail   bool
	waitSleep   time.Duration
	waitFail    bool
	commandFail bool
	listFail    bool
	listEmpty   bool
	logFail     bool
}

func (mc *mockClient) ListImages(all bool) ([]*dockerclient.Image, error) {
	if mc.listFail {
		return nil, errors.New("List Failed")
	}
	if mc.listEmpty {
		return nil, nil
	}

	// Let's return some more or less realistic images.
	return []*dockerclient.Image{
		&dockerclient.Image{
			Id:          "1",
			ParentId:    "0",       // Not used
			Size:        0,         // Not used
			VirtualSize: 254515796, // 254.5 MB
			RepoTags:    []string{"opensuse:latest"},
			Created:     time.Now().UnixNano(),
		},
		&dockerclient.Image{
			Id:          "2",
			ParentId:    "0",       // Not used
			Size:        0,         // Not used
			VirtualSize: 254515796, // 254.5 MB
			RepoTags:    []string{"opensuse:13.2"},
			Created:     time.Now().UnixNano(),
		},
		&dockerclient.Image{
			Id:          "3",
			ParentId:    "0",       // Not used
			Size:        0,         // Not used
			VirtualSize: 254515796, // 254.5 MB
			RepoTags:    []string{"ubuntu:latest"},
			Created:     time.Now().UnixNano(),
		},
		&dockerclient.Image{
			Id:          "4",
			ParentId:    "0",        // Not used
			Size:        0,          // Not used
			VirtualSize: 254515796,  // 254.5 MB
			RepoTags:    []string{}, // Invalid image
			Created:     time.Now().UnixNano(),
		},
	}, nil
}

func (mc *mockClient) CreateContainer(config *dockerclient.ContainerConfig, name string) (string, error) {
	if mc.createFail {
		return "", errors.New("Create failed")
	}
	name = fmt.Sprintf("zypper-docker-private-%s", config.Image)

	return name, nil
}

func (mc *mockClient) StartContainer(id string, config *dockerclient.HostConfig) error {
	if mc.startFail {
		return errors.New("Start failed")
	}
	if id == "zypper-docker-private-3" {
		// Ubuntu doesn't have zypper: fail.
		return errors.New("Start failed")
	}
	return nil
}

func (mc *mockClient) RemoveContainer(id string, force, volume bool) error {
	if mc.removeFail {
		return errors.New("Remove failed")
	}
	log.Printf("Removed container %v", id)
	return nil
}

func (mc *mockClient) Wait(id string) <-chan dockerclient.WaitResult {
	ch := make(chan dockerclient.WaitResult)

	go func() {
		time.Sleep(mc.waitSleep)
		if mc.waitFail {
			err := errors.New("Wait failed")
			ch <- dockerclient.WaitResult{ExitCode: -1, Error: err}
		} else {
			if mc.commandFail {
				ch <- dockerclient.WaitResult{ExitCode: 1, Error: nil}
			} else {
				ch <- dockerclient.WaitResult{ExitCode: 0, Error: nil}
			}
		}
	}()
	return ch
}

type closingBuffer struct {
	*bytes.Buffer
}

func (cb *closingBuffer) Close() error {
	return nil
}

func (mc *mockClient) ContainerLogs(id string, options *dockerclient.LogOptions) (io.ReadCloser, error) {
	if mc.logFail {
		return nil, fmt.Errorf("Fake log failure")
	}
	cb := &closingBuffer{bytes.NewBuffer([]byte{})}
	_, err := cb.WriteString("streaming buffer initialized\n")
	return cb, err
}
