#!/bin/bash

realpath() {
    [[ $1 = /* ]] && echo "$1" || echo "$PWD/${1#./}"
}

me=$(realpath $0)
here=$(dirname $me)
rootDir=${here}/..

function showChanges() {
	last=$(git tag | sort -r | head -n 1)
	prev=$(git tag | sort -r | head -n 2 | tail -n 1)
	echo "## $1 $last"
	echo
	git log ${prev}..${last} '--pretty=format:%s' | grep -i -e '^new:' -e '^fix:' -e '^change:' | sort -r | sed 's/^/* /'
}

showChanges "Backend"
echo 

(
cd $rootDir/web
showChanges "Web"
)