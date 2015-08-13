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

	"github.com/SUSE/dockerclient"
	"github.com/mssola/capture"
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
	if _, err := runCommandInContainer("fail", []string{}, false); err == nil {
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
	if ret := checkCommandInImage("fail", ""); ret {
		t.Fatal("It should've failed\n")
	}

	// The only logged stuff is that the created container has been removed.
	lines := strings.Split(buffer.String(), "\n")
	if len(lines) != 3 {
		t.Fatal("Wrong number of lines")
	}
	if !strings.Contains(buffer.String(), "Removed container") {
		t.Fatal("It should've logged something expected\n")
	}
	if !strings.Contains(buffer.String(), "Start failed") {
		t.Fatal("It should've logged something expected\n")
	}
}

func TestRunCommandInContainerContainerLogsFailure(t *testing.T) {
	dockerClient = &mockClient{logFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	_, err := runCommandInContainer("opensuse", []string{"zypper"}, true)
	if err == nil {
		t.Fatal("It should have failed\n")
	}

	if err.Error() != "Fake log failure" {
		t.Fatal("Should have failed because of a logging issue")
	}
}

func TestRunCommandInContainerStreaming(t *testing.T) {
	mock := mockClient{}
	dockerClient = &mock

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)

	var err error
	res := capture.All(func() {
		_, err = runCommandInContainer("opensuse", []string{"foo"}, true)
	})

	if err != nil {
		t.Fatal("It shouldn't have failed\n")
	}

	if !strings.Contains(string(res.Stdout), "streaming buffer initialized") {
		t.Fatal("The streaming buffer should have been initialized\n")
	}
}

func TestRunCommandInContainerCommandFailure(t *testing.T) {
	dockerClient = &mockClient{commandFail: true}

	var err error

	capture.All(func() {
		_, err = runCommandInContainer("busybox", []string{"zypper"}, false)
	})

	if err == nil {
		t.Fatal("It should've failed\n")
	}

	if err.Error() != "Command exited with status 1" {
		t.Fatal("Wrong type of error received")
	}
}

