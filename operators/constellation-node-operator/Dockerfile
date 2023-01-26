# Build the manager binary
FROM golang:1.19.5@sha256:a0b51fe882f269828b63e7f69e6925f85afc548cf7cf967ecbfbcce6afe6f235 as builder

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
FROM gcr.io/distroless/static:nonroot@sha256:26d07ba1f954c02943786e352bc2c8f4eac719ae2f76a0ced68a953bed93a779 as release
WORKDIR /
COPY --from=builder /workspace/operators/constellation-node-operator/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
