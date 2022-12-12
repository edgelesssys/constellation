#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

# Usage: find-image.sh

set -euo pipefail
shopt -s inherit_errexit

ref="-"
stream="stable"

POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
  -r | --ref)
    ref="$2"
    shift # past argument
    shift # past value
    ;;
  -s | --stream)
    stream="$2"
    shift # past argument
    shift # past value
    ;;
  -*)
    echo "Unknown option $1"
    exit 1
    ;;
  *)
    POSITIONAL_ARGS+=("$1") # save positional arg
    shift                   # past argument
    ;;
  esac
done

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

ref=$(echo -n "${ref}" | tr -c '[:alnum:]' '-')

base_url="https://cdn.confidential.cloud"
latest_path="constellation/v1/ref/${ref}/stream/${stream}/versions/latest/image.json"
latest_url="${base_url}/${latest_path}"
latest_status=$(curl -s -o /dev/null -w "%{http_code}" "${latest_url}")
if [[ ${latest_status} != "200" ]]; then
  echo "No image found for ref ${ref} and stream ${stream} (${latest_status})"
  exit 1
fi
latest_version=$(curl -sL "${latest_url}" | jq -r '.version')

shortname=""
if [[ ${ref} != "-" ]]; then
  shortname+="ref/${ref}/"
fi
if [[ ${stream} != "stable" ]]; then
  shortname+="stream/${stream}/"
fi
shortname+="${latest_version}"

echo "${shortname}"
