FROM fedora@sha256:486fd5578f93fbc57a519e34ad4b7cac927c3f8a95409baedf0c19e9f287c207 AS deploy
RUN dnf -y update && \
    dnf -y install dnf-plugins-core \
        libvirt-daemon-config-network \
        libvirt-daemon-kvm \
        qemu-kvm \
        swtpm \
        swtpm-tools \
        libvirt-client && \
    dnf clean all

# Prevent cgroup issues on Fedora and configure libvirt
RUN echo "cgroup_controllers = []" >> /etc/libvirt/qemu.conf && \
    echo "listen_tls = 0" >> /etc/libvirt/libvirtd.conf && \
    echo "listen_tcp = 1" >> /etc/libvirt/libvirtd.conf && \
    echo "tcp_port = \"16599\"" >> /etc/libvirt/libvirtd.conf && \
    echo "listen_addr = \"localhost\"" >> /etc/libvirt/libvirtd.conf && \
    echo "auth_tcp = \"none\"" >> /etc/libvirt/libvirtd.conf

# Copy nvram templates
COPY ./cli/internal/libvirt/nvram/constellation_vars.testing.fd /usr/share/OVMF/constellation_vars.testing.fd
# TODO: Uncomment this line when we have a production template
# COPY ./cli/internal/libvirt/nvram/constellation_vars.production.fd /usr/share/OVMF/constellation_vars.production.fd

COPY --chmod=755 ./cli/internal/libvirt/start.sh /start.sh

ENTRYPOINT ["/start.sh"]
