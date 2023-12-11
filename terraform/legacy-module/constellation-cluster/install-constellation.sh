#!/usr/bin/env bash
if [[ -f ./constellation ]]; then
  echo "constellation CLI is already available."
  exit 0
fi

os=$(uname -s)
arch=$(uname -m)
version=$1
url=""

echo "Fetching constellation ${version}"

if [[ ${os} == "Darwin" ]]; then
  if [[ ${arch} == "arm64" ]]; then
    url="https://github.com/edgelesssys/constellation/releases/${version}/download/constellation-darwin-arm64"
  elif [[ ${arch} == "x86_64" ]]; then
    url="https://github.com/edgelesssys/constellation/releases/${version}/download/constellation-darwin-amd64"
  fi
elif [[ ${os} == "Linux" ]]; then
  if [[ ${arch} == "x86_64" ]]; then
    url="https://github.com/edgelesssys/constellation/releases/${version}/download/constellation-linux-amd64"
  elif [[ ${arch} == "arm64" ]]; then
    url="https://github.com/edgelesssys/constellation/releases/${version}/download/constellation-linux-arm64"
  fi
fi

if [[ -z ${url} ]]; then
  echo "os \"${os}\" and/or architecture \"${arch}\" is not supported."
  exit 1
else
  curl -o constellation -L "${url}"
  chmod +x constellation
fi
