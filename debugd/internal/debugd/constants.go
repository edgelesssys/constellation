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
	DiscoverDebugdInterval          = 30 * time.Second
	DownloadRetryBackoff            = 1 * time.Minute
	BootstrapperDeployFilename      = "/run/state/bin/bootstrapper"
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
	GettyAutologinOverrideFilename     = "/run/systemd/system/serial-getty@ttyS0.service.d/autologin.conf"
	GettyAutologinOverrideUnitContents = `[Service]
ExecStart=
ExecStart=-/sbin/agetty -o '-p -f -- \\u' --autologin root --keep-baud 115200,57600,38400,9600 - $TERM`
)
