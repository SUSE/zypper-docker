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
	"log"

	"github.com/docker/engine-api/types"
)

// ContainerWithState adds a message to the types.Container type. This way we
// can have more information on the status of the container.
type ContainerWithState struct {
	Container types.Container
	Message   string
}

// ContainersState is a struct representing the state of all the running
// containers. There are three states:
//	1. Containers that have an updated image.
//	2. Containers that have been ignored (see `ContainerWithState`)
//  3. Containers that have an unknown state (see `ContainerWithState`)
type ContainersState struct {
	Updated []types.Container
	Ignored []ContainerWithState
	Unknown []ContainerWithState
}

// addNotAnalyzed adds a container that could not be analyzed to the unknown
// field with the proper message.
func (st *ContainersState) addNotAnalyzed(container types.Container) {
	st.Unknown = append(st.Unknown, ContainerWithState{
		Container: container,
		Message:   "container could not be analyzed",
	})
}

// addUnknown adds a container that has an unknown state.
func (st *ContainersState) addUnknown(container types.Container) {
	st.Unknown = append(st.Unknown, ContainerWithState{
		Container: container,
		Message:   "container has an unknown state",
	})
}

// addUnsupported adds a container that has an unsupported backend.
func (st *ContainersState) addUnsupported(container types.Container, reason string) {
	st.Ignored = append(st.Ignored, ContainerWithState{
		Container: container,
		Message:   fmt.Sprintf("container has no supported backend: %v", reason),
	})
}

// ListContainers fetches all the running containers in the system. If there
// are no running containers, then both returned values are nil. The `ignore`
// parameter will tell this function whether or not to ignore failures when
// inspecting containers. Note that signals are already being handled
// gracefully by listening to the `KillChannel` channel.
func ListContainers(ignore bool) (*ContainersState, error) {
	client := getDockerClient()
	containers, err := client.ContainerList(types.ContainerListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not fetch running containers: %v", err)
	}

	if len(containers) == 0 {
		return nil, nil
	}

	return inspectContainers(containers, ignore)
}

// inspectContainers inspects the given containers. The `ignore` parameter
// tells this function whether failures when analyzing containers can be
// ignored or not. This function also listens to the `KillChannel` channel, so
// you don't have to bother about handling signals gracefully.
func inspectContainers(containers []types.Container, ignore bool) (*ContainersState, error) {
	state := &ContainersState{}
	cache := getCacheFile()

	for _, container := range containers {
		select {
		case <-KillChannel:
			return nil, errors.New("interruped")
		default:
			imageID, err := getImageID(container.Image)
			if err != nil {
				str := fmt.Sprintf("cannot analyze container %s [%s]: %s", container.ID, container.Image, err)
				if ignore {
					log.Print(str)
				} else {
					return nil, errors.New(str)
				}
				state.addNotAnalyzed(container)
				continue
			}

			if exists, suse := cache.idExists(imageID); exists && !suse {
				state.addUnsupported(container, "only zypper is supported")
			} else if cache.isImageOutdated(imageID) {
				state.Updated = append(state.Updated, container)
			} else {
				state.addUnknown(container)
			}
		}
	}
	return state, nil
}
