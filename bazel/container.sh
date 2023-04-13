#!/usr/bin/env bash

containerImage="ghcr.io/malt3/bazel-container:v6.1.0-0"
containerName="bazeld"

hostWorkspaceDir="$(git rev-parse --show-toplevel)"
containerWorkDir=/workspace/$(realpath --relative-base="${hostWorkspaceDir}" .)

function setup {
  # Ensure that the cache directories exist, so they are not created by docker with root permissions.
  mkdir -p "${HOME}/.cache/bazel"
  mkdir -p "${HOME}/.cache/shared_bazel_repository_cache"
  mkdir -p "${HOME}/.cache/shared_bazel_action_cache"
}

function startBazelServer {
  echo Starting bazel container as daemon...
  echo You can stop this command using:
  echo docker kill "${containerName}"
  docker run \
    --rm \
    -d \
    --name "${containerName}" \
    -v "${hostWorkspaceDir}":/workspace \
    -v "${HOME}/.cache/bazel":"/home/builder/.cache/bazel" \
    -v "${HOME}/.cache/shared_bazel_repository_cache":"/home/builder/.cache/shared_bazel_repository_cache" \
    -v "${HOME}/.cache/shared_bazel_action_cache":"/home/builder/.cache/shared_bazel_action_cache" \
    --entrypoint=/bin/sleep \
    "${containerImage}" \
    infinity
}

function bazel {
  docker exec \
    -it \
    --workdir "${containerWorkDir}" \
    "${containerName}" \
    bazel "$@"
}
