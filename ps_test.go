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

// PS

func TestPsCommand(t *testing.T) {
	cases := testCases{
		{"List fail", &mockClient{listFail: true}, 1, []string{}, true,
			"Error while fetching running containers: Fake failure while listing containers", ""},
		{"Empty list of containers", &mockClient{listEmpty: true}, 0, []string{}, false, "",
			"There are no running containers"},
	}
	cases.run(t, psCmd, "", "")
}

// Special checks for the PS command.

func TestPsCommandNoMatches(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() { psCmd(testContext([]string{}, false)) })

	if strings.Contains(buffer.String(), "Running containers whose images have been updated") {
		t.Fatal("It should not have found matches")
	}
	if exitInvocations != 0 {
		t.Fatalf("Should not have exited with an error")
	}
	if lastCode != 0 {
		t.Fatalf("Exit status should be 1, %v given", lastCode)
	}
}

func TestPsCommandMatches(t *testing.T) {
	cacheFile := getCacheFile()
	cacheFile.Outdated = []string{"2"} // this is the Id of the opensuse:13.2 image
	cacheFile.Other = []string{"3"}    // this is the Id of the ubuntu:latest image
	cacheFile.flush()

	setupTestExitStatus()
	dockerClient = &mockClient{}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	rec := capture.All(func() { psCmd(testContext([]string{}, false)) })

	if !strings.Contains(buffer.String(), "Cannot analyze container 4 [foo]") {
		t.Fatal("Wrong message")
	}
	if !strings.Contains(string(rec.Stdout), "Running containers whose images have been updated") {
		t.Fatal("Wrong message")
	}
	if !strings.Contains(string(rec.Stdout), "The following containers have an unknown state") &&
		!strings.Contains(string(rec.Stdout), "busybox") &&
		!strings.Contains(string(rec.Stdout), "foo") {
		t.Fatal("Wrong message")
	}
	if !strings.Contains(string(rec.Stdout), "The following containers have been ignored") &&
		!strings.Contains(string(rec.Stdout), "ubuntu") {
		t.Fatal("Wrong message")
	}
	if exitInvocations != 0 {
		t.Fatalf("Should not have exited with an error")
	}
	if lastCode != 0 {
		t.Fatalf("Exit status should be 1, %v given", lastCode)
	}
}
