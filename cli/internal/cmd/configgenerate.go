/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/cmd/pathprefix"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

func newConfigGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate {aws|azure|gcp|openstack|qemu|stackit}",
		Short: "Generate a default configuration file",
		Long:  "Generate a default configuration file for your selected cloud provider.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			isCloudProvider(0),
		),
		ValidArgsFunction: generateCompletion,
		RunE:              runConfigGenerate,
	}
	cmd.Flags().StringP("kubernetes", "k", semver.MajorMinor(config.Default().KubernetesVersion), "Kubernetes version to use in format MAJOR.MINOR")
	cmd.Flags().StringP("attestation", "a", "", fmt.Sprintf("attestation variant to use %s. If not specified, the default for the cloud provider is used", printFormattedSlice(variant.GetAvailableAttestationVariants())))

	return cmd
}

type generateFlags struct {
	pf                 pathprefix.PathPrefixer
	k8sVersion         string
	attestationVariant variant.Variant
}

type configGenerateCmd struct {
	log debugLog
}

func runConfigGenerate(cmd *cobra.Command, args []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	fileHandler := file.NewHandler(afero.NewOsFs())
	provider := cloudprovider.FromString(args[0])
	cg := &configGenerateCmd{log: log}
	return cg.configGenerate(cmd, fileHandler, provider, args[0])
}

func (cg *configGenerateCmd) configGenerate(cmd *cobra.Command, fileHandler file.Handler, provider cloudprovider.Provider, rawProvider string) error {
	flags, err := parseGenerateFlags(cmd)
	if err != nil {
		return err
	}

	cg.log.Debugf("Parsed flags as %v", flags)
	cg.log.Debugf("Using cloud provider %s", provider.String())
	conf, err := createConfigWithAttestationVariant(provider, rawProvider, flags.attestationVariant)
	if err != nil {
		return fmt.Errorf("creating config: %w", err)
	}
	conf.KubernetesVersion = flags.k8sVersion
	cg.log.Debugf("Writing YAML data to configuration file")
	if err := fileHandler.WriteYAML(constants.ConfigFilename, conf, file.OptMkdirAll); err != nil {
		return err
	}

	cmd.Println("Config file written to", flags.pf.PrefixPrintablePath(constants.ConfigFilename))
	cmd.Println("Please fill in your CSP-specific configuration before proceeding.")
	cmd.Println("For more information refer to the documentation:")
	cmd.Println("\thttps://docs.edgeless.systems/constellation/getting-started/first-steps")

	return nil
}

// createConfigWithAttestationVariant creates a config file for the given provider.
func createConfigWithAttestationVariant(provider cloudprovider.Provider, rawProvider string, attestationVariant variant.Variant) (*config.Config, error) {
	conf := config.Default().WithOpenStackProviderDefaults(rawProvider)
	conf.RemoveProviderExcept(provider)

	// set a lower default for QEMU's state disk
	if provider == cloudprovider.QEMU {
		for groupName, group := range conf.NodeGroups {
			group.StateDiskSizeGB = 10
			conf.NodeGroups[groupName] = group
		}
	}

	if provider == cloudprovider.Unknown {
		return conf, nil
	}
	if attestationVariant.Equal(variant.Dummy{}) {
		attestationVariant = variant.GetDefaultAttestation(provider)
		if attestationVariant.Equal(variant.Dummy{}) {
			return nil, fmt.Errorf("provider %s does not have a default attestation variant", provider)
		}
	} else if !variant.ValidProvider(provider, attestationVariant) {
		return nil, fmt.Errorf("provider %s does not support attestation variant %s", provider, attestationVariant)
	}
	conf.SetAttestation(attestationVariant)
	return conf, nil
}

// createConfig creates a config file for the given provider.
func createConfig(provider cloudprovider.Provider) *config.Config {
	// rawProvider can be hardcoded as it only matters for OpenStack
	res, _ := createConfigWithAttestationVariant(provider, "", variant.Dummy{})
	return res
}

// supportedVersions prints the supported version without v prefix and without patch version.
// Should only be used when accepting Kubernetes versions from --kubernetes.
func supportedVersions() string {
	builder := strings.Builder{}
	for i, version := range versions.SupportedK8sVersions() {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(strings.TrimPrefix(semver.MajorMinor(version), "v"))
	}
	return builder.String()
}

func parseGenerateFlags(cmd *cobra.Command) (generateFlags, error) {
	workDir, err := cmd.Flags().GetString("workspace")
	if err != nil {
		return generateFlags{}, fmt.Errorf("parsing workspace flag: %w", err)
	}
	k8sVersion, err := cmd.Flags().GetString("kubernetes")
	if err != nil {
		return generateFlags{}, fmt.Errorf("parsing kuberentes flag: %w", err)
	}
	resolvedVersion, err := resolveK8sVersion(k8sVersion)
	if err != nil {
		return generateFlags{}, fmt.Errorf("resolving kuberentes version from flag: %w", err)
	}

	attestationString, err := cmd.Flags().GetString("attestation")
	if err != nil {
		return generateFlags{}, fmt.Errorf("parsing attestation flag: %w", err)
	}

	var attestationVariant variant.Variant
	// if no attestation variant is specified, use the default for the cloud provider
	if attestationString == "" {
		attestationVariant = variant.Dummy{}
	} else {
		attestationVariant, err = variant.FromString(attestationString)
		if err != nil {
			return generateFlags{}, fmt.Errorf("invalid attestation variant: %s", attestationString)
		}
	}
	return generateFlags{
		pf:                 pathprefix.New(workDir),
		k8sVersion:         resolvedVersion,
		attestationVariant: attestationVariant,
	}, nil
}

// generateCompletion handles the completion of the create command. It is frequently called
// while the user types arguments of the command to suggest completion.
func generateCompletion(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return []string{"aws", "gcp", "azure", "qemu", "stackit"}, cobra.ShellCompDirectiveNoFileComp
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}

// resolveK8sVersion takes the user input from --kubernetes and transforms a MAJOR.MINOR definition into a supported
// MAJOR.MINOR.PATCH release.
func resolveK8sVersion(k8sVersion string) (string, error) {
	prefixedVersion := compatibility.EnsurePrefixV(k8sVersion)
	if !semver.IsValid(prefixedVersion) {
		return "", fmt.Errorf("kubernetes flag does not specify a valid semantic version: %s", k8sVersion)
	}

	extendedVersion := config.K8sVersionFromMajorMinor(prefixedVersion)
	if extendedVersion == "" {
		return "", fmt.Errorf("--kubernetes (%s) does not specify a valid Kubernetes version. Supported versions: %s", strings.TrimPrefix(k8sVersion, "v"), supportedVersions())
	}

	return extendedVersion, nil
}

func printFormattedSlice[T any](input []T) string {
	return fmt.Sprintf("{%s}", strings.Join(toString(input), "|"))
}

func toString[T any](t []T) []string {
	var res []string
	for _, v := range t {
		res = append(res, fmt.Sprintf("%v", v))
	}
	return res
}
