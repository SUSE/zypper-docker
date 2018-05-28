#!/bin/bash

VERSIONS=( 1.9 1.10.x tip )

for version in "${VERSIONS[@]}"; do
    echo "Running unit tests inside of Go ${version} ..."

    if [ $version == "tip" ]; then
        gimme 1.10.x
    fi
    gimme $version
    source ~/.gimme/envs/go${version}.env
    go test -v
done
