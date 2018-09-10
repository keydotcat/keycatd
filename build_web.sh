#!/bin/bash

webTag="${WEB_TAG:-0.0.1}"

here=$(python -c 'import os; print os.path.realpath(os.getcwd())')
webDir=${here}/web
test -d ${webDir} || git clone https://github.com/keydotcat/web.git ${webDir}

(
cd ${webDir}
git fetch
[ $(git describe --abbrev=8 --dirty --always --tags) == "${webTag}" ] || git checkout ${webTag} -b auto-${webTag}
echo 'Set web version to' ${webTag}
yarn install
yarn run build:web
)

ln -sf ${webDir}/dist/web ${here}/data/web 

