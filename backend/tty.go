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
	"syscall"
	"unsafe"

	"github.com/SUSE/zypper-docker/logger"
	"github.com/docker/engine-api/types"
)

// Resize the TTY of the container with the given id to the size of the current
// TTY.
func resizeTty(id string) {
	height, width := getTtySize()
	if height == 0 && width == 0 {
		return
	}

	client := getDockerClient()
	err := client.ContainerResize(types.ResizeOptions{
		ID:     id,
		Height: height,
		Width:  width,
	})

	if err != nil {
		logger.Printf("could not resize container: %v", err)
	}
}

// Get the size of the current TTY. On error, the returned values will be zero
// and the error itself will be logged.
func getTtySize() (int, int) {
	size := &struct {
		Height uint16
		Width  uint16

		// Not needed but we may avoid some random crashes.
		x uint16
		y uint16
	}{}

	// Call ioctl, requesting the size of the window.
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, 1,
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(size)))

	if errno != 0 {
		return 0, 0
	}
	return int(size.Height), int(size.Width)
}
