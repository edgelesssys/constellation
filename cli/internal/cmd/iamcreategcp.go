/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// NewIAMCreateGCPCmd returns a new cobra.Command for the iam create gcp command.
func newIAMCreateGCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcp",
		Short: "Create IAM configuration on GCP for your Constellation cluster",
		Long:  "Create IAM configuration on GCP for your Constellation cluster.",
		Args:  cobra.ExactArgs(0),
		RunE:  runIAMCreateGCP,
	}

	cmd.Flags().String("zone", "", "GCP zone the cluster will be deployed in (required)\n"+
		"Find a list of available zones here: https://cloud.google.com/compute/docs/regions-zones#available")
	must(cobra.MarkFlagRequired(cmd.Flags(), "zone"))
	cmd.Flags().String("serviceAccountID", "", "ID for the service account that will be created (required)\n"+
		"Must be 6 to 30 lowercase letters, digits, or hyphens.")
	must(cobra.MarkFlagRequired(cmd.Flags(), "serviceAccountID"))
	cmd.Flags().String("projectID", "", "ID of the GCP project the configuration will be created in (required)\n"+
		"Find it on the welcome screen of your project: https://console.cloud.google.com/welcome")
	must(cobra.MarkFlagRequired(cmd.Flags(), "projectID"))

	return cmd
}

func runIAMCreateGCP(cmd *cobra.Command, _ []string) error {
	creator := &gcpIAMCreator{}
	if err := creator.flags.parse(cmd.Flags()); err != nil {
		return err
	}
	return runIAMCreate(cmd, creator, cloudprovider.GCP)
}

// gcpIAMCreateFlags contains the parsed flags of the iam create gcp command.
type gcpIAMCreateFlags struct {
	rootFlags
	serviceAccountID string
	zone             string
	region           string
	projectID        string
}

func (f *gcpIAMCreateFlags) parse(flags *pflag.FlagSet) error {
	var err error
	if err = f.rootFlags.parse(flags); err != nil {
		return err
	}

	f.zone, err = flags.GetString("zone")
	if err != nil {
		return fmt.Errorf("getting 'zone' flag: %w", err)
	}
	if !zoneRegex.MatchString(f.zone) {
		return fmt.Errorf("invalid zone string: %s", f.zone)
	}

	// Infer region from zone.
	zoneParts := strings.Split(f.zone, "-")
	f.region = fmt.Sprintf("%s-%s", zoneParts[0], zoneParts[1])
	if !regionRegex.MatchString(f.region) {
		return fmt.Errorf("invalid region string: %s", f.region)
	}

	f.projectID, err = flags.GetString("projectID")
	if err != nil {
		return fmt.Errorf("getting 'projectID' flag: %w", err)
	}
	if !gcpIDRegex.MatchString(f.projectID) {
		return fmt.Errorf("projectID %q doesn't match %s", f.projectID, gcpIDRegex)
	}

	f.serviceAccountID, err = flags.GetString("serviceAccountID")
	if err != nil {
		return fmt.Errorf("getting 'serviceAccountID' flag: %w", err)
	}
	if !gcpIDRegex.MatchString(f.serviceAccountID) {
		return fmt.Errorf("serviceAccountID %q doesn't match %s", f.serviceAccountID, gcpIDRegex)
	}
	return nil
}

// gcpIAMCreator implements the providerIAMCreator interface for GCP.
type gcpIAMCreator struct {
	flags gcpIAMCreateFlags
}

func (c *gcpIAMCreator) getIAMConfigOptions() *cloudcmd.IAMConfigOptions {
	return &cloudcmd.IAMConfigOptions{
		GCP: cloudcmd.GCPIAMConfig{
			Zone:             c.flags.zone,
			Region:           c.flags.region,
			ProjectID:        c.flags.projectID,
			ServiceAccountID: c.flags.serviceAccountID,
		},
	}
}

func (c *gcpIAMCreator) printConfirmValues(cmd *cobra.Command) {
	cmd.Printf("Project ID:\t\t%s\n", c.flags.projectID)
	cmd.Printf("Service Account ID:\t%s\n", c.flags.serviceAccountID)
	cmd.Printf("Region:\t\t\t%s\n", c.flags.region)
	cmd.Printf("Zone:\t\t\t%s\n\n", c.flags.zone)
}

func (c *gcpIAMCreator) printOutputValues(cmd *cobra.Command, _ cloudcmd.IAMOutput) {
	cmd.Printf("projectID:\t\t%s\n", c.flags.projectID)
	cmd.Printf("region:\t\t\t%s\n", c.flags.region)
	cmd.Printf("zone:\t\t\t%s\n", c.flags.zone)
	cmd.Printf("serviceAccountKeyPath:\t%s\n\n", c.flags.pathPrefixer.PrefixPrintablePath(constants.GCPServiceAccountKeyFilename))
}

func (c *gcpIAMCreator) writeOutputValuesToConfig(conf *config.Config, _ cloudcmd.IAMOutput) {
	conf.Provider.GCP.Project = c.flags.projectID
	conf.Provider.GCP.ServiceAccountKeyPath = constants.GCPServiceAccountKeyFilename // File was created in workspace, so only the filename is needed.
	conf.Provider.GCP.Region = c.flags.region
	conf.Provider.GCP.Zone = c.flags.zone
	for groupName, group := range conf.NodeGroups {
		group.Zone = c.flags.zone
		conf.NodeGroups[groupName] = group
	}
}

func (c *gcpIAMCreator) parseAndWriteIDFile(iamFile cloudcmd.IAMOutput, fileHandler file.Handler) error {
	// GCP needs to write the service account key to a file.
	tmpOut, err := parseIDFile(iamFile.GCPOutput.ServiceAccountKey)
	if err != nil {
		return err
	}

	return fileHandler.WriteJSON(constants.GCPServiceAccountKeyFilename, tmpOut, file.OptNone)
}

func (c *gcpIAMCreator) validateConfigWithFlagCompatibility(conf config.Config) error {
	return validateConfigWithFlagCompatibility(cloudprovider.GCP, conf, c.flags.zone)
}
