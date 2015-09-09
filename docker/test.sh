#!/bin/bash

read -d '' usage << 'EOF'
Usage: test [OPTIONS]...\\n
Options:\\n
  -v=,\\t-version=\\t\\tSpecify the Go Version for the test enviroment.
EOF

if [[ $# -eq 0 ]]; then
    echo "No arguments supplied"
    echo -e $usage
    exit 1
else
    for flag in "$@"; do
        case $flag in
            -v=*|-version=*)
                version=`echo $flag | sed 's/-v[a-z]*=//'`
            ;;
            *)
                echo -e $usage
                exit 1
            ;;
        esac
   done
fi
if [ $version == "tip" ]; then
    gimme 1.4
fi
gimme $version
source ~/.gimme/envs/go${version}.env
godep go test -race ./...
climate -open=false -threshold=80.0 -errcheck -vet -ft

