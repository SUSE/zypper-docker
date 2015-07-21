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
	"log"
	"time"

	"github.com/mssola/dockerclient"
)

// This interface lists all the functions that we use from Docker clients. Take
// a look at http://godoc.org/github.com/samalba/dockerclient if you want to
// read the documentation for each function.
type DockerClient interface {
	ListImages(all bool) ([]*dockerclient.Image, error)

	CreateContainer(config *dockerclient.ContainerConfig, name string) (string, error)
	StartContainer(id string, config *dockerclient.HostConfig) error
	RemoveContainer(id string, force, volume bool) error

	Wait(id string) <-chan dockerclient.WaitResult
}

// This global variable holds the instance to the docker client.
var dockerClient DockerClient

const (
	// The path in which the Docker client is listening to.
	dockerSocket = "unix:///var/run/docker.sock"

	// The timeout in which the container is allowed to run a command as given
	// to the `runCommandInContainer` function.
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

// Run the given command in a container based on the given image. The given
// image string is just the ID of said image. It returns true if the command
// was successful, false otherwise.
func runCommandInContainer(img string, cmd []string) bool {
	client := getDockerClient()

	// First of all we create a container in which we will run the command.
	config := &dockerclient.ContainerConfig{Image: img, Entrypoint: cmd}
	id, err := client.CreateContainer(config, "")
	if err != nil {
		log.Println(err)
		return false
	}
	defer removeContainer(client, id)

	// Second step: start the container, wait for it to finish and return the
	// results.

	err = client.StartContainer(id, &dockerclient.HostConfig{})
	if err != nil {
		// Silently fail, since it might be "zypper" not existing and we don't
		// want to add noise to the log.
		return false
	}

	select {
	case res := <-client.Wait(id):
		if res.Error != nil {
			log.Println(res.Error)
		} else {
			return res.ExitCode == 0
		}
	case <-time.After(containerTimeout):
		log.Printf("Timed out when waiting for a container.\n")
	}
	return false
}

// Safely remove the given container. It will deal with the error by logging
// it.
func removeContainer(client DockerClient, id string) {
	if err := client.RemoveContainer(id, true, true); err != nil {
		log.Println(err)
	}
}
