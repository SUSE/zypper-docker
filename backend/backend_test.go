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

package backend

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

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

type closingBuffer struct {
	*bytes.Buffer
}

func (cb *closingBuffer) Close() error {
	return nil
}
