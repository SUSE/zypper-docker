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
	"path/filepath"
	"strings"
	"testing"

	"github.com/mssola/capture"
)

func TestPsCommandListContaienrsFailure(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{listFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() { psCmd(testContext([]string{}, false)) })

	if !strings.Contains(buffer.String(), "Cannot list running containers") {
		t.Fatal("It should've logged something\n")
	}
	if exitInvocations != 1 {
		t.Fatalf("Expected to have exited with error")
	}

}

func TestPsCommandListContaienrsEmpty(t *testing.T) {
	setupTestExitStatus()
	dockerClient = &mockClient{listEmpty: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	capture.All(func() { psCmd(testContext([]string{}, false)) })

	if !strings.Contains(buffer.String(), "There are no running containers") {
		t.Fatal("It should've logged something\n")
	}
	if exitInvocations != 0 {
		t.Fatalf("Should not have exited with an error")
	}
}

func TestPsCommandNoMatches(t *testing.T) {
	cache := os.Getenv("XDG_CACHE_HOME")
	abs, _ := filepath.Abs(".")
	test := filepath.Join(abs, "test")

	defer func() {
		_ = os.Setenv("XDG_CACHE_HOME", cache)
	}()

	_ = os.Setenv("XDG_CACHE_HOME", test)

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
}

func TestPsCommandMatches(t *testing.T) {
	cache := os.Getenv("XDG_CACHE_HOME")
	abs, _ := filepath.Abs(".")
	test := filepath.Join(abs, "test")

	defer func() {
		_ = os.Setenv("XDG_CACHE_HOME", cache)
	}()

	_ = os.Setenv("XDG_CACHE_HOME", test)

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
}
