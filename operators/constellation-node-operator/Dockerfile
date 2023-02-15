# Build the manager binary
FROM golang:1.20.1@sha256:d19ee8512191c8b8e967246d4a7d0de4f4133a30bd9ce982e9b70a0c596dbf18 as builder

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
FROM gcr.io/distroless/static:nonroot@sha256:6b01107391648040c796967b49b7973188b7c9a6b1d49d06090db349248eba39 as release
WORKDIR /
COPY --from=builder /workspace/operators/constellation-node-operator/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
