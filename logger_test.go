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
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/mssola/capture"
)

func TestSetupLoggerDebug(t *testing.T) {
	// Set debug mode.
	set := flag.NewFlagSet("test", 0)
	set.Bool("debug", true, "doc")
	c := cli.NewContext(nil, set, nil)

	res := capture.All(func() {
		_ = setupLogger(c)
		log.Printf("Test")
	})
	if !strings.HasSuffix(string(res.Stdout), "Test\n") {
		t.Fatalf("'%v' expected to have logged 'Test'\n", string(res.Stdout))
	}
}

func TestSetupLoggerHome(t *testing.T) {
	abs, err := filepath.Abs("test")
	if err != nil {
		t.Fatalf("Could not setup the test suite: %v\n", err)
	}
	home := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", home)
		_ = os.Remove(filepath.Join(abs, logFileName))
	}()
	_ = os.Setenv("HOME", abs)

	res := capture.All(func() {
		_ = setupLogger(testContext([]string{}, false))
		log.Printf("Test")
	})
	if len(res.Stdout) != 0 {
		t.Fatal("Nothing should've been printed to stdout\n")
	}
	contents, err := ioutil.ReadFile(filepath.Join(abs, logFileName))
	if err != nil {
		t.Fatalf("Could not read contents of the log: %v\n", err)
	}
	if !strings.HasSuffix(string(contents), "Test\n") {
		t.Fatalf("'%v' expected to have logged 'Test'\n", string(contents))
	}
}

func TestSetupLoggerWrongHome(t *testing.T) {
	home := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", home)
	}()
	abs, err := filepath.Abs("does_not_exist")
	if err != nil {
		t.Fatalf("Could not setup the test suite: %v\n", err)
	}
	_ = os.Setenv("HOME", abs)

	res := capture.All(func() {
		_ = setupLogger(testContext([]string{}, false))
		log.Printf("Test")
	})
	if strings.Index(string(res.Stdout), "Could not open log file") == -1 {
		t.Fatalf("An error should've been printed\n")
	}
	if strings.Index(string(res.Stdout), "Test") == -1 {
		t.Fatalf("There should be a 'Test' string\n")
	}
}
