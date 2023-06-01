FROM fedora:37@sha256:ab2fd8de428f5dbdb147d43135902a776a6c64d2243fd7ac99629df3b47c9256 AS build

ARG LOGSTASH_VER=8.6.1

RUN curl -fsSLO https://artifacts.opensearch.org/logstash/logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz
RUN tar -zxvf logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz

FROM fedora:37@sha256:ab2fd8de428f5dbdb147d43135902a776a6c64d2243fd7ac99629df3b47c9256 AS release

COPY --from=build logstash-* /usr/share/logstash

COPY debugd/internal/debugd/logcollector/logstash/config/ /usr/share/logstash/config/
COPY debugd/internal/debugd/logcollector/logstash/templates/ /usr/share/constellogs/templates/

RUN chmod -R 777 /usr/share/logstash/data/

RUN useradd -s /bin/bash logstash
USER logstash

ENTRYPOINT ["/usr/share/logstash/bin/logstash"]
