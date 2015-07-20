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
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/codegangsta/cli"
)

func TestImagesCmdFail(t *testing.T) {
	dockerClient = &mockClient{listFail: true}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	imagesCmd(&cli.Context{})

	lines := strings.Split(buffer.String(), "\n")
	if len(lines) != 2 {
		t.Fatal("Wrong number of lines")
	}
	if !strings.Contains(buffer.String(), "List Failed") {
		t.Fatal("It should've logged something expected\n")
	}
}

func TestImagesListEmpty(t *testing.T) {
	dockerClient = &mockClient{listEmpty: true}

	temp, err := ioutil.TempFile("", "zypper")
	if err != nil {
		t.Fatal("Could not setup test")
	}
	original := os.Stdout
	os.Stdout = temp

	imagesCmd(&cli.Context{})
	b, err := ioutil.ReadFile(temp.Name())
	if err != nil {
		t.Fatal("Could not read temporary file")
	}

	temp.Close()
	os.Remove(temp.Name())
	os.Stdout = original

	lines := strings.Split(string(b), "\n")
	if len(lines) != 2 {
		t.Fatal("Wrong number of lines")
	}
	if !strings.HasPrefix(lines[0], "REPOSITORY") {
		t.Fatal("Wrong contents")
	}
}

func TestImagesListOk(t *testing.T) {
	dockerClient = &mockClient{
		inspectSleep:     100 * time.Millisecond,
		monitoringStatus: "die",
	}

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	temp, err := ioutil.TempFile("", "zypper")
	if err != nil {
		t.Fatal("Could not setup test")
	}
	original := os.Stdout
	os.Stdout = temp

	imagesCmd(&cli.Context{})
	b, err := ioutil.ReadFile(temp.Name())
	if err != nil {
		t.Fatal("Could not read temporary file")
	}

	temp.Close()
	os.Remove(temp.Name())
	os.Stdout = original

	lines := strings.Split(string(b), "\n")
	if len(lines) != 4 {
		t.Fatal("Wrong number of lines")
	}
	if !strings.HasPrefix(lines[0], "REPOSITORY") {
		t.Fatal("Wrong contents")
	}
	str := "opensuse            latest              1                   Less than a second ago   254.5 MB"
	if lines[1] != str {
		t.Fatal("Wrong contents")
	}
	str = "opensuse            13.2                2                   Less than a second ago   254.5 MB"
	if lines[2] != str {
		t.Fatal("Wrong contents")
	}
}
