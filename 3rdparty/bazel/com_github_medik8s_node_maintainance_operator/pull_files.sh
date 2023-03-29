#!/usr/bin/env bash

###### script header ######

lib=$(realpath @@BASE_LIB@@) || exit 1
stat "${lib}" >> /dev/null || exit 1

# shellcheck source=../../../bazel/sh/lib.bash
if ! source "${lib}"; then
  echo "Error: could not find import"
  exit 1
fi

operator_source=(@@API_SRC@@)

###### script body ######

operator_real_source=()
for file in "${operator_source[@]}"; do
  if [[ ${file} =~ _test\.go$ ]]; then
    continue
  fi
  operator_real_source+=("$(realpath "${file}")")
done

cd "${BUILD_WORKSPACE_DIRECTORY}" # needs to be done after realpath

targetDir="3rdparty/node-maintenance-operator/api/v1beta1"

rm -rf "${targetDir}"/*.go
cp "${operator_real_source[@]}" "${targetDir}"
