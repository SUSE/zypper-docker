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
	"syscall"
	"testing"
	"time"

	"github.com/mssola/capture"
)

// TODO
func TestMain(m *testing.M) {
	status := 0

	home, umask := os.Getenv("HOME"), syscall.Umask(0)
	abs, _ := filepath.Abs(".")
	test := filepath.Join(abs, "test")

	defer func() {
		_ = os.Setenv("HOME", home)
		syscall.Umask(umask)
		_ = os.Remove(filepath.Join(test, ".cache", cacheName))
		os.Exit(status)
	}()

	_ = os.Setenv("HOME", test)

	status = m.Run()
}

// IMAGES
// TODO

func TestImagesCommand(t *testing.T) {
	cases := testCases{
		{"List fail", &mockClient{listFail: true}, 1, []string{}, false, "Cannot proceed safely: List Failed", ""},
		{"Empty list of images", &mockClient{listEmpty: true}, 0, []string{}, false, "", "REPOSITORY"},
	}
	cases.run(t, imagesCmd, "", "")
}

func TestImagesCommandList(t *testing.T) {
	safeClient.client = &mockClient{waitSleep: 100 * time.Millisecond}
	setupTestExitStatus()

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)

	res := capture.All(func() { imagesCmd(testContext([]string{}, false)) })

	testReaderData(t, bytes.NewBuffer(res.Stdout), []string{
		"REPOSITORY",
		"opensuse            latest              1",
		"opensuse            tag                 1",
		"opensuse            13.2                2",
		"busybox             latest              5",
	})
	if exitInvocations != 1 && lastCode != 0 {
		t.Fatal("Wrong exit code")
	}
}

// Special tests for the IMAGES command.

func TestImagesListUsingCache(t *testing.T) {
	safeClient.client = &mockClient{waitSleep: 100 * time.Millisecond}
	setupTestExitStatus()

	// Dump some dummy value.
	cd := getCacheFile()
	cd.Suse = []string{"1"}
	cd.Other = []string{"3"}
	cd.flush()

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)

	res := capture.All(func() { imagesCmd(testContext([]string{}, false)) })

	testReaderData(t, bytes.NewBuffer(res.Stdout), []string{
		"REPOSITORY",
		"opensuse            latest              1",
		"opensuse            tag                 1",
		"opensuse            13.2                2",
		"busybox             latest              5",
	})

	if exitInvocations != 1 && lastCode != 0 {
		t.Fatal("Wrong exit code")
	}
}

func TestImagesForce(t *testing.T) {
	safeClient.client = &mockClient{waitSleep: 100 * time.Millisecond}
	setupTestExitStatus()

	// Dump some dummy value.
	cd := getCacheFile()
	cd.Suse = []string{"1234"}
	cd.flush()

	// Check that they are really written there.
	cd = getCacheFile()
	if cd.Suse[0] != "1234" {
		t.Fatal("Unexpected value")
	}

	// Luke, use the force!
	capture.All(func() { imagesCmd(testContext([]string{}, true)) })
	cd = getCacheFile()

	if !cd.Valid {
		t.Fatal("It should be valid")
	}
	if len(cd.Suse) != 5 {
		t.Fatalf("Expected 5 SUSE images, got %v", len(cd.Suse))
	}

	for i, v := range []string{"1234", "1", "2", "4", "5"} {
		if cd.Suse[i] != v {
			t.Fatal("Unexpected value")
		}
	}
	if len(cd.Other) != 1 || cd.Other[0] != "3" {
		t.Fatal("Unexpected value")
	}
	if exitInvocations != 1 && lastCode != 0 {
		t.Fatal("Wrong exit code")
	}
}

// Helper functions in the images.go file.

func TestCheckImageListFail(t *testing.T) {
	safeClient.client = &mockClient{listFail: true}

	var err error

	capture.All(func() {
		_, err = checkImageExists("opensuse", "bar")
	})

	if err == nil {
		t.Fatal("Error did not occur")
	}
}

