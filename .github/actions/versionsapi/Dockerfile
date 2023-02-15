FROM golang:1.20.1@sha256:d19ee8512191c8b8e967246d4a7d0de4f4133a30bd9ce982e9b70a0c596dbf18 as builder

# Download project root dependencies
WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY . .

# Build
WORKDIR /workspace/internal/versionsapi/cli
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o versionsapi .

FROM scratch as release

COPY --from=builder /workspace/internal/versionsapi/cli/versionsapi .

CMD ["/notIntendedToBeExecuted"]
