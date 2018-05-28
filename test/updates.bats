#!/usr/bin/bats -t
# Copyright (c) 2018 SUSE LLC. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

load helpers

@test "zypper-docker update" {
  zypperdocker up $TESTIMAGE:$TAG $TESTIMAGE:updated
  [ "$status" -eq 0 ]
  [[ "$output" =~ "$TESTIMAGE:updated successfully created"+ ]]

  zypperdocker up $TESTIMAGE:$TAG $TESTIMAGE:updated
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Cannot overwrite an existing image. Please use a different repository/tag"+ ]]

  docker rmi -f $TESTIMAGE:updated
}

@test "zypper-docker list-updates" {
  zypperdocker lu $TESTIMAGE:$TAG
  [ "$status" -eq 0 ]

  zypperdocker lu alpine:latest
  [ "$status" -eq 127 ]
  [[ "${lines[0]}" =~ "/bin/sh: zypper: not found" ]]
}

@test "zypper-docker list-updates-container" {
  # run a container
  run_container
  zypperdocker luc $CONTAINER_ID
  [ "$status" -eq 0 ]
  [[ "$output" =~ "WARNING: Only the source image from this container will be inspected. Manually installed packages won't be taken into account."+ ]]

  zypperdocker -f luc $CONTAINER_ID
  [ "$status" -eq 0 ]
  [[ "$output" =~ "WARNING: Force flag used. Manually installed packages will be analyzed as well."+ ]]

  # stop the container
  stop_container
  zypperdocker luc $CONTAINER_ID
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Checking stopped container"+ ]]

  zypperdocker luc blahblub
  [ "$status" -eq 1 ]
  [[ "$output" =~ "container blahblub does not exist"+ ]]

  # remove the container before exiting
  remove_container
}
