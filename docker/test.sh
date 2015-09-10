#!/bin/bash

VERSIONS=( 1.4 1.5 tip )

for version in "${VERSIONS[@]}"; do
    echo "Running unit tests inside of Go ${version} ..."

    if [ $version == "tip" ]; then
        gimme 1.4
    fi
    gimme $version
    source ~/.gimme/envs/go${version}.env
    godep go test -race -v ./...
    climate -open=false -threshold=80.0 -errcheck -vet -fmt .
done
