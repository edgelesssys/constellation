#!/usr/bin/env bash

# update-csi-charts updates the Helm charts for the CSI drivers in the CLI.

set -euo pipefail
set -o errtrace
shopt -s inherit_errexit

echo "Updating CSI Helm charts..."

# Required tools
if ! command -v git &> /dev/null; then
  echo "git could not be found"
  exit 1
fi

if ! command -v yq &> /dev/null; then
  echo "yq could not be found"
  exit 1
fi

# download_chart downloads the Helm chart for the given CSI driver and version.
#
# Arguments:
#   $1: URL of the git repo containing the Helm chart
#   $2: branch or tag of the git repo
#   $3: path to the Helm chart in the git repo
#   $4: name of the Helm chart
download_chart() {
  chart_url=$1
  branch=$2
  chart_dir=$3
  chart_name=$4

  callDir=$(pwd)
  repo_tmp_dir=$(mktemp -d)

  csi_chart_path="charts/edgeless/csi"
  chart_base_path="${csi_chart_path}/charts"

  cd "${repo_tmp_dir}"
  git clone \
    --filter=blob:none \
    --no-checkout \
    --sparse \
    --depth 1 \
    --branch="${branch}" \
    "${chart_url}" "${repo_tmp_dir}"

  git sparse-checkout add "${chart_dir}"
  git checkout
  cd "${callDir}"

  # remove old chart
  rm -rf -- "${chart_base_path:?}/${chart_name}"

  # move new chart
  mkdir -p "${chart_base_path}/${chart_name}"
  cp -r "${repo_tmp_dir}/${chart_dir}"/* "${chart_base_path}/${chart_name}"

  rm -r -- "${repo_tmp_dir}"

  # get new version from Chart.yaml
  new_version=$(yq '.version' "${chart_base_path}/${chart_name}/Chart.yaml")

  # update dependency version in parent Chart.yaml
  yq -i "(.dependencies[] | select( .name== \"${chart_name}\").version) = \"${new_version}\"" "${csi_chart_path}/Chart.yaml"
}

## AWS CSI Driver
download_chart "https://github.com/edgelesssys/constellation-aws-ebs-csi-driver" "v1.1.1" "charts/aws-ebs-csi-driver" "aws-csi-driver"

## Azure CSI Driver
download_chart "https://github.com/edgelesssys/constellation-azuredisk-csi-driver" "v1.3.0" "charts/edgeless" "azuredisk-csi-driver"

## GCP CSI Driver
download_chart "https://github.com/edgelesssys/constellation-gcp-compute-persistent-disk-csi-driver" "v1.3.0" "charts" "gcp-compute-persistent-disk-csi-driver"

## OpenStack CSI Driver (cinder)
download_chart "https://github.com/edgelesssys/constellation-cloud-provider-openstack" "v1.0.1" "charts/cinder-csi-plugin" "openstack-cinder-csi"

echo # final newline
