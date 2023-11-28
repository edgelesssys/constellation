{ pkgs
, pkgsLinux
, stdenv
}:
let
  passwd = pkgs.writeTextDir "etc/passwd" ''
    root:x:0:0:root:/root:/bin/sh
    bin:x:1:1:bin:/bin:/sbin/nologin
    daemon:x:2:2:daemon:/sbin:/sbin/nologin
    adm:x:3:4:adm:/var/adm:/sbin/nologin
    lp:x:4:7:lp:/var/spool/lpd:/sbin/nologin
    sync:x:5:0:sync:/sbin:/bin/sync
    shutdown:x:6:0:shutdown:/sbin:/sbin/shutdown
    halt:x:7:0:halt:/sbin:/sbin/halt
    nobody:x:65534:65534:Kernel Overflow User:/:/sbin/nologin
    tss:x:59:59:Account used for TPM access:/:/usr/sbin/nologin
    saslauth:x:998:76:Saslauthd user:/run/saslauthd:/sbin/nologin
    polkitd:x:996:996:User for polkitd:/:/sbin/nologin
    dnsmasq:x:994:994:Dnsmasq DHCP and DNS server:/var/lib/dnsmasq:/usr/sbin/nologin
    rpc:x:32:32:Rpcbind Daemon:/var/lib/rpcbind:/sbin/nologin
    rpcuser:x:29:29:RPC Service User:/var/lib/nfs:/sbin/nologin
    qemu:x:107:107:qemu user:/:/sbin/nologin
  '';
  group = pkgs.writeTextDir "etc/group" ''
    root:x:0:
    bin:x:1:
    daemon:x:2:
    sys:x:3:
    adm:x:4:
    tty:x:5:
    disk:x:6:
    lp:x:7:
    mem:x:8:
    kmem:x:9:
    wheel:x:10:
    lock:x:54:
    users:x:100:
    nobody:x:65534:
    tss:x:59:
    utmp:x:22:
    utempter:x:35:
    saslauth:x:76:saslauth
    input:x:104:
    kvm:x:36:qemu
    sgx:x:106:
    polkitd:x:996:
    dnsmasq:x:994:
    rpc:x:32:
    rpcuser:x:29:
    qemu:x:107:
    libvirt:x:990:
  '';
  libvirtdConf = pkgs.writeTextDir "etc/libvirt/libvirtd.conf" ''
    listen_tls = 0
    listen_tcp = 1
    tcp_port = "16599"
    listen_addr = "localhost"
    auth_tcp = "none"
  '';
  qemuConf = pkgs.writeTextDir "var/lib/libvirt/qemu.conf" ''
    cgroup_controllers = []
  '';
  startScript = pkgsLinux.writeShellApplication {
    name = "start.sh";
    runtimeInputs = with pkgsLinux; [
      shadow
      coreutils
      libvirt
      qemu
      swtpm
    ];
    text = ''
      set -euo pipefail
      shopt -s inherit_errexit

      # Assign qemu the GID of the host system's 'kvm' group to avoid permission issues for environments defaulting to 660 for /dev/kvm (e.g. Debian-based distros)
      KVM_HOST_GID="$(stat -c '%g' /dev/kvm)"

      groupadd -o -g "''${KVM_HOST_GID}" host-kvm || true
      usermod -a -G host-kvm qemu || true

      # Start libvirt daemon
      libvirtd -f /etc/libvirt/libvirtd.conf --daemon --listen
      virtlogd --daemon

      sleep infinity
    '';
  };
  ovmf = stdenv.mkDerivation {
    name = "OVMF";
    postInstall = ''
      mkdir -p $out/usr/share/
      ln -s ${pkgsLinux.OVMFFull.fd}/FV  $out/usr/share/OVMF
    '';
    propagatedBuildInputs = with pkgsLinux; [
      OVMF
    ];
    dontUnpack = true;
  };
in
pkgs.dockerTools.buildImage {
  name = "ghcr.io/edgelesssys/constellation/libvirtd-base";
  copyToRoot = with pkgsLinux.dockerTools; [
    passwd
    group
    libvirtdConf
    qemuConf
    ovmf
    startScript
    usrBinEnv
    caCertificates
    pkgsLinux.busybox
  ];
  config = {
    Cmd = [ "/bin/start.sh" ];
  };
  runAsRoot = ''
    #!${pkgs.runtimeShell}
    mkdir -p /tmp
    mkdir -p /run
    mkdir -p /var/lock
    mkdir -p /var/log/libvirt
    mkdir -p /var/lib/swtpm-localca
    mkdir -p /var/lib/libvirt/boot
    mkdir -p /var/lib/libvirt/dnsmasq
    mkdir -p /var/lib/libvirt/filesystems
    mkdir -p /var/lib/libvirt/images
    mkdir -p /var/lib/libvirt/libxl
    mkdir -p /var/lib/libvirt/lxc
    mkdir -p /var/lib/libvirt/network
    mkdir -p /var/lib/libvirt/qemu
    mkdir -p /var/lib/libvirt/swtpm

    chmod 1777 /tmp
    chown -R tss:root /var/lib/swtpm-localca
    chown -R qemu:qemu /var/lib/libvirt/qemu
    chown -R root:libvirt /var/log/libvirt/
  '';
}
