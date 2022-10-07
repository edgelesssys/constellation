FROM fedora:36@sha256:486fd5578f93fbc57a519e34ad4b7cac927c3f8a95409baedf0c19e9f287c207 as build

RUN dnf -y update && \
    dnf -y install libvirt-devel @development-tools pkg-config wget git && \
    dnf clean all

ARG GO_VER=1.19.2
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

FROM fedora:36@sha256:486fd5578f93fbc57a519e34ad4b7cac927c3f8a95409baedf0c19e9f287c207 as release
RUN  dnf -y install libvirt-devel && \
    dnf clean all
COPY --from=build /qemu-metadata-api/hack/qemu-metadata-api/api-server /server
ENTRYPOINT [ "/server" ]
