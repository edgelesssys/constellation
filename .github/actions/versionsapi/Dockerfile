FROM golang:1.20.0@sha256:61dafe97cc20b12faef7a744a1a3e43411c45b9908c3ccf9cadd05566e473e1d as builder

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
