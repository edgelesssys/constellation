FROM fedora@sha256:486fd5578f93fbc57a519e34ad4b7cac927c3f8a95409baedf0c19e9f287c207 as build

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
RUN CGO_ENABLED=0 go build -o /constellation/build/access_manager -ldflags "-s -w" .

# Copy the access_manager from build into a scratch container, which is eventually deployed into the cluster
FROM scratch as release
COPY --from=build /constellation/build/access_manager /access_manager
ENTRYPOINT [ "/access_manager" ]
