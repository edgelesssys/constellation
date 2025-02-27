#!/usr/bin/env bash
# no -e since we need to collect errors later
# no -u since it interferes with checking associative arrays
set -o pipefail
shopt -s extglob

declare -A errors

for directory in "$1"/system/!(mkosi_wrapper.sh); do
  dirname="$(basename "$directory")"
  attestationVariant="$(echo "$dirname" | cut -d_ -f2)"

  echo "Their measurements for $attestationVariant:"
  ts "  " < "$attestationVariant"_their-measurements.json
  echo "Own measurements for $attestationVariant:"
  ts "  " < "$attestationVariant"_own-measurements.json

  diff="$(jd ./"$attestationVariant"_their-measurements.json ./"$attestationVariant"_own-measurements.json)"
  if [[ ! -z "$diff" ]]; then
    errors["$attestationVariant"]="$diff"
  fi
done

for attestationVariant in "${!errors[@]}"; do
  echo "Failed to reproduce measurements for $attestationVariant:"
  echo "${errors["$attestationVariant"]}" | ts "  "
done

if [[ "${#errors[@]}" -ne 0 ]]; then
  exit 1
fi
