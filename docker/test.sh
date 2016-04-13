#!/bin/bash

VERSIONS=( 1.5 1.6 tip )

for version in "${VERSIONS[@]}"; do
    echo "Running unit tests inside of Go ${version} ..."

    if [ $version == "tip" ]; then
        gimme 1.6
    fi
    gimme $version
    source ~/.gimme/envs/go${version}.env
    go test -v
done
