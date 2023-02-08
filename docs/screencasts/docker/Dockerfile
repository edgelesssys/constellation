FROM ubuntu:20.04

# Install requirements
RUN apt update
RUN apt install -y software-properties-common
RUN apt-add-repository ppa:zanchey/asciinema
RUN apt update
RUN apt install -y curl expect asciinema sudo
RUN curl -LO https://go.dev/dl/go1.19.5.linux-amd64.tar.gz && rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.5.linux-amd64.tar.gz
RUN echo 'export PATH="${PATH}:/usr/local/go/bin:/root/go/bin"' >> /root/.bashrc
ENV PATH="${PATH}:/usr/local/go/bin:/root/go/bin"

# Install cosign & rekor (speedup in recording)
RUN go install github.com/sigstore/cosign/cmd/cosign@latest
RUN go install github.com/sigstore/rekor/cmd/rekor-cli@latest

# Set prompt
ENV PS1='$ '

# Copy install scripts
COPY ./*.sh /root/

WORKDIR /root
ENTRYPOINT ["/usr/bin/expect", "-f"]
CMD ["verify-cli.sh", "/recordings/verify-cli.cast"]
