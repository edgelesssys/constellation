# Build the manager binary
FROM golang:1.20.2@sha256:2101aa981e68ab1e06e3d4ac35ae75ed122f0380e5331e3ae4ba7e811bf9d256 as builder

# Download project root dependencies
WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Download operator dependencies
WORKDIR /workspace/operators/constellation-node-operator
COPY operators/constellation-node-operator/go.mod ./
COPY operators/constellation-node-operator/go.sum ./
RUN go mod download all

WORKDIR /workspace
COPY . .

# Build
WORKDIR /workspace/operators/constellation-node-operator
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot@sha256:149531e38c7e4554d4a6725d7d70593ef9f9881358809463800669ac89f3b0ec as release
WORKDIR /
COPY --from=builder /workspace/operators/constellation-node-operator/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
