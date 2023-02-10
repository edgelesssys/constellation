#!/usr/bin/env bash

docker build -t screenrecodings docker

# Generate cast to verify CLI
docker run -it -v "$(pwd)"/recordings:/recordings screenrecodings
cp recordings/verify-cli.cast ../static/assets/verify-cli.cast

# Generate cast to check SBOM
# docker run -it -v "$(pwd)"/recordings:/recordings screenrecodings check-sbom.sh /recordings/check-sbom.cast
# cp recordings/check-sbom.cast ../static/assets/check-sbom.cast

# docker rm -f recorder || true
# docker build -t screenrecodings docker
# docker run --name recorder -d -v "$(pwd)"/recordings:/recordings screenrecodings
# docker exec recorder /bin/bash < . check-sbom.sh /recordings/check-sbom.cast
