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

@test "zypper-docker images" {
  zypperdocker -f images
  [ "$status" -eq 0 ]
  [[ "${lines[0]}" =~ "REPOSITORY"+.+"TAG"+.+"IMAGE ID"+.+"CREATED"+ ]]
  [[ "$output" =~ "$TESTIMAGE"+.+"$TAG"+ ]]
  [[ "$output" != *"alpine"*.*"latest"* ]]
}

