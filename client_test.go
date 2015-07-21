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
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mssola/dockerclient"
)

func TestMockClient(t *testing.T) {
	dockerClient = &mockClient{}

	client := getDockerClient()
	to := reflect.TypeOf(client)
	if to.String() != "*main.mockClient" {
		t.Fatal("Wrong type for the client")
	}

	iface := reflect.TypeOf((*DockerClient)(nil)).Elem()
	if !to.Implements(iface) {
		t.Fatal("The mock type does not implement the DockerClient interface!")
	}
}

// This is the only test that will check for the actual real connection, so for
// safety reasons we do `dockerClient = nil` before and after the actual test.
func TestDockerClient(t *testing.T) {
	dockerClient = nil

	// This test will work even if docker is not running. Take a look at the
	// implementation of it for more details.
	client := getDockerClient()

	docker, ok := client.(*dockerclient.DockerClient)
	if !ok {
		t.Fatal("Could not cast to dockerclient.DockerClient")
	}

	if docker.URL.Scheme != "http" {
		t.Fatalf("Unexpected scheme: %v\n", docker.URL.Scheme)
	}
	if docker.URL.Host != "unix.sock" {
		t.Fatalf("Unexpected host: %v\n", docker.URL.Host)
	}

	dockerClient = nil
}

func TestRunCommandInContainerCreateFailure(t *testing.T) {
	dockerClient = &mockClient{createFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if res := runCommandInContainer("fail", []string{}); res {
		t.Fatal("It should've failed\n")
	}
	if !strings.Contains(buffer.String(), "Create failed") {
		t.Fatal("It should've logged something expected\n")
	}
}

func TestRunCommandInContainerStartFailure(t *testing.T) {
	dockerClient = &mockClient{startFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if res := runCommandInContainer("fail", []string{}); res {
		t.Fatal("It should've failed\n")
	}

	// The only logged stuff is that the created container has been removed.
	lines := strings.Split(buffer.String(), "\n")
	if len(lines) != 2 {
		t.Fatal("Wrong number of lines")
	}
	if !strings.Contains(buffer.String(), "Removed container") {
		t.Fatal("It should've logged something expected\n")
	}
}

func TestRunCommandInContainerWaitFailed(t *testing.T) {
	dockerClient = &mockClient{
		waitFail:  true,
		waitSleep: 100 * time.Millisecond,
	}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if res := runCommandInContainer("fail", []string{}); res {
		t.Fatal("It should've failed\n")
	}

	lines := strings.Split(buffer.String(), "\n")
	if len(lines) != 3 {
		t.Fatal("Wrong number of lines")
	}
	if !strings.Contains(buffer.String(), "Wait failed") {
		t.Fatal("It should've logged something expected\n")
	}
	if !strings.Contains(buffer.String(), "Removed container zypper-docker-private-fail") {
		t.Fatal("It should've logged something expected\n")
	}
}

func TestRunCommandInContainerWaitTimedOut(t *testing.T) {
	dockerClient = &mockClient{waitSleep: containerTimeout * 2}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if res := runCommandInContainer("fail", []string{}); res {
		t.Fatal("It should've failed\n")
	}

	lines := strings.Split(buffer.String(), "\n")
	if len(lines) != 3 {
		t.Fatal("Wrong number of lines")
	}
	if !strings.Contains(buffer.String(), "Timed out when waiting for a container.") {
		t.Fatal("It should've logged something expected\n")
	}
	if !strings.Contains(buffer.String(), "Removed container zypper-docker-private-fail") {
		t.Fatal("It should've logged something expected\n")
	}
}

func TestRunCommandInContainerSuccess(t *testing.T) {
	dockerClient = &mockClient{waitSleep: 100 * time.Millisecond}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if res := runCommandInContainer("ok", []string{}); !res {
		t.Fatal("It should've been ok\n")
	}

	lines := strings.Split(buffer.String(), "\n")
	if len(lines) != 2 {
		t.Fatal("Wrong number of lines")
	}
	if !strings.Contains(buffer.String(), "Removed container zypper-docker-private-ok") {
		t.Fatal("It should've logged something expected\n")
	}
}

func TestRemoveContainerFail(t *testing.T) {
	dockerClient = &mockClient{removeFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	removeContainer(dockerClient, "fail")
	if !strings.Contains(buffer.String(), "Remove failed") {
		t.Fatal("It should've logged something expected\n")
	}

	// Making sure that the logger has not print the "success" message
	// from the mock type.
	lines := strings.Split(buffer.String(), "\n")
	if len(lines) != 2 {
		t.Fatal("Wrong number of lines")
	}
}
