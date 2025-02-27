#!/usr/bin/env bash
set -euo pipefail
shopt -s extglob

for directory in "$1"/system/!(mkosi_wrapper.sh); do
  dirname="$(basename "$directory")"
  csp="$(echo "$dirname" | cut -d_ -f1)"
  attestationVariant="$(echo "$dirname" | cut -d_ -f2)"

  # This jq filter selects the measurements for the correct CSP and attestation variant
  # and then removes all `warnOnly: true` measurements.
  jq --arg attestation_variant "$attestationVariant" --arg csp "$csp" \
    '
      .list.[]
      | select(
        .attestationVariant == $attestation_variant
        and (.csp | ascii_downcase) == $csp
      )
      | .measurements
      | to_entries
      | map(select(.value.warnOnly | not))
      | from_entries
      | del(.[] .warnOnly)
  ' \
    measurements.json > "$attestationVariant"_their-measurements.json

  sudo --preserve-env "$1/measured-boot/cmd/cmd_/cmd" "$directory/constellation" /dev/stdout | jq '.measurements' > ./"$attestationVariant"_own-measurements.json
done
