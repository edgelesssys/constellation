#!/usr/bin/env bash

###### script header ######

lib=$(realpath @@BASE_LIB@@) || exit 1
stat "${lib}" >> /dev/null || exit 1

# shellcheck source=../../../bazel/sh/lib.bash
if ! source "${lib}"; then
  echo "Error: could not find import"
  exit 1
fi

controller_policy_source="@@POLICY_SRC@@"

###### script body ######

controller_policy_real_source=$(realpath "${controller_policy_source}")

cd "${BUILD_WORKSPACE_DIRECTORY}" # needs to be done after realpath

targetDir="terraform/infrastructure/iam/aws/alb_policy.json"

cp "${controller_policy_real_source}" "${targetDir}"
