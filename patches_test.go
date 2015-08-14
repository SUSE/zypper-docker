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
	"flag"
	"log"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/mssola/capture"
)

func testPatchContext(source, destination string) *cli.Context {
	set := flag.NewFlagSet("test", 0)
	c := cli.NewContext(nil, set, nil)
	args := []string{}

	if source != "" {
		args = append(args, source)
	}
	if destination != "" {
		args = append(args, destination)
	}

	if err := set.Parse(args); err != nil {
		log.Fatalf("Cannot parse args: %s", err)
	}
	return c
}

func testCommand() string {
	cmd := dockerClient.(*mockClient).lastCmd
	if len(cmd) != 1 {
		return ""
	}

	// The command is basically: "zypper ref && actual command".
	return strings.TrimSpace(strings.Split(cmd[0], "&&")[1])
}

func TestListPatchesNoImageSpecified(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() { listPatchesCmd(testContext([]string{}, false)) })

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

func TestListPatchesCommandFailure(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{commandFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)

	capture.All(func() {
		listPatchesCmd(testContext([]string{"opensuse:13.2"}, false))
	})

	if testCommand() != "zypper lp" {
		t.Fatalf("Wrong command!")
	}
	if !strings.Contains(buffer.String(), "Error: Command exited with status 1") {
		t.Fatal("It should've logged something\n")
	}
	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
}

func TestPatchCommandWrongInvocation(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() { patchCmd(testPatchContext("", "")) })

	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
	if !strings.Contains(buffer.String(), "Wrong invocation") {
		t.Fatal("It should've logged something\n")
	}
}

func TestPatchCommandImageOverwriteDetected(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{listFail: true}

	capture.All(func() { patchCmd(testPatchContext("ori", "new:1.0.0")) })

	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
}

func TestPatchCommandRunAndCommitFailure(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{startFail: true}

	capture.All(func() { patchCmd(testPatchContext("ori", "new:1.0.0")) })

	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
}

func TestPatchCommandCommitSuccess(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() { patchCmd(testPatchContext("ori", "new:1.0.0")) })

	if exitInvocations != 0 {
		t.Fatalf("Expected to have exited successfully")
	}
	if !strings.Contains(buffer.String(), "new:1.0.0 successfully created") {
		t.Fatal("It should've logged something\n")
	}
}

func TestListPatchesContainerCheckContainerFailure(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{listFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)

	capture.All(func() {
		listPatchesContainerCmd(testContext([]string{"opensuse:13.2"}, false))
	})

	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
}

func TestListPatchesContainerCheckContainerSuccess(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)

	capture.All(func() {
		listPatchesContainerCmd(testContext([]string{"suse"}, false))
	})

	if exitInvocations != 0 {
		t.Fatalf("Should not have exited with error")
	}
}
