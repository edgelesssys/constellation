FROM golang:1.20.7@sha256:37c7d8580e0616e96914a67291b9693fe038ab677eb9b5deb81e541b2322f20f as builder

# Download project root dependencies
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
COPY operators/constellation-node-operator/api/go.mod ./operators/constellation-node-operator/api/go.mod
COPY operators/constellation-node-operator/api/go.sum ./operators/constellation-node-operator/api/go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY . .

# Build
WORKDIR /workspace/internal/api/versionsapi/cli
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o versionsapi .

FROM scratch as release

COPY --from=builder /workspace/internal/api/versionsapi/cli/versionsapi .

CMD ["/notIntendedToBeExecuted"]
