#!/bin/bash

SOURCE="$0"
while [ -h "$SOURCE" ]; do
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
SOURCE="$(readlink "$SOURCE")"
[[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE"
done
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

cd $DIR

go build \
      -trimpath \
      -ldflags \
      "-s -w" \
      -o ./release/light_minio_client src/*.go