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

// Generates all the man pages from the given Markdown documents.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cpuguy83/go-md2man/md2man"
)

// The directory in which the generated files will be saved.
const m1Dir = "man1"

// A map containing the commands which have multiple names.
var duplicates map[string]string

// Returns the alternative name for the given command. If no alternative
// exists, it just returns an empty string.
func duplicateName(name string) string {
	re := regexp.MustCompile(`^zypper-docker-([a-z]+)\.1$`)
	matches := re.FindStringSubmatch(name)
	if len(matches) != 2 {
		return ""
	}

	if val, ok := duplicates[matches[1]]; ok {
		return "zypper-docker-" + val + ".1"
	}
	return ""
}

// Generate the man page for the given command (alternative name included).
func generateMan(info os.FileInfo, name string) {
	contents, err := ioutil.ReadFile(info.Name())
	if err != nil {
		fmt.Printf("Could not read file: %v\n", err)
		os.Exit(1)
	}
	out := md2man.Render(contents)
	if len(out) == 0 {
		fmt.Println("Failed to render")
		os.Exit(1)
	}
	complete := filepath.Join(m1Dir, name)
	if err := ioutil.WriteFile(complete, out, info.Mode()); err != nil {
		fmt.Printf("Could not write file: %v\n", err)
		os.Exit(1)
	}

	// Check duplicates (e.g. lu and list-updates)
	name = duplicateName(name)
	if name != "" {
		complete = filepath.Join(m1Dir, name)
		if err := ioutil.WriteFile(complete, out, info.Mode()); err != nil {
			fmt.Printf("Could not write file: %v\n", err)
			os.Exit(1)
		}
	}
}

func main() {
	// First of all, get all the files and create the directory in which man
	// pages will be saved.
	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Printf("Could not read directory: %v\n", err)
		os.Exit(1)
	}
	if err = os.MkdirAll(m1Dir, 0744); err != nil {
		fmt.Printf("Could not create man/%v directory: %v\n", m1Dir, err)
	}

	duplicates = map[string]string{
		"lu":    "list-updates",
		"luc":   "list-updates-container",
		"up":    "update",
		"lp":    "list-patches",
		"lpc":   "list-patches-container",
		"pchk":  "patch-check",
		"pchkc": "patch-check-container",
	}

	// Filter markdown files that actually contain a man page and generate it.
	re := regexp.MustCompile(`^([a-z\-]+\.1)\.md$`)
	for _, f := range files {
		matches := re.FindStringSubmatch(f.Name())
		if len(matches) == 2 {
			generateMan(f, matches[1])
		}
	}
}
