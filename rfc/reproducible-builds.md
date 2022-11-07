# Reproducible Builds
To ensure the security of constellation's supply chain, we need to make our software builds reproducible.
This is the only way to ensure a verifiable path from source code to binary.
Every step of the build process must be deterministic.

## Definition of Goals
1. All our OCIs executed in the cluster must be reproducible bit by bit .
1. For that, every compiled executable has to be deterministically compiled.
1. If we change parts of our codebase (i.e. add several other programming languages) we should not have to make major changes to the image build system.
1. Moving from `Docker` to a reproducible build system should have a minimum overhead. Ideally, we can reuse our existing `Dockerfile`s.
1. The tool that builds the OCIs should be battle proven and reliable.
1. The image should be lightweight.

## Steps to Achieve Goals

### Executables
* Remove metadata from binaries, namely:
  * Timestamps
  * Build IDs
  * Symbols
  * Version numbers
  * Paths
  * Anything that might change for another host OS / time / ...
* Eliminate dependencies on libraries (make executable static)

Striping metadata from the binary can be done in the building process.
This can be achieved by setting the appropriate compiler flags.

```bash
$ CGO_ENABLED=0 go build -o <out_name> -buildvcs=false -trimpath -ldflags "-s -w -buildid=''"
```

### OCIs
For the OCIs to be deterministic, each component of the image has to be deterministic as well.
This includes:
* The base image used to build the software must be the same for each build of a version. Pin the version with its `sha256` hash checksum.
* The timestamps of the files in the image (creation, modification) must be identical for each build.
* Every component that is shipped with the image has to be identical.

To achieve this we will use [buildah](https://github.com/containers/buildah). It currently meets all the requirements mentioned above.

To ensure that the final image is deterministic, a pattern such as the following should be followed:

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
This is done as follows with `buildah:`

```sh
buildah build \
    --timestamp 0 \
    -t <image_name>
```

The result is an image with one deterministic layer (pinned by the hash) and deterministic build artifacts.
Hence, the entire build is reproducible.
