FROM fedora:37@sha256:f2c083c0b7d2367a375f15e002c2dc7baaca2b3181ace61f9d5113a8fe2f6b44 as build

RUN dnf -y update && \
    dnf install -y iproute iputils wget git && \
    dnf clean all

# Install Go
ARG GO_VER=1.19.4
RUN wget -q https://go.dev/dl/go${GO_VER}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VER}.linux-amd64.tar.gz && \
    rm go${GO_VER}.linux-amd64.tar.gz
ENV PATH ${PATH}:/usr/local/go/bin

# Download go dependencies
WORKDIR /constellation/
COPY go.mod ./
COPY go.sum ./
RUN go mod download all

# Copy Repo
COPY . /constellation
RUN rm -rf ./hack/

WORKDIR /constellation/joinservice
ARG PROJECT_VERSION=0.0.0
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o join-service -trimpath -buildvcs=false -ldflags "-s -w -buildid='' -X github.com/edgelesssys/constellation/v2/internal/constants.VersionInfo=${PROJECT_VERSION}" ./cmd/

# Use gcr.io/distroless/static here since we need CA certificates to be installed for aTLS operations on GCP.
FROM gcr.io/distroless/static@sha256:ebd8cc37d22551dce0957ba8e58f03b22a8448bbf844c8c9ded4feef883b36bc as release
COPY --from=build /constellation/joinservice/join-service /joinservice
ENTRYPOINT [ "/joinservice" ]
