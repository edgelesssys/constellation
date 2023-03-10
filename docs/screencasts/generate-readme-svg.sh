#!/usr/bin/env bash
#
# This script prepares the environment for expect scripts to be recorded in,
# executes all scripts, and copies the .cast files to our doc's asset folder.
#
# Note: A cluster is created in GCP. Therefore you are expected to be logged in
# via `gcloud` CLI. You credentials at $HOME/.config/gcloud are mounted into the
# screenrecordings container. A full script run takes ~20min.
#

docker build -t screenrecodings docker

# Create cast
docker run -it \
  -v "${HOME}"/.config/gcloud:/root/.config/gcloud \
  -v "$(pwd)"/recordings:/recordings \
  -v "$(pwd)"/constellation:/constellation \
  screenrecodings /scripts/github-readme.expect

# Fix meta data: width and height are always zero in Docker produced cast files.
# Header is the first lint of cast file in JSON format, we read, fix and write it.
head recordings/github-readme.cast -n 1 | yq e -M '.width = 95 | .height = 17' - > new_header.cast
# Then append everything, expect first line from original cast file.
tail -n+2 recordings/github-readme.cast >> new_header.cast

# Then render cast into svg using:
#   https://github.com/nbedos/termtosvg
termtosvg render new_header.cast readme.svg -t window-frame.svg

# Copy and cleanup
cp readme.svg ../static/img/shell-windowframe.svg
rm readme.svg new_header.cast
