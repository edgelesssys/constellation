/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubecmd"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewStatusCmd returns a new cobra.Command for the statuus command.
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of a Constellation cluster",
		Long: "Show the status of a constellation cluster.\n\n" +
			"Shows microservice, image, and Kubernetes versions installed in the cluster. Also shows status of current version upgrades.",
		Args: cobra.NoArgs,
		RunE: runStatus,
	}
	return cmd
}

// runStatus runs the terminate command.
func runStatus(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()

	flags, err := parseStatusFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	fileHandler := file.NewHandler(afero.NewOsFs())

	helmClient, err := helm.NewReleaseVersionClient(constants.AdminConfFilename, log)
	if err != nil {
		return fmt.Errorf("setting up helm client: %w", err)
	}
	helmVersionGetter := func() (fmt.Stringer, error) {
		return helmClient.Versions()
	}

	fetcher := attestationconfigapi.NewFetcher()
	conf, err := config.New(fileHandler, constants.ConfigFilename, fetcher, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	variant := conf.GetAttestationConfig().GetVariant()

	kubeClient, err := kubecmd.New(cmd.OutOrStdout(), constants.AdminConfFilename, fileHandler, log)
	if err != nil {
		return fmt.Errorf("setting up kubernetes client: %w", err)
	}

	output, err := status(cmd.Context(), helmVersionGetter, kubeClient, variant)
	if err != nil {
		return fmt.Errorf("getting status: %w", err)
	}

	cmd.Print(output)
	return nil
}

// status queries the cluster for the relevant status information and returns the output string.
func status(ctx context.Context, getHelmVersions func() (fmt.Stringer, error), kubeClient kubeCmd, attestVariant variant.Variant,
) (string, error) {
	nodeVersion, err := kubeClient.GetConstellationVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("getting constellation version: %w", err)
	}

	attestationConfig, err := kubeClient.GetClusterAttestationConfig(ctx, attestVariant)
	if err != nil {
		return "", fmt.Errorf("getting attestation config: %w", err)
	}
	prettyYAML, err := yaml.Marshal(attestationConfig)
	if err != nil {
		return "", fmt.Errorf("marshalling attestation config: %w", err)
	}

	serviceVersions, err := getHelmVersions()
	if err != nil {
		return "", fmt.Errorf("getting service versions: %w", err)
	}

	status, err := kubeClient.ClusterStatus(ctx)
	if err != nil {
		return "", fmt.Errorf("getting cluster status: %w", err)
	}

	return statusOutput(nodeVersion, serviceVersions, status, string(prettyYAML)), nil
}

// statusOutput creates the status cmd output string by formatting the received information.
func statusOutput(
	nodeVersion kubecmd.NodeVersion, serviceVersions fmt.Stringer,
	status map[string]kubecmd.NodeStatus, rawAttestationConfig string,
) string {
	builder := strings.Builder{}

	builder.WriteString(targetVersionsString(nodeVersion))
	builder.WriteString(serviceVersions.String())
	builder.WriteString(fmt.Sprintf("Cluster status: %s\n", nodeVersion.ClusterStatus()))
	builder.WriteString(nodeStatusString(status, nodeVersion))
	builder.WriteString(fmt.Sprintf("Attestation config:\n%s", indentEntireStringWithTab(rawAttestationConfig)))
	return builder.String()
}

func indentEntireStringWithTab(input string) string {
	lines := strings.Split(input, "\n")
	for i, line := range lines[:len(lines)-1] {
		lines[i] = "\t" + line
	}
	return strings.Join(lines, "\n")
}

// nodeStatusString creates the node status part of the output string.
func nodeStatusString(status map[string]kubecmd.NodeStatus, targetVersions kubecmd.NodeVersion) string {
	var upToDateImages int
	var upToDateK8s int
	for _, node := range status {
		if node.KubeletVersion() == targetVersions.KubernetesVersion() {
			upToDateK8s++
		}
		if node.ImageVersion() == targetVersions.ImageReference() {
			upToDateImages++
		}
	}

	builder := strings.Builder{}
	if upToDateImages != len(status) || upToDateK8s != len(status) {
		builder.WriteString(fmt.Sprintf("\tImage: %d/%d\n", upToDateImages, len(status)))
		builder.WriteString(fmt.Sprintf("\tKubernetes: %d/%d\n", upToDateK8s, len(status)))
	}

	return builder.String()
}

// targetVersionsString creates the target versions part of the output string.
func targetVersionsString(target kubecmd.NodeVersion) string {
	builder := strings.Builder{}
	builder.WriteString("Target versions:\n")
	builder.WriteString(fmt.Sprintf("\tImage: %s\n", target.ImageVersion()))
	builder.WriteString(fmt.Sprintf("\tKubernetes: %s\n", target.KubernetesVersion()))

	return builder.String()
}

type statusFlags struct {
	workspace string
	force     bool
}

func parseStatusFlags(cmd *cobra.Command) (statusFlags, error) {
	workspace, err := cmd.Flags().GetString("workspace")
	if err != nil {
		return statusFlags{}, fmt.Errorf("getting config flag: %w", err)
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return statusFlags{}, fmt.Errorf("getting config flag: %w", err)
	}
	return statusFlags{
		workspace: workspace,
		force:     force,
	}, nil
}

type kubeCmd interface {
	ClusterStatus(ctx context.Context) (map[string]kubecmd.NodeStatus, error)
	GetConstellationVersion(ctx context.Context) (kubecmd.NodeVersion, error)
	GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, error)
}
