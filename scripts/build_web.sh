#!/bin/bash


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
webTag="${WEB_TAG:-$(git describe --abbrev=0 --tags origin/master)}"
[ $(git describe --abbrev=8 --dirty --always --tags) == "${webTag}" ] || git checkout ${webTag} -b auto-${webTag}
git submodule update --init --recursive --remote
echo 'Set web version to' ${webTag}
npm install
npm run build
)

rm -f ${rootDir}/data/web
ln -sf ${webDir}/dist ${rootDir}/data/web 

