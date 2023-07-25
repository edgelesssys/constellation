#!/usr/bin/env bash

function setup {
  # Ensure that the cache directories exist, so they are not created by docker with root permissions.
  mkdir -p "${HOME}/.cache/bazel"
  mkdir -p "${HOME}/.cache/shared_bazel_repository_cache"
  mkdir -p "${HOME}/.cache/shared_bazel_action_cache"

  if [[ ! -f "${HOME}/.docker/config.json" ]]; then
    echo "ERROR: ${HOME}/.docker/config.json does not exist."
    echo "Please login into your container registry to create it."
    echo "echo <TOKEN> | docker login ghcr.io -u <USERNAME> --password-stdin"
    exit 1
  fi
}

function startBazelServer {
  local containerImage="ghcr.io/edgelesssys/bazel-container:v6.2.1-0"
  local containerName="bazeld"

  setup

  local hostWorkspaceDir
  hostWorkspaceDir="$(git rev-parse --show-toplevel)"
  if [[ $? -ne 0 ]]; then
    echo Could not find git repository root. Are you in a git repository?
    return 1
  fi

  echo Starting bazel container as daemon...
  echo You can stop this command using:
  echo docker kill "${containerName}"

  docker run \
    --rm \
    --detach \
    --name "${containerName}" \
    -v "${hostWorkspaceDir}":/workspace \
    -v "${HOME}/.cache/bazel":"/home/builder/.cache/bazel" \
    -v "${HOME}/.cache/shared_bazel_repository_cache":"/home/builder/.cache/shared_bazel_repository_cache" \
    -v "${HOME}/.cache/shared_bazel_action_cache":"/home/builder/.cache/shared_bazel_action_cache" \
    -v "${HOME}/.docker/config.json":"/home/builder/.docker/config.json" \
    --entrypoint=/bin/sleep \
    "${containerImage}" \
    infinity || return $?
}

function stopBazelServer {
  local containerName="bazeld"

  echo Stopping bazel container...

  docker kill "${containerName}" || return $?
}

function bazel {
  local containerName="bazeld"

  local hostWorkspaceDir
  hostWorkspaceDir="$(git rev-parse --show-toplevel)"
  if [[ $? -ne 0 ]]; then
    echo Could not find git repository root. Are you in a git repository?
    return 1
  fi

  local containerWorkDir
  containerWorkDir=$(realpath -m "/workspace/$(realpath --relative-base="${hostWorkspaceDir}" .)")
  if [[ $? -ne 0 ]]; then
    echo Could not determine container work directory.
    return 1
  fi

  docker exec \
    -it \
    --workdir "${containerWorkDir}" \
    --env "HOST_CACHE=${HOME}/.cache" \
    "${containerName}" \
    bazel "$@" || return $?
}
