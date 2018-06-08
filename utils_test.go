// Copyright (c) 2018 SUSE LLC. All rights reserved.
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
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/mssola/capture"
)

var exitInvocations, lastCode int

func setupTestExitStatus() {
	exitInvocations = 0
	lastCode = 0
	debugMode = false

	exitWithCode = func(code int) {
		lastCode = code
		exitInvocations++
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

func compareStringSlices(actual, expected []string) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("different size, actual is %d while expected is %d",
			len(actual), len(expected))
	}

	for pos, item := range actual {
		if item != expected[pos] {
			return fmt.Errorf("item at position %d are different, expected %s, got %s",
				pos, expected[pos], item)
		}
	}

	return nil
}

// testReaderData scans the data available in the given reader and matches each
// line with the given messages. Since the messages most surely come from a log
// message, the comparison will be done with the strings.Contains function,
// instead of a full match.
func testReaderData(t *testing.T, reader io.Reader, messages []string) {
	scanner := bufio.NewScanner(reader)
	idx := 0
	read := 0

	for ; scanner.Scan(); idx++ {
		if idx == len(messages) {
			t.Fatalf("More than %v messages! Next message: %v", len(messages), scanner.Text())
		}
		if txt := scanner.Text(); !strings.Contains(txt, messages[idx]) {
			t.Fatalf("Expected the text \"%s\" in: %s", messages[idx], txt)
		}
		read++
	}
	if read != len(messages) {
		t.Fatalf("Expected %v messages, but we have read %v.", len(messages), read)
	}
}

// Fetch the last command that has been executed. Note that this evaluates that
// the command has been executed inside of a container, it doesn't care whether
// the command exited before trying to start a container.
func testCommand() string {
	cmd := safeClient.client.(*mockClient).lastCmd
	if len(cmd) != 1 {
		return ""
	}

	// The command is basically: "zypper ref && actual command".
	return strings.TrimSpace(strings.Split(cmd[0], "&&")[1])
}

// testCase represents anything that can be tested for a command while using
// a table for testing all cases.
type testCase struct {
	// Short description of the test case.
	desc string

	// The docker client to be used.
	client *mockClient

	// The exit code.
	code int

	// The arguments passed to the command.
	args []string

	// Whether the stdout matches the log message.
	logAndStdout bool

	// The message being printed in the logs/stdout.
	msg string

	// The message being printed to stdout. If logAndStdout is false and this
	// is empty, it will mean that no relevant stdout has to be printed.
	stdout string
}

// testCases is a collection of test cases for a single command.
type testCases []testCase

// run the test cases.
func (cases testCases) run(t *testing.T, cmd func(*cli.Context), command, debug string) {
	for _, test := range cases {
		// Skip if this not the one being debugged.
		if debug != "" && debug != test.desc {
			continue
		}

		setupTestExitStatus()
		safeClient.client = test.client

		buffer := bytes.NewBuffer([]byte{})
		log.SetOutput(buffer)
		captured := capture.All(func() { cmd(testContext(test.args, false)) })

		// Exit status code
		if lastCode != test.code {
			t.Fatalf("[%s] Expected to have exited with code %v, %v was received.",
				test.desc, test.code, lastCode)
		}

		// Log
		if test.msg == "" {
			lines := strings.Split(buffer.String(), "\n")
			// The first line might be the cache failing to be loaded.
			if !(len(lines) == 1 || len(lines) == 2) || lines[len(lines)-1] != "" {
				t.Fatalf("Should've logged nothing, logged:\n%s\n", buffer.String())
			}
		} else {
			if !strings.Contains(buffer.String(), test.msg) {
				t.Fatalf("[%s] Wrong logged message.\nExpecting:\n%s\n===\nReceived:\n%s\n",
					test.desc, test.msg, buffer.String())
			}
		}

		// Stdout
		if test.logAndStdout {
			test.stdout = test.msg
		}
		if test.stdout != "" {
			if !strings.Contains(string(captured.Stdout), test.stdout) {
				t.Fatalf("[%s] Wrong stdout.\nExpecting:\n%s\n===\nReceived:\n%s\n",
					test.desc, test.stdout, string(captured.Stdout))
			}
		}
		if lastCode == 0 && command != "" {
			if command != testCommand() {
				t.Fatalf("[%s] Wrong command. Expecting '%s', '%s' received.\n",
					test.desc, command, testCommand())
			}
		}
	}
}
