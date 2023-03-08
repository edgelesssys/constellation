FROM fedora:37@sha256:a9fed38b343ea8a2722c78d5ad97d691421bf46f20f20076d34dd6948a2a792d as build

RUN dnf -y update && \
    dnf -y install libvirt-devel @development-tools pkg-config wget git && \
    dnf clean all

ARG GO_VER=1.20.2
RUN wget -q https://go.dev/dl/go${GO_VER}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VER}.linux-amd64.tar.gz && \
    rm go${GO_VER}.linux-amd64.tar.gz
ENV PATH ${PATH}:/usr/local/go/bin

WORKDIR /qemu-metadata-api
# Necessary to make `go mod download all` work while having a local replace rule in the root-go.mod.
COPY operators/constellation-node-operator/api/go.mod ./operators/constellation-node-operator/api/go.mod
COPY operators/constellation-node-operator/api/go.sum ./operators/constellation-node-operator/api/go.sum
COPY go.mod ./
COPY go.sum ./
RUN go mod download all

WORKDIR /qemu-metadata-api/hack
COPY hack/go.mod ./go.mod
COPY hack/go.sum ./go.sum
RUN go mod download all

COPY . /qemu-metadata-api
WORKDIR /qemu-metadata-api/hack/qemu-metadata-api
RUN go build -o api-server .

FROM fedora:37@sha256:a9fed38b343ea8a2722c78d5ad97d691421bf46f20f20076d34dd6948a2a792d as release
RUN  dnf -y install libvirt-devel && \
    dnf clean all
COPY --from=build /qemu-metadata-api/hack/qemu-metadata-api/api-server /server
ENTRYPOINT [ "/server" ]
