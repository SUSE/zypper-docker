#!/bin/bash
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

INTEGRATION_ROOT=$(dirname "$(readlink -f "$BASH_SOURCE")")
ZYPPERDOCKER="${ZYPPERDOCKER:-${INTEGRATION_ROOT}/../zypper-docker}"
CONTAINER_ID=""
TESTIMAGE="zypper-docker-integration-tests"
TAG="latest"

function sane_run() {
  local cmd="$1"
  shift

  run "$cmd" "$@"

  # Some debug information to make life easier.
  echo "$(basename "$cmd") $@ (status=$status)" >&2
  echo "$output" >&2
}

function zypperdocker() {
  local args=()

  # Set the first argument (the subcommand).
  args+=("$1")

  shift
  args+=("$@")
  sane_run "$ZYPPERDOCKER" "${args[@]}"
}

function run_container() {
  ID=$(docker ps | grep $TESTIMAGE | awk '{print $1}' | head -n1)
  if [[ $ID != "" ]]; then
    CONTAINER_ID=$ID
  else
    CONTAINER_ID=$(docker run -t -d $TESTIMAGE:$TAG)
  fi
}

function stop_container() {
  docker stop $CONTAINER_ID > /dev/null
}

function remove_container() {
  docker rm -f $CONTAINER_ID > /dev/null
  CONTAINER_ID=""
}
