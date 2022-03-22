package cmd

import (
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/attestation/gcp"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newVerifyGCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "gcp IP PORT",
		Short:             "Verify the confidential properties of your Constellation on Google Cloud Platform.",
		Long:              "Verify the confidential properties of your Constellation on Google Cloud Platform.",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: verifyCompletion,
		RunE:              runVerifyGCP,
	}

	return cmd
}

func runVerifyGCP(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	devConfigName, err := cmd.Flags().GetString("dev-config")
	if err != nil {
		return err
	}
	config, err := config.FromFile(fileHandler, devConfigName)
	if err != nil {
		return err
	}

	validators, err := getGCPValidator(cmd, *config.Provider.GCP.PCRs)
	if err != nil {
		return err
	}

	return runVerify(cmd, args, *config.Provider.GCP.PCRs, validators)
}

// getValidators returns a GCP validator.
func getGCPValidator(cmd *cobra.Command, pcrs map[uint32][]byte) (atls.Validator, error) {
	if err := prepareValidator(cmd, pcrs); err != nil {
		return nil, err
	}

	return gcp.NewValidator(pcrs), nil
}
