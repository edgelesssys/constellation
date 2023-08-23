# Use this dockerfile to build the image used in the Kubernetes job that runs the malicious join test.
FROM scratch

# Make sure to move the built binary to the build context before, e.g. via:
# `go test -c github.com/edgelesssys/constellation/v2/e2e/malicious-join`
COPY malicious-join.test /

ENTRYPOINT [ "/malicious-join.test" ]
