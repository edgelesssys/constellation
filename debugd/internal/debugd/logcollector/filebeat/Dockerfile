FROM fedora:37@sha256:67870e45f3548cecd23234cff374c1003ccb2141447b50acd0acd8db1afcbd24 AS release

RUN dnf install -y https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.6.2-x86_64.rpm

COPY debugd/internal/debugd/logcollector/filebeat/filebeat.yml /usr/share/filebeat/filebeat.yml

COPY debugd/internal/debugd/logcollector/filebeat/inputs.yml /usr/share/filebeat/inputs.d/inputs.yml

ENTRYPOINT ["/usr/share/filebeat/bin/filebeat", "-e", "--path.home", "/usr/share/filebeat", "--path.data", "/usr/share/filebeat/data"]
