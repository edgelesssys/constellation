FROM ubuntu:24.04@sha256:72297848456d5d37d1262630108ab308d3e9ec7ed1c3286a32fe09856619a782 as build

# Install packages
RUN apt-get update && apt-get install -y \
    wget \
    git

# Install Go
ARG GO_VER=1.23.5
RUN wget -q https://go.dev/dl/go${GO_VER}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VER}.linux-amd64.tar.gz && \
    rm go${GO_VER}.linux-amd64.tar.gz
ENV PATH ${PATH}:/usr/local/go/bin

# Download go dependencies
WORKDIR /src/
COPY go.mod ./
COPY go.sum ./
RUN go mod download all

# Build
ARG VERSION=latest
COPY . /src
WORKDIR /src/google_guest_agent
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${VERSION}" -mod=readonly
WORKDIR /

FROM scratch as run
COPY --from=build /src/google_guest_agent/google_guest_agent /
COPY instance_configs.cfg /etc/default/
ENTRYPOINT [ "/google_guest_agent" ]
