/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clientcodec "k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/yaml"

	"github.com/edgelesssys/constellation/v2/internal/file"
)

// NewInitCmd returns a new cobra.Command for the init command.
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the Constellation cluster",
		Long: "Initialize the Constellation cluster.\n\n" +
			"Start your confidential Kubernetes.",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Define flags for apply backend that are not set by init
			cmd.Flags().Bool("yes", false, "")
			// We always want to skip the infrastructure phase here, to be aligned with the
			// functionality of the old init command.
			cmd.Flags().StringSlice("skip-phases", []string{string(skipInfrastructurePhase)}, "")
			cmd.Flags().Duration("helm-timeout", 10*time.Minute, "")
			return runApply(cmd, args)
		},
		Deprecated: "use 'constellation apply' instead.",
	}
	cmd.Flags().Bool("conformance", false, "enable conformance mode")
	cmd.Flags().Bool("skip-helm-wait", false, "install helm charts without waiting for deployments to be ready")
	cmd.Flags().Bool("merge-kubeconfig", false, "merge Constellation kubeconfig file with default kubeconfig file in $HOME/.kube/config")
	return cmd
}

func writeRow(wr io.Writer, col1 string, col2 string) {
	fmt.Fprint(wr, col1, "\t", col2, "\n")
}

type configMerger interface {
	mergeConfigs(configPath string, fileHandler file.Handler) error
	kubeconfigEnvVar() string
}

type kubeconfigMerger struct {
	log debugLog
}

func (c *kubeconfigMerger) mergeConfigs(configPath string, fileHandler file.Handler) error {
	constellConfig, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("loading admin kubeconfig: %w", err)
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.Precedence = []string{
		clientcmd.RecommendedHomeFile,
		configPath, // our config should overwrite the default config
	}
	c.log.Debug(fmt.Sprintf("Kubeconfig file loading precedence: %v", loadingRules.Precedence))

	// merge the kubeconfigs
	cfg, err := loadingRules.Load()
	if err != nil {
		return fmt.Errorf("loading merged kubeconfig: %w", err)
	}

	// Set the current context to the cluster we just created
	cfg.CurrentContext = constellConfig.CurrentContext
	c.log.Debug(fmt.Sprintf("Set current context to %s", cfg.CurrentContext))

	json, err := runtime.Encode(clientcodec.Codec, cfg)
	if err != nil {
		return fmt.Errorf("encoding merged kubeconfig: %w", err)
	}

	mergedKubeconfig, err := yaml.JSONToYAML(json)
	if err != nil {
		return fmt.Errorf("converting merged kubeconfig to YAML: %w", err)
	}

	if err := fileHandler.Write(clientcmd.RecommendedHomeFile, mergedKubeconfig, file.OptOverwrite); err != nil {
		return fmt.Errorf("writing merged kubeconfig to file: %w", err)
	}
	c.log.Debug(fmt.Sprintf("Merged kubeconfig into default config file: %s", clientcmd.RecommendedHomeFile))
	return nil
}

func (c *kubeconfigMerger) kubeconfigEnvVar() string {
	return os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
}

type grpcDialer interface {
	Dial(ctx context.Context, target string) (*grpc.ClientConn, error)
}
