/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/rogpeppe/go-internal/diff"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func newUpgradeApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply an upgrade to a Constellation cluster",
		Long:  "Apply an upgrade to a Constellation cluster by applying the chosen configuration.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Define flags for apply backend that are not set by upgrade-apply
			cmd.Flags().Bool("merge-kubeconfig", false, "")
			return runApply(cmd, args)
		},
		Deprecated: "use 'constellation apply' instead.",
	}

	cmd.Flags().BoolP("yes", "y", false, "run upgrades without further confirmation\n"+
		"WARNING: might delete your resources in case you are using cert-manager in your cluster. Please read the docs.\n"+
		"WARNING: might unintentionally overwrite measurements in the running cluster.")
	cmd.Flags().Duration("helm-timeout", 10*time.Minute, "change helm upgrade timeout\n"+
		"Might be useful for slow connections or big clusters.")
	cmd.Flags().Bool("conformance", false, "enable conformance mode")
	cmd.Flags().Bool("skip-helm-wait", false, "install helm charts without waiting for deployments to be ready")
	cmd.Flags().StringSlice("skip-phases", nil, "comma-separated list of upgrade phases to skip\n"+
		"one or multiple of { infrastructure | helm | image | k8s }")
	must(cmd.Flags().MarkHidden("helm-timeout"))

	return cmd
}

func diffAttestationCfg(currentAttestationCfg config.AttestationCfg, newAttestationCfg config.AttestationCfg) (string, error) {
	// cannot compare structs directly with go-cmp because of unexported fields in the attestation config
	currentYml, err := yaml.Marshal(currentAttestationCfg)
	if err != nil {
		return "", fmt.Errorf("marshalling remote attestation config: %w", err)
	}
	newYml, err := yaml.Marshal(newAttestationCfg)
	if err != nil {
		return "", fmt.Errorf("marshalling local attestation config: %w", err)
	}
	diff := string(diff.Diff("current", currentYml, "new", newYml))
	return diff, nil
}

type kubernetesUpgrader interface {
	UpgradeNodeVersion(ctx context.Context, conf *config.Config, force, skipImage, skipK8s bool) error
	ExtendClusterConfigCertSANs(ctx context.Context, alternativeNames []string) error
	GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, error)
	ApplyJoinConfig(ctx context.Context, newAttestConfig config.AttestationCfg, measurementSalt []byte) error
	BackupCRs(ctx context.Context, crds []apiextensionsv1.CustomResourceDefinition, upgradeDir string) error
	BackupCRDs(ctx context.Context, upgradeDir string) ([]apiextensionsv1.CustomResourceDefinition, error)
}
