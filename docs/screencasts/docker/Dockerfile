FROM ubuntu:22.04@sha256:58b87898e82351c6cf9cf5b9f3c20257bb9e2dcf33af051e12ce532d7f94e3fe

# Install requirements
RUN apt-get update && apt-get install -y software-properties-common &&\
    apt-add-repository ppa:zanchey/asciinema && apt-get update &&\
    apt-get install -y curl expect asciinema sudo unzip &&\
    rm -rf /var/lib/apt/lists/*

RUN curl -fsSLO https://github.com/mikefarah/yq/releases/download/v4.30.8/yq_linux_amd64 &&\
    sudo install yq_linux_amd64 /usr/local/bin/yq && rm yq_linux_amd64

RUN curl -fsSLO https://dl.k8s.io/release/v1.26.0/bin/linux/amd64/kubectl &&\
    sudo install kubectl /usr/local/bin/kubectl && rm kubectl

RUN curl -fsSLO https://releases.hashicorp.com/terraform/1.3.8/terraform_1.3.8_linux_amd64.zip &&\
    unzip terraform_1.3.8_linux_amd64.zip &&\
    sudo install terraform /usr/local/bin/terraform && rm terraform terraform_1.3.8_linux_amd64.zip

RUN curl -fsSLO https://github.com/edgelesssys/constellation/releases/latest/download/constellation-linux-amd64 &&\
    sudo install constellation-linux-amd64 /usr/local/bin/constellation &&\
    rm constellation-linux-amd64

# As mount point for $HOME/.config/gcloud
RUN mkdir /root/.config

# Disable spinner when running Constellation CLI commands
ENV CONSTELL_NO_SPINNER=1
# Enable RGB colors in PS1
ENV TERM=xterm-256color
# Set width of terminal, default is ~80 and leads to broken lines for long lines,
# e.g., curl & cosign commands.
ENV COLUMNS=512
# For PS1 to work shell needs to specified
ENV SHELL=/bin/bash
# ANSI color codes are used to control PS1 prompt. We use "\033[38;2;<r>;<g>;<b>m"
# to control the foreground color with RBG colors [1]. Non-printable characters
# need to  be escaped with additional \[ and \], see [2].
# [1]: https://stackoverflow.com/a/33206814/2306355
# [2]: https://stackoverflow.com/a/19501528/2306355
RUN echo 'export PS1="\[\033[38;2;139;4;221m\]$\[\033[0m\] "' >> /root/.bashrc

# Copy install scripts
COPY ./*.expect /scripts/

WORKDIR /constellation
ENTRYPOINT ["/usr/bin/expect", "-f"]
CMD ["/scripts/verify-cli.expect"]
