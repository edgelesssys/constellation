#!/usr/bin/env bash

# Script removes the .terraform directorys with the providers that were
# downloaded by terraform init.

set -euo pipefail
shopt -s inherit_errexit

dirs=$(find . -type d -name "*.terraform" | sort -ud)

if [[ -z ${dirs} ]]; then
  echo "No .terraform directories found"
  exit 0
fi

echo "Size of $(pwd): $(du -hs | cut -f1)"
echo
echo "Removing Terraform providers from the following directories:"
echo
echo "${dirs}"
echo

read -p "Should we delete them all? [y/n] " -n 1 -r
if [[ ! ${REPLY} =~ ^[Yy]$ ]]; then
  echo "Aborting."
  [[ $0 == "${BASH_SOURCE[0]}" ]] && exit 1 || return 1 # handle exits from shell or function but don't exit interactive shell
fi
echo

for dir in ${dirs}; do
  echo "Removing ${dir}"
  rm -rf "${dir}"
done

echo
echo "Size of $(pwd): $(du -hs | cut -f1)"
