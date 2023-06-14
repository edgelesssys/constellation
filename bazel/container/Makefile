export DOCKER_BUILDKIT=1

CONTAINER_NAME := ghcr.io/edgelesssys/bazel-container
CONTAINER_VERSION := v6.2.1-0

.PHONY: container container-push

container:
	docker buildx build --platform=linux/amd64,linux/arm64 . -f Containerfile -t "$(CONTAINER_NAME):$(CONTAINER_VERSION)" --push
