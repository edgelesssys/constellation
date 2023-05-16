/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/siderolabs/talos/pkg/machinery/config/encoder"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

func newConfigGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate {aws|azure|gcp|openstack|qemu}",
		Short: "Generate a default configuration file",
		Long:  "Generate a default configuration file for your selected cloud provider.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			isCloudProvider(0),
		),
		ValidArgsFunction: generateCompletion,
		RunE:              runConfigGenerate,
	}
	cmd.Flags().StringP("file", "f", constants.ConfigFilename, "path to output file, or '-' for stdout")
	cmd.Flags().StringP("kubernetes", "k", semver.MajorMinor(config.Default().KubernetesVersion), "Kubernetes version to use in format MAJOR.MINOR")
	cmd.Flags().StringP("attestation", "a", "", fmt.Sprintf("Attestation type to use %s. If not specified, the default for the cloud provider is used.", printFormattedSlice(config.GetAvailableAttestationTypes())))

	return cmd
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

type generateFlags struct {
	file            string
	k8sVersion      string
	attestationType config.AttestationType
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
	return cg.configGenerate(cmd, fileHandler, provider)
}

func (cg *configGenerateCmd) configGenerate(cmd *cobra.Command, fileHandler file.Handler, provider cloudprovider.Provider) error {
	flags, err := parseGenerateFlags(cmd)
	if err != nil {
		return err
	}

	cg.log.Debugf("Parsed flags as %v", flags)
	cg.log.Debugf("Using cloud provider %s", provider.String())
	conf, err := createConfigWithAttestationType(provider, flags.attestationType)
	if err != nil {
		return err
	}
	conf.KubernetesVersion = flags.k8sVersion
	if flags.file == "-" {
		content, err := encoder.NewEncoder(conf).Encode()
		if err != nil {
			return fmt.Errorf("encoding config content: %w", err)
		}

		cg.log.Debugf("Writing YAML data to stdout")
		_, err = cmd.OutOrStdout().Write(content)
		return err
	}

	cg.log.Debugf("Writing YAML data to configuration file")
	if err := fileHandler.WriteYAML(flags.file, conf, file.OptMkdirAll); err != nil {
		return err
	}

	cmd.Println("Config file written to", flags.file)
	cmd.Println("Please fill in your CSP-specific configuration before proceeding.")
	cmd.Println("For more information refer to the documentation:")
	cmd.Println("\thttps://docs.edgeless.systems/constellation/getting-started/first-steps")

	return nil
}

// createConfig creates a config file for the given provider.
func createConfigWithAttestationType(provider cloudprovider.Provider, attestationType config.AttestationType) (*config.Config, error) {
	conf := config.Default()
	conf.RemoveProviderExcept(provider)

	// set a lower default for QEMU's state disk
	if provider == cloudprovider.QEMU {
		conf.StateDiskSizeGB = 10
	}

	if provider == cloudprovider.Unknown {
		return conf, nil // TODO tests use Unknown provider... the CLI doesn't allow it.. why do we do that?
	}
	if attestationType == "" {
		attestationType = config.GetDefaultAttestationType(provider)
	} else {
		if !attestationType.ValidProvider(provider) {
			return nil, fmt.Errorf("provider %s does not support attestation type %s", provider, attestationType)
		}
	}
	conf.SetAttestation(attestationType)
	return conf, nil
}

// createConfig creates a config file for the given provider.
func createConfig(provider cloudprovider.Provider) *config.Config {
	res, _ := createConfigWithAttestationType(provider, config.AttestationType(""))
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
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return generateFlags{}, fmt.Errorf("parsing file flag: %w", err)
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

	var attestationType config.AttestationType
	// if no attestation type is specified, use the default for the cloud provider
	if attestationString == "" {
		attestationType = config.AttestationType("")
	} else {
		resType := validateAttestationType(attestationString)
		if resType == nil {
			return generateFlags{}, fmt.Errorf("invalid attestation type: %s", attestationString)
		} else {
			attestationType = *resType
		}
	}
	return generateFlags{
		file:            file,
		k8sVersion:      resolvedVersion,
		attestationType: attestationType,
	}, nil
}

func validateAttestationType(attestationType string) *config.AttestationType {
	for _, aType := range []config.AttestationType{config.AttestationTypeAWSNitroTPM, config.AttestationTypeAzureSEVSNP, config.AttestationTypeAzureTrustedLaunch, config.AttestationTypeGCPSEVES, config.AttestationTypeQEMUVTPM} {
		if attestationType == string(aType) {
			return &aType
		}
	}
	return nil
}

// generateCompletion handles the completion of the create command. It is frequently called
// while the user types arguments of the command to suggest completion.
func generateCompletion(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return []string{"aws", "gcp", "azure", "qemu"}, cobra.ShellCompDirectiveNoFileComp
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
