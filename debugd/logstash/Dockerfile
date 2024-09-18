FROM fedora:40@sha256:d0207dbb078ee261852590b9a8f1ab1f8320547be79a2f39af9f3d23db33735e AS build

ARG LOGSTASH_VER=8.6.1

RUN curl -fsSLO https://artifacts.opensearch.org/logstash/logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz
RUN tar -zxvf logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz

FROM fedora:40@sha256:d0207dbb078ee261852590b9a8f1ab1f8320547be79a2f39af9f3d23db33735e AS release

COPY --from=build logstash-* /usr/share/logstash

COPY debugd/logstash/config/ /usr/share/logstash/config/
COPY debugd/logstash/templates/ /usr/share/constellogs/templates/

RUN chmod -R 777 /usr/share/logstash/data/

RUN useradd -s /bin/bash logstash
USER logstash

ENTRYPOINT ["/usr/share/logstash/bin/logstash"]
