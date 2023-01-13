#!/usr/bin/env bash

set -x

for arg in "$@"; do
  if [[ ${arg} != "" ]]; then
    args[${#args[@]}]="${arg}"
  fi
done

out=$(/versionsapi "${args[@]}")
if [[ $? -ne 0 ]]; then
  exit 1
fi

if [[ ${GITHUB_ACTIONS} == "true" ]]; then
  out="${out//'%'/'%25'}"
  out="${out//$'\n'/'%0A'}"
  out="${out//$'\r'/'%0D'}"
  echo "${out}" | tee "${GITHUB_OUTPUT}"
else
  echo "${out}"
fi
