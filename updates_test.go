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

func testListUpdatesContext(image string) *cli.Context {
	set := flag.NewFlagSet("test", 0)
	err := set.Parse([]string{image})
	if err != nil {
		log.Fatal("Cannot parse cli options", err)
	}
	return cli.NewContext(nil, set, nil)
}

func TestListUpdatesNoImageSpecified(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() { listUpdatesCmd(testListUpdatesContext("")) })

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
		listUpdatesCmd(testListUpdatesContext("opensuse:13.2"))
	})

	if !strings.Contains(buffer.String(), "Error: Command exited with status 1") {
		t.Fatal("It should've logged something\n")
	}
	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}
}
