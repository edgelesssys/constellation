FROM fedora:42@sha256:aa7befe5cfd1f0e062728c16453cd1c479d4134c7b85eac00172f3025ab0d522 AS build

ARG LOGSTASH_VER=8.6.1

RUN curl -fsSLO https://artifacts.opensearch.org/logstash/logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz
RUN tar -zxvf logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz

FROM fedora:42@sha256:aa7befe5cfd1f0e062728c16453cd1c479d4134c7b85eac00172f3025ab0d522 AS release

COPY --from=build logstash-* /usr/share/logstash

COPY debugd/logstash/config/ /usr/share/logstash/config/
COPY debugd/logstash/templates/ /usr/share/constellogs/templates/

RUN chmod -R 777 /usr/share/logstash/data/

RUN useradd -s /bin/bash logstash
USER logstash

ENTRYPOINT ["/usr/share/logstash/bin/logstash"]
