FROM fedora:40@sha256:7cdd2b48396929bb8723ea2fa60e03bee39cc22e2a853cbd891587fab4eb1bc9 AS release

RUN dnf install -y https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.6.2-x86_64.rpm

RUN dnf install -y systemd-libs

COPY debugd/filebeat/templates/ /usr/share/constellogs/templates/

ENTRYPOINT ["/usr/share/filebeat/bin/filebeat", "-e", "--path.home", "/usr/share/filebeat", "--path.data", "/usr/share/filebeat/data"]
