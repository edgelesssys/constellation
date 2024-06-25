#!/usr/bin/env bash

set -euo pipefail

set -x

function cleanup {
  kill -SIGTERM "$(cat "${PWD}"/qemu-dnsmasq-br0.pid)" || true
  rm "${PWD}"/qemu-dnsmasq-br0.pid || true

  kill -SIGTERM "$(cat "${PWD}"/swtpm0.pid)" || true
  kill -SIGTERM "$(cat "${PWD}"/swtpm1.pid)" || true

  ip l delete br0 || true
  ip l delete tap0 || true
  ip l delete tap1 || true

  rm -r "${PWD}"/tpm0 || true
  rm -r "${PWD}"/tpm1 || true

  rm OVMF_VARS_0.fd || true
  rm OVMF_VARS_1.fd || true

  rm dnsmasq.leases || true
  rm dnsmasq.log || true

  rm constellation-mastersecret.json || true
  rm constellation-admin.conf || true
  rm constellation-cluster.log || true
  rm constellation-debug.log || true
  rm constellation-state.yaml || true
  rm -r constellation-upgrade || true

  docker stop metadata-server || true
}

trap cleanup EXIT

get_mac() {
  printf '52:54:%02X:%02X:%02X:%02X' $((RANDOM % 256)) $((RANDOM % 256)) $((RANDOM % 256)) $((RANDOM % 256))
}

mac_0=$(get_mac)
mac_1=$(get_mac)

# Regarding network setup see: https://bbs.archlinux.org/viewtopic.php?id=207907

dd if=/dev/zero of=disk0.img iflag=fullblock bs=1M count=10000 && sync
dd if=/dev/zero of=disk1.img iflag=fullblock bs=1M count=10000 && sync

DEFAULT_INTERFACE=$(ip r show default | cut -d' ' -f5)

ip link add name br0 type bridge || true
ip addr add 10.42.0.1/16 dev br0 || true
ip link set br0 up

dnsmasq \
  --pid-file="${PWD}"/qemu-dnsmasq-br0.pid \
  --interface=br0 \
  --bind-interfaces \
  --log-facility="${PWD}"/dnsmasq.log \
  --dhcp-range=10.42.0.2,10.42.255.254 \
  --dhcp-leasefile="${PWD}"/dnsmasq.leases \
  --dhcp-host="${mac_0}",10.42.1.1,control-plane0 \
  --dhcp-host="${mac_1}",10.42.2.1,worker0

password=$(tr -dc 'A-Za-z0-9!?%=' < /dev/urandom | head -c 32) || true
password_hex=$(echo -n "${password}" | xxd -p -u -c 256)
echo "${password_hex}"

# htpasswd from apache2-utils
password_bcrypt=$(htpasswd -bnBC 10 "" "${password}" | tr -d ':\n')

docker run \
  -dit \
  --rm \
  --name metadata-server \
  --net=host \
  --mount type=bind,source="$(pwd)"/dnsmasq.leases,target=/dnsmasq.leases \
  ghcr.io/edgelesssys/constellation/qemu-metadata-api:v2.17.0-pre.0.20240603111213-d7ce6af383f2 \
  --dnsmasq-leases /dnsmasq.leases --initsecrethash "${password_bcrypt}"

cat > ./constellation-state.yaml <<- EOM
version: v1 # Schema version of this state file.
# State of the cluster's cloud resources. These values are retrieved during
infrastructure:
    uid: qemu # Unique identifier the cluster's cloud resources are tagged with.
    clusterEndpoint: 10.42.1.1 # Endpoint the cluster can be reached at. This is the endpoint that is being used by the CLI.
    inClusterEndpoint: 10.42.1.1 # The Cluster uses to reach itself. This might differ from the ClusterEndpoint in case e.g.,
    initSecret: "${password_hex}" # Secret used to authenticate the bootstrapping node.
    # List of Subject Alternative Names (SANs) to add to the Kubernetes API server certificate.
    apiServerCertSANs:
        - 10.42.1.1
    name: mini-qemu # Name used in the cluster's named resources.
    ipCidrNode: 10.42.0.0/16 # CIDR range of the cluster's nodes.
# DO NOT EDIT. State of the Constellation Kubernetes cluster.
clusterValues:
    clusterID: "" # Unique identifier of the cluster.
    ownerID: "" # Unique identifier of the owner of the cluster.
    measurementSalt: "" # Salt used to generate the ClusterID on the bootstrapping node.
EOM

sysctl net.ipv4.ip_forward=1
sysctl net.ipv6.conf.default.forwarding=1
sysctl net.ipv6.conf.all.forwarding=1

iptables -t nat -C POSTROUTING -o "${DEFAULT_INTERFACE}" -j MASQUERADE || iptables -t nat -I POSTROUTING -o "${DEFAULT_INTERFACE}" -j MASQUERADE
iptables -C FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT || iptables -I FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
iptables -P FORWARD ACCEPT

ip tuntap add dev tap0 mode tap user "${USER}" || true
ip link set tap0 up promisc on
ip link set tap0 master br0

iptables -C FORWARD -i tap0 -o "${DEFAULT_INTERFACE}" -j ACCEPT || iptables -I FORWARD -i tap0 -o "${DEFAULT_INTERFACE}" -j ACCEPT

