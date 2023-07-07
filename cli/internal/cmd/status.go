/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubernetes"
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
	helmClient, err := helm.NewClient(kubectl.New(), constants.AdminConfFilename, constants.HelmNamespace, log)
	if err != nil {
		return fmt.Errorf("setting up helm client: %w", err)
	}

	stableClient := kubernetes.NewStableClient(kubeClient)
	output, err := status(cmd.Context(), kubeClient, stableClient, helmClient, kubernetes.NewNodeVersionClient(unstructuredClient))
	if err != nil {
		return fmt.Errorf("getting status: %w", err)
	}

	cmd.Print(output)
	return nil
}

// status queries the cluster for the relevant status information and returns the output string.
func status(ctx context.Context, kubeClient kubeClient, stableClient kubernetes.StableInterface, helmClient helmClient, dynamicInterface kubernetes.DynamicInterface) (string, error) {
	nodeVersion, err := kubernetes.GetConstellationVersion(ctx, dynamicInterface)
	if err != nil {
		return "", fmt.Errorf("getting constellation version: %w", err)
	}
	if len(nodeVersion.Status.Conditions) != 1 {
		return "", fmt.Errorf("expected exactly one condition, got %d", len(nodeVersion.Status.Conditions))
	}

	// attestation version
	joinConfig, err := stableClient.GetCurrentConfigMap(ctx, constants.JoinConfigMap)
	rawAttestationConfig, ok := joinConfig.Data[constants.AttestationConfigFilename]
	if !ok {
		return "", fmt.Errorf("attestationConfig not found in %s", constants.JoinConfigMap)
	}
	prettyYAML, err := convertJSONToYAML([]byte(rawAttestationConfig), err)
	if err != nil {
		return "", fmt.Errorf("converting attestation config to yaml: %w", err)
	}
	targetVersions, err := kubernetes.NewTargetVersions(nodeVersion)
	if err != nil {
		return "", fmt.Errorf("getting configured versions: %w", err)
	}

	serviceVersions, err := helmClient.Versions()
	if err != nil {
		return "", fmt.Errorf("getting service versions: %w", err)
	}

	status, err := kubernetes.ClusterStatus(ctx, kubeClient)
	if err != nil {
		return "", fmt.Errorf("getting cluster status: %w", err)
	}

	return statusOutput(targetVersions, serviceVersions, status, nodeVersion, string(prettyYAML)), nil
}

func convertJSONToYAML(rawJSON []byte, err error) ([]byte, error) {
	var jsonMap map[string]interface{}
	err = json.Unmarshal(rawJSON, &jsonMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling raw json: %w", err)
	}

	yamlFormatted, err := yaml.Marshal(jsonMap)
	if err != nil {
		return nil, fmt.Errorf("marshalling yaml: %w", err)
	}
	return yamlFormatted, nil
}

// statusOutput creates the status cmd output string by formatting the received information.
func statusOutput(targetVersions kubernetes.TargetVersions, serviceVersions helm.ServiceVersions, status map[string]kubernetes.NodeStatus, nodeVersion v1alpha1.NodeVersion, rawAttestationConfig string) string {
	builder := strings.Builder{}

	builder.WriteString(targetVersionsString(targetVersions))
	builder.WriteString(serviceVersionsString(serviceVersions))
	builder.WriteString(fmt.Sprintf("Cluster status: %s\n", nodeVersion.Status.Conditions[0].Message))
	builder.WriteString(nodeStatusString(status, targetVersions))
	builder.WriteString(fmt.Sprintf("Attestation config:\n%s\n", indentEntireStringWithTab(rawAttestationConfig)))

	return builder.String()
}

func indentEntireStringWithTab(input string) string {
	lines := strings.Split(input, "\n")
	for i, line := range lines {
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

// serviceVersionsString creates the service versions part of the output string.
func serviceVersionsString(versions helm.ServiceVersions) string {
	builder := strings.Builder{}
	builder.WriteString("Installed service versions:\n")
	builder.WriteString(fmt.Sprintf("\tCilium: %s\n", versions.Cilium()))
	builder.WriteString(fmt.Sprintf("\tcert-manager: %s\n", versions.CertManager()))
	builder.WriteString(fmt.Sprintf("\tconstellation-operators: %s\n", versions.ConstellationOperators()))
	builder.WriteString(fmt.Sprintf("\tconstellation-services: %s\n", versions.ConstellationServices()))
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

type kubeClient interface {
	GetNodes(ctx context.Context) ([]corev1.Node, error)
}

type configMapClient interface {
	GetCurrentConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error)
}
type helmClient interface {
	Versions() (helm.ServiceVersions, error)
}
