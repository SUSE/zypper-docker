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

func TestListUpdatesNoImageSpecified(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() { listUpdatesCmd(testContext([]string{}, false)) })

	if testCommand() != "" {
		t.Fatalf("The command should not have been executed")
	}
	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
	if !strings.Contains(buffer.String(), "Error: no image name specified") {
		t.Fatal("It should've logged something\n")
	}
}

func TestListUpdatesCommandFailure(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{commandFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)

	capture.All(func() {
		listUpdatesCmd(testContext([]string{"opensuse:13.2"}, false))
	})

	if !strings.Contains(buffer.String(), "Error: Command exited with status 1") {
		t.Fatal("It should've logged something\n")
	}
	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
}

func TestUpdateCommandWrongInvocation(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() { updateCmd(testContext([]string{}, false)) })

	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
	if !strings.Contains(buffer.String(), "Wrong invocation") {
		t.Fatal("It should've logged something\n")
	}
}

func TestUpdateCommandImageOverwriteDetected(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{listFail: true}

	capture.All(func() { updateCmd(testContext([]string{"ori", "new:1.0.0"}, false)) })

	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
}

func TestUpdateCommandRunAndCommitFailure(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{startFail: true}

	capture.All(func() { updateCmd(testContext([]string{"ori", "new:1.0.0"}, false)) })

	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
}

func TestUpdateCommandCommitSuccess(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() { updateCmd(testContext([]string{"ori", "new:1.0.0"}, false)) })

	if exitInvocations != 0 {
		t.Fatalf("Expected to have exited successfully")
	}
	if !strings.Contains(buffer.String(), "new:1.0.0 successfully created") {
		t.Fatal("It should've logged something\n")
	}
}

func TestListUpdatessContainerCheckContainerFailure(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{listFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)

	capture.All(func() {
		listUpdatesContainerCmd(testContext([]string{"opensuse:13.2"}, false))
	})

	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
}

func TestListUpdatesContainerCheckContainerSuccess(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)

	capture.All(func() {
		listUpdatesContainerCmd(testContext([]string{"1"}, false))
	})

	if exitInvocations != 0 {
		t.Fatalf("Should not have exited with error")
	}
}
