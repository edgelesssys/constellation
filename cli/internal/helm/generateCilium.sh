#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

calldir=$(pwd)
ciliumTmpDir=$(mktemp -d)
pushd "${ciliumTmpDir}"
git clone --filter=blob:none --no-checkout --sparse --depth 1 -b 1.12.1 https://github.com/cilium/cilium.git
pushd cilium

git sparse-checkout add install/kubernetes/cilium
git checkout

git apply "${calldir}"/cilium.patch
cp -r install/kubernetes/cilium "${calldir}"/charts

popd
popd
rm -r "${ciliumTmpDir}"
