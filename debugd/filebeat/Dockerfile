FROM fedora:42@sha256:6c53109bbf345c2612ada5bd06e8e352b45d85c0a2d770167b61c8d1d047e124 AS release

RUN dnf install -y https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.6.2-x86_64.rpm

RUN dnf install -y systemd-libs

COPY debugd/filebeat/templates/ /usr/share/constellogs/templates/

ENTRYPOINT ["/usr/share/filebeat/bin/filebeat", "-e", "--path.home", "/usr/share/filebeat", "--path.data", "/usr/share/filebeat/data"]
