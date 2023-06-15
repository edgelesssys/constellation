#!/usr/bin/env bash

rm -f rpm/repo.yaml
bazel run //:bazeldnf -- init \
  --fc 38 \
  --arch x86_64 \
  --output rpm/repo.yaml
