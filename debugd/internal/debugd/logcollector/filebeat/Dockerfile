FROM fedora:37@sha256:ab2fd8de428f5dbdb147d43135902a776a6c64d2243fd7ac99629df3b47c9256 AS release

RUN dnf install -y https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.6.2-x86_64.rpm

COPY debugd/internal/debugd/logcollector/filebeat/filebeat.yml /usr/share/filebeat/filebeat.yml

COPY debugd/internal/debugd/logcollector/filebeat/inputs.yml /usr/share/filebeat/inputs.d/inputs.yml

ENTRYPOINT ["/usr/share/filebeat/bin/filebeat", "-e", "--path.home", "/usr/share/filebeat", "--path.data", "/usr/share/filebeat/data"]
