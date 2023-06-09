FROM golang:1.20.5@sha256:e88f338421aa37d614c3cf0fea40a978a943b5cc1b1b34ecff3cb7e1a6e79090 as builder

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
