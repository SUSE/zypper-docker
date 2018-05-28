#!/usr/bin/env bats -t
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

@test "zypper-docker --version" {
  zypperdocker --version
  [ "$status" -eq 0 ]
  [[ "$output" =~ "zypper-docker version "+ ]]

  zypperdocker -v
  [ "$status" -eq 0 ]
  [[ "$output" =~ "zypper-docker version "+ ]]
}

@test "zypper-docker --help" {
  zypperdocker --help
  [ "$status" -eq 0 ]
  [[ "${lines[0]}" =~ "NAME:"+ ]]
  [[ "${lines[1]}" =~ "zypper-docker - Patching Docker images safely"+ ]]

  zypperdocker -h
  [ "$status" -eq 0 ]
  [[ "${lines[0]}" =~ "NAME:"+ ]]
  [[ "${lines[1]}" =~ "zypper-docker - Patching Docker images safely"+ ]]

  zypperdocker help
  [ "$status" -eq 0 ]
  [[ "${lines[0]}" =~ "NAME:"+ ]]
  [[ "${lines[1]}" =~ "zypper-docker - Patching Docker images safely"+ ]]

  zypperdocker h
  [ "$status" -eq 0 ]
  [[ "${lines[0]}" =~ "NAME:"+ ]]
  [[ "${lines[1]}" =~ "zypper-docker - Patching Docker images safely"+ ]]
}

@test "zypper-docker [command] --help" {

  # zypper-docker images
  zypperdocker images --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker images"+ ]]

  zypperdocker images -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker images"+ ]]

  # zypper-docker list-updates
  zypperdocker list-updates --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-updates"+ ]]

  zypperdocker list-updates -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-updates"+ ]]

  zypperdocker lu --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-updates"+ ]]

  zypperdocker lu -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-updates"+ ]]

  # zypper-docker list-updates-container
  zypperdocker list-updates-container --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-updates-container"+ ]]

  zypperdocker list-updates-container -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-updates-container"+ ]]

  zypperdocker luc --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-updates-container"+ ]]

  zypperdocker luc -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-updates-container"+ ]]

  # zypper-docker update
  zypperdocker update --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker update"+ ]]

  zypperdocker update -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker update"+ ]]

  zypperdocker up --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker update"+ ]]

  zypperdocker up -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker update"+ ]]

  # zypper-docker list-patches
  zypperdocker list-patches --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-patches"+ ]]

  zypperdocker list-patches -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-patches"+ ]]

  zypperdocker lp --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-patches"+ ]]

  zypperdocker lp -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-patches"+ ]]

  # zypper-docker list-patches-container
  zypperdocker list-patches-container --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-patches-container"+ ]]

  zypperdocker list-patches-container -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-patches-container"+ ]]

  zypperdocker lpc --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-patches-container"+ ]]

  zypperdocker lpc -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker list-patches-container"+ ]]

  # zypper-docker patch
  zypperdocker patch --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker patch"+ ]]

  zypperdocker patch -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker patch"+ ]]

  # zypper-docker patch-check
  zypperdocker patch-check --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker patch-check"+ ]]

  zypperdocker patch-check -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker patch-check"+ ]]

  zypperdocker pchk --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker patch-check"+ ]]

  zypperdocker pchk -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker patch-check"+ ]]

  # zypper-docker patch-check-container
  zypperdocker patch-check-container --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker patch-check-container"+ ]]

  zypperdocker patch-check-container -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker patch-check-container"+ ]]

  zypperdocker pchkc --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker patch-check-container"+ ]]

  zypperdocker pchkc -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker patch-check-container"+ ]]

  # zypper-docker ps
  zypperdocker ps --help
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker ps"+ ]]

  zypperdocker ps -h
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" =~ "zypper-docker ps"+ ]]
}
