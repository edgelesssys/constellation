FROM golang:1.21.6@sha256:6fbd2d3398db924f8d708cf6e94bd3a436bb468195daa6a96e80504e0a9615f2 as builder

# Download project root dependencies
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
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
