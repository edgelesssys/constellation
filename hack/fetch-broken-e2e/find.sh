#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

TO_DELETE=$(grep -lr "\"uid\": \"${1}\"" . || true)
if [[ -z "${TO_DELETE}" ]]
then
    printf "Unable to find '%s'\n" "${1}"
else
    printf "Statefile found. You should run:\n\n"
    printf "cd %s\n" "${TO_DELETE}"
    printf "constellation terminate --yes\n\n"
fi
