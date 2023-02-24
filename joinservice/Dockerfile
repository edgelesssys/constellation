FROM fedora:37@sha256:a9fed38b343ea8a2722c78d5ad97d691421bf46f20f20076d34dd6948a2a792d as build

RUN dnf -y update && \
    dnf install -y iproute iputils wget git && \
    dnf clean all

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

WORKDIR /constellation/joinservice
ARG PROJECT_VERSION=0.0.0
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o join-service -trimpath -buildvcs=false -ldflags "-s -w -buildid='' -X github.com/edgelesssys/constellation/v2/internal/constants.VersionInfo=${PROJECT_VERSION}" ./cmd/

# Use gcr.io/distroless/static here since we need CA certificates to be installed for aTLS operations on GCP.
FROM gcr.io/distroless/static@sha256:5b2fa762fb6ebf66ff88ae1db2dc4ad8fc6ddf1164477297dfac1a09f20e7339 as release
COPY --from=build /constellation/joinservice/join-service /joinservice
ENTRYPOINT [ "/joinservice" ]
