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

// TODO: this package assumes that drivers are all CLI based, which is not
// necessarily true. We should use the `drivers.needsCLI` function and act
// accordingly.

// Initialize initializes the backend of zypper-docker.
func Initialize() {
	listenSignals()

	// TODO: available backends and so on
}

func isSupported(image string) bool {
	// TODO: improve once we have more drivers.
	cache := getCacheFile()
	return cache.isSUSE(image)
}
