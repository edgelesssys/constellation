FROM fedora:37@sha256:a9fed38b343ea8a2722c78d5ad97d691421bf46f20f20076d34dd6948a2a792d as build

RUN dnf -y update && \
    dnf install -y wget git

# Install Go
ARG GO_VER=1.20.1
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

# Build
RUN mkdir -p /constellation/build
WORKDIR /constellation/keyservice/cmd
ARG PROJECT_VERSION=0.0.0
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o /constellation/build/keyservice -trimpath -buildvcs=false -ldflags "-s -w -buildid='' -X github.com/edgelesssys/constellation/v2/internal/constants.versionInfo=${PROJECT_VERSION}"

FROM gcr.io/distroless/static:nonroot@sha256:21e5d22dbe956542e93c28d3b01037fd42236aeef2d4efe3bd7fb48f11e126db as release
COPY --from=build /constellation/build/keyservice /keyservice
ENTRYPOINT ["/keyservice"]
