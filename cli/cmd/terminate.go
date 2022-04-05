package cmd

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"

	azure "github.com/edgelesssys/constellation/cli/azure/client"
	ec2 "github.com/edgelesssys/constellation/cli/ec2/client"
	"github.com/edgelesssys/constellation/cli/file"
	gcp "github.com/edgelesssys/constellation/cli/gcp/client"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
)

func newTerminateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminate",
		Short: "Terminate an existing Constellation.",
		Long:  "Terminate an existing Constellation. The Constellation can't be started again, and all persistent storage will be lost.",
		Args:  cobra.NoArgs,
		RunE:  runTerminate,
	}
	return cmd
}

// runTerminate runs the terminate command.
func runTerminate(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	devConfigName, err := cmd.Flags().GetString("dev-config")
	if err != nil {
		return err
	}
	config, err := config.FromFile(fileHandler, devConfigName)
	if err != nil {
		return err
	}
	return terminate(cmd, fileHandler, config)
}

func terminate(cmd *cobra.Command, fileHandler file.Handler, config *config.Config) error {
	var stat state.ConstellationState
	if err := fileHandler.ReadJSON(*config.StatePath, &stat); err != nil {
		return err
	}

	cmd.Println("Terminating ...")

	if len(stat.EC2Instances) != 0 || stat.EC2SecurityGroup != "" {
		ec2client, err := ec2.NewFromDefault(cmd.Context())
		if err != nil {
			return err
		}
		if err := terminateEC2(cmd, ec2client, stat); err != nil {
			return err
		}
	}
	// TODO: improve check, also look for other resources that might need to be terminated
	if len(stat.GCPNodes) != 0 {
		gcpclient, err := gcp.NewFromDefault(cmd.Context())
		if err != nil {
			return err
		}
		if err := terminateGCP(cmd, gcpclient, stat); err != nil {
			return err
		}
	}

	if len(stat.AzureResourceGroup) != 0 {
		azureclient, err := azure.NewFromDefault(stat.AzureSubscription, stat.AzureTenant)
		if err != nil {
			return err
		}
		if err := terminateAzure(cmd, azureclient, stat); err != nil {
			return err
		}
	}

	cmd.Println("Your Constellation was terminated successfully.")

	var retErr error
	if err := fileHandler.Remove(*config.StatePath); err != nil {
		retErr = multierr.Append(err, fmt.Errorf("failed to remove file '%s', please remove manually", *config.StatePath))
	}

	if err := fileHandler.Remove(*config.AdminConfPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		retErr = multierr.Append(err, fmt.Errorf("failed to remove file '%s', please remove manually", *config.AdminConfPath))
	}

	if err := fileHandler.Remove(*config.WGQuickConfigPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		retErr = multierr.Append(err, fmt.Errorf("failed to remove file '%s', please remove manually", *config.WGQuickConfigPath))
	}

	return retErr
}

func terminateAzure(cmd *cobra.Command, cl azureclient, stat state.ConstellationState) error {
	if err := cl.SetState(stat); err != nil {
		return fmt.Errorf("failed to terminate the Constellation: %w", err)
	}

	if err := cl.TerminateServicePrincipal(cmd.Context()); err != nil {
		return err
	}
	return cl.TerminateResourceGroup(cmd.Context())
}

func terminateGCP(cmd *cobra.Command, cl gcpclient, stat state.ConstellationState) error {
	if err := cl.SetState(stat); err != nil {
		return fmt.Errorf("failed to terminate the Constellation: %w", err)
	}

	if err := cl.TerminateInstances(cmd.Context()); err != nil {
		return err
	}
	if err := cl.TerminateFirewall(cmd.Context()); err != nil {
		return err
	}
	if err := cl.TerminateVPCs(cmd.Context()); err != nil {
		return err
	}
	return cl.TerminateServiceAccount(cmd.Context())
}

// terminateEC2 and remove the existing Constellation form the state file.
func terminateEC2(cmd *cobra.Command, cl ec2client, stat state.ConstellationState) error {
	if err := cl.SetState(stat); err != nil {
		return fmt.Errorf("failed to terminate the Constellation: %w", err)
	}

	if err := cl.TerminateInstances(cmd.Context()); err != nil {
		return fmt.Errorf("failed to terminate the Constellation: %w", err)
	}

	return cl.DeleteSecurityGroup(cmd.Context())
}
