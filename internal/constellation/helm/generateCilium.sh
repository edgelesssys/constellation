#!/usr/bin/env bash

set -euo pipefail
set -o errtrace
shopt -s inherit_errexit

echo "Pulling Cilium Helm chart..."

function cleanup {
  rm -rf -- "${ciliumTmpDir}"
}

trap cleanup EXIT

calldir=$(pwd)
ciliumTmpDir=$(mktemp -d)
cd "${ciliumTmpDir}"

git clone \
  --filter=blob:none \
  --no-checkout \
  --sparse \
  --depth 1 \
  -b v1.15.8-edg.0 \
  https://github.com/edgelesssys/cilium.git
cd cilium

git sparse-checkout add install/kubernetes/cilium
git checkout

cp -r install/kubernetes/cilium "${calldir}/charts"

echo # final newline
