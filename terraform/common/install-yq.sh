#!/usr/bin/env bash
version="v4.35.2"
if [[ -f ./yq ]] && ./yq --version | grep -q "${version}"; then
  echo "yq is already available and up to date."
  exit 0
fi
if [[ -f ./yq ]]; then
  echo "yq is already available but not at the required version. Replacing with ${version}."
  rm -f yq
fi

echo "Fetching yq ${version}"
os=$(uname -s)
arch=$(uname -m)
url=""

if [[ ${os} == "Darwin" ]]; then
  if [[ ${arch} == "arm64" ]]; then
    url="https://github.com/mikefarah/yq/releases/download/${version}/yq_darwin_arm64"
  elif [[ ${arch} == "x86_64" ]]; then
    url="https://github.com/mikefarah/yq/releases/download/${version}/yq_darwin_amd64"
  fi
elif [[ ${os} == "Linux" ]]; then
  if [[ ${arch} == "x86_64" ]]; then
    url="https://github.com/mikefarah/yq/releases/download/${version}/yq_linux_amd64"
  elif [[ ${arch} == "arm64" ]]; then
    url="https://github.com/mikefarah/yq/releases/download/${version}/yq_linux_arm64"
  fi
fi

if [[ -z ${url} ]]; then
  echo "os \"${os}\" and/or architecture \"${arch}\" is not supported."
  exit 1
else
  echo "Downloading yq from ${url}"
  curl -o yq -L "${url}"
  chmod +x ./yq
  ./yq --version
  if ! ./yq --version | grep -q "${version}"; then # check that yq was installed correctly
    echo "Version is incorrect"
    exit 1
  fi
fi
