# RFC 006: Reproducible Builds

To ensure the security of constellation's supply chain, we need to make our software builds reproducible.
This is the only way to ensure a verifiable path from source code to binary.
Every step of the build process must be deterministic.

## Definition of Goals

1. All our OCIs executed in the cluster must be reproducible bit by bit.
1. For that, every compiled executable has to be deterministically compiled.
1. If we change parts of our codebase (i.e. add several other programming languages) we should not have to make major changes to the image build system.
1. Since docker does not offer built-in options to remove timestamps (which breaks reproducibility) we have to move to another OCI build system.
1. The tool that builds the OCIs should be battle proven and reliable.
1. The image should be lightweight.

## OCI build alternatives

This is a list comparing different OCI builders.
This list does not claim to be complete, since we only focus on points relevant to our current needs.

## [`buildah/podman`](https://github.com/containers/buildah)

Since podman internally [uses](https://github.com/containers/buildah#buildah-and-podman-relationship) buildah to build the image and this rfc only covers building and not execution the names are used synonymously.
With buildah, Containerfiles/Dockerfiles can be used as usual. This means we can adjust the build to include necessary libraries for i.e CGO builds. Only the build command itself has to be adjusted to omit timestamps.

## [`ko`](https://github.com/ko-build/ko)

Ko is limited to building OCI images for go applications. By default images such as [distroless](https://github.com/GoogleContainerTools/distroless) are used as base images.
These are minimal and hence very small images, that are stripped of anything but runtime dependencies.
Problems arise when the default images do not satisfy our dependency needs (as they currently do with the `disk-mapper` which relies on a dynamically linked `libcryptsetup` library).
To solve this issue we have three options:

1. Build our own base images independent from or inspired by distroless
2. Fork distroless, edit underlying [bazel dependencies](https://github.com/GoogleContainerTools/distroless/blob/main/debian_archives.bzl), build the image
3. Use `apko` to build minimal Alpine images. These images can be configured via `apko` and a declarative `*.yaml` config file.

Option `1.`: Results in a similar maintenance work as using `buildah`.

Option `2.`: Results in an even bigger maintenance overhead since we currently do not use `bazel` at all.

Option `3.`: For our current use cases very easy to configure.

## [`kaniko`](https://github.com/GoogleContainerTools/kaniko)

Over time, issues complaining about breaking/inconsistent reproducibility accumulated.
This seems to happen more or less regularly. We should try to avoid a build system having these issues.

## Steps to Achieve Goals

## Executables

* Remove metadata from binaries, namely:
  * Timestamps
  * Build IDs
  * Symbols
  * Version numbers
  * Paths
  * Anything that might change for another host OS / time / ...
* Eliminate dependencies on libraries (make executable static)

Striping metadata from the binary can be done in the building process.
This can be achieved by setting the appropriate compiler and linker flags (see [`go tool link`](https://pkg.go.dev/cmd/link) and [`go help build`](https://pkg.go.dev/cmd/go)).

* `buildvcs=false`: Omit version control information
* `-trimpath`: Remove file system paths from executable
* `-s`: Remove the symbol table
* `-w`: Disable [DWARF](https://en.wikipedia.org/wiki/DWARF) generation
* `-buildid=""`: Unset build ID

A reference compilation could look like this:

```bash
$ CGO_ENABLED=0 go build -o <out_name> -buildvcs=false -trimpath -ldflags "-s -w -buildid=''"
```

## OCIs

* For the OCIs to be deterministic, each component of the image has to be deterministic as well.
This includes:
* The base image used to build the software must be the same for each build of a version. Pin the version with its `sha256` hash checksum. For that it has to be guaranteed, that the image is available as long as we need it.
* The timestamps of the files in the image (creation, modification) must be identical for each build.
* Every component that is shipped with the image has to be identical.
* We must ensure, that the pinned images are always available. Since we probably cannot use stock images due to our dependencies, this is a step we have to take anyway.

### `buildah`

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

### `apko` / `ko`

To include c libraries into a distroless minimal image, we have to rebuild the base image.
For that, we can use `apko`.
It can be configured using a `*.yaml` file and is easy to use. An exemplary image definition could look like this:

```yaml
contents:
  repositories:
    - https://dl-cdn.alpinelinux.org/alpine/edge/main
    - https://dl-cdn.alpinelinux.org/alpine/edge/community  # for community packages
  packages:
    - alpine-base
    - <custom-package>

environment:
  PATH: /usr/sbin:/sbin:/usr/bin:/bin
```

To build this image, use the official docker container as [recommended](https://github.com/chainguard-dev/apko#installation) by `chainguard`.
This produces a container image that can be pushed to a remote registry and a [tarred](https://docs.podman.io/en/latest/markdown/podman-save.1.html) export of the image locally.

```sh
docker run -v "$PWD":/work cgr.dev/chainguard/apko build <modified-base-image>.yaml <image-name>:<tag> <image-name>.tar
```

Then in our `.ko.yaml`, we can use the newly created image as a base image, also just for certain build ids:

```yaml
baseImageOverrides:
  github.com/edgelesssys/constellation/v2/keyservice/cmd: edgelesssys/alpine-custom:base
```

The result is also a reproducible OCI image with reproducible artifacts.

## Considerations

Finally we can conclude, that both `buildah` and `ko` get the job done.
`buildah` constructs the images in a procedural way such as we are used to by writing `Dockerfile`s, while `ko`/`apko` configures the images in a declarative way.
Since `ko`/`apko` are very easy to use and we currently only use `go` in our microservices, `ko`/`apko` can do everything we need right now.
Further, the creation of minimal images is easier with `apko` than with `Containerfile`s.
