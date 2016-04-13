// Copyright (C) 2015-2016 Miquel Sabaté Solà <mikisabate@gmail.com>
// This file is licensed under the MIT license.
// See the LICENSE file.

// Package capture safely captures the stdout and stderr of some given code.
package capture

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Result stores what has been captured by the `All` function.
type Result struct {
	// The captured output from stdout.
	Stdout []byte

	// The captured output from stderr.
	Stderr []byte

	// The value of the error when something wrong happens.
	Error error
}

// Close and remove the given open file.
func cleanup(files ...*os.File) {
	for _, file := range files {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}
}

// Setup the mocking files for the stdout and the stderr.
func setupFiles() (*os.File, *os.File) {
	outFile, _ := ioutil.TempFile("", "outputmock")
	errFile, _ := ioutil.TempFile("", "errormock")
	return outFile, errFile
}

// All safely captures all the output from both the stdout and the stderr for
// the given function. This function will recover from any panic that occurres
// inside the given function.
func All(f func()) (captured *Result) {
	var oldStdout, oldStderr *os.File

	// Setup the thing.
	captured = &Result{}
	outFile, errFile := setupFiles()

	// It recovers from panics and it tears down all the temporary stuff that
	// has been created inside this function.
	defer func() {
		if r := recover(); r != nil {
			captured.Error = fmt.Errorf("Panic: %v", r)
		}
		cleanup(outFile, errFile)
		os.Stdout, os.Stderr = oldStdout, oldStderr
	}()

	// Exchange the output channels and call the function.
	oldStdout, oldStderr = os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outFile, errFile
	f()
	captured.Stdout, _ = ioutil.ReadFile(outFile.Name())
	captured.Stderr, _ = ioutil.ReadFile(errFile.Name())
	return captured
}
