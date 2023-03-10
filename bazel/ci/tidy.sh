#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

# shellcheck disable=SC2153
gazelle=$(realpath "${GAZELLE}")
go=$(realpath "${GO}")

cd "${BUILD_WORKSPACE_DIRECTORY}"

submodules=$(${go} list -f '{{.Dir}}' -m)

for mod in ${submodules}; do
  ${go} mod tidy -C "${mod}"
done

${gazelle} update-repos \
  -from_file=go.work \
  -to_macro=toolchains/go_module_deps.bzl%go_dependencies \
  -build_file_proto_mode=disable_global \
  -build_file_generation=on \
  -prune
