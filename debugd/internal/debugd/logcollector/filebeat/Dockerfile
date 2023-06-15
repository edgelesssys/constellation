FROM fedora:38@sha256:a29e3f593f5eb60a0a52d34bdc0147da22c97381d3667c5588dbe02ca35e1514 AS release

RUN dnf install -y https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.6.2-x86_64.rpm

COPY debugd/internal/debugd/logcollector/filebeat/filebeat.yml /usr/share/filebeat/filebeat.yml

COPY debugd/internal/debugd/logcollector/filebeat/inputs.yml /usr/share/filebeat/inputs.d/inputs.yml

ENTRYPOINT ["/usr/share/filebeat/bin/filebeat", "-e", "--path.home", "/usr/share/filebeat", "--path.data", "/usr/share/filebeat/data"]
