FROM golang:1.20.7@sha256:37c7d8580e0616e96914a67291b9693fe038ab677eb9b5deb81e541b2322f20f AS builder

WORKDIR /workspace
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go test -tags e2e -c github.com/edgelesssys/constellation/v2/e2e/malicious-join

FROM scratch AS release

COPY --from=builder /workspace/malicious-join.test /

ENTRYPOINT [ "/malicious-join.test" ]
