FROM fedora:38@sha256:a1347432284b0361f472ebc7c00874cf084e93bc2d9c78fe19f9840da9514789 AS release

RUN dnf install -y https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.6.2-x86_64.rpm

COPY debugd/internal/debugd/logcollector/filebeat/filebeat.yml /usr/share/filebeat/filebeat.yml

COPY debugd/internal/debugd/logcollector/filebeat/inputs.yml /usr/share/filebeat/inputs.d/inputs.yml

ENTRYPOINT ["/usr/share/filebeat/bin/filebeat", "-e", "--path.home", "/usr/share/filebeat", "--path.data", "/usr/share/filebeat/data"]
