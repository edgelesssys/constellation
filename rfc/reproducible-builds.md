# Reproducible Builds
To ensure the security of constellation's supply chain, we need to make our software builds reproducible.
Only this guarantees a verifiable path from source code to binary.
Every step of the build process has to be deterministic.

## Definition of Goals
1. All OCIs that are run in the cluster must be bit by bit reproducible.
1. For that, every compiled executable has to be deterministically compiled.
1. If we change parts of our codebase (i.e. add several other programming languages) we should not have to make major changes to the image build system.
1. Switching from `Docker` to a reproducible build system should have a minimum overhead. Ideally we can reuse our existing `Dockerfile`s.
1. The tool that builds the OCIs should be battle proven and reliable.
1. The image should be lightweight.

## Steps to Achieve Goals

### Executables
* Strip binaries of metadata, namely
  * Timestamps
  * Build ID's
  * Symbols
  * Version numbers
  * Paths
  * Anything that could change for another host OS / date / time ...
* Eliminate reliance on libraries (make it static)
* This can be done by setting the appropriate compiler flags.

```bash
$ CGO_ENABLED=0 go build -o <program> -buildvcs=false -trimpath -ldflags "-s -w -buildid=''"
```

### OCIs
For the OCIs to be deterministic, every component of the image has to be deterministic as well.
This includes:
* The base image, the software is build with, ash to be the same for every build of one version. Pin the version with i.e. it's `sha256` hash checksum.
* The timestamps of files in the image (creation, modification) have to be the same in every build
* Every component that is shipped with the image has to be identical.

To achieve this we will use [buildah](https://github.com/containers/buildah). Currently, it satisfies all our needs.

To guarantee that the final image is deterministic, a pattern such as the one below should be followed:

```Dockerfile
FROM <image>@sha256:<hash> as build
RUN  <install_deps>
RUN  <get_sources>
RUN  <build_deterministic>
RUN  ...

FROM <clean_base_image>@sha256:<hash>
COPY --from=build <artifacts> <path>
CMD  [<executable>]
```

This `Containerfile` must then be built reproducible.
This is done using the `--reproducible` flag for `buildah`:

```sh
buildah build \
    --timestamp 0 \
    -t <image_name>
```

The result is an image with one deterministic layer (pinned by the hash) and deterministic build artifacts.
Hence, the entire build is reproducible.