func TestCheckImageExistsEmptyList(t *testing.T) {
	var found bool
	var err error

	safeClient.client = &mockClient{listEmpty: true}

	capture.All(func() {
		found, err = checkImageExists("suse/sles11sp3", "latest")
	})

	if err != nil {
		t.Fatal("Unexpected error")
	}
	if found == true {
		t.Fatal("The image should not have been found")
	}
}

func TestCheckImageExists(t *testing.T) {
	var found bool
	var err error

	safeClient.client = &mockClient{waitSleep: 100 * time.Millisecond}

	expected := []string{"latest", "13.2"}
	for _, e := range expected {
		capture.All(func() {
			found, err = checkImageExists("opensuse", e)
		})

		if err != nil {
			t.Fatal("Unexpected error")
		}
		if found != true {
			t.Fatal("The image should have been found")
		}
	}

	notExpected := []string{"unexpected_tag"}
	for _, unexpected := range notExpected {
		capture.All(func() {
			found, err = checkImageExists("opensuse", unexpected)
		})

		if err != nil {
			t.Fatal("Unexpected error")
		}
		if found != false {
			t.Fatal("The image should not have been found")
		}
	}
}

// PATCH-CHECK

func TestPatchCheckCommand(t *testing.T) {
	cases := testCases{
		{"Image not specified", &mockClient{}, 1, []string{}, true, "Error: no image name specified.", ""},
		{"Invalid error", &mockClient{commandFail: true, commandExit: 2}, 1, []string{"opensuse:13.2"}, false,
			"Could not execute command 'zypper pchk' successfully in image 'opensuse:13.2': Command exited with status 2.",
			"streaming buffer initialized"},
		{"Supported non-zero exit", &mockClient{commandFail: true, commandExit: 100}, 100, []string{"opensuse:13.2"}, false,
			"Removed container zypper-docker-private-opensuse:13.2",
			"streaming buffer initialized"},
		{"Ok", &mockClient{}, 0, []string{"opensuse:13.2"}, false, "Removed container zypper-docker-private-opensuse:13.2",
			"streaming buffer initialized"},
	}
	cases.run(t, patchCheckCmd, "zypper pchk", "")
}

// PATCH-CHECK-CONTAINER

func TestPatchCheckContainerCommand(t *testing.T) {
	cases := testCases{
		{"List Command fails", &mockClient{listFail: true}, 1, []string{"opensuse:13.2"}, true,
			"Error while fetching running containers: Fake failure while listing containers.", ""},
		{"Ok", &mockClient{}, 0, []string{"suse"}, false, "Removed container zypper-docker-private-opensuse:13.2",
			"streaming buffer initialized"},
	}
	cases.run(t, patchCheckContainerCmd, "zypper pchk", "")
}

// PATCH

func TestPatchCommand(t *testing.T) {
	cases := testCases{
		{"Wrong number of arguments", &mockClient{}, 1, []string{}, true, "Wrong invocation: expected 2 arguments, 0 given.", ""},
		{"Wrong format of image name", &mockClient{}, 1, []string{"ori", "dollar$$"}, true, "Could not parse 'dollar$$': invalid reference format", ""},
		{"List Command fails", &mockClient{listFail: true}, 1, []string{"ori", "opensuse:13.2"}, true, "Cannot proceed safely: List Failed.", ""},
		{"Overwrite detected", &mockClient{}, 1, []string{"ori", "opensuse:13.2"}, true, "Cannot overwrite an existing image. Please use a different repository/tag.", ""},
		{"Start fail on commit", &mockClient{startFail: true}, 1, []string{"ori", "new:1.0.0"}, true, "Could not commit to the new image: Start failed.", ""},
		{"Cannot update cache", &mockClient{}, 1, []string{"ori", "new:1.0.0"}, false, "Cannot add image details to zypper-docker cache", ""},
		{"Cannot inspect", &mockClient{inspectFail: true}, 1, []string{"opensuse:13.2", "new:1.0.0"}, true, "could not inspect image 'opensuse:13.2': inspect fail", ""},
		{"Patch success", &mockClient{listReturnOneImage: true}, 0, []string{"opensuse:13.2", "new:1.0.0"}, true, "new:1.0.0 successfully created", ""},
	}
	cases.run(t, patchCmd, "zypper -n patch", "")
}

