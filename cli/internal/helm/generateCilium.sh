#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

calldir=$(pwd)
ciliumTmpDir=$(mktemp -d)
cd "${ciliumTmpDir}" || exit 1
git clone --depth 1 -b 1.12.1 https://github.com/cilium/cilium.git
cd cilium || exit 1
git apply "${calldir}"/cilium.patch
cp -r install/kubernetes/cilium "${calldir}"/charts
rm -r "${ciliumTmpDir}"
