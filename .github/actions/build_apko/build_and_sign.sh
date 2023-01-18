#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

# buildImage <apko_config_path>
function buildImage() {
  local imageConfig=$1

  echo "Building image for ${imageConfig}"

  local imageName
  imageName=$(basename "${imageConfig}" | cut -d. -f1)
  registryPath="${REGISTRY}/edgelesssys/apko-${imageName}"
  outTar="${imageName}.tar"

  mkdir -p "sboms/${imageName}"

  # build the image
  docker run \
    -v "${PWD}":/work \
    cgr.dev/chainguard/apko:"${APKO_TAG}" \
    build \
    "${imageConfig}" \
    --build-arch "${APKO_ARCH}" \
    --sbom \
    "${registryPath}" \
    "${outTar}"

  # push container
  docker load < "${outTar}"
  docker push "${registryPath}"
  imageDigest=$(docker inspect --format='{{index .RepoDigests 0}}' "${registryPath}")
  echo "${imageDigest}" >> "${GITHUB_STEP_SUMMARY}"

  # cosign the container and push to registry
  cosign sign \
    --key env://COSIGN_PRIVATE_KEY \
    "${imageDigest}" \
    -y

  # move sboms to folder
  mv sbom-*.* "sboms/${imageName}/"
}

mkdir "sboms"

if [[ -n ${APKO_CONFIG} ]]; then
  buildImage "${APKO_CONFIG}"
  exit 0
fi

echo "Building all images in image"
for imageConfig in apko/*.yaml; do
  buildImage "${imageConfig}"
done
