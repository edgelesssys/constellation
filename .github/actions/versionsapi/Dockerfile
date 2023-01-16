FROM golang:1.19.4@sha256:766625f2182dacec4c8774355a65a81a3b73acb0b4287b6a32a8efc185aede2c as builder

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
