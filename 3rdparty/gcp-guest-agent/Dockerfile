FROM ubuntu:24.04@sha256:a08e551cb33850e4740772b38217fc1796a66da2506d312abe51acda354ff061 as build

# Install packages
RUN apt-get update && apt-get install -y \
    wget \
    git

# Install Go
ARG GO_VER=1.24.6
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
