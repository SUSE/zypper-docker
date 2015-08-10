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

	"github.com/codegangsta/cli"
)

var exitInvocations, lastCode int

func setupTestExitStatus() {
	exitInvocations = 0
	lastCode = 0

	if exitWithCode == nil {
		exitWithCode = func(code int) {
			lastCode = code
			exitInvocations += 1
		}
	}
}

type closingBuffer struct {
	*bytes.Buffer
}

func (cb *closingBuffer) Close() error {
	return nil
}

func testContext(args []string, force bool) *cli.Context {
	set := flag.NewFlagSet("test", 0)
	c := cli.NewContext(nil, set, nil)
	set.Bool("force", force, "doc")
	err := set.Parse(args)
	if err != nil {
		log.Fatal("Cannot parse cli options", err)
	}
	return c
}
