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

package utils

// JoinAsArray joins the given array of commands so it's compatible to what is
// expected from a dockerfile syntax.
func JoinAsArray(cmds []string, emptyArray bool) string {
	if emptyArray && len(cmds) == 0 {
		return ""
	}

	str := "["
	for i, v := range cmds {
		str += "\"" + v + "\""
		if i < len(cmds)-1 {
			str += ", "
		}
	}
	return str + "]"
}

// ArrayIncludeString returns whether the given string is in the given array.
func ArrayIncludeString(arr []string, s string) bool {
	for _, i := range arr {
		if i == s {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate entries from an array of strings. Should
// the resulting array be empty, it does not return nil but an empty array.
func RemoveDuplicates(elements []string) []string {
	seen := make(map[string]bool)
	var res []string

	for _, v := range elements {
		if seen[v] {
			continue
		} else {
			seen[v] = true
			res = append(res, v)
		}
	}

	// make sure not to return nil
	if res == nil {
		return []string{}
	}

	return res
}

// ExitWithCode is a function that has to be set when initializing the program.
// This is done so it can be configured in tests, but the usual implementation
// consists of just calling os.Exit on the given code.
var ExitWithCode func(code int)
