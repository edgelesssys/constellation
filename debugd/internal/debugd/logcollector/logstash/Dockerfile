FROM docker.io/opensearchproject/logstash-oss-with-opensearch-output-plugin:7.16.3

RUN rm -f /usr/share/logstash/pipeline/logstash.conf

COPY config/ /usr/share/logstash/config/

COPY templates/ /usr/share/constellogs/templates/

ENTRYPOINT ["/usr/share/logstash/bin/logstash"]
