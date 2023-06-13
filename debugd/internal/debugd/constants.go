/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package debugd

import "time"

// Debugd internal constants.
const (
	DebugdMetadataFlag              = "constellation-debugd"
	GRPCTimeout                     = 5 * time.Minute
	DiscoverDebugdInterval          = 10 * time.Second
	DownloadRetryBackoff            = 1 * time.Minute
	BinaryAccessMode                = 0o755 // -rwxr-xr-x
	BootstrapperDeployFilename      = "/run/state/bin/bootstrapper"
	UpgradeAgentDeployFilename      = "/run/state/bin/upgrade-agent"
	Chunksize                       = 1024
	BootstrapperSystemdUnitName     = "bootstrapper.service"
	BootstrapperSystemdUnitContents = `[Unit]
Description=Constellation Bootstrapper
Wants=network-online.target
After=network-online.target
[Service]
Type=simple
RemainAfterExit=yes
Restart=on-failure
EnvironmentFile=/run/constellation.env
Environment=PATH=/run/state/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin
ExecStart=/run/state/bin/bootstrapper
[Install]
WantedBy=multi-user.target
`
)
