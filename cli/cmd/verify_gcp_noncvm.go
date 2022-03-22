package cmd

import (
	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/attestation/gcp"
	"github.com/spf13/cobra"
)

// TODO: Remove this command once we no longer use non cvms.
func newVerifyGCPNonCVMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "gcp-non-cvm IP PORT",
		Short:             "Verify the TPM attestation of your shielded VM Constellation on Google Cloud Platform.",
		Long:              "Verify the TPM attestation of your shielded VM Constellation on Google Cloud Platform.",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: verifyCompletion,
		RunE:              runVerifyGCPNonCVM,
	}

	return cmd
}

func runVerifyGCPNonCVM(cmd *cobra.Command, args []string) error {
	pcrs := map[uint32][]byte{}
	validator, err := getGCPNonCVMValidator(cmd, pcrs)
	if err != nil {
		return err
	}

	return runVerify(cmd, args, pcrs, validator)
}

// getGCPNonCVMValidator returns a GCP validator for regular shielded VMs.
func getGCPNonCVMValidator(cmd *cobra.Command, pcrs map[uint32][]byte) (atls.Validator, error) {
	if err := prepareValidator(cmd, pcrs); err != nil {
		return nil, err
	}

	return gcp.NewNonCVMValidator(pcrs), nil
}
