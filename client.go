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
	"log"
	"time"

	"github.com/samalba/dockerclient"
)

// This interface lists all the functions that we use from Docker clients. Take
// a look at http://godoc.org/github.com/samalba/dockerclient if you want to
// read the documentation for each function.
type DockerClient interface {
	ListImages(all bool) ([]*dockerclient.Image, error)

	CreateContainer(config *dockerclient.ContainerConfig, name string) (string, error)
	StartContainer(id string, config *dockerclient.HostConfig) error
	RemoveContainer(id string, force, volume bool) error
	InspectContainer(id string) (*dockerclient.ContainerInfo, error)

	StartMonitorEvents(cb dockerclient.Callback, ec chan error, args ...interface{})
	StopAllMonitorEvents()
}

var (
	// This global variable holds the instance to the docker client.
	dockerClient DockerClient

	// This map contains all the containers that are waiting for an event. The
	// key for each entry is the ID of the container, and the value is a
	// channel that will be supplied when an event has occurred.
	containers map[string]chan bool = make(map[string]chan bool)
)

const (
	// The path in which the Docker client is listening to.
	dockerSocket = "unix:///var/run/docker.sock"

	// The name of the temporary container created in the
	// `runCommandInContainer` function.
	temporaryName = "zypper-docker-private"

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
//
// TODO: (mssola) this function is using the monitoring API, which is a bit of
// an overkill. We should be using the `Wait` endpoint from the API. This can
// be done once the PR https://github.com/samalba/dockerclient/pull/140 gets
// merged.
func runCommandInContainer(img string, cmd []string) bool {
	client := getDockerClient()

	// First of all we create a container in which we will run the command.
	config := &dockerclient.ContainerConfig{Image: img, Cmd: cmd}
	name := fmt.Sprintf("%s-%s", temporaryName, img)
	id, err := client.CreateContainer(config, name)
	if err != nil {
		log.Println(err)
		return false
	}
	defer removeContainer(client, id)

	// Second step: start the container and wait for an event in it.

	err = client.StartContainer(id, &dockerclient.HostConfig{})
	if err != nil {
		// Silently fail, since it might be "zypper" not existing and we don't
		// want to add noise to the log.
		return false
	}

	containers[id] = make(chan bool, 0)
	errors := make(chan error)
	client.StartMonitorEvents(
		func(event *dockerclient.Event, ec chan error, args ...interface{}) {
			if event.Status == "die" || event.Status == "destroy" {
				containers[event.Id] <- true
			}
		}, errors)

	select {
	case <-containers[id]:
		// Event received, remove it from the queue.
		delete(containers, id)
	case errStr := <-errors:
		log.Printf("%v\n", errStr)
	case <-time.After(containerTimeout):
		log.Printf("Timed out when waiting for a container.\n")
	}
	client.StopAllMonitorEvents()

	// Finally, we check for the exit code as given by the command and the
	// temporary container gets removed.
	info, err := client.InspectContainer(id)
	if err != nil {
		log.Println(err)
		return false
	}
	return info.State.ExitCode == 0
}

// Safely remove the given container. It will deal with the error by logging
// it.
func removeContainer(client DockerClient, id string) {
	if err := client.RemoveContainer(id, true, true); err != nil {
		log.Println(err)
	}
}
