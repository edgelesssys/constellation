FROM fedora@sha256:36af84ba69e21c9ef86a0424a090674c433b2b80c2462e57503886f1d823abe8 as build

RUN dnf -y update && \
    dnf -y install @development-tools pkg-config iproute iputils wget git jq openssl-devel cryptsetup-libs cryptsetup-devel && \
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

# Build the access_manager
WORKDIR /constellation/access_manager/
RUN CGO_ENABLED=0 go build -o /constellation/build/access_manager -ldflags "-s -w" .

# Copy the access_manager from build into a scratch container, which is eventually deployed into the cluster
FROM scratch as release
COPY --from=build /constellation/build/access_manager /access_manager
ENTRYPOINT [ "/access_manager" ]
