/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"runtime/debug"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/spf13/cobra"
)

// NewVersionCmd returns a new cobra.Command for the verify command.
func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display version of this CLI",
		Long:  "Display version of this CLI.",
		Args:  cobra.NoArgs,
		Run:   runVersion,
	}
	return cmd
}

func runVersion(cmd *cobra.Command, args []string) {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		cmd.PrintErrf("Unable to retrieve build info. Is buildvcs enabled?")
		return
	}

	commit, state, date, goVersion, compiler, platform := parseBuildInfo(buildInfo)

	cmd.Printf("Version:\t%s (%s)\n", constants.VersionInfo(), constants.VersionBuild)
	cmd.Printf("GitCommit:\t%s\n", commit)
	cmd.Printf("GitTreeState:\t%s\n", state)
	cmd.Printf("BuildDate:\t%s\n", date)
	cmd.Printf("GoVersion:\t%s\n", goVersion)
	cmd.Printf("Compiler:\t%s\n", compiler)
	cmd.Printf("Platform:\t%s\n", platform)
}

func parseBuildInfo(info *debug.BuildInfo) (commit, state, date, goVersion, compiler, platform string) {
	var arch, os string
	for idx := range info.Settings {
		key := info.Settings[idx].Key
		value := info.Settings[idx].Value

		switch key {
		case "-compiler":
			compiler = value
		case "GOARCH":
			arch = value
		case "GOOS":
			os = value
		case "vcs.time":
			date = value
		case "vcs.modified":
			if value == "true" {
				state = "dirty"
			} else {
				state = "clean"
			}
		case "vcs.revision":
			commit = value
		}
	}

	platform = os + "/" + arch
	goVersion = info.GoVersion
	return commit, state, date, goVersion, compiler, platform
}
