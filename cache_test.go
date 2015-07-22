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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// NOTE: some functions are already covered in other places of this test suite,
// so there's no point to add more tests in this specific file.

func TestCachePath(t *testing.T) {
	cache, data := os.Getenv("XDG_CACHE_HOME"), os.Getenv("XDG_DATA_DIRS")
	umask := syscall.Umask(0)

	defer func() {
		syscall.Umask(umask)
		_ = os.Setenv("XDG_CACHE_HOME", cache)
		_ = os.Setenv("XDG_DATA_DIRS", data)
	}()

	_ = os.Setenv("XDG_CACHE_HOME", "")
	abs, _ := filepath.Abs(".")
	_ = os.Setenv("XDG_DATA_DIRS", abs)

	file := cachePath()
	if file == nil {
		t.Fatal("The given file should be ok")
	}
	info, err := file.Stat()
	if err != nil {
		t.Fatal("I should be able to stat the given file")
	}
	name, mode := file.Name(), info.Mode().Perm()
	_ = file.Close()
	_ = os.Remove(name)
	if name != filepath.Join(abs, cacheName) {
		t.Fatal("Unexpected name")
	}
	if mode != 0666 {
		t.Fatal("Given file does not come from hell ;)")
	}
}

func TestCachePathFail(t *testing.T) {
	cache, data := os.Getenv("XDG_CACHE_HOME"), os.Getenv("XDG_DATA_DIRS")
	home := os.Getenv("HOME")
	umask := syscall.Umask(0)

	defer func() {
		syscall.Umask(umask)
		_ = os.Setenv("XDG_CACHE_HOME", cache)
		_ = os.Setenv("XDG_DATA_DIRS", data)
		_ = os.Setenv("HOME", home)
	}()

	_ = os.Setenv("XDG_CACHE_HOME", "")
	_ = os.Setenv("XDG_DATA_DIRS", "")
	_ = os.Setenv("HOME", "")
	file, _ := os.OpenFile(filepath.Join("/tmp", cacheName), os.O_RDONLY|os.O_CREATE, 0000)
	_ = file.Close()

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	cacheFile := getCacheFile()
	if cacheFile.Valid {
		t.Fatal("Cache should not be valid")
	}
	if !strings.Contains(buffer.String(), "Could not find path for the cache!") {
		t.Fatal("Wrong log")
	}
}

func TestCacheBadJson(t *testing.T) {
	cache := os.Getenv("XDG_CACHE_HOME")
	abs, _ := filepath.Abs(".")
	test := filepath.Join(abs, "test")

	defer func() {
		_ = os.Setenv("XDG_CACHE_HOME", cache)
		_ = os.Rename(filepath.Join(test, cacheName), filepath.Join(test, "bad.json"))
	}()

	_ = os.Setenv("XDG_CACHE_HOME", test)
	_ = os.Rename(filepath.Join(test, "bad.json"), filepath.Join(test, cacheName))

	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	file := getCacheFile()
	if !file.Valid {
		t.Fatal("It should be valid")
	}
	if !strings.Contains(buffer.String(), "Cache file has a bad format!") {
		t.Fatal("Wrong log")
	}
}

func TestCacheGoodJson(t *testing.T) {
	cache := os.Getenv("XDG_CACHE_HOME")
	abs, _ := filepath.Abs(".")
	test := filepath.Join(abs, "test")

	defer func() {
		_ = os.Setenv("XDG_CACHE_HOME", cache)
		_ = os.Rename(filepath.Join(test, cacheName), filepath.Join(test, "ok.json"))
	}()

	_ = os.Setenv("XDG_CACHE_HOME", test)
	_ = os.Rename(filepath.Join(test, "ok.json"), filepath.Join(test, cacheName))

	file := getCacheFile()
	if !file.Valid {
		t.Fatal("It should be valid")
	}
	if file.Path != filepath.Join(test, cacheName) {
		t.Fatal("Wrong path")
	}
	elements := append(file.Suse, file.Other[0], file.Other[1])
	for i, value := range elements {
		if value != fmt.Sprintf("%v", i+1) {
			t.Fatal("Wrong value")
		}
	}
}

func TestFlush(t *testing.T) {
	abs, _ := filepath.Abs(".")
	test := filepath.Join(abs, "test")
	path := filepath.Join(test, "testflush.json")

	cd := &cachedData{
		Path:  path,
		Valid: false,
		Suse:  []string{},
		Other: []string{},
	}

	// Now put some contents there.
	err := ioutil.WriteFile(path, []byte("{\"suse\": [\"1\"], \"other\": []}"), 0666)
	if err != nil {
		t.Fatal("Failed on writing a file")
	}

	// It's invalid, flush will do nothing.
	cd.flush()

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal("Failed on reading a file")
	}
	if strings.TrimSpace(string(contents)) != "{\"suse\": [\"1\"], \"other\": []}" {
		t.Fatal("Wrong contents")
	}

	// Now it will overwrite the file.
	cd.Valid = true
	cd.flush()
	contents, err = ioutil.ReadFile(path)
	if err != nil {
		t.Fatal("Failed on reading a file")
	}
	if strings.TrimSpace(string(contents)) != "{\"suse\":[],\"other\":[]}" {
		t.Fatal("Wrong contents")
	}

	// If we remove the file and try to access it, it will print a proper
	// error.
	buffer := bytes.NewBuffer([]byte{})
	log.SetOutput(buffer)
	if err := os.Remove(path); err != nil {
		t.Fatal("Could not remove temporary file")
	}
	cd.flush()
	if !strings.Contains(buffer.String(), "Cannot write to the cache file") {
		t.Fatal("Didn't logged what it was expected")
	}
}
