FROM fedora:38@sha256:a29e3f593f5eb60a0a52d34bdc0147da22c97381d3667c5588dbe02ca35e1514 AS build

ARG LOGSTASH_VER=8.6.1

RUN curl -fsSLO https://artifacts.opensearch.org/logstash/logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz
RUN tar -zxvf logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz

FROM fedora:38@sha256:a29e3f593f5eb60a0a52d34bdc0147da22c97381d3667c5588dbe02ca35e1514 AS release

COPY --from=build logstash-* /usr/share/logstash

COPY debugd/internal/debugd/logcollector/logstash/config/ /usr/share/logstash/config/
COPY debugd/internal/debugd/logcollector/logstash/templates/ /usr/share/constellogs/templates/

RUN chmod -R 777 /usr/share/logstash/data/

RUN useradd -s /bin/bash logstash
USER logstash

ENTRYPOINT ["/usr/share/logstash/bin/logstash"]
