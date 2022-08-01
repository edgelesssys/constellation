package debugd

import "time"

const (
	DebugdMetadataFlag               = "constellation-debugd"
	GRPCTimeout                      = 5 * time.Minute
	SSHCheckInterval                 = 30 * time.Second
	DiscoverDebugdInterval           = 30 * time.Second
	BootstrapperDownloadRetryBackoff = 1 * time.Minute
	BootstrapperDeployFilename       = "/opt/bootstrapper"
	Chunksize                        = 1024
	BootstrapperSystemdUnitName      = "bootstrapper.service"
	BootstrapperSystemdUnitContents  = `[Unit]
Description=Constellation Bootstrapper
Wants=network-online.target
After=network-online.target
[Service]
Type=simple
RemainAfterExit=yes
Restart=on-failure
EnvironmentFile=/etc/constellation.env
ExecStartPre=-setenforce Permissive
ExecStartPre=/usr/bin/mkdir -p /opt/cni/bin/
# merge all CNI binaries in writable folder until containerd can use multiple CNI bins: https://github.com/containerd/containerd/issues/6600
ExecStartPre=/bin/sh -c "/usr/bin/cp /usr/libexec/cni/* /opt/cni/bin/"
ExecStart=/opt/bootstrapper
[Install]
WantedBy=multi-user.target
`
)
