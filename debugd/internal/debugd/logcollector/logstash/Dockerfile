FROM fedora:37@sha256:f2c083c0b7d2367a375f15e002c2dc7baaca2b3181ace61f9d5113a8fe2f6b44 AS build

ARG LOGSTASH_VER=8.4.0

RUN curl -sLO https://artifacts.opensearch.org/logstash/logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz
RUN tar -zxvf logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz

FROM fedora:37@sha256:f2c083c0b7d2367a375f15e002c2dc7baaca2b3181ace61f9d5113a8fe2f6b44 AS release

COPY --from=build logstash-* /usr/share/logstash

COPY debugd/internal/debugd/logcollector/logstash/config/ /usr/share/logstash/config/
COPY debugd/internal/debugd/logcollector/logstash/templates/ /usr/share/constellogs/templates/

RUN chmod -R 777 /usr/share/logstash/data/

RUN useradd -s /bin/bash logstash
USER logstash

ENTRYPOINT ["/usr/share/logstash/bin/logstash"]
