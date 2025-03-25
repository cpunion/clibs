#!/bin/bash

set -e

CURDIR=$(pwd)

cd $(dirname "$0")/build

go install ./cmd/llgo_clibs

cd $CURDIR

llgo_clibs "$@"
