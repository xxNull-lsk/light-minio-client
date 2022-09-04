#!/bin/bash

SOURCE="$0"
while [ -h "$SOURCE" ]; do
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
SOURCE="$(readlink "$SOURCE")"
[[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE"
done
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

cd $DIR

export CGO_ENABLED=0
os_list=(linux windows darwin)
arch_list=(amd64 arm64)
for os in ${os_list[*]}; do
  for arch in ${arch_list[*]}; do
      echo build for ${os} ${arch} ...
      filename=light_minio_client.${os}.${arch}
      if [ "$os" == "windows" ]; then
          filename=light_minio_client.${os}.${arch}.exe
      fi
      GOOS=${os} GOARCH=${arch} go build \
                                  -trimpath \
                                  -ldflags \
                                  "-s -w" \
                                  -o ./release/${filename} src/*.go

      if [ "$os" == "windows" ]; then
          zip -q -o $filename.zip $filename
      else
          tar -czf $filename.tar.gz $filename
      fi
      rm $filename
  done
done