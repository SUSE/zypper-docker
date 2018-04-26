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
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/client"
	"github.com/mssola/capture"
)

func TestMockClient(t *testing.T) {
	safeClient.client = &mockClient{}

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
// safety reasons we do `safeClient.client = nil` before and after the actual test.
func TestDockerClient(t *testing.T) {
	safeClient.client = nil

	// This test will work even if docker is not running. Take a look at the
	// implementation of it for more details.
	cl := getDockerClient()

	if _, ok := cl.(*client.Client); !ok {
		t.Fatal("Could not cast to *client.Client")
	}

	safeClient.client = nil
}

func TestRunCommandInContainerCreateFailure(t *testing.T) {
	safeClient.client = &mockClient{createFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if _, err := runCommandInContainer("fail", []string{}, nil); err == nil {
		t.Fatal("It should've failed\n")
	}

	testReaderData(t, buffer, []string{"Create failed"})
}

func TestCreateContainerWarnings(t *testing.T) {
	safeClient.client = &mockClient{createWarnings: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if _, err := createContainer("image", []string{"such", "command", "wow"}); err != nil {
		t.Fatalf("We've got the error %v", err)
	}

	testReaderData(t, buffer, []string{"warning"})
}

func TestRunCommandInContainerStartFailure(t *testing.T) {
	safeClient.client = &mockClient{startFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if ret := checkCommandInImage("fail", ""); ret {
		t.Fatal("It should've failed\n")
	}

	testReaderData(t, buffer, []string{"Start failed", "Removed container"})
}

func TestRunCommandInContainerContainerLogsFailure(t *testing.T) {
	safeClient.client = &mockClient{logFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	_, err := runCommandInContainer("opensuse", []string{"zypper"}, os.Stdout)
	if err == nil {
		t.Fatal("It should have failed\n")
	}

	testReaderData(t, buffer, []string{"Fake log failure"})
}

func TestRunCommandInContainerStreaming(t *testing.T) {
	safeClient.client = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)

	var err error
	resp := capture.All(func() {
		_, err = runCommandInContainer("opensuse", []string{"foo"}, os.Stdout)
	})
	if err != nil {
		t.Fatal("It shouldn't have failed")
	}

	testReaderData(t, buffer, []string{})

	str := "streaming buffer initialized"
	if !strings.Contains(string(resp.Stdout), str) {
		t.Fatalf("Expected the text \"%s\" in: %s", str, resp.Stdout)
	}
}

func TestRunCommandInContainerCommandFailure(t *testing.T) {
	safeClient.client = &mockClient{commandFail: true}

	var err error

	capture.All(func() {
		_, err = runCommandInContainer("busybox", []string{"zypper"}, nil)
	})

	if err == nil {
		t.Fatal("It should've failed\n")
	}

	if err.Error() != "Command exited with status 1" {
		t.Fatal("Wrong type of error received")
	}
}

func TestCheckCommandInImageWaitFailed(t *testing.T) {
	safeClient.client = &mockClient{
		waitFail:  true,
		waitSleep: 100 * time.Millisecond,
	}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if res := checkCommandInImage("fail", ""); res {
		t.Fatal("It should've failed\n")
	}

	testReaderData(t, buffer, []string{"Wait failed", "Removed container zypper-docker-private-fail"})
}

func TestCheckCommandInImageWaitTimedOut(t *testing.T) {
	safeClient.client = &mockClient{waitSleep: containerTimeout * 2}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if res := checkCommandInImage("fail", ""); res {
		t.Fatal("It should've failed\n")
	}

	testReaderData(t, buffer, []string{"Timed out when waiting for a container",
		"Removed container zypper-docker-private-fail"})
}

func TestCheckCommandInImageSuccess(t *testing.T) {
	safeClient.client = &mockClient{waitSleep: 100 * time.Millisecond}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if res := checkCommandInImage("ok", ""); !res {
		t.Fatal("It should've been ok\n")
	}

	testReaderData(t, buffer, []string{"Removed container zypper-docker-private-ok"})
}

func TestRemoveContainerFail(t *testing.T) {
	safeClient.client = &mockClient{removeFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	removeContainer("fail")
	testReaderData(t, buffer, []string{"Remove failed"})
}

func TestHandleSignalWhileContainerRuns(t *testing.T) {
	// create the killChannel, make it buffered and put already a message inside
	// of it
	killChannel = make(chan bool, 1)
	killChannel <- true

	exitInvocations = 0
	exitWithCode = func(code int) {
		exitInvocations++
	}

	safeClient.client = &mockClient{}

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
		exitInvocations++
	}

	safeClient.client = &mockClient{killFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() { checkCommandInImage("kill", "") })

	if exitInvocations != 1 {
		t.Fatal("os.Exit should have been called by the client code\n")
	}
}

func TestRunCommandAndCommitToImageSuccess(t *testing.T) {
	safeClient.client = &mockClient{}
	var err error

	capture.All(func() {
		_, err = runCommandAndCommitToImage(
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
	safeClient.client = &mockClient{startFail: true}
	var err error

	capture.All(func() {
		_, err = runCommandAndCommitToImage(
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
	safeClient.client = &mockClient{commitFail: true}
	var err error

	capture.All(func() {
		_, err = runCommandAndCommitToImage(
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
	safeClient.client = &mockClient{listFail: true}

	_, err := checkContainerRunning("1")

	if err == nil {
		t.Fatal("Was supposed to have an error")
	}

	if !strings.Contains(err.Error(), "Fake failure while listing containers") {
		t.Fatal("Unexpected error message")
	}
}

func TestCheckContainerRunningNoRunningContainer(t *testing.T) {
	safeClient.client = &mockClient{listEmpty: true}

	_, err := checkContainerRunning("35ae93c88cf8")

	if err == nil {
		t.Fatal("Was supposed to have an error")
	}

	if !strings.Contains(err.Error(), "Cannot find running container") {
		t.Fatal("Unexpected error message")
	}
}

func TestCheckContainerRunningWrongContainer(t *testing.T) {
	safeClient.client = &mockClient{}

	_, err := checkContainerRunning("not running")

	if err == nil {
		t.Fatal("Was supposed to have an error")
	}

	if !strings.Contains(err.Error(), "Cannot find running container") {
		t.Fatal("Unexpected error message")
	}
}

func TestCheckContainerRunningNotSUSESystem(t *testing.T) {
	safeClient.client = &mockClient{startFail: true}

	_, err := checkContainerRunning("not_suse")

	if err == nil {
		t.Fatal("Was supposed to have an error")
	}

	if !strings.Contains(err.Error(), "which is not a SUSE system") {
		t.Fatal("Unexpected error message")
	}
}

func TestCheckContainerRunningByNameSuccess(t *testing.T) {
	safeClient.client = &mockClient{}

	container, err := checkContainerRunning("suse")

	if err != nil {
		t.Fatal("Wasn't supposed to have an error")
	}

	if container.ID != "35ae93c88cf8ab18da63bb2ad2dfd2399d745f292a344625fbb65892b7c25a01" {
		t.Fatal("Wrong container found")
	}
}

func TestCheckContainerRunningByFullIDSuccess(t *testing.T) {
	safeClient.client = &mockClient{}

	container, err := checkContainerRunning("35ae93c88cf8ab18da63bb2ad2dfd2399d745f292a344625fbb65892b7c25a01")

	if err != nil {
		t.Fatal("Wasn't supposed to have an error")
	}

	if container.ID != "35ae93c88cf8ab18da63bb2ad2dfd2399d745f292a344625fbb65892b7c25a01" {
		t.Fatal("Wrong container found")
	}
}

func TestCheckContainerRunningByShortIDSuccess(t *testing.T) {
	safeClient.client = &mockClient{}

	container, err := checkContainerRunning("35ae93c88cf8")

	if err != nil {
		t.Fatal("Wasn't supposed to have an error")
	}

	if container.ID != "35ae93c88cf8ab18da63bb2ad2dfd2399d745f292a344625fbb65892b7c25a01" {
		t.Fatal("Wrong container found")
	}
}

func TestHostConfig(t *testing.T) {
	hc := getHostConfig()
	if len(hc.ExtraHosts) != 0 {
		t.Fatalf("Wrong number of extra hosts: %v; Expected: 1", len(hc.ExtraHosts))
	}

	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		currentContext = nil
	}()
	os.Args = []string{"exe", "--add-host", "host:ip", "test"}

	app := newApp()
	app.Commands = []cli.Command{{Name: "test", Action: getCmd("test", func(*cli.Context) {})}}
	capture.All(func() { app.RunAndExitOnError() })

	hc = getHostConfig()
	if len(hc.ExtraHosts) != 1 {
		t.Fatalf("Wrong number of extra hosts: %v; Expected: 1", len(hc.ExtraHosts))
	}
	if hc.ExtraHosts[0] != "host:ip" {
		t.Fatalf("Did not expect %v", hc.ExtraHosts[0])
	}
}
