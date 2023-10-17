FROM fedora:38@sha256:8285246bd5fad4e76e17a71c88dee34c49e2f227dab4ce7df704b592f8e72d41 AS release

RUN dnf install -y https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.6.2-x86_64.rpm

COPY debugd/filebeat/templates/ /usr/share/constellogs/templates/

ENTRYPOINT ["/usr/share/filebeat/bin/filebeat", "-e", "--path.home", "/usr/share/filebeat", "--path.data", "/usr/share/filebeat/data"]
