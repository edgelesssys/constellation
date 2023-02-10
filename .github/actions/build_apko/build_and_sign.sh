#!/usr/bin/env bash

set -exuo pipefail
shopt -s inherit_errexit

# buildImage <apko_config_path>
buildImage() {
  local imageConfig=$1

  echo "Building image for ${imageConfig}"

  local imageName
  imageName=$(basename "${imageConfig}" | cut -d. -f1)
  local registryPath
  registryPath="${REGISTRY}/edgelesssys/apko-${imageName}"
  local outTar
  outTar="${imageName}.tar"

  mkdir -p "sboms/${imageName}"

  # build the image
  docker run \
    -v "${PWD}":/work \
    cgr.dev/chainguard/apko@sha256:8952f4f3ce58052b7df5e46f230f7192b42b220d3e46c8b06178cc25fd700846 \
    build \
    "${imageConfig}" \
    --build-arch "${APKO_ARCH}" \
    --sbom \
    "${registryPath}" \
    "${outTar}"

  docker load < "${outTar}"

  for tag in ${CONTAINER_TAGS}; do
    tagSanitized=${tag//\//-}

    docker image tag "${registryPath}" "${registryPath}:${tagSanitized}"
    docker push "${registryPath}:${tagSanitized}"

    imageDigest=$(docker inspect --format='{{index .RepoDigests 0}}' "${registryPath}")

    # write full image as Markdown code block to step summary
    cat << EOF >> "${GITHUB_STEP_SUMMARY}"
\`\`\`
${imageDigest%%@*}:${tagSanitized}@${imageDigest##*@}
\`\`\`
EOF
  done

  # cosign the container and push to registry
  cosign sign \
    --key env://COSIGN_PRIVATE_KEY \
    "${imageDigest}" \
    -y

  # move sboms to folder
  mv sbom-*.* "sboms/${imageName}/"
}

if [[ -n ${APKO_CONFIG} ]]; then
  buildImage "${APKO_CONFIG}"
  exit 0
fi

echo "Building all images in image"
for imageConfig in ./*.yaml; do
  buildImage "${imageConfig}"
done
