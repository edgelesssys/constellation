FROM fedora:40@sha256:7cdd2b48396929bb8723ea2fa60e03bee39cc22e2a853cbd891587fab4eb1bc9 AS build

ARG LOGSTASH_VER=8.6.1

RUN curl -fsSLO https://artifacts.opensearch.org/logstash/logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz
RUN tar -zxvf logstash-oss-with-opensearch-output-plugin-$LOGSTASH_VER-linux-x64.tar.gz

FROM fedora:40@sha256:7cdd2b48396929bb8723ea2fa60e03bee39cc22e2a853cbd891587fab4eb1bc9 AS release

COPY --from=build logstash-* /usr/share/logstash

COPY debugd/logstash/config/ /usr/share/logstash/config/
COPY debugd/logstash/templates/ /usr/share/constellogs/templates/

RUN chmod -R 777 /usr/share/logstash/data/

RUN useradd -s /bin/bash logstash
USER logstash

ENTRYPOINT ["/usr/share/logstash/bin/logstash"]