// LIST PATCHES

func TestListPatchesCommand(t *testing.T) {
	cases := testCases{
		{"No image specified", &mockClient{}, 1, []string{}, true, "no image name specified", ""},
		{"Command fail", &mockClient{commandFail: true}, 1, []string{"opensuse:13.2"}, false, "Error: Command exited with status 1", ""},
		{"List patches", &mockClient{}, 0, []string{"opensuse:13.2"}, false, "Removed container zypper-docker-private-opensuse:13.2", "streaming buffer initialized"},
	}
	cases.run(t, listPatchesCmd, "zypper lp", "")
}

// LIST PATCHES CONTAINER

func TestListPatchesContainerCommand(t *testing.T) {
	cases := testCases{
		{"List fails on list patch container", &mockClient{listFail: true}, 1, []string{"opensuse:13.2"}, true, "Error while fetching running containers: Fake failure while listing containers", ""},
		{"Patches container successfully", &mockClient{}, 0, []string{"suse"}, false, "Removed container zypper-docker-private-opensuse:13.2", "streaming buffer initialized"},
	}
	cases.run(t, listPatchesContainerCmd, "zypper lp", "")
}

// UPDATE

func TestUpdateCommand(t *testing.T) {
	cases := testCases{
		{"Wrong number of arguments", &mockClient{}, 1, []string{}, true, "Wrong invocation: expected 2 arguments, 0 given.", ""},
		{"Wrong format of image name", &mockClient{}, 1, []string{"ori", "dollar$$"}, true, "Could not parse 'dollar$$': invalid reference format", ""},
		{"List Command fails", &mockClient{listFail: true}, 1, []string{"ori", "opensuse:13.2"}, true, "Cannot proceed safely: List Failed.", ""},
		{"Overwrite detected", &mockClient{}, 1, []string{"ori", "opensuse:13.2"}, true, "Cannot overwrite an existing image. Please use a different repository/tag.", ""},
		{"Start fail on commit", &mockClient{startFail: true}, 1, []string{"ori", "new:1.0.0"}, true, "Could not commit to the new image: Start failed.", ""},
		{"Cannot update cache", &mockClient{}, 1, []string{"ori", "new:1.0.0"}, false, "Cannot add image details to zypper-docker cache", ""},
		{"Update success", &mockClient{listReturnOneImage: true}, 0, []string{"opensuse:13.2", "new:1.0.0"}, true, "new:1.0.0 successfully created", ""},
	}
	cases.run(t, updateCmd, "zypper -n up", "")
}

// LIST UPDATES

func TestListUpdatesCommand(t *testing.T) {
	cases := testCases{
		{"No image specified", &mockClient{}, 1, []string{}, true, "no image name specified", ""},
		{"Command failure", &mockClient{commandFail: true}, 1, []string{"opensuse:13.2"}, false, "Command exited with status 1", ""},
	}
	cases.run(t, listUpdatesCmd, "zypper lu", "")
}

// LIST UPDATES CONTAINER

func TestListUpdatesContainerCommand(t *testing.T) {
	cases := testCases{
		{"List fails on list update container", &mockClient{listFail: true}, 1, []string{"opensuse:13.2"}, true, "Error while fetching running containers: Fake failure while listing containers", ""},
		{"Updates container successfully", &mockClient{}, 0, []string{"suse"}, false, "Removed container zypper-docker-private-opensuse:13.2", ""},
	}
	cases.run(t, listUpdatesContainerCmd, "zypper lu", "")
}

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
	safeClient.client = &mockClient{}

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
	safeClient.client = &mockClient{}

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
