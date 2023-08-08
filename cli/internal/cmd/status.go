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
	"github.com/edgelesssys/constellation/v2/cli/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
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

	kubeClient := kubectl.New()

	flags, err := parseStatusFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	fileHandler := file.NewHandler(afero.NewOsFs())
	kubeConfig, err := fileHandler.Read(constants.AdminConfFilename)
	if err != nil {
		return fmt.Errorf("reading admin.conf: %w", err)
	}

	// need kubectl client to fetch nodes.
	if err := kubeClient.Initialize(kubeConfig); err != nil {
		return err
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	if err != nil {
		return fmt.Errorf("creating k8s client config from kubeconfig: %w", err)
	}
	// need unstructed client to fetch NodeVersion CRD.
	unstructuredClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("setting up custom resource client: %w", err)
	}

	// need helm client to fetch service versions.
	// The client used here, doesn't need to know the current workspace.
	// It may be refactored in the future for easier usage.
	helmClient, err := helm.NewUpgradeClient(kubectl.New(), constants.UpgradeDir, constants.AdminConfFilename, constants.HelmNamespace, log)
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

	stableClient, err := kubernetes.NewStableClient(constants.AdminConfFilename)
	if err != nil {
		return fmt.Errorf("setting up stable client: %w", err)
	}
	output, err := status(cmd.Context(), kubeClient, stableClient, helmVersionGetter, kubernetes.NewNodeVersionClient(unstructuredClient), variant)
	if err != nil {
		return fmt.Errorf("getting status: %w", err)
	}

	cmd.Print(output)
	return nil
}

// status queries the cluster for the relevant status information and returns the output string.
func status(
	ctx context.Context, kubeClient kubeClient, cmClient configMapClient, getHelmVersions func() (fmt.Stringer, error),
	dynamicInterface kubernetes.DynamicInterface, attestVariant variant.Variant,
) (string, error) {
	nodeVersion, err := kubernetes.GetConstellationVersion(ctx, dynamicInterface)
	if err != nil {
		return "", fmt.Errorf("getting constellation version: %w", err)
	}
	if len(nodeVersion.Status.Conditions) != 1 {
		return "", fmt.Errorf("expected exactly one condition, got %d", len(nodeVersion.Status.Conditions))
	}

	attestationConfig, err := getAttestationConfig(ctx, cmClient, attestVariant)
	if err != nil {
		return "", fmt.Errorf("getting attestation config: %w", err)
	}
	prettyYAML, err := yaml.Marshal(attestationConfig)
	if err != nil {
		return "", fmt.Errorf("marshalling attestation config: %w", err)
	}

	targetVersions, err := kubernetes.NewTargetVersions(nodeVersion)
	if err != nil {
		return "", fmt.Errorf("getting configured versions: %w", err)
	}

	serviceVersions, err := getHelmVersions()
	if err != nil {
		return "", fmt.Errorf("getting service versions: %w", err)
	}

	status, err := kubernetes.ClusterStatus(ctx, kubeClient)
	if err != nil {
		return "", fmt.Errorf("getting cluster status: %w", err)
	}

	return statusOutput(targetVersions, serviceVersions, status, nodeVersion, string(prettyYAML)), nil
}

func getAttestationConfig(ctx context.Context, cmClient configMapClient, attestVariant variant.Variant) (config.AttestationCfg, error) {
	joinConfig, err := cmClient.GetCurrentConfigMap(ctx, constants.JoinConfigMap)
	if err != nil {
		return nil, fmt.Errorf("getting current config map: %w", err)
	}
	rawAttestationConfig, ok := joinConfig.Data[constants.AttestationConfigFilename]
	if !ok {
		return nil, fmt.Errorf("attestationConfig not found in %s", constants.JoinConfigMap)
	}
	attestationConfig, err := config.UnmarshalAttestationConfig([]byte(rawAttestationConfig), attestVariant)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling attestation config: %w", err)
	}
	return attestationConfig, nil
}

// statusOutput creates the status cmd output string by formatting the received information.
func statusOutput(
	targetVersions kubernetes.TargetVersions, serviceVersions fmt.Stringer,
	status map[string]kubernetes.NodeStatus, nodeVersion v1alpha1.NodeVersion, rawAttestationConfig string,
) string {
	builder := strings.Builder{}

	builder.WriteString(targetVersionsString(targetVersions))
	builder.WriteString(serviceVersions.String())
	builder.WriteString(fmt.Sprintf("Cluster status: %s\n", nodeVersion.Status.Conditions[0].Message))
	builder.WriteString(nodeStatusString(status, targetVersions))
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
func nodeStatusString(status map[string]kubernetes.NodeStatus, targetVersions kubernetes.TargetVersions) string {
	var upToDateImages int
	var upToDateK8s int
	for _, node := range status {
		if node.KubeletVersion() == targetVersions.Kubernetes() {
			upToDateK8s++
		}
		if node.ImageVersion() == targetVersions.ImagePath() {
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
func targetVersionsString(target kubernetes.TargetVersions) string {
	builder := strings.Builder{}
	builder.WriteString("Target versions:\n")
	builder.WriteString(fmt.Sprintf("\tImage: %s\n", target.Image()))
	builder.WriteString(fmt.Sprintf("\tKubernetes: %s\n", target.Kubernetes()))

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

type kubeClient interface {
	GetNodes(ctx context.Context) ([]corev1.Node, error)
}

type configMapClient interface {
	GetCurrentConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error)
}
