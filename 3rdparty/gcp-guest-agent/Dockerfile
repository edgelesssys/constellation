FROM ubuntu:22.04@sha256:58b87898e82351c6cf9cf5b9f3c20257bb9e2dcf33af051e12ce532d7f94e3fe as build

# Install packages
RUN apt-get update && apt-get install -y \
    wget \
    git

# Install Go
ARG GO_VER=1.22.3
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
