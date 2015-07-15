package main

import (
	"log"

	"github.com/samalba/dockerclient"
)

// This interface lists all the functions that we use from Docker clients.
type DockerClient interface {
	ListImages(all bool) ([]*dockerclient.Image, error)
	InspectImage(id string) (*dockerclient.ImageInfo, error)
}

var dockerClient DockerClient

const (
	dockerSocket = "unix:///var/run/docker.sock"
)

func getDockerClient() DockerClient {
	if dockerClient != nil {
		return dockerClient
	}

	// TODO: (mssola) tls client
	dockerClient, err := dockerclient.NewDockerClient(dockerSocket, nil)
	if err != nil {
		log.Fatalf("client: Could not connect to Docker!\n")
	}
	return dockerClient
}
