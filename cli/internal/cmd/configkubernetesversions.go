/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/cobra"
)

func newConfigKubernetesVersionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubernetes-versions",
		Short: "Print the Kubernetes versions supported by this CLI",
		Long:  "Print the Kubernetes versions supported by this CLI.",
		Args:  cobra.ArbitraryArgs,
		Run:   printSupportedKubernetesVersions,
	}

	return cmd
}

func printSupportedKubernetesVersions(cmd *cobra.Command, _ []string) {
	cmd.Printf("Supported Kubernetes Versions:\n\t%s\n", formatKubernetesVersions())
}

func formatKubernetesVersions() string {
	return strings.Join(versions.SupportedK8sVersions(), "\n\t")
}
