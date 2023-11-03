#!/usr/bin/env bash
if [[ -f ./constellation ]]; then
  echo "constellation CLI is already available."
  exit 0
fi

echo "Fetching constellation ${VERSION}"
OS=$(uname -s)
ARCH=$(uname -m)
VERSION="latest"
URL=""

if [[ ${OS} == "Darwin" ]]; then
  if [[ ${ARCH} == "arm64" ]]; then
    URL="https://github.com/edgelesssys/constellation/releases/${VERSION}/download/constellation-darwin-arm64"
  elif [[ ${ARCH} == "x86_64" ]]; then
    URL="https://github.com/edgelesssys/constellation/releases/${VERSION}/download/constellation-darwin-amd64"
  fi
elif [[ ${OS} == "Linux" ]]; then
  if [[ ${ARCH} == "x86_64" ]]; then
    URL="https://github.com/edgelesssys/constellation/releases/${VERSION}/download/constellation-linux-amd64"
  elif [[ ${ARCH} == "arm64" ]]; then
    URL="https://github.com/edgelesssys/constellation/releases/${VERSION}/download/constellation-linux-arm64"
  fi
fi

if [[ -z ${URL} ]]; then
  echo "OS \"${OS}\" and/or architecture \"${ARCH}\" is not supported."
  exit 1
else
  curl -o constellation -L "${URL}"
  chmod +x constellation
fi
