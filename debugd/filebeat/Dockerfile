FROM fedora:38@sha256:3f01c8f79691df76331cb4bb0944794a60850475e859c15e49513fcbe0a3d88a AS release

RUN dnf install -y https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.6.2-x86_64.rpm

COPY debugd/filebeat/templates/ /usr/share/constellogs/templates/

ENTRYPOINT ["/usr/share/filebeat/bin/filebeat", "-e", "--path.home", "/usr/share/filebeat", "--path.data", "/usr/share/filebeat/data"]
