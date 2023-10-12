#!/usr/bin/env bash
script_path=$(realpath $0)

bazel build //bazel/settings:tag
pseudo_version=$(cat ../../../bazel-bin/bazel/settings/_tag.tags.txt)

docker build -t mint . -f Dockerfile
if [[ "$?" -ne 0 ]]; then
  echo "Failed to build docker image"
  exit 1
fi

tag=ghcr.io/edgelesssys/constellation/mint:"$pseudo_version"
docker tag mint:latest "$tag"

if [[ "$?" -eq 0 ]]; then
  echo "Successfully built docker image: $tag"
fi
