FROM fedora:36@sha256:455fec9590de794fbc21f61dbc7e90bf9918b58492d2a03fa269c09db47b43f6 as build

RUN dnf -y update && \
    dnf -y install @development-tools pkg-config iproute iputils wget git jq openssl-devel cryptsetup-libs cryptsetup-devel && \
    dnf clean all

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

# Build the access_manager
WORKDIR /constellation/access_manager/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o /constellation/build/access_manager -ldflags "-s -w" .

# Copy the access_manager from build into a scratch container, which is eventually deployed into the cluster
FROM scratch as release
COPY --from=build /constellation/build/access_manager /access_manager
ENTRYPOINT [ "/access_manager" ]
