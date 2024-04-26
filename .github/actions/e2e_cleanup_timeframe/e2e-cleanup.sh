#!/bin/bash

# get_e2e_test_ids_on_date gets all workflow IDs of workflows that contain "e2e" on a specific date.
function get_e2e_test_ids_on_date {
  ids="$(gh run list --created "$1" --status failure --json createdAt,workflowName,databaseId --jq '.[] | select(.workflowName | contains("e2e") and (contains("MiniConstellation") | not)) | .databaseId' -L1000 -R edgelesssys/constellation || exit 1)"
  echo "$ids"
}

# download_tfstate_artifact downloads all artifacts matching the pattern terraform-state-* from a given workflow ID.
function download_tfstate_artifact {
  gh run download "$1" -p "terraform-state-*" -R edgelesssys/constellation > /dev/null
}

# delete_resources runs terraform destroy on the constellation-terraform subfolder of a given folder.
function delete_resources {
  if [ -d "$1/constellation-terraform" ]; then
    cd "$1/constellation-terraform" || exit 1
    terraform init > /dev/null || exit 1 # first, install plugins
    terraform destroy -auto-approve || exit 1
    cd ../../ || exit 1
  fi
}

# delete_iam_config runs terraform destroy on the constellation-iam-terraform subfolder of a given folder.
function delete_iam_config {
  if [ -d "$1/constellation-iam-terraform" ]; then
    cd "$1/constellation-iam-terraform" || exit 1
    terraform init > /dev/null || exit 1 # first, install plugins
    terraform destroy -auto-approve || exit 1
    cd ../../ || exit 1
  fi
}

# check if the password for artifact decryption was given
if [[ -z $ENCRYPTION_SECRET ]]; then
  echo "ENCRYPTION_SECRET is not set. Please set an environment variable with that secret."
  exit 1
fi

artifact_pwd=$ENCRYPTION_SECRET

shopt -s nullglob

start_date=$(date "+%Y-%m-%d")
end_date=$(date --date "-7 day" "+%Y-%m-%d")
dates_to_clean=()

# get all dates of the last week
while [[ $end_date != "$start_date" ]]; do
  dates_to_clean+=("$end_date")
  end_date=$(date --date "$end_date +1 day" "+%Y-%m-%d")
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
  if [[ $id == *[^[:space:]]* ]]; then
    echo "    downloading from workflow $id"
    download_tfstate_artifact "$id"
  fi
done

echo "[*] extracting artifacts"
for directory in ./terraform-state-*; do
  echo "    extracting $directory"

  # extract and decrypt the artifact
  unzip -d "${directory}" -P "$artifact_pwd" "$directory/archive.zip" > /dev/null || exit 1
done

# create terraform caching directory
mkdir "$HOME/tf_plugin_cache"
export TF_PLUGIN_CACHE_DIR="$HOME/tf_plugin_cache"
echo "[*] created terraform cache directory $TF_PLUGIN_CACHE_DIR"

echo "[*] deleting resources"
for directory in ./terraform-state-*; do
  echo "    deleting resources in $directory"
  delete_resources "$directory"
  echo "    deleting IAM configuration in $directory"
  delete_iam_config "$directory"
  echo "    deleting directory $directory"
  rm -rf "$directory"
done

exit 0
