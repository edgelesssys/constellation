/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package debugd

import "time"

const (
	DebugdMetadataFlag               = "constellation-debugd"
	GRPCTimeout                      = 5 * time.Minute
	SSHCheckInterval                 = 30 * time.Second
	DiscoverDebugdInterval           = 30 * time.Second
	BootstrapperDownloadRetryBackoff = 1 * time.Minute
	BootstrapperDeployFilename       = "/run/state/bin/bootstrapper"
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
EnvironmentFile=/run/constellation.env
Environment=PATH=/run/state/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin
ExecStartPre=-setenforce Permissive
ExecStart=/run/state/bin/bootstrapper
[Install]
WantedBy=multi-user.target
`
)
