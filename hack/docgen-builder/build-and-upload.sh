#!/usr/bin/env bash

# Usage: ./build-and-upload.sh [dryrun]

set -euo pipefail
set -o errtrace
shopt -s inherit_errexit

talosRepo="https://github.com/siderolabs/talos"
talosHash="94c24ca64e70f227da29cd02bd367d3c2701b96c"
s3CASPath="s3://cdn-constellation-backend/constellation/cas/sha256"
publicCASPath="https://cdn.confidential.cloud/constellation/cas/sha256"

function cleanup {
  echo "Cleaning up"
  rm -rf "${tmpDir}"
}

trap cleanup EXIT

# Set flags to --dryrun if arg 1 is "dryrun"
awsFlags=()
if [[ ${1-} == "dryrun" ]]; then
  awsFlags+=("--dryrun")
fi

# Create a temp dir to work in
tmpDir=$(mktemp -d)
pushd "${tmpDir}"

# Get the talos source code
wget -qO- "${talosRepo}/archive/${talosHash}.tar.gz" | tar -xz
cp -r "talos-${talosHash}/hack/docgen" .
pushd "docgen"

# Build and upload the talos-docgen binary
echo
for arch in "amd64" "arm64"; do
  for os in "linux" "darwin"; do
    echo "Building and uploading talos-docgen-${os}-${arch}"
    CGO_ENABLED="0" GOWORK="" GOOS="${os}" GOARCH="${arch}" go build -trimpath -ldflags="-buildid=" -o "talos-docgen-${os}-${arch}" .
    sum=$(shasum -a 256 "talos-docgen-${os}-${arch}" | cut -d ' ' -f1) && echo "Binary sha256: ${sum}"
    file "talos-docgen-${os}-${arch}"
    aws s3 "${awsFlags[@]}" cp "./talos-docgen-${os}-${arch}" "${s3CASPath}/${sum}"
    echo
    cat << EOF >> "bazelout.txt"
    http_file(
        name = "com_github_siderolabs_talos_hack_docgen_${os}_${arch}",
        urls = [
            "${publicCASPath}/${sum}",
        ],
        executable = True,
        sha256 = "${sum}",
    )
EOF
  done
done

# Print the bazel output
cat bazelout.txt
echo
