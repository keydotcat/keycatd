#!/bin/bash

rootDir=$(dirname $0)/..

GIT_VERSION=${GIT_VERSION:-$(git describe --abbrev=8 --dirty --always --tags 2>/dev/null)}

#make -C $rootDir web 
make -C $rootDir static
relDir=$rootDir/bin/releases/${GIT_VERSION}
mkdir -p $relDir

for plat in linux linux:arm linux:arm64 darwin windows
do
	os=$(echo $plat | sed 's/:.*$//')
	arch=$(echo $plat | sed 's/^.*://')
	if [ "$arch" == "$os" ]; then
		arch=amd64
	fi
	GOOS=$os GOARCH=$arch make -C $rootDir keycatd
	if [ "$os" == "windows" ]; then
	    zip -9 $rootDir/bin/keycatd.windows.$arch.${GIT_VERSION}.zip $rootDir/bin/keycatd.windows.$arch.exe
			mv $rootDir/bin/keycatd.windows.$arch.${GIT_VERSION}.zip $relDir
	else
		gzip -9 -S .${GIT_VERSION}.gz $rootDir/bin/keycatd.$os.$arch
		mv $rootDir/bin/keycatd.$os.$arch.${GIT_VERSION}.gz $relDir
	fi
done

GOOS=linux make -C $rootDir keycatd
(cd $rootDir && docker build --label keycat/keycatd:${GIT_VERSION} . )
