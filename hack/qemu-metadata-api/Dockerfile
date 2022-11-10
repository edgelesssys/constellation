FROM fedora:38@sha256:c7dfa518d9db440fb02362c0f9b014c0e1b8e04bc0f6bf540d1d5ac2ecb43453 as build

RUN dnf -y update && \
    dnf -y install libvirt-devel @development-tools pkg-config wget git && \
    dnf clean all

ARG GO_VER=1.19.3
RUN wget -q https://go.dev/dl/go${GO_VER}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VER}.linux-amd64.tar.gz && \
    rm go${GO_VER}.linux-amd64.tar.gz
ENV PATH ${PATH}:/usr/local/go/bin

WORKDIR /qemu-metadata-api
COPY go.mod ./
COPY go.sum ./
RUN go mod download all

WORKDIR /qemu-metadata-api/hack
COPY hack/go.mod ./go.mod
COPY hack/go.sum ./go.sum
RUN go mod download all

COPY . /qemu-metadata-api
WORKDIR /qemu-metadata-api/hack/qemu-metadata-api
RUN go build -o api-server ./main.go

FROM fedora:38@sha256:c7dfa518d9db440fb02362c0f9b014c0e1b8e04bc0f6bf540d1d5ac2ecb43453 as release
RUN  dnf -y install libvirt-devel && \
    dnf clean all
COPY --from=build /qemu-metadata-api/hack/qemu-metadata-api/api-server /server
ENTRYPOINT [ "/server" ]
