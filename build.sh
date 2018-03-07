#!/usr/bin/env bash

set -e

cd $(dirname $(basename $0))
echo "Build directory is $(pwd)"
OS=("darwin" "linux")
ARCH="amd64"
NAME="aws-ps"

for o in ${OS[@]}; do
    echo "Build start for ${o}-${ARCH}"
    GOOS=$o GOARCH=$ARCH go build -a -o bin/$NAME
    tar -czvf bin/$NAME-$o-$ARCH.tar.gz -C bin $NAME
    echo "Build succeed !! for ${o}-${ARCH}"
    rm bin/$NAME
done

echo $(ls -al bin)