FROM fedora@sha256:455fec9590de794fbc21f61dbc7e90bf9918b58492d2a03fa269c09db47b43f6 as build

RUN dnf -y update && \
    dnf install -y wget git

# Install Go
ARG GO_VER=1.19.3
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
RUN CGO_ENABLED=0 go build -o /constellation/build/kmsserver -trimpath -buildvcs=false -ldflags "-s -w -buildid='' -X github.com/edgelesssys/constellation/v2/internal/constants.VersionInfo=${PROJECT_VERSION}"

FROM gcr.io/distroless/static:nonroot@sha256:d8afc7d6973f357162e2283551cf3347b2bb847a03d24510ee837f289505f8e3 as release
COPY --from=build /constellation/build/kmsserver /kmsserver
ENTRYPOINT ["/kmsserver"]
