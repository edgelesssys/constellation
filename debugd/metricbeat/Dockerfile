FROM fedora:38@sha256:3f01c8f79691df76331cb4bb0944794a60850475e859c15e49513fcbe0a3d88a AS release

RUN dnf install -y https://artifacts.elastic.co/downloads/beats/metricbeat/metricbeat-8.9.2-x86_64.rpm

COPY debugd/metricbeat/templates/ /usr/share/constellogs/templates/

# -e enables logging to stderr
# -E http.enabled=true enables http endpoint as seen in https://github.com/elastic/helm-charts/blob/main/metricbeat/templates/daemonset.yaml
# --path.home and --path.data set the paths to the metricbeat binary and data directory
ENTRYPOINT [ "/usr/share/metricbeat/bin/metricbeat", \
    "-e", \
    "-E", "http.enabled=true", \
    "--path.home", "/usr/share/metricbeat", \
    "--path.data", "/usr/share/metricbeat/data" \
]
