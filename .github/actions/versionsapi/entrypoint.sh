#!/usr/bin/env bash

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
  echo "output=${out}" | tee "${GITHUB_ENV}"
else
  echo "${out}"
fi
