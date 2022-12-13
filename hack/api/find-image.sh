#!/usr/bin/env bash
# Copyright (c) Edgeless Systems GmbH
#
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail
shopt -s inherit_errexit

ref="-"
stream="stable"
json=false
cdn_url="https://cdn.confidential.cloud"

function usage() {
  cat << 'EOF'
Usage: find-image.sh [options] [command]

Options:
  -r, --ref <ref>        Ref to search for (default: "-")
  -s, --stream <stream>  Stream to search for (default: "stable")
  --json                 Output JSON instead of shortname(s)
  --help                 Show this help

Commands:
  latest                 Find latest image for ref and stream
  list                   List all images for ref and stream
EOF
}

POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
  -r | --ref)
    # Canonicalize ref format (e.g. "feat/foo/bar" -> "feat-foo-bar")
    ref=$(echo -n "$2" | tr -c '[:alnum:]' '-')
    shift # past argument
    shift # past value
    ;;
  -s | --stream)
    stream="$2"
    shift # past argument
    shift # past value
    ;;
  --json)
    json=true
    shift # past argument
    ;;
  --help)
    usage
    exit 0
    ;;
  -*)
    echo "Unknown option $1"
    echo
    usage
    exit 1
    ;;
  *)
    POSITIONAL_ARGS+=("$1") # save positional arg
    shift                   # past argument
    ;;
  esac
done

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

# shortname <ref> <stream> <version>
function shortname() {
  shortname=""

  if [[ ${ref} != "-" ]]; then
    shortname+="ref/${1}/"
  fi

  if [[ ${stream} != "stable" ]]; then
    shortname+="stream/${2}/"
  fi

  shortname+="${3}"

  echo "${shortname}"
}

function latest() {
  latest_path="constellation/v1/ref/${ref}/stream/${stream}/versions/latest/image.json"
  latest_url="${cdn_url}/${latest_path}"

  latest_status=$(curl -s -o /dev/null -w "%{http_code}" "${latest_url}")
  if [[ ${latest_status} != "200" ]]; then
    echo "[Error] No image found for ref ${ref} and stream ${stream} (${latest_status})"
    exit 1
  fi

  latest_json=$(curl -sL "${latest_url}")

  if [[ ${json} == true ]]; then
    jq <<< "${latest_json}"
    exit 0
  fi

  latest_version=$(echo "${latest_json}" | jq -r '.version')

  shortname "${ref}" "${stream}" "${latest_version}"
  exit 0
}

function list() {
  major="v2"
  list_path="constellation/v1/ref/${ref}/stream/${stream}/versions/major/${major}/image.json"
  list_url="${cdn_url}/${list_path}"

  list_status=$(curl -s -o /dev/null -w "%{http_code}" "${list_url}")
  if [[ ${list_status} != "200" ]]; then
    echo "[Error] No minor image list found for ref ${ref} and stream ${stream} and version ${major} (${list_status})"
    exit 1
  fi

  minor_list=$(curl -sL "${list_url}" | jq -r '.versions[]')

  for minor in ${minor_list}; do
    list_path="constellation/v1/ref/${ref}/stream/${stream}/versions/minor/${minor}/image.json"
    list_url="${cdn_url}/${list_path}"

    list_status=$(curl -s -o /dev/null -w "%{http_code}" "${list_url}")
    if [[ ${list_status} != "200" ]]; then
      echo "[Error] No patch image list found for ref ${ref} and stream ${stream} and version ${minor} (${list_status})"
      exit 1
    fi

    patch_list="${patch_list-""} $(curl -sL "${list_url}" | jq -r '.versions[]')"
  done

  if [[ ${json} == true ]]; then
    out="{}"
    out=$(jq <<< "${out}" --arg ref "${ref}" '.ref = $ref')
    out=$(jq <<< "${out}" --arg stream "${stream}" '.stream = $stream')
    for patch in ${patch_list}; do
      out=$(jq <<< "${out}" --arg patch "${patch}" '.versions += [$patch]')
    done
    jq <<< "${out}"
    exit 0
  fi

  for version in ${patch_list}; do
    shortname "${ref}" "${stream}" "${version}"
  done
  exit 0
}

case ${1-"latest"} in
"list")
  list
  ;;
"latest")
  latest
  ;;
*)
  echo "Unknown command $1"
  ;;
esac
