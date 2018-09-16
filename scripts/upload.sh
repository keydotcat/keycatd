#!/bin/bash

rootDir=$(dirname $0)/..

GIT_VERSION=${GIT_VERSION:-$(git describe --abbrev=8 --dirty --always --tags 2>/dev/null)}

(cd $rootDir; docker build -t keycat/keycatd:${GIT_VERSION} .)
docker push keycat/keycatd:${GIT_VERSION}
