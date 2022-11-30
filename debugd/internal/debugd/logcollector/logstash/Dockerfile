FROM docker.io/opensearchproject/logstash-oss-with-opensearch-output-plugin:7.16.2 AS release

RUN rm -f /usr/share/logstash/pipeline/logstash.conf

COPY debugd/internal/debugd/logcollector/logstash/config/ /usr/share/logstash/config/

COPY debugd/internal/debugd/logcollector/logstash/templates/ /usr/share/constellogs/templates/

ENTRYPOINT ["/usr/share/logstash/bin/logstash"]
