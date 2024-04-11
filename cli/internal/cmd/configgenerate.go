/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/mod/semver"
)

func newConfigGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate {aws|azure|gcp|openstack|qemu|stackit}",
		Short: "Generate a default configuration and state file",
		Long:  "Generate a default configuration and state file for your selected cloud provider.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			isCloudProvider(0),
		),
		ValidArgsFunction: generateCompletion,
		RunE:              runConfigGenerate,
	}
	cmd.Flags().StringP("kubernetes", "k", semver.MajorMinor(string(config.Default().KubernetesVersion)), "Kubernetes version to use in format MAJOR.MINOR")
	cmd.Flags().StringP("attestation", "a", "", fmt.Sprintf("attestation variant to use %s. If not specified, the default for the cloud provider is used", printFormattedSlice(variant.GetAvailableAttestationVariants())))
	cmd.Flags().StringP("tags", "t", "", "Additional tags for created resources.")

	return cmd
}

type generateFlags struct {
	rootFlags
	k8sVersion         versions.ValidK8sVersion
	attestationVariant variant.Variant
	tags cloudprovider.Tags
}

func (f *generateFlags) parse(flags *pflag.FlagSet) error {
	if err := f.rootFlags.parse(flags); err != nil {
		return err
	}

	k8sVersion, err := parseK8sFlag(flags)
	if err != nil {
		return err
	}
	f.k8sVersion = k8sVersion

	variant, err := parseAttestationFlag(flags)
	if err != nil {
		return err
	}
	f.attestationVariant = variant

	tags, err := parseTagsFlags(flags)
	if err != nil {
		return err
	}
	f.tags = tags

	return nil
}

type configGenerateCmd struct {
	flags generateFlags
	log   debugLog
}

func runConfigGenerate(cmd *cobra.Command, args []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}

	fileHandler := file.NewHandler(afero.NewOsFs())
	provider := cloudprovider.FromString(args[0])

	cg := &configGenerateCmd{log: log}
	if err := cg.flags.parse(cmd.Flags()); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}
	log.Debug("Using flags", "k8sVersion", cg.flags.k8sVersion, "attestationVariant", cg.flags.attestationVariant)

	return cg.configGenerate(cmd, fileHandler, provider, args[0])
}

func (cg *configGenerateCmd) configGenerate(cmd *cobra.Command, fileHandler file.Handler, provider cloudprovider.Provider, rawProvider string) error {
	cg.log.Debug(fmt.Sprintf("Using cloud provider %q", provider.String()))

	// Config creation
	conf, err := createConfigWithAttestationVariant(provider, rawProvider, cg.flags.attestationVariant)
	if err != nil {
		return fmt.Errorf("creating config: %w", err)
	}
	conf.KubernetesVersion = cg.flags.k8sVersion
	cg.log.Debug("Writing YAML data to configuration file")
	if err := fileHandler.WriteYAML(constants.ConfigFilename, conf, file.OptMkdirAll); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	cmd.Println("Config file written to", cg.flags.pathPrefixer.PrefixPrintablePath(constants.ConfigFilename))
	cmd.Println("Please fill in your CSP-specific configuration before proceeding.")

	// State-file creation
	stateFile := state.New()
	switch provider {
	case cloudprovider.GCP:
		stateFile.SetInfrastructure(state.Infrastructure{GCP: &state.GCP{}})
	case cloudprovider.Azure:
		stateFile.SetInfrastructure(state.Infrastructure{Azure: &state.Azure{}})
	}
	if err = stateFile.WriteToFile(fileHandler, constants.StateFilename); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}
	cmd.Println("State file written to", cg.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename))

	cmd.Println("For more information refer to the documentation:")
	cmd.Println("\thttps://docs.edgeless.systems/constellation/getting-started/first-steps")

	return nil
}

// createConfigWithAttestationVariant creates a config file for the given provider.
func createConfigWithAttestationVariant(provider cloudprovider.Provider, rawProvider string, attestationVariant variant.Variant) (*config.Config, error) {
	conf := config.Default().WithOpenStackProviderDefaults(provider, rawProvider)
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

	conf.SetCSPNodeGroupDefaults(provider)
	return conf, nil
}

// createConfig creates a config file for the given provider.
func createConfig(provider cloudprovider.Provider) *config.Config {
	// rawProvider can be hardcoded as it only matters for OpenStack
	res, _ := createConfigWithAttestationVariant(provider, "", variant.Dummy{})
	return res
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

func parseK8sFlag(flags *pflag.FlagSet) (versions.ValidK8sVersion, error) {
	versionString, err := flags.GetString("kubernetes")
	if err != nil {
		return "", fmt.Errorf("getting kubernetes flag: %w", err)
	}
	resolvedVersion, err := versions.ResolveK8sPatchVersion(versionString)
	if err != nil {
		return "", fmt.Errorf("resolving kubernetes patch version from flag: %w", err)
	}
	k8sVersion, err := versions.NewValidK8sVersion(resolvedVersion, true)
	if err != nil {
		return "", fmt.Errorf("resolving Kubernetes version from flag: %w", err)
	}
	return k8sVersion, nil
}

func parseAttestationFlag(flags *pflag.FlagSet) (variant.Variant, error) {
	attestationString, err := flags.GetString("attestation")
	if err != nil {
		return nil, fmt.Errorf("getting attestation flag: %w", err)
	}

	var attestationVariant variant.Variant
	// if no attestation variant is specified, use the default for the cloud provider
	if attestationString == "" {
		attestationVariant = variant.Dummy{}
	} else {
		attestationVariant, err = variant.FromString(attestationString)
		if err != nil {
			return nil, fmt.Errorf("invalid attestation variant: %s", attestationString)
		}
	}

	return attestationVariant, nil
}

func parseTagsFlags(flags *pflag.FlagSet) (cloudprovider.Tags, error) {
	tagsString, err := flags.GetString("tags")
	if err != nil {
		return nil, fmt.Errorf("getting tags flag: %w", err)
	}
	tagsString = strings.ReplaceAll(tagsString, " ", "")
	tagsStringSplit := strings.Split(tagsString, ",")

	tags := make(cloudprovider.Tags)
	for _, tag := range tagsStringSplit {
		tagSplit := strings.Split(tag, "=")
		if len(tagSplit) != 2 {
			return nil, fmt.Errorf("wrong format of tags")
		}

		tags[tagSplit[0]] = tagSplit[1]
	}

	return tags, nil
}
