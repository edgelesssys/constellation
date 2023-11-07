#!/usr/bin/env bash
VERSION="v4.35.2"
if [[ -f ./yq ]] && ./yq --version | grep -q "${VERSION}"; then
  echo "yq is already available and up to date."
  exit 0
fi
if [[ -f ./yq ]]; then
  echo "yq is already available but not at the required version. Replacing with ${VERSION}."
  rm -f yq
fi

echo "Fetching yq ${VERSION}"
OS=$(uname -s)
ARCH=$(uname -m)
URL=""

if [[ ${OS} == "Darwin" ]]; then
  if [[ ${ARCH} == "arm64" ]]; then
    URL="https://github.com/mikefarah/yq/releases/download/${VERSION}/yq_darwin_arm64"
  elif [[ ${ARCH} == "x86_64" ]]; then
    URL="https://github.com/mikefarah/yq/releases/download/${VERSION}/yq_darwin_amd64"
  fi
elif [[ ${OS} == "Linux" ]]; then
  if [[ ${ARCH} == "x86_64" ]]; then
    URL="https://github.com/mikefarah/yq/releases/download/${VERSION}/yq_linux_amd64"
  elif [[ ${ARCH} == "arm64" ]]; then
    URL="https://github.com/mikefarah/yq/releases/download/${VERSION}/yq_linux_arm64"
  fi
fi

if [[ -z ${URL} ]]; then
  echo "OS \"${OS}\" and/or architecture \"${ARCH}\" is not supported."
  exit 1
else
  echo "Downloading yq from ${URL}"
  curl -o yq -L "${URL}"
  chmod +x ./yq
  ./yq --version
  if ! ./yq --version | grep -q "${VERSION}"; then # check that yq was installed correctly
    echo "Version is incorrect"
    exit 1
  fi
fi
