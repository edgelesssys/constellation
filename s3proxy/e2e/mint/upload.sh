#!/usr/bin/env bash

if [[ ! "$1" =~ ^ghcr.io/edgelesssys/constellation/mint:v.*$ ]]; then
  echo "Error: invalid tag, expected input to match pattern '^ghcr.io\/edgelesssys\/constellation\/mint:v*$'"
  exit 1
fi

docker push "$1"
