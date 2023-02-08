#!/usr/bin/env bash

docker build -t screenrecodings docker

# Generate cast to verify CLI
docker run -it -v "$(pwd)"/recordings:/recordings screenrecodings
cp recordings/verify-cli.cast ../static/assets/verify-cli.cast

# Generate cast to check SBOM
docker run -it -v "$(pwd)"/recordings:/recordings screenrecodings check-sbom.sh /recordings/check-sbom.cast
cp recordings/check-sbom.cast ../static/assets/check-sbom.cast
