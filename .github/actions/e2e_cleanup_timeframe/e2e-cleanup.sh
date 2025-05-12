#!/bin/bash

# get_e2e_test_ids_on_date gets all workflow IDs of workflows that contain "e2e" on a specific date.
function get_e2e_test_ids_on_date {
  ids="$(gh run list --created "$1" --status failure --json createdAt,workflowName,databaseId --jq '.[] | select(.workflowName | (contains("e2e") or contains("Release")) and (contains("MiniConstellation") | not)) | .databaseId' -L1000 -R edgelesssys/constellation || exit 1)"
  echo "${ids}"
}

# download_tfstate_artifact downloads all artifacts matching the pattern terraform-state-* from a given workflow ID.
function download_tfstate_artifact {
  gh run download "$1" -p "terraform-state-*" -R edgelesssys/constellation > /dev/null
}

# delete_terraform_resources runs terraform destroy on the given folder.
function delete_terraform_resources {
  delete_err=0
  if pushd "${1}/${2}"; then
    # Workaround for cleaning up Azure resources
    # We include a data source that is only used to generate output
    # If this data source is deleted before we call terraform destroy,
    # terraform will first try to evaluate the data source and fail,
    # causing the destroy to fail as well.
    sed -i '/data "azurerm_user_assigned_identity" "uaid" {/,/}/d' main.tf
    sed -i '/output "user_assigned_identity_client_id" {/,/}/d' outputs.tf

    terraform init > /dev/null || delete_err=1 # first, install plugins
    terraform destroy -auto-approve || delete_err=1
    popd || exit 1
  fi
  return "${delete_err}"
}

# check if the password for artifact decryption was given
if [[ -z ${ENCRYPTION_SECRET} ]]; then
  echo "ENCRYPTION_SECRET is not set. Please set an environment variable with that secret."
  exit 1
fi

artifact_pwd=${ENCRYPTION_SECRET}

shopt -s nullglob

start_date=$(date "+%Y-%m-%d")
end_date=$(date --date "-4 day" "+%Y-%m-%d")
dates_to_clean=()

# get all dates of the last week
while [[ ${end_date} != "${start_date}" ]]; do
  dates_to_clean+=("${end_date}")
  end_date=$(date --date "${end_date} +1 day" "+%Y-%m-%d")
done

echo "[*] retrieving run IDs for cleanup"
database_ids=()
for d in "${dates_to_clean[@]}"; do
  echo "    retrieving run IDs from $d"
  mapfile -td " " tmp < <(get_e2e_test_ids_on_date "$d")
  database_ids+=("${tmp[*]}")
done

# cleanup database_ids
mapfile -t database_ids < <(echo "${database_ids[@]}")
mapfile -td " " database_ids < <(echo "${database_ids[@]}")

echo "[*] downloading terraform state artifacts"
for id in "${database_ids[@]}"; do
  if [[ ${id} == *[^[:space:]]* ]]; then
    echo "    downloading from workflow ${id}"
    download_tfstate_artifact "${id}"
  fi
done

echo "[*] extracting artifacts"
for directory in ./terraform-state-*; do
  echo "    extracting ${directory}"

  # extract and decrypt the artifact
  7zz x -t7z -p"${artifact_pwd}" -o"${directory}" "${directory}/archive.7z" > /dev/null || exit 1
done

# create terraform caching directory
mkdir "${HOME}/tf_plugin_cache"
export TF_PLUGIN_CACHE_DIR="${HOME}/tf_plugin_cache"
echo "[*] created terraform cache directory ${TF_PLUGIN_CACHE_DIR}"

echo "[*] deleting resources"
error_occurred=0
for directory in ./terraform-state-*; do
  echo "    deleting resources in ${directory}"
  if ! delete_terraform_resources "${directory}" "constellation-terraform"; then
    echo "[!] deleting resources failed"
    error_occurred=1
  fi
  echo "    deleting IAM configuration in ${directory}"
  if ! delete_terraform_resources "${directory}" "constellation-iam-terraform"; then
    echo "[!] deleting IAM resources failed"
    error_occurred=1
  fi
  echo "    deleting directory ${directory}"
  rm -rf "${directory}"
done

if [[ ${error_occurred} -ne 0 ]]; then
  echo "[!] Errors occurred during resource deletion."
  exit 1
fi

exit 0
