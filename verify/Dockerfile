FROM fedora@sha256:36af84ba69e21c9ef86a0424a090674c433b2b80c2462e57503886f1d823abe8 as build

RUN dnf -y update && \
    dnf install -y iproute iputils wget git && \
    dnf clean all

# Install Go
ARG GO_VER=1.18
RUN wget https://go.dev/dl/go${GO_VER}.linux-amd64.tar.gz && \
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

WORKDIR /constellation/verify
ARG PROJECT_VERSION=0.0.0
RUN CGO_ENABLED=0 go build -o verify-service -trimpath -buildvcs=false -ldflags "-s -w -buildid='' -X github.com/edgelesssys/constellation/internal/constants.VersionInfo=${PROJECT_VERSION}" ./cmd/

FROM scratch AS release
COPY --from=build /constellation/verify/verify-service /verify
ENTRYPOINT [ "/verify" ]
