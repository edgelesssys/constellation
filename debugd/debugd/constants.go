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
ExecStart=/opt/coordinator
[Install]
WantedBy=multi-user.target
`
)
