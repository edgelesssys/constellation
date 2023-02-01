FROM fedora:37@sha256:99aa8919afd1880064ec915dba44cdc5b52808667717f605750329d55006538a AS release
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
COPY ./cli/internal/libvirt/nvram/constellation_vars.production.fd /usr/share/OVMF/constellation_vars.production.fd

COPY --chmod=755 ./cli/internal/libvirt/start.sh /start.sh

ENTRYPOINT ["/start.sh"]
