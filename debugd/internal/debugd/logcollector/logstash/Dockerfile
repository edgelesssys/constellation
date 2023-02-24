FROM fedora:37@sha256:a9fed38b343ea8a2722c78d5ad97d691421bf46f20f20076d34dd6948a2a792d AS build

ARG LOGSTASH_VER=8.4.0

RUN curl -fsSLO https://artifacts.opensearch.org/logstash/logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz
RUN tar -zxvf logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz

FROM fedora:37@sha256:a9fed38b343ea8a2722c78d5ad97d691421bf46f20f20076d34dd6948a2a792d AS release

COPY --from=build logstash-* /usr/share/logstash

COPY debugd/internal/debugd/logcollector/logstash/config/ /usr/share/logstash/config/
COPY debugd/internal/debugd/logcollector/logstash/templates/ /usr/share/constellogs/templates/

RUN chmod -R 777 /usr/share/logstash/data/

RUN useradd -s /bin/bash logstash
USER logstash

ENTRYPOINT ["/usr/share/logstash/bin/logstash"]