func TestCheckCommandInImageWaitFailed(t *testing.T) {
	dockerClient = &mockClient{
		waitFail:  true,
		waitSleep: 100 * time.Millisecond,
	}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if res := checkCommandInImage("fail", ""); res {
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

func TestCheckCommandInImageWaitTimedOut(t *testing.T) {
	dockerClient = &mockClient{waitSleep: containerTimeout * 2}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if res := checkCommandInImage("fail", ""); res {
		t.Fatal("It should've failed\n")
	}

	lines := strings.Split(buffer.String(), "\n")
	if len(lines) != 4 {
		t.Fatal("Wrong number of lines")
	}
	if !strings.Contains(buffer.String(), "Timed out when waiting for a container.") {
		t.Fatal("It should've logged something expected\n")
	}
	if !strings.Contains(buffer.String(), "Removed container zypper-docker-private-fail") {
		t.Fatal("It should've logged something expected\n")
	}
}

func TestCheckCommandInImageSuccess(t *testing.T) {
	dockerClient = &mockClient{waitSleep: 100 * time.Millisecond}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if res := checkCommandInImage("ok", ""); !res {
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
	removeContainer("fail")
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

func TestHandleSignalWhileContainerRuns(t *testing.T) {
	// create the killChannel, make it buffered and put already a message inside
	// of it
	killChannel = make(chan bool, 1)
	killChannel <- true

	exitInvocations = 0
	exitWithCode = func(code int) {
		exitInvocations += 1
	}

	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	checkCommandInImage("kill", "")

	if exitInvocations != 1 {
		t.Fatal("os.Exit should have been called by the client code\n")
	}
}

func TestHandleSignalWhileContainerRunsEvenWhenKillContainerFails(t *testing.T) {
	// create the killChannel, make it buffered and put already a message inside
	// of it
	killChannel = make(chan bool, 1)
	killChannel <- true

	exitInvocations = 0
	exitWithCode = func(code int) {
		exitInvocations += 1
	}

	dockerClient = &mockClient{killFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() { checkCommandInImage("kill", "") })

	if exitInvocations != 1 {
		t.Fatal("os.Exit should have been called by the client code\n")
	}
}

func TestRunCommandAndCommitToImageSuccess(t *testing.T) {
	dockerClient = &mockClient{}
	var err error

	capture.All(func() {
		err = runCommandAndCommitToImage(
			"source",
			"new_repo",
			"new_tag",
			"touch foo",
			"comment",
			"author")
	})

	if err != nil {
		t.Fatalf("Unexpected error")
	}
}

func TestRunCommandAndCommitToImageRunFailure(t *testing.T) {
	dockerClient = &mockClient{startFail: true}
	var err error

	capture.All(func() {
		err = runCommandAndCommitToImage(
			"source",
			"new_repo",
			"new_tag",
			"touch foo",
			"comment",
			"author")
	})

	if err == nil {
		t.Fatalf("No error")
	}
	if !strings.Contains(err.Error(), "Start failed") {
		t.Fatalf("Wrong error")
	}
}

func TestRunCommandAndCommitToImageCommitFailure(t *testing.T) {
	dockerClient = &mockClient{commitFail: true}
	var err error

	capture.All(func() {
		err = runCommandAndCommitToImage(
			"source",
			"new_repo",
			"new_tag",
			"touch foo",
			"comment",
			"author")
	})

	if err == nil {
		t.Fatalf("No error")
	}
	if !strings.Contains(err.Error(), "Fake failure while committing container") {
		t.Fatalf("Wrong error")
	}
}

func TestCheckContainerRunningListContainersFailure(t *testing.T) {
	dockerClient = &mockClient{listFail: true}

	container, err := checkContainerRunning("1")

	if container != nil {
		t.Fatal("Wasn't supposed to find container")
	}

	if err == nil {
		t.Fatal("Was supposed to have an error")
	}

	if !strings.Contains(err.Error(), "Fake failure while listing containers") {
		t.Fatal("Unexpected error message")
	}
}

func TestCheckContainerRunningNoRunningContainer(t *testing.T) {
	dockerClient = &mockClient{listEmpty: true}

	container, err := checkContainerRunning("1")

	if container != nil {
		t.Fatal("Wasn't supposed to find container")
	}

	if err == nil {
		t.Fatal("Was supposed to have an error")
	}

	if !strings.Contains(err.Error(), "Cannot find running container") {
		t.Fatal("Unexpected error message")
	}
}

func TestCheckContainerRunningWrongContainer(t *testing.T) {
	dockerClient = &mockClient{}

	container, err := checkContainerRunning("not running")

	if container != nil {
		t.Fatal("Wasn't supposed to find container")
	}

	if err == nil {
		t.Fatal("Was supposed to have an error")
	}

	if !strings.Contains(err.Error(), "Cannot find running container") {
		t.Fatal("Unexpected error message")
	}
}

func TestCheckContainerRunningNotSUSESystem(t *testing.T) {
	dockerClient = &mockClient{startFail: true}

	container, err := checkContainerRunning("not_suse")

	if container != nil {
		t.Fatal("Wasn't supposed to find container")
	}

	if err == nil {
		t.Fatal("Was supposed to have an error")
	}

	if !strings.Contains(err.Error(), "which is not a SUSE system") {
		t.Fatal("Unexpected error message")
	}
}

func TestCheckContainerRunningSuccess(t *testing.T) {
	dockerClient = &mockClient{}

	container, err := checkContainerRunning("suse")

	if container == nil {
		t.Fatal("Was supposed to find container")
	}

	if err != nil {
		t.Fatal("Wasn't supposed to have an error")
	}

	if container.Id != "1" {
		t.Fatal("Wrong container found")
	}
}
