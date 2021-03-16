#!/bin/sh

# disable go modules
export GOPATH=""

# disable cgo
export CGO_ENABLED=0

set -e
set -x

# linux
GOOS=linux GOARCH=amd64 go build -o release/linux/amd64/drone-runner-docker
GOOS=linux GOARCH=arm64 go build -o release/linux/arm64/drone-runner-docker
GOOS=linux GOARCH=arm   go build -o release/linux/arm/drone-runner-docker
GOOS=linux GOARCH=ppc64le go build -o release/linux/ppc64le/drone-runner-docker

# windows
GOOS=windows go build -o release/windows/amd64/drone-runner-docker.exe
