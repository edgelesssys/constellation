FROM golang:1.20.4@sha256:31a8f92b17829b3ccddf0add184f18203acfd79ccc1bcb5c43803ab1c4836cca as builder

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
WORKDIR /workspace/internal/versionsapi/cli
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o versionsapi .

FROM scratch as release

COPY --from=builder /workspace/internal/versionsapi/cli/versionsapi .

CMD ["/notIntendedToBeExecuted"]
