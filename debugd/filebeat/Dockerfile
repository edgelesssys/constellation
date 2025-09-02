FROM fedora:42@sha256:3da64cb89971a1cdbc6046e307eeebcb54f7281c0a606ee48d9995473f6b88d5 AS release

RUN dnf install -y https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.6.2-x86_64.rpm

RUN dnf install -y systemd-libs

COPY debugd/filebeat/templates/ /usr/share/constellogs/templates/

ENTRYPOINT ["/usr/share/filebeat/bin/filebeat", "-e", "--path.home", "/usr/share/filebeat", "--path.data", "/usr/share/filebeat/data"]
