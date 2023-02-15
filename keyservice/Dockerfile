FROM fedora:37@sha256:23c63666eefb64cd2364e6d8ad327f06abf9eb1f34b621e9fd6c1602e142244b as build

RUN dnf -y update && \
    dnf install -y wget git

# Install Go
ARG GO_VER=1.19.5
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
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o /constellation/build/keyservice -trimpath -buildvcs=false -ldflags "-s -w -buildid='' -X github.com/edgelesssys/constellation/v2/internal/constants.VersionInfo=${PROJECT_VERSION}"

FROM gcr.io/distroless/static:nonroot@sha256:6b01107391648040c796967b49b7973188b7c9a6b1d49d06090db349248eba39 as release
COPY --from=build /constellation/build/keyservice /keyservice
ENTRYPOINT ["/keyservice"]
