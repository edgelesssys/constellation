#!/usr/bin/env bash

# updates the Helm charts for the AWS Load Balancer Controller in the CLI.
# script is mostly copied from cli/internal/helm/update-csi-charts.sh

set -euo pipefail
set -o errtrace
shopt -s inherit_errexit

echo "Updating AWS Load Balancer Controller Helm chart..."
branch="v0.0.140" # releases can update the AWS load-balancer-controller chart
# Required tools
if ! command -v git &> /dev/null; then
  echo "git could not be found"
  exit 1
fi


callDir=$(pwd)
repo_tmp_dir=$(mktemp -d)

chart_base_path="charts/edgeless/constellation-services/charts"
chart_name="aws-load-balancer-controller"

chart_url="https://github.com/aws/eks-charts"
chart_dir="stable/aws-load-balancer-controller"
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
rm -r "${chart_base_path:?}/${chart_name}"

# move new chart
mkdir -p "${chart_base_path}/${chart_name}"
cp -r "${repo_tmp_dir}/${chart_dir}"/* "${chart_base_path}/${chart_name}"
