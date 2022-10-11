FROM fedora@sha256:486fd5578f93fbc57a519e34ad4b7cac927c3f8a95409baedf0c19e9f287c207 as build

RUN dnf -y update && \
    dnf install -y iproute iputils wget git && \
    dnf clean all

# Install Go
ARG GO_VER=1.19.2
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

WORKDIR /constellation/verify
ARG PROJECT_VERSION=0.0.0
RUN CGO_ENABLED=0 go build -o verify-service -trimpath -buildvcs=false -ldflags "-s -w -buildid='' -X github.com/edgelesssys/constellation/v2/internal/constants.VersionInfo=${PROJECT_VERSION}" ./cmd/

FROM scratch AS release
COPY --from=build /constellation/verify/verify-service /verify
ENTRYPOINT [ "/verify" ]
