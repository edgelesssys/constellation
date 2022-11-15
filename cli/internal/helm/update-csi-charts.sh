#!/usr/bin/env bash

# update-csi-charts updates the Helm charts for the CSI drivers in the CLI.

set -euo pipefail
shopt -s inherit_errexit

# Required tools
if ! command -v git &> /dev/null; then
    echo "git could not be found"
    exit 1
fi

download_chart() {
    chart_url=$1
    branch=$2
    chart_dir=$3
    chart_name=$4

    repo_tmp_dir=$(mktemp -d)

    chart_base_path="charts/edgeless/constellation-services/charts/"

    pushd "${repo_tmp_dir}"
    git clone --filter=blob:none --no-checkout --sparse --depth 1 --branch="${branch}" "${chart_url}" "${repo_tmp_dir}"

    git sparse-checkout add "${chart_dir}"
    git checkout
    popd

    # remove old chart
    rm -r "${chart_base_path}${chart_name}"
    
    # move new chart
    mkdir -p "${chart_base_path}/${chart_name}"
    cp -r "${repo_tmp_dir}/${chart_dir}"/* "${chart_base_path}${chart_name}"
    rm -r "${repo_tmp_dir}"

    return
}


## GCP CSI Driver
# TODO: clone from main branch once we rebase on upstream
download_chart "https://github.com/edgelesssys/constellation-gcp-compute-persistent-disk-csi-driver" "ref/chart" "charts" "csi-gcp-pd"


## Azure CSI Driver
# TODO: https://github.com/edgelesssys/constellation/pull/548
