/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	zoneRegex         = regexp.MustCompile(`^\w+-\w+-[abc]$`)
	regionRegex       = regexp.MustCompile(`^\w+-\w+[0-9]$`)
	projectIDRegex    = regexp.MustCompile(`^[a-z][-a-z0-9]{4,28}[a-z0-9]{1}$`)
	serviceAccIDRegex = regexp.MustCompile(`^[a-z](?:[-a-z0-9]{4,28}[a-z0-9])$`)
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

	cmd.Flags().String("zone", "", "GCP zone the cluster will be deployed in. Find a list of available zones here: https://cloud.google.com/compute/docs/regions-zones#available")
	must(cobra.MarkFlagRequired(cmd.Flags(), "zone"))
	cmd.Flags().String("serviceAccountID", "", "ID for the service account that will be created. Must match ^[a-z](?:[-a-z0-9]{4,28}[a-z0-9])$")
	must(cobra.MarkFlagRequired(cmd.Flags(), "serviceAccountID"))
	cmd.Flags().String("projectID", "", "ID of the GCP project the configuration will be created in. Find it on the welcome screen of your project: https://console.cloud.google.com/welcome")
	must(cobra.MarkFlagRequired(cmd.Flags(), "projectID"))
	cmd.Flags().Bool("yes", false, "Create the IAM configuration without further confirmation")

	return cmd
}

func runIAMCreateGCP(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	spinner := newSpinner(cmd.ErrOrStderr())
	defer spinner.Stop()
	creator := cloudcmd.NewIAMCreator(spinner)

	return iamCreateGCP(cmd, spinner, fileHandler, creator)
}

func iamCreateGCP(cmd *cobra.Command, spinner spinnerInterf, fileHandler file.Handler, creator iamCreator) error {
	// Get input variables.
	gcpFlags, err := parseGCPFlags(cmd)
	if err != nil {
		return err
	}

	// Confirmation.
	if !gcpFlags.yesFlag {
		cmd.Printf("The following IAM configuration will be created:\n")
		cmd.Printf("Project ID:\t%s\n", gcpFlags.projectID)
		cmd.Printf("Service Account ID:\t%s\n", gcpFlags.serviceAccountID)
		cmd.Printf("Region:\t%s\n", gcpFlags.region)
		cmd.Printf("Zone:\t%s\n", gcpFlags.zone)
		ok, err := askToConfirm(cmd, "Do you want to create the configuration?")
		if err != nil {
			return err
		}
		if !ok {
			cmd.Println("The creation of the configuration was aborted.")
			return nil
		}
	}

	// Creation.
	spinner.Start("Creating", false)
	iamFile, err := creator.Create(cmd.Context(), cloudprovider.GCP, &cloudcmd.IAMConfig{
		GCP: cloudcmd.GCPIAMConfig{
			ServiceAccountID: gcpFlags.serviceAccountID,
			Region:           gcpFlags.region,
			Zone:             gcpFlags.zone,
			ProjectID:        gcpFlags.projectID,
		},
	})
	spinner.Stop()
	if err != nil {
		return err
	}

	// Write back values.
	tmpOut, err := parseIDFile(iamFile.GCPOutput.ServiceAccountKey)
	if err != nil {
		return err
	}

	if err := fileHandler.WriteJSON(constants.GCPServiceAccountKeyFile, tmpOut, file.OptNone); err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("serviceAccountKeyPath:\t%s", constants.GCPServiceAccountKeyFile))
	cmd.Println("Your IAM configuration was created successfully. Please fill the above values into your configuration file.")

	return nil
}

func parseIDFile(serviceAccountKeyBase64 string) (map[string]string, error) {
	dec, err := base64.StdEncoding.DecodeString(serviceAccountKeyBase64)
	if err != nil {
		return nil, err
	}

	out := make(map[string]string)
	if err = json.Unmarshal(dec, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// parseGCPFlags parses and validates the flags of the iam create gcp command.
func parseGCPFlags(cmd *cobra.Command) (gcpFlags, error) {
	zone, err := cmd.Flags().GetString("zone")
	if err != nil {
		return gcpFlags{}, fmt.Errorf("parsing zone string: %w", err)
	}
	if !zoneRegex.MatchString(zone) {
		return gcpFlags{}, fmt.Errorf("invalid zone string: %s", zone)
	}

	// Infer region from zone.
	zoneParts := strings.Split(zone, "-")
	region := fmt.Sprintf("%s-%s", zoneParts[0], zoneParts[1])
	if !regionRegex.MatchString(region) {
		return gcpFlags{}, fmt.Errorf("invalid region string: %s", region)
	}

	projectID, err := cmd.Flags().GetString("projectID")
	if err != nil {
		return gcpFlags{}, fmt.Errorf("parsing projectID string: %w", err)
	}
	// Source for regex: https://cloud.google.com/resource-manager/reference/rest/v1/projects.
	if !projectIDRegex.MatchString(projectID) {
		return gcpFlags{}, fmt.Errorf("invalid projectID string: %s", projectID)
	}

	serviceAccID, err := cmd.Flags().GetString("serviceAccountID")
	if err != nil {
		return gcpFlags{}, fmt.Errorf("parsing serviceAccountID string: %w", err)
	}
	if !serviceAccIDRegex.MatchString(serviceAccID) {
		return gcpFlags{}, fmt.Errorf("invalid serviceAccountID string: %s", serviceAccID)
	}

	yesFlag, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return gcpFlags{}, fmt.Errorf("parsing yes bool: %w", err)
	}

	return gcpFlags{
		serviceAccountID: serviceAccID,
		zone:             zone,
		region:           region,
		projectID:        projectID,
		yesFlag:          yesFlag,
	}, nil
}

// gcpFlags contains the parsed flags of the iam create gcp command.
type gcpFlags struct {
	serviceAccountID string
	zone             string
	region           string
	projectID        string
	yesFlag          bool
}
