#!/bin/bash

webTag="${WEB_TAG:-v0.0.9}"

realpath() {
    [[ $1 = /* ]] && echo "$1" || echo "$PWD/${1#./}"
}

me=$(realpath $0)
here=$(dirname $me)
rootDir=${here}/..
webDir=${rootDir}/web

test -d ${webDir} || git clone https://github.com/keydotcat/web.git ${webDir}

(
cd ${webDir}
git fetch
[ $(git describe --abbrev=8 --dirty --always --tags) == "${webTag}" ] || git checkout ${webTag} -b auto-${webTag}
git submodule update --init --recursive --remote
echo 'Set web version to' ${webTag}
yarn install
yarn run build:web
)

rm -f ${rootDir}/data/web
ln -sf ${webDir}/dist/web ${rootDir}/data/web 

