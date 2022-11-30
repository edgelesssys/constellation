FROM fedora:37@sha256:f2c083c0b7d2367a375f15e002c2dc7baaca2b3181ace61f9d5113a8fe2f6b44

RUN dnf install -y https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.5.0-x86_64.rpm systemd-devel

COPY filebeat.yml /usr/share/filebeat/filebeat.yml

RUN chmod 600 /usr/share/filebeat/filebeat.yml

ENTRYPOINT ["/usr/share/filebeat/bin/filebeat", "-e", "--path.home", "/usr/share/filebeat", "--path.data", "/usr/share/filebeat/data"]
