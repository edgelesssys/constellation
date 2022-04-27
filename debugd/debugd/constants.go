package debugd

import "time"

const (
	DebugdMetadataFlag              = "constellation-debugd"
	DebugdPort                      = "4000"
	GRPCTimeout                     = 5 * time.Minute
	SSHCheckInterval                = 30 * time.Second
	DiscoverDebugdInterval          = 30 * time.Second
	CoordinatorDownloadRetryBackoff = 1 * time.Minute
	CoordinatorDeployFilename       = "/opt/coordinator"
	Chunksize                       = 1024
	CoordinatorSystemdUnitName      = "coordinator.service"
	CoordinatorSystemdUnitContents  = `[Unit]
Description=Constellation Coordinator
Wants=network-online.target
After=network-online.target
[Service]
Type=simple
EnvironmentFile=/etc/constellation.env
ExecStartPre=-setenforce Permissive
ExecStartPre=/usr/bin/mkdir -p /opt/cni/bin/
# merge all CNI binaries in writable folder until containerd can use multiple CNI bins: https://github.com/containerd/containerd/issues/6600
ExecStartPre=/bin/sh -c "/usr/bin/cp /usr/libexec/cni/* /opt/cni/bin/"
ExecStart=/opt/coordinator
[Install]
WantedBy=multi-user.target
`
)
