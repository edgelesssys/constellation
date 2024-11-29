#!/usr/bin/env bash

set -euo pipefail

function info() {
  echo "$@" 1>&2
}

function error() {
  echo "::err::$*"
  exit 1
}

allCCMVersions=$(git tag | grep ccm || test $? = 1)
if [[ -z ${allCCMVersions} ]]; then
  error "No CCM tags found in git"
fi

allMajorVersions=()

for ver in ${allCCMVersions}; do
  major=${ver#ccm/v} # remove "ccm/v" prefix
  major=${major%%.*} # remove everything after the first dot

  if [[ ${major} -eq 0 ]]; then
    continue # skip major version 0
  fi

  # Check if this major version is already in the list.
  for existingMajor in "${allMajorVersions[@]}"; do
    if [[ ${existingMajor} -eq ${major} ]]; then
      continue 2
    fi
  done

  info "Found major version ${major}"
  allMajorVersions+=("${major}")
done

if [[ ${#allMajorVersions[@]} -eq 0 ]]; then
  error "No major versions found in CCM tags"
fi

existingContainerVersions=$(crane ls "ghcr.io/edgelesssys/cloud-provider-gcp")
if [[ -z ${existingContainerVersions} ]]; then
  info "No existing container versions found"
fi

versionsToBuild=()

for major in "${allMajorVersions[@]}"; do
  # Get the latest released version with this major version.
  latest=$(echo "${allCCMVersions[@]}" | grep "${major}" | sort -V | tail -n 1)
  latest=${latest#ccm/} # remove "ccm/" prefix, keep v
  if [[ -z ${latest} ]]; then
    error "Could not determine latest version with major ${major}"
  fi
  info "Latest ${major} version is ${latest}"

  # Find the latest version with this major version.
  majorVerRegexp="v${major}.[0-9]+.[0-9]+"
  allExistingWithMajor=$(grep -E "${majorVerRegexp}" <<< "${existingContainerVersions}" || test $? = 1)
  latestExistingWithMinor=$(echo "${allExistingWithMajor}" | sort -V | tail -n 1)

  # If there is no existing version with this major version, build the latest released version.
  if [[ -z ${latestExistingWithMinor} ]]; then
    info "No existing version with major ${major}, adding ${latest} to versionsToBuild"
    versionsToBuild+=("${latest}")
    continue
  fi
  info "Latest existing version with major ${major} is ${latestExistingWithMinor}"

  newerVer=$(echo -e "${latest}\n${latestExistingWithMinor}" | sort -V | tail -n 1)
  if [[ ${newerVer} == "${latestExistingWithMinor}" ]]; then
    info "Existing version ${latestExistingWithMinor} is up to date, skipping"
    continue
  fi

  info "Newer version ${latest} is available, existing version is ${latestExistingWithMinor}."
  info "Adding ${latest} to versionsToBuild"
  versionsToBuild+=("${latest}")
done

# Print one elem per line | quote elems | create array | remove empty elems and print compact.
printf '%s\n' "${versionsToBuild[@]}" | jq -R | jq -s | jq -c 'map(select(length > 0))'
