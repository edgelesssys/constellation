FROM golang:1.21.2 AS gobuild

COPY ./build/aws-sdk-go /aws-sdk-go
COPY ./build/versioning /versioning

WORKDIR /aws-sdk-go
RUN CGO_ENABLED=0 go build --ldflags "-s -w"

WORKDIR /versioning
RUN CGO_ENABLED=0 go build --ldflags "-s -w"

FROM openjdk:8-alpine3.9 as javabuild

ENV ANT_VERSION=1.10.12
RUN wget http://archive.apache.org/dist/ant/binaries/apache-ant-${ANT_VERSION}-bin.tar.gz \
    && tar xvfvz apache-ant-${ANT_VERSION}-bin.tar.gz -C /opt \
    && ln -sfn /opt/apache-ant-${ANT_VERSION} /opt/ant \
    && sh -c 'echo ANT_HOME=/opt/ant >> /etc/environment' \
    && ln -sfn /opt/ant/bin/ant /usr/bin/ant \
    && rm apache-ant-${ANT_VERSION}-bin.tar.gz \
    && apk update && apk add bash

COPY ./build/aws-sdk-java /aws-sdk-java
RUN /aws-sdk-java/install.sh

FROM alpine:3.18.4

ENV LANG C.UTF-8
ENV MINT_ROOT_DIR /mint
ENV MINT_DATA_DIR=/mint/data

RUN apk update && \
    apk add git python3 py3-pip openjdk8-jre ruby ruby-dev ruby-bundler jq bash patch openssl

COPY . /mint

RUN /mint/create-data-files.sh

# Putting these two into separate stages would allow for more parallelism.
RUN /mint/build/aws-sdk-ruby/install.sh
RUN /mint/build/awscli/install.sh

COPY --from=gobuild /aws-sdk-go/aws-sdk-go /mint/run/core/aws-sdk-go/aws-sdk-go
COPY --from=gobuild /versioning/tests /mint/run/core/versioning/tests
COPY --from=javabuild /aws-sdk-java/build/jar/FunctionalTests.jar /mint/run/core/aws-sdk-java/FunctionalTests.jar

WORKDIR /mint

ENTRYPOINT ["/mint/entrypoint.sh"]
