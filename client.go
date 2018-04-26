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
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// zypperExitCode is used as zypper-docker's exit code.
var zypperExitCode int64

// rootUser is the explicit value to use for the USER directive to specify a user
// as being root. Oddly, specifying the default value ("") doesn't work even though
// images that have the default value set for .Config.User run as root.
const rootUser = "0:0"

// dockerError encapsulates an exit status that has a value different than 0.
// This is done this way because, for some commands, zypper might set an exit
// code different than 0, even if there was no error. For example, the patch-check
// command can set the exit code 100, to determine that there are patches to be
// installed. In this case, the caller can decide what to do depending on the
// error returned by zypper. Therefore, the caller of functions such as
// `startContainer` should only care about this type if the command being
// implemented has this kind of behavior.
type dockerError struct {
	exitCode int64
	err      error
}

func (de dockerError) Error() string {
	return fmt.Sprintf("Command exited with status %d", de.exitCode)
}

// DockerClient is an interface listing all the functions that we use from
// Docker clients.
type DockerClient interface {
	ContainerCommit(ctx context.Context, container string, options types.ContainerCommitOptions) (types.IDResponse, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerKill(ctx context.Context, containerID, signal string) error
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
	ContainerResize(ctx context.Context, containerID string, options types.ResizeOptions) error
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error)

	ImageInspectWithRaw(ctx context.Context, imageID string) (types.ImageInspect, []byte, error)
	ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)
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
		log.Printf("Could not get a docker client: %v", err)
	} else {
		safeClient.client = dc
		return dc
	}

	// The return statement is just to make golint happy about this and for
	// compliance with the API.
	exitWithCode(1)
	return nil
}

// humanizeCommandError tries to print an explicit and useful message for a
// failing command inside of a docker container (based on the given image).
func humanizeCommandError(cmd, image string, err error) {
	var reason string

	switch err.(type) {
	case dockerError:
		de := err.(dockerError)
		if de.exitCode == 127 {
			reason = "command not found"
		} else {
			reason = err.Error()
		}
	default:
		reason = err.Error()
	}

	msg := "Could not execute command '%s' successfully in image '%s': %v.\n"
	log.Printf(msg, cmd, image, reason)
}

// Looks for the specified command inside of a Docker image.
// The given image string is just the ID of said image, which is
// then passed to startContainer, which starts the container.
// It returns true if the command was successful, false otherwise.
func checkCommandInImage(img, cmd string) bool {

	containerID, err := createContainer(img, []string{cmd})
	containerID, err = startContainer(containerID, false, nil)

	defer removeContainer(containerID)

	if err != nil {
		humanizeCommandError(cmd, img, err)
		return false
	}
	return true
}

// runStreamedCommand is a convenient wrapper of the `runCommandInContainer`
// for functions that just need to run a command on streaming without the
// burden of removing the resulting container, etc.
//
// The image has to be provided, otherwise this function will exit with 1 as
// the status code and it will both log and print that no image was provided.
// The given command will be executed as "zypper ref && zypper <command>".
//
// If getError is set to false, then this function will always return nil.
// Otherwise, it will return the error as given by the `runCommandInContainer`
// function.
func runStreamedCommand(img, cmd string, getError bool) error {
	if img == "" {
		logAndFatalf("Error: no image name specified.\n")
		return nil
	}

	cmd = formatZypperCommand("ref", cmd)
	id, err := runCommandInContainer(img, []string{cmd}, os.Stdout)
	removeContainer(id)

	if getError {
		return err
	}
	if err != nil {
		log.Printf("Error: %s\n", err)
		fmt.Println(err)
		exitWithCode(1)
	}
	return nil
}

// Run the given command in a container based on the given image. The given
// image string is just the ID of said image.
// The STDOUT and STDERR of the container can be streamed by providing a
// destination in `dst`.
// It returns the ID of the container spawned from the image.
// Note well: the container is NOT deleted when the given command terminates.
func runCommandInContainer(img string, cmd []string, dst io.Writer) (string, error) {
	id, err := createContainer(img, cmd)
	if err != nil {
		log.Println(err)
		return "", err
	}

	for i := 0; i < 16; i++ {
		id, err = startContainer(id, true, dst)
		switch err.(type) {
		case dockerError:
			de := err.(dockerError)
			if isZypperExitCodeSevere(int(de.exitCode)) {
				return "", err
			}
			if de.exitCode == zypperExitInfRestartNeeded {
				continue
			}
		default:
		}
		return id, err
	}
	return "", err
}

