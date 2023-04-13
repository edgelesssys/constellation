export DOCKER_BUILDKIT=1

CONTAINER_NAME := ghcr.io/edgelesssys/bazel-container
CONTAINER_VERSION := v6.1.0-0

.PHONY: container container-push

container:
	docker build -f Containerfile -t "$(CONTAINER_NAME):$(CONTAINER_VERSION)" .

container-push:
	docker push "$(CONTAINER_NAME):$(CONTAINER_VERSION)"
