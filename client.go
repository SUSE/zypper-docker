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
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/mssola/dockerclient"
)

// dockerError encapsulates a dockerclient.WaitResult that has an exit status
// different than 0. This is done this way because, for some commands, zypper
// might set an exit code different than 0, even if there was no error. For
// example, the patch-check command can set the exit code 100, to determine
// that there are patches to be installed. In this case, the caller can decide
// what to do depending on the error returned by zypper. Therefore, the caller
// of functions such as `startContainer` should only care about this type if
// the command being implemented has this kind of behavior.
type dockerError struct {
	dockerclient.WaitResult
}

func (de dockerError) Error() string {
	return fmt.Sprintf("Command exited with status %d", de.ExitCode)
}

// The DockerClient interface lists all the functions that we use from Docker clients. Take
// a look at http://godoc.org/github.com/samalba/dockerclient if you want to
// read the documentation for each function.
type DockerClient interface {
	ListImages(all bool) ([]*dockerclient.Image, error)

	CreateContainer(config *dockerclient.ContainerConfig, name string) (string, error)
	StartContainer(id string, config *dockerclient.HostConfig) error
	RemoveContainer(id string, force, volume bool) error
	KillContainer(id, signal string) error

	Wait(id string) <-chan dockerclient.WaitResult

	ContainerLogs(id string, options *dockerclient.LogOptions) (io.ReadCloser, error)
}

// This global variable holds the instance to the docker client.
var dockerClient DockerClient

const (
	// The path in which the Docker client is listening to.
	dockerSocket = "unix:///var/run/docker.sock"

	// The timeout in which the container is allowed to run a command as given
	// to the `startContainer` function.
	containerTimeout = 2 * time.Second
)

// Use this function to safely retrieve the singleton instance of the Docker
// client. In order to guarantee such safety, the instance has to be
// initialized when no goroutines are being executed concurrently (e.g. the
// `init` or the `main` function).
func getDockerClient() DockerClient {
	if dockerClient != nil {
		return dockerClient
	}

	// We can safely discard the error. The connection will be started
	// successfully because internally `NewDockerClientTimeout` will handle the
	// connection as a dial for the http package. Therefore, it won't fail even
	// if the given URL does not exist. This is ok, since this possible error
	// will appear later on with subsequent commands.
	//
	// The only time it will return an error will be if the given URL has a bad
	// format, which won't happen.
	dockerClient, _ = dockerclient.NewDockerClientTimeout(dockerSocket, nil,
		containerTimeout)
	return dockerClient
}

// Looks for the specified command inside of a Docker image.
// The given image string is just the ID of said image.
// It returns true if the command was successful, false otherwise.
func checkCommandInImage(img, cmd string) bool {
	containerId, err := startContainer(img, []string{cmd}, false, false)

	defer removeContainer(containerId)

	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// Run the given command in a container based on the given image. The given
// image string is just the ID of said image.
// The STDOUT and STDERR of the container can be streamed to the host's STDOUT
// by setting the `streaming` parameter to true.
// It returns the ID of the container spawned from the image.
// Note well: the container is NOT deleted when the given command terminates.
func runCommandInContainer(img string, cmd []string, streaming bool) (string, error) {
	return startContainer(img, cmd, streaming, true)
}

// Start a container from the specified image and then runs the given command
// inside of it. The given image string is just the ID of said image.
// The STDOUT and STDERR of the container can be streamed to the host's STDOUT
// by setting the `streaming` parameter to true.
// When `wait` is set to true the function will wait untill the container exits,
// otherwise it will timeout raising an error.
// Note well: the container is NOT deleted when the given command terminates.
// This is again up to the caller.
//
// It returns the ID of the container spawned from the image. The error
// returned can be of type dockerError. This only happens when the container
// has run normally (no signals, no timeout), but the exit code is not 0. Read
// the documentation on the `dockerError` command on why we do this.
func startContainer(img string, cmd []string, streaming, wait bool) (string, error) {
	id, err := createContainer(img, cmd)
	if err != nil {
		log.Println(err)
		return "", err
	}

	client := getDockerClient()
	if err = client.StartContainer(id, &dockerclient.HostConfig{}); err != nil {
		// Silently fail, since it might be "zypper" not existing and we don't
		// want to add noise to the log.
		return id, err
	}

	sc := make(chan bool)

	if streaming {
		// setup logging
		rc, err := dockerClient.ContainerLogs(id, &dockerclient.LogOptions{
			Follow: true, // required to keep streaming
			Stdout: true,
			Stderr: true,
		})
		if err != nil {
			log.Println(err)
			return id, err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				log.Print(err)
			}
		}()
		go func() {
			if _, err := io.Copy(os.Stdout, rc); err != nil {
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
	case res := <-client.Wait(id):
		if streaming {
			<-sc
		}
		if res.Error != nil {
			return id, res.Error
		} else if res.ExitCode != 0 {
			return id, dockerError{res}
		}
	case <-timeout:
		return id, fmt.Errorf("Timed out when waiting for a container.\n")
	case <-killChannel:
		if err := client.KillContainer(id, "KILL"); err != nil {
			fmt.Println("Error while killing running container:", err)
		} else {
			removeContainer(id)
		}
		exitWithCode(1)
	}

	return id, nil
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
	config := &dockerclient.ContainerConfig{
		Image:        img,
		Entrypoint:   cmd,
		AttachStdout: true,
		AttachStderr: true,
		// required to avoid garbage when cmd overwrites the terminal
		// like "zypper ref" does
		Tty: true,
	}
	id, err := client.CreateContainer(config, "")
	if err != nil {
		return "", err
	}
	return id, nil
}

// Safely remove the given container. It will deal with the error by logging
// it.
func removeContainer(id string) {
	client := getDockerClient()
	if err := client.RemoveContainer(id, true, true); err != nil {
		log.Println(err)
	}
}
