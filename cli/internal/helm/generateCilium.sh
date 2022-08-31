#!/bin/bash

CALLDIR=$(pwd)
CILIUMTMPDIR=$(mktemp -d)
cd $CILIUMTMPDIR
git clone --depth 1 -b 1.12.1 https://github.com/cilium/cilium.git
cd cilium
git apply $CALLDIR/cilium.patch
cp -r install/kubernetes/cilium $CALLDIR/charts
rm -r $CILIUMTMPDIR