// Start a container from the specified ID and then runs the given command inside of it.
// The given ID string is just the ID of the container created by createContainer.
// When `wait` is set to true the function will wait untill the container exits,
// otherwise it will timeout raising an error.
// The STDOUT and STDERR of the container can be streamed by providing a
// destination in `dst`.
// Note well: the container is NOT deleted when the given command terminates.
// This is again up to the caller.
//
// It returns the ID of the started container. The error
// returned can be of type dockerError. This only happens when the container
// has run normally (no signals, no timeout), but the exit code is not 0. Read
// the documentation on the `dockerError` command on why we do this.
func startContainer(containerID string, wait bool, dst io.Writer) (string, error) {

	client := getDockerClient()
	if err := client.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{}); err != nil {
		// Silently fail, since it might be "zypper" not existing and we don't
		// want to add noise to the log.
		return containerID, err
	}
	resizeTty(containerID)

	sc := make(chan bool)

	if dst != nil {
		// setup logging
		rc, err := client.ContainerLogs(context.Background(), containerID, types.ContainerLogsOptions{
			Follow:     true, // required to keep streaming
			ShowStdout: true,
			ShowStderr: true,
		})
		if err != nil {
			log.Println(err)
			return containerID, err
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

	var waitErr error
	ctx := context.Background()
	if !wait {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, containerTimeout)
		defer cancel()
	}

	statusCh, errCh := client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	select {
	case <-ctx.Done():
		return containerID, fmt.Errorf("Timed out when waiting for a container")

	case <-killChannel:
		if err := client.ContainerKill(context.Background(), containerID, "KILL"); err != nil {
			fmt.Println("Error while killing running container:", err)
		} else {
			removeContainer(containerID)
		}
		exitWithCode(1)

	case err := <-errCh:
		waitErr = err

	case exitCode := <-statusCh:
		zypperExitCode = exitCode.StatusCode
		if dst != nil {
			<-sc
		}
	}
	if waitErr != nil {
		return containerID, waitErr
	} else if zypperExitCode != 0 {
		return containerID, dockerError{
			exitCode: zypperExitCode,
			err:      nil}
	}
	return containerID, waitErr
}

// Get the host config to be used for starting containers.
func getHostConfig() *container.HostConfig {
	if currentContext == nil {
		return &container.HostConfig{}
	}
	return &container.HostConfig{ExtraHosts: currentContext.GlobalStringSlice("add-host")}
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
		Cmd:          cmd,
		Entrypoint:   []string{"/bin/sh", "-c"},
		AttachStdout: true,
		AttachStderr: true,
		// We need to run as root in order to run zypper commands.
		User: rootUser,
		// required to avoid garbage when cmd overwrites the terminal
		// like "zypper ref" does
		Tty: true,
	}
	resp, err := client.ContainerCreate(context.Background(), config, getHostConfig(), nil, "")
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
func removeContainer(containerID string) {
	client := getDockerClient()

	err := client.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{
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
	info, _, err := client.ImageInspectWithRaw(context.Background(), img)
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
		"ENTRYPOINT " + joinAsArray(info.Config.Entrypoint, false),
		"CMD " + joinAsArray(info.Config.Cmd, true),
	}

	// And we commit into the new image.
	resp, err := client.ContainerCommit(context.Background(), containerID, types.ContainerCommitOptions{
		Reference: repo + ":" + tag,
		Comment:   comment,
		Author:    author,
		Changes:   changes,
		Config:    &container.Config{},
	})
	return resp.ID, err
}

// Spawns a container from the specified image, runs the specified command inside
// of it and commits the results to a new image.
// The name of the new image is specified via target_repo and target_tag.
// The container is always deleted.
// If something goes wrong an error message is returned.
// Returns the ID of the new image on success.
func runCommandAndCommitToImage(img, targetRepo, targetTag, cmd, comment, author string) (string, error) {
	containerID, err := runCommandInContainer(img, []string{cmd}, os.Stdout)
	if err != nil {
		return "", err
	}

	imageID, err := commitContainerToImage(img, containerID, targetRepo, targetTag, comment, author)

	// always remove the container
	removeContainer(containerID)

	return imageID, err
}

// Looks for the specified running container and makes sure it's running either
// SUSE or openSUSE.
func checkContainerRunning(id string) (types.Container, error) {
	client := getDockerClient()
	var container types.Container

	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return container, fmt.Errorf("Error while fetching running containers: %v", err)
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
		if arrayIncludeString(c.Names, "/"+id) {
			container = c
			found = true
			break
		}
	}

	if !found {
		return container, fmt.Errorf("Cannot find running container: %s", id)
	}

	cache := getCacheFile()
	if !cache.isSUSE(container.Image) {
		return container, fmt.Errorf(
			"The container %s is based on the Docker image %s which is not a SUSE system",
			id, container.Image)
	}

	return container, nil
}
