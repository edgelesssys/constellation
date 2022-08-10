#!/bin/bash

set -euo pipefail

TO_DELETE=$(grep -lr "\"uid\": \"${1}\"" . || true)
if [ -z "$TO_DELETE" ]
then
    printf "Unable to find '${1}'\n"
else
    printf "Statefile found. You should run:\n\n"
    printf "cd %s\n" $TO_DELETE
    printf "constellation terminate\n\n"
fi
