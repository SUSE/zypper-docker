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
	"errors"
	"log"
	"time"

	"github.com/samalba/dockerclient"
)

type mockClient struct {
	id               string
	createFail       bool
	removeFail       bool
	startFail        bool
	inspectFail      bool
	inspectSleep     time.Duration
	monitoringStatus string
	listFail         bool
	listEmpty        bool
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
	mc.id = id
	return nil
}

func (mc *mockClient) RemoveContainer(id string, force, volume bool) error {
	if mc.removeFail {
		return errors.New("Remove failed")
	}
	log.Printf("Removed container %v", id)
	return nil
}

func (mc *mockClient) InspectContainer(id string) (*dockerclient.ContainerInfo, error) {
	if mc.inspectFail {
		return nil, errors.New("Inspect fail")
	}
	return &dockerclient.ContainerInfo{
		State: &dockerclient.State{ExitCode: 0},
	}, nil
}

func (mc *mockClient) StartMonitorEvents(cb dockerclient.Callback, ec chan error, args ...interface{}) {
	// Just create a fake event and get out.
	event := &dockerclient.Event{Id: mc.id, Status: mc.monitoringStatus}
	go func() {
		// This sleep is needed, otherwise we enter in a deadlock with the
		// containers[id] channel...
		time.Sleep(mc.inspectSleep)
		if event.Status == "error" {
			ec <- errors.New("Start monitor errored")
		} else {
			cb(event, ec, args...)
		}
	}()
}

func (mc *mockClient) StopAllMonitorEvents() {
	// Doing nothing.
}
