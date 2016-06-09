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
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/SUSE/zypper-docker/backend/drivers"
	"github.com/SUSE/zypper-docker/logger"
	"github.com/SUSE/zypper-docker/utils"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/engine-api/types/strslice"
)

// rootUser is the explicit value to use for the USER directive to specify a user
// as being root. Oddly, specifying the default value ("") doesn't work even though
// images that have the default value set for .Config.User run as root.
const rootUser = "0:0"

// dockerError encapsulates a WaitResult that has an exit status
// different than 0. This is done this way because, for some commands, zypper
// might set an exit code different than 0, even if there was no error. For
// example, the patch-check command can set the exit code 100, to determine
// that there are patches to be installed. In this case, the caller can decide
// what to do depending on the error returned by zypper. Therefore, the caller
// of functions such as `startContainer` should only care about this type if
// the command being implemented has this kind of behavior.
type dockerError struct {
	waitResult
}

func (de dockerError) Error() string {
	return fmt.Sprintf("command exited with status %d", de.exitCode)
}

// DockerClient is an interface listing all the functions that we use from
// Docker clients.
type DockerClient interface {
	ContainerCommit(options types.ContainerCommitOptions) (types.ContainerCommitResponse, error)
	ContainerCreate(config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (types.ContainerCreateResponse, error)
	ContainerKill(containerID, signal string) error
	ContainerList(options types.ContainerListOptions) ([]types.Container, error)
	ContainerLogs(options types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerRemove(options types.ContainerRemoveOptions) error
	ContainerResize(options types.ResizeOptions) error
	ContainerStart(id string) error
	ContainerWait(containerID string) (int, error)

	ImageInspectWithRaw(imageID string, getSize bool) (types.ImageInspect, []byte, error)
	ImageList(options types.ImageListOptions) ([]types.Image, error)
}

// The timeout in which the container is allowed to run a command as given
// to the `startContainer` function.
const containerTimeout = 15 * time.Second

// safeClient holds the instance of the docker client and the mutex protecting
// it from concurrent accesses. Do *not* use this global variable directly to
// fetch the docker client. For that use the `getDockerClient` function.
var safeClient struct {
	sync.Mutex
	client DockerClient
}

// getDockerClient safely returns the singleton instance of the Docker client.
func getDockerClient() DockerClient {
	safeClient.Lock()
	defer safeClient.Unlock()

	if safeClient.client != nil {
		return safeClient.client
	}

	if dc, err := client.NewEnvClient(); err != nil {
		logger.Fatalf("Could not get a docker client: %v", err)
	} else {
		safeClient.client = dc
		return dc
	}
	return nil
}

// runStreamedCommand is a convenient wrapper of the `runCommandInContainer`
// for functions that just need to run a command on streaming without the
// burden of removing the resulting container, etc.
func runStreamedCommand(img, cmd string) error {
	if img == "" {
		return errors.New("no image name specified")
	}

	id, err := runCommandInContainer(img, []string{cmd}, os.Stdout)
	removeContainer(id)
	return err
}

// Run the given command in a container based on the given image. The given
// image string is just the ID of said image.
// The STDOUT and STDERR of the container can be streamed by providing a
// destination in `dst`.
// It returns the ID of the container spawned from the image.
// Note well: the container is NOT deleted when the given command terminates.
func runCommandInContainer(img string, cmd []string, dst io.Writer) (string, error) {
	return startContainer(img, cmd, true, dst)
}

// Start a container from the specified image and then runs the given command
// inside of it. The given image string is just the ID of said image.
// When `wait` is set to true the function will wait untill the container exits,
// otherwise it will timeout raising an error.
// The STDOUT and STDERR of the container can be streamed by providing a
// destination in `dst`.
// Note well: the container is NOT deleted when the given command terminates.
// This is again up to the caller.
//
// It returns the ID of the container spawned from the image. The error
// returned can be of type dockerError. This only happens when the container
// has run normally (no signals, no timeout), but the exit code is not 0. Read
// the documentation on the `dockerError` command on why we do this.
func startContainer(img string, cmd []string, wait bool, dst io.Writer) (string, error) {
	id, err := createContainer(img, cmd)
	if err != nil {
		return "", err
	}

	client := getDockerClient()
	if err = client.ContainerStart(id); err != nil {
		// Silently fail, since it might be "zypper" not existing and we don't
		// want to add noise to the log.
		return id, err
	}
	resizeTty(id)

	sc := make(chan bool)

	if dst != nil {
		// setup logging
		rc, err := client.ContainerLogs(types.ContainerLogsOptions{
			ContainerID: id,
			Follow:      true, // required to keep streaming
			ShowStdout:  true,
			ShowStderr:  true,
		})
		if err != nil {
			return id, err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				log.Print(err)
			}
		}()
		go func() {
			if _, err := io.Copy(dst, rc); err != nil {
				log.Print(err)
			}
			sc <- true
		}()
	}

	timeout := make(chan int)
	go func() {
		if !wait {
			time.Sleep(containerTimeout)
			timeout <- 1
		}
	}()

	select {
	case res := <-containerWait(id):
		if dst != nil {
			<-sc
		}
		if res.err != nil {
			return id, res.err
		} else if res.exitCode != 0 {
			return id, dockerError{res}
		}
	case <-timeout:
		return id, fmt.Errorf("timed out when waiting for a container")
	case <-KillChannel:
		if err := client.ContainerKill(id, "KILL"); err != nil {
			fmt.Println("error while killing running container:", err)
		} else {
			removeContainer(id)
		}
		utils.ExitWithCode(1)
	}

	return id, nil
}

// waitResult encapsulates the result of the client.ContainerWait function.
// Defined for the convenience of the containerWait function.
type waitResult struct {
	exitCode int
	err      error
}

// containerWait is a tiny wrapper on top of the client.ContainerWait function
// so it returns a channel instead.
func containerWait(id string) chan waitResult {
	wr := make(chan waitResult)
	client := getDockerClient()

	go func() {
		code, err := client.ContainerWait(id)
		wr <- waitResult{exitCode: code, err: err}
	}()
	return wr
}

// Creates a container based on the given image. The given image string is just
// the ID of said image. The command specified is set as the entry point of the
// container.
// It returns the ID of the spawned container when successful, nil otherwise.
// The error is set accordingly when it's not possible to create the container.
// Note well: the container is not running at this time, it must be started via
// the `startContainer` function.
func createContainer(img string, cmd []string) (string, error) {
	client := getDockerClient()

	// First of all we create a container in which we will run the command.
	config := &container.Config{
		Image:        img,
		Cmd:          strslice.New(cmd...),
		Entrypoint:   strslice.New("/bin/sh", "-c"),
		AttachStdout: true,
		AttachStderr: true,
		// We need to run as root in order to run zypper commands.
		User: rootUser,
		// required to avoid garbage when cmd overwrites the terminal
		// like "zypper ref" does
		Tty: true,
	}
	resp, err := client.ContainerCreate(config, drivers.GetHostConfig(), nil, "")
	if err != nil {
		return "", err
	}

	for _, warning := range resp.Warnings {
		log.Print(warning)
	}
	return resp.ID, nil
}

// Safely remove the given container. It will deal with the error by logging
// it.
func removeContainer(id string) {
	client := getDockerClient()

	err := client.ContainerRemove(types.ContainerRemoveOptions{
		ContainerID:   id,
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		log.Println(err)
	}
}

// commitContainerToImage commits the container with the given containerID
// that is based on the given img into a new image. The given repo should also
// contain the namespace. Returns the id of the created image.
func commitContainerToImage(img, containerID, repo, tag, comment, author string) (string, error) {
	client := getDockerClient()

	// First of all, we inspect the parent image and fetch the values for the
	// entrypoint and the cmd. We do this to preserve them on the committed
	// image. See issue: https://github.com/SUSE/zypper-docker/issues/75.
	info, _, err := client.ImageInspectWithRaw(img, false)
	if err != nil {
		return "", fmt.Errorf("could not inspect image '%s': %v", img, err)
	}

	user := info.Config.User
	if user == "" {
		// While images what have .Config.User set to "" run as root, we cannot
		// explicity set it to this value. Instead, just set it to rootUser.
		user = rootUser
	}

	changes := []string{
		"USER " + user,
		"ENTRYPOINT " + utils.JoinAsArray(info.Config.Entrypoint.Slice(), false),
		"CMD " + utils.JoinAsArray(info.Config.Cmd.Slice(), true),
	}

	// And we commit into the new image.
	resp, err := client.ContainerCommit(types.ContainerCommitOptions{
		ContainerID:    containerID,
		RepositoryName: repo,
		Tag:            tag,
		Comment:        comment,
		Author:         author,
		Changes:        changes,
		Config:         &container.Config{},
	})
	return resp.ID, err
}

// CheckContainer looks for the specified running container and makes sure it's
// running on a supported driver.
func CheckContainer(id string) (types.Container, error) {
	client := getDockerClient()
	var container types.Container

	containers, err := client.ContainerList(types.ContainerListOptions{})
	if err != nil {
		return container, fmt.Errorf("error while fetching running containers: %v", err)
	}

	found := false
	for _, c := range containers {
		if id == c.ID {
			container = c
			found = true
			break
		}
		// look also for the short version of the container ID
		if len(id) >= 12 && strings.Index(c.ID, id) == 0 {
			container = c
			found = true
			break
		}
		// for some reason the daemon has all the names prefixed by "/"
		if utils.ArrayIncludeString(c.Names, "/"+id) {
			container = c
			found = true
			break
		}
	}

	if !found {
		return container, fmt.Errorf("cannot find running container: %s", id)
	}

	if !isSupported(container.Image) {
		return container, fmt.Errorf(
			"the container %s is based on the Docker image %s which is not supported",
			id, container.Image)
	}
	return container, nil
}
