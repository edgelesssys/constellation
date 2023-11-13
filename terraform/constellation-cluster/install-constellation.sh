#!/usr/bin/env bash
if [[ -f ./constellation ]]; then # needed to allow using devbuilds
  echo "constellation CLI is already available."
  exit 0
fi
version="v2.13.0"
os=$(uname -s)
arch=$(uname -m)
url=""
if [[ ${os} == "Darwin" ]]; then
  if [[ ${arch} == "arm64" ]]; then
    url="https://github.com/edgelesssys/constellation/releases/download/${version}/constellation-darwin-arm64"
  elif [[ ${arch} == "x86_64" ]]; then
    url="https://github.com/edgelesssys/constellation/releases/download/${version}/constellation-darwin-amd64"
  fi
elif [[ ${os} == "Linux" ]]; then
  if [[ ${arch} == "x86_64" ]]; then
    url="https://github.com/edgelesssys/constellation/releases/download/${version}/constellation-linux-amd64"
  elif [[ ${arch} == "arm64" ]]; then
    url="https://github.com/edgelesssys/constellation/releases/download/${version}/constellation-linux-arm64"
  fi
fi

echo "Fetching constellation ${version}"
if [[ -z ${url} ]]; then
  echo "OS \"${os}\" and/or architecture \"${arch}\" is not supported."
  exit 1
else
  curl -o constellation -L "${url}"
  chmod +x constellation
fi
