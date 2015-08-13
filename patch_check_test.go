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
	"strings"
	"testing"

	"github.com/mssola/capture"
)

func TestPatchCheckNoImageSpecified(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() { patchCheckCmd(testContext([]string{}, false)) })

	if testCommand() != "" {
		t.Fatalf("The command should not have been executed")
	}
	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
	if !strings.Contains(buffer.String(), "Error: no image name specified") {
		t.Fatal("It should've logged something\n")
	}
	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
}

func TestPatchCheckInvalidError(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{commandFail: true, commandExit: 2}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() {
		patchCheckCmd(testContext([]string{"opensuse:13.2"}, false))
	})

	if testCommand() != "zypper pchk" {
		t.Fatalf("Wrong command!")
	}
	if !strings.Contains(buffer.String(), "Error: Command exited with status 2") {
		t.Fatalf("Wrong error message")
	}
	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
}

func TestPatchCheckSupportedNonZeroExit(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{commandFail: true, commandExit: 100}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() {
		patchCheckCmd(testContext([]string{"opensuse:13.2"}, false))
	})

	if testCommand() != "zypper pchk" {
		t.Fatalf("Wrong command!")
	}
	if len(strings.Split(buffer.String(), "\n")) != 2 {
		t.Fatalf("Something went wrong")
	}
}

func TestPatchCheckOk(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() {
		patchCheckCmd(testContext([]string{"opensuse:13.2"}, false))
	})

	if testCommand() != "zypper pchk" {
		t.Fatalf("Wrong command!")
	}
	if len(strings.Split(buffer.String(), "\n")) != 2 {
		t.Fatalf("Something went wrong")
	}
}

func TestPatchCheckContainerFailure(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{listFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)

	capture.All(func() {
		patchCheckContainerCmd(testContext([]string{"opensuse:13.2"}, false))
	})

	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
}

func TestPatchCheckContainerCheckContainerSuccess(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)

	capture.All(func() {
		patchCheckContainerCmd(testContext([]string{"1"}, false))
	})

	if exitInvocations != 0 {
		t.Fatalf("Should not have exited with error")
	}
}
