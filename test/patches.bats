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

@test "zypper-docker patch" {
  zypperdocker patch $TESTIMAGE:$TAG $TESTIMAGE:patched
  [ "$status" -eq 0 ]
  [[ "$output" =~ "$TESTIMAGE:patched successfully created"+ ]]

  zypperdocker patch $TESTIMAGE:$TAG $TESTIMAGE:patched
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Cannot overwrite an existing image. Please use a different repository/tag"+ ]]

  # remove created image
  docker rmi -f $TESTIMAGE:patched

  zypperdocker patch --no-recommends $TESTIMAGE:$TAG $TESTIMAGE:patched
  [ "$status" -eq 0 ]
  [[ "$output" =~ "$TESTIMAGE:patched successfully created"+ ]]

  # remove created image
  docker rmi -f $TESTIMAGE:patched
}

@test "zypper-docker list-patches" {
  zypperdocker lp $TESTIMAGE:$TAG
  [ "$status" -eq 0 ]

  zypperdocker lp alpine:latest
  [ "$status" -eq 127 ]
  [[ "${lines[0]}" =~ "/bin/sh: zypper: not found" ]]
}

@test "zypper-docker list-patches-container" {
  # run a container
  run_container
  zypperdocker lpc --base $CONTAINER_ID
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Base image $IMAGE_ID of container $CONTAINER_ID will be analyzed. Manually installed packages won't be taken into account."+ ]]

  zypperdocker lpc $CONTAINER_ID
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Checking running container"+ ]]

  # stop the container
  stop_container
  zypperdocker lpc $CONTAINER_ID
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Checking stopped container"+ ]]

  zypperdocker lpc blahblub
  [ "$status" -eq 1 ]
  [[ "$output" =~ "container blahblub does not exist"+ ]]

  # remove the container before exiting
  remove_container
}
