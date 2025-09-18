FROM fedora:44@sha256:61bc3614a3aeae53af1817b0506e57d6b8ec9adf27203f6b1510c753b294b6f1 AS release

RUN dnf install -y https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.6.2-x86_64.rpm

RUN dnf install -y systemd-libs

COPY debugd/filebeat/templates/ /usr/share/constellogs/templates/

ENTRYPOINT ["/usr/share/filebeat/bin/filebeat", "-e", "--path.home", "/usr/share/filebeat", "--path.data", "/usr/share/filebeat/data"]