ip tuntap add dev tap1 mode tap user "${USER}" || true
ip link set tap1 up promisc on
ip link set tap1 master br0

iptables -C FORWARD -i tap1 -o "${DEFAULT_INTERFACE}" -j ACCEPT || iptables -I FORWARD -i tap1 -o "${DEFAULT_INTERFACE}" -j ACCEPT

#
# ovmf
#

cp AMDSEV/usr/local/share/qemu/OVMF_VARS.fd OVMF_VARS_0.fd
cp AMDSEV/usr/local/share/qemu/OVMF_VARS.fd OVMF_VARS_1.fd

#
# swtpm
#

mkdir "${PWD}"/tpm0 || true
swtpm_setup --tpm2 --tpmstate "${PWD}/tpm0" --create-ek-cert --create-platform-cert --allow-signing --overwrite --pcr-banks - --logfile "${PWD}/tpm0/setup.log"
swtpm socket --tpm2 --tpmstate dir="${PWD}/tpm0",mode=0600 --ctrl type=unixio,path="${PWD}/tpm0/swtpm-sock" --log file="${PWD}/tpm0/tpm.log",level=20,truncate --pid file="${PWD}/swtpm0.pid" &

mkdir "${PWD}"/tpm1 || true
swtpm_setup --tpm2 --tpmstate "${PWD}/tpm1" --create-ek-cert --create-platform-cert --allow-signing --overwrite --pcr-banks - --logfile "${PWD}/tpm1/setup.log"
swtpm socket --tpm2 --tpmstate dir="${PWD}/tpm1",mode=0600 --ctrl type=unixio,path="${PWD}/tpm1/swtpm-sock" --log file="${PWD}/tpm1/tpm.log",level=20,truncate --pid file="${PWD}/swtpm1.pid" &

tmux new-session -d -s const-sess

tmux split-window
tmux split-window

launch_cmd_base_sev="AMDSEV/usr/local/bin/qemu-system-x86_64 \
	-enable-kvm \
	-cpu EPYC-v4 \
	-machine q35,smm=off \
	-smp 4,maxcpus=255 \
	-m 2048M,slots=5,maxmem=$((2048 + 8192))M \
	-no-reboot \
	-bios AMDSEV/usr/local/share/qemu/OVMF_CODE.fd \
	-drive file=./image.raw,if=none,id=disk1,format=raw,readonly=on \
	-device virtio-blk-pci,drive=disk1,id=virtio-disk1,disable-legacy=on,iommu_platform=true,bootindex=1 \
	-machine memory-encryption=sev0,vmport=off \
	-object memory-backend-memfd,id=ram1,size=2048M,share=true,prealloc=false \
	-machine memory-backend=ram1 \
	-object sev-snp-guest,id=sev0,cbitpos=51,reduced-phys-bits=1 \
	-nographic  \
	-device virtio-blk-pci,drive=disk2,id=virtio-disk2 \
	-tpmdev emulator,id=tpm0,chardev=chrtpm \
	-device tpm-crb,tpmdev=tpm0"

# shellcheck disable=2034
launch_cmd_base_no_sev="AMDSEV/usr/local/bin/qemu-system-x86_64 \
	-enable-kvm \
	-cpu EPYC-v4 \
	-machine q35 \
	-smp 1,maxcpus=255 \
	-m 2048M,slots=5,maxmem=10240M \
	-no-reboot \
	-drive if=pflash,format=raw,unit=0,file=${PWD}/OVMF_CODE.fd,readonly=true \
	-drive file=./image.raw,if=none,id=disk1,format=raw,readonly=on \
	-device virtio-blk-pci,drive=disk1,id=virtio-disk1,disable-legacy=on,iommu_platform=true,bootindex=1 \
	-nographic  \
	-device virtio-blk-pci,drive=disk2,id=virtio-disk2 \
	-tpmdev emulator,id=tpm0,chardev=chrtpm \
	-device tpm-crb,tpmdev=tpm0"

launch_cmd_0="${launch_cmd_base_sev} \
	-drive if=pflash,format=raw,unit=0,file=${PWD}/OVMF_VARS_0.fd \
	-device virtio-net,netdev=network0,mac=${mac_0} \
	-netdev tap,id=network0,ifname=tap0,script=no,downscript=no \
	-drive file=./disk0.img,id=disk2,if=none,format=raw \
	-chardev socket,id=chrtpm,path=${PWD}/tpm0/swtpm-sock"
launch_cmd_1="${launch_cmd_base_sev} \
	-drive if=pflash,format=raw,unit=0,file=${PWD}/OVMF_VARS_1.fd \
	-device virtio-net,netdev=network0,mac=${mac_1} \
	-netdev tap,id=network0,ifname=tap1,script=no,downscript=no \
	-drive file=./disk1.img,id=disk2,if=none,format=raw \
	-chardev socket,id=chrtpm,path=${PWD}/tpm1/swtpm-sock"

init_cmd="./constellation apply --skip-phases infrastructure"

tmux send -t const-sess:0.0 "${launch_cmd_0}" ENTER
sleep 3
tmux send -t const-sess:0.1 "${launch_cmd_1}" ENTER
tmux send -t const-sess:0.2 "${init_cmd}" ENTER

tmux a -t const-sess
