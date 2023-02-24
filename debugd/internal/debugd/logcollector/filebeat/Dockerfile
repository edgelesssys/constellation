FROM fedora:37@sha256:a9fed38b343ea8a2722c78d5ad97d691421bf46f20f20076d34dd6948a2a792d AS release

RUN dnf install -y https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.5.0-x86_64.rpm

COPY debugd/internal/debugd/logcollector/filebeat/filebeat.yml /usr/share/filebeat/filebeat.yml

COPY debugd/internal/debugd/logcollector/filebeat/inputs.yml /usr/share/filebeat/inputs.d/inputs.yml

ENTRYPOINT ["/usr/share/filebeat/bin/filebeat", "-e", "--path.home", "/usr/share/filebeat", "--path.data", "/usr/share/filebeat/data"]
