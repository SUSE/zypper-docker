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
	"fmt"

	"github.com/SUSE/zypper-docker/backend/drivers"
)

// UpdateKind represents the kind of update to be executed.
type UpdateKind int

const (
	// General represents a general update.
	General UpdateKind = iota

	// Security represents a security update.
	Security
)

func uniqueUpdatedName(image string) (string, string, error) {
	repo, tag, err := parseImageName(image)
	if err != nil {
		return "", "", err
	}
	if err = preventImageOverwrite(repo, tag); err != nil {
		return "", "", err
	}
	return repo, tag, nil
}

func fetchCommand(kind UpdateKind) (string, error) {
	if kind == General {
		return drivers.Current().GeneralUpdate()
	} else if kind == Security {
		return drivers.Current().SecurityUpdate()
	}

	// TODO: in the future this will be meant for those backends which
	// don't support patching.
	return drivers.Current().GeneralUpdate()
}

// PerformUpdate performs an update operation to the given `original` image and
// saves it into the given `dest` new image. This function will prevent clients
// to overwrite an existing image.
func PerformUpdate(kind UpdateKind, original, dest, comment, author string) (string, string, error) {
	repo, tag, err := uniqueUpdatedName(dest)
	if err != nil {
		return "", "", err
	}

	cmd, err := fetchCommand(kind)
	if err != nil {
		return "", "", err
	}

	newImgID, err := runCommandAndCommitToImage(original, repo, tag, cmd, comment, author)
	if err != nil {
		return "", "", err
	}

	cache := getCacheFile()
	if err := cache.updateCacheAfterUpdate(original, newImgID); err != nil {
		return "", "", fmt.Errorf("failed to write to cache: %v", err)
	}
	return repo, tag, nil
}

// ListUpdates lists the updates available for the given image.
func ListUpdates(kind UpdateKind, image string) error {
	var cmd string
	var err error

	if kind == Security {
		cmd, err = drivers.Current().ListSecurityUpdates()
	} else {
		cmd, err = drivers.Current().ListGeneralUpdates()
	}
	if err != nil {
		return err
	}
	return runStreamedCommand(image, cmd)
}

// HasPatches returns true if the given image has pending patches.
// TODO: improve with a "Severity" return value or something
func HasPatches(image string) (bool, bool, error) {
	cmd, err := drivers.Current().CheckPatches()
	if err != nil {
		return false, false, err
	}

	err = runStreamedCommand(image, cmd)
	if err == nil {
		return false, false, nil
	}

	switch err.(type) {
	case dockerError:
		// According to zypper's documentation:
		// 	100 - There are patches available for installation.
		// 	101 - There are security patches available for installation.
		de := err.(dockerError)
		if de.exitCode == 100 {
			return true, false, nil
		} else if de.exitCode == 101 {
			return false, true, nil
		}
	}
	// TODO: nope!
	humanizeCommandError("zypper pchk", image, err)
	return false, false, err
}
