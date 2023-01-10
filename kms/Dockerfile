FROM fedora:37@sha256:99aa8919afd1880064ec915dba44cdc5b52808667717f605750329d55006538a as build

RUN dnf -y update && \
    dnf install -y wget git

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

# Build
RUN mkdir -p /constellation/build
WORKDIR /constellation/kms/cmd
ARG PROJECT_VERSION=0.0.0
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o /constellation/build/kmsserver -trimpath -buildvcs=false -ldflags "-s -w -buildid='' -X github.com/edgelesssys/constellation/v2/internal/constants.VersionInfo=${PROJECT_VERSION}"

FROM gcr.io/distroless/static:nonroot@sha256:1fa522fe6cfe020d50341f1ca561c099487bd44f8eb98d25d1920b07e05e40be as release
COPY --from=build /constellation/build/kmsserver /kmsserver
ENTRYPOINT ["/kmsserver"]
