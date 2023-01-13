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
  echo 'output<<EOF' >> "${GITHUB_OUTPUT}"
  echo "test" | tee "${GITHUB_OUTPUT}"
  echo 'EOF' >> "${GITHUB_OUTPUT}"
else
  echo "${out}"
fi
