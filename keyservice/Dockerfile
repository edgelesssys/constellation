FROM fedora:37@sha256:3f987b7657e944cf87a129cc262982d4f80e38bd98f7db313ccaf90ca7069dd2 as build

RUN dnf -y update && \
    dnf install -y wget git

# Install Go
ARG GO_VER=1.20.2
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

FROM gcr.io/distroless/static:nonroot@sha256:149531e38c7e4554d4a6725d7d70593ef9f9881358809463800669ac89f3b0ec as release
COPY --from=build /constellation/build/keyservice /keyservice
ENTRYPOINT ["/keyservice"]
