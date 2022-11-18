/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// NewCreateCmd returns a new cobra.Command for the create command.
func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create instances on a cloud platform for your Constellation cluster",
		Long:  "Create instances on a cloud platform for your Constellation cluster.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(0),
		),
		RunE: runCreate,
	}
	cmd.Flags().String("name", "constell", "create the cluster with the specified name")
	cmd.Flags().BoolP("yes", "y", false, "create the cluster without further confirmation")
	cmd.Flags().IntP("control-plane-nodes", "c", 0, "number of control-plane nodes (required)")
	must(cobra.MarkFlagRequired(cmd.Flags(), "control-plane-nodes"))
	cmd.Flags().IntP("worker-nodes", "w", 0, "number of worker nodes (required)")
	must(cobra.MarkFlagRequired(cmd.Flags(), "worker-nodes"))
	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	spinner := newSpinner(cmd.ErrOrStderr())
	defer spinner.Stop()
	creator := cloudcmd.NewCreator(spinner)

	return create(cmd, creator, fileHandler, spinner)
}

func create(cmd *cobra.Command, creator cloudCreator, fileHandler file.Handler, spinner spinnerInterf,
) (retErr error) {
	flags, err := parseCreateFlags(cmd)
	if err != nil {
		return err
	}

	if err := checkDirClean(fileHandler); err != nil {
		return err
	}

	conf, err := config.New(fileHandler, flags.configPath)
	if err != nil {
		return displayConfigValidationErrors(cmd.ErrOrStderr(), err)
	}

	var printedAWarning bool
	if conf.IsDebugImage() {
		cmd.PrintErrln("Configured image doesn't look like a released production image. Double check image before deploying to production.")
		printedAWarning = true
	}

	if conf.IsDebugCluster() {
		cmd.PrintErrln("WARNING: Creating a debug cluster. This cluster is not secure and should only be used for debugging purposes.")
		cmd.PrintErrln("DO NOT USE THIS CLUSTER IN PRODUCTION.")
		printedAWarning = true
	}

	// Print an extra new line later to separate warnings from the prompt message of the create command
	if printedAWarning {
		cmd.PrintErrln("")
	}

	provider := conf.GetProvider()
	var instanceType string
	switch provider {
	case cloudprovider.AWS:
		instanceType = conf.Provider.AWS.InstanceType
		if len(flags.name) > 10 {
			return fmt.Errorf("cluster name on AWS must not be longer than 10 characters")
		}
	case cloudprovider.Azure:
		instanceType = conf.Provider.Azure.InstanceType
	case cloudprovider.GCP:
		instanceType = conf.Provider.GCP.InstanceType
	case cloudprovider.QEMU:
		cpus := conf.Provider.QEMU.VCPUs
		instanceType = fmt.Sprintf("%d-vCPU", cpus)
	}

	if !flags.yes {
		// Ask user to confirm action.
		cmd.Printf("The following Constellation cluster will be created:\n")
		cmd.Printf("%d control-planes nodes of type %s will be created.\n", flags.controllerCount, instanceType)
		cmd.Printf("%d worker nodes of type %s will be created.\n", flags.workerCount, instanceType)
		ok, err := askToConfirm(cmd, "Do you want to create this cluster?")
		if err != nil {
			return err
		}
		if !ok {
			cmd.Println("The creation of the cluster was aborted.")
			return nil
		}
	}

	spinner.Start("Creating", false)
	idFile, err := creator.Create(cmd.Context(), provider, conf, flags.name, instanceType, flags.controllerCount, flags.workerCount)
	spinner.Stop()
	if err != nil {
		return translateCreateErrors(cmd, err)
	}

	if err := fileHandler.WriteJSON(constants.ClusterIDsFileName, idFile, file.OptNone); err != nil {
		return err
	}

	cmd.Println("Your Constellation cluster was created successfully.")
	return nil
}

// parseCreateFlags parses the flags of the create command.
func parseCreateFlags(cmd *cobra.Command) (createFlags, error) {
	controllerCount, err := cmd.Flags().GetInt("control-plane-nodes")
	if err != nil {
		return createFlags{}, fmt.Errorf("parsing number of control-plane nodes: %w", err)
	}
	if controllerCount < constants.MinControllerCount {
		return createFlags{}, fmt.Errorf("number of control-plane nodes must be at least %d", constants.MinControllerCount)
	}

	workerCount, err := cmd.Flags().GetInt("worker-nodes")
	if err != nil {
		return createFlags{}, fmt.Errorf("parsing number of worker nodes: %w", err)
	}
	if workerCount < constants.MinWorkerCount {
		return createFlags{}, fmt.Errorf("number of worker nodes must be at least %d", constants.MinWorkerCount)
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return createFlags{}, fmt.Errorf("parsing name argument: %w", err)
	}
	if len(name) > constants.ConstellationNameLength {
		return createFlags{}, fmt.Errorf(
			"name for Constellation cluster too long, maximum length is %d, got %d: %s",
			constants.ConstellationNameLength, len(name), name,
		)
	}

	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return createFlags{}, fmt.Errorf("%w; Set '-yes' without a value to automatically confirm", err)
	}

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return createFlags{}, fmt.Errorf("parsing config path argument: %w", err)
	}

	return createFlags{
		controllerCount: controllerCount,
		workerCount:     workerCount,
		name:            name,
		configPath:      configPath,
		yes:             yes,
	}, nil
}

// createFlags contains the parsed flags of the create command.
type createFlags struct {
	controllerCount int
	workerCount     int
	name            string
	configPath      string
	yes             bool
}

// checkDirClean checks if files of a previous Constellation are left in the current working dir.
func checkDirClean(fileHandler file.Handler) error {
	if _, err := fileHandler.Stat(constants.AdminConfFilename); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("file '%s' already exists in working directory, run 'constellation terminate' before creating a new one", constants.AdminConfFilename)
	}
	if _, err := fileHandler.Stat(constants.MasterSecretFilename); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("file '%s' already exists in working directory. Constellation won't overwrite previous master secrets. Move it somewhere or delete it before creating a new cluster", constants.MasterSecretFilename)
	}
	if _, err := fileHandler.Stat(constants.ClusterIDsFileName); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("file '%s' already exists in working directory. Constellation won't overwrite previous cluster IDs. Move it somewhere or delete it before creating a new cluster", constants.ClusterIDsFileName)
	}

	return nil
}

func translateCreateErrors(cmd *cobra.Command, err error) error {
	switch {
	case errors.Is(err, terraform.ErrTerraformWorkspaceDifferentFiles):
		cmd.PrintErrln("\nYour current working directory contains an existing Terraform workspace which does not match the expected state.")
		cmd.PrintErrln("This can be due to a mix up between providers, versions or an otherwise corrupted workspace.")
		cmd.PrintErrln("Before creating a new cluster, try \"constellation terminate\".")
		cmd.PrintErrf("If this does not work, either move or delete the directory %q.\n", constants.TerraformWorkingDir)
		cmd.PrintErrln("Please only delete the directory if you made sure that all created cloud resources have been terminated.")
		return err
	case errors.Is(err, terraform.ErrTerraformWorkspaceExistsWithDifferentVariables):
		cmd.PrintErrln("\nYour current working directory contains an existing Terraform workspace which was initiated with different input variables.")
		cmd.PrintErrln("This can be the case if you have tried to create a cluster before with different options which did not complete, or the workspace is corrupted.")
		cmd.PrintErrln("Before creating a new cluster, try \"constellation terminate\".")
		cmd.PrintErrf("If this does not work, either move or delete the directory %q.\n", constants.TerraformWorkingDir)
		cmd.PrintErrln("Please only delete the directory if you made sure that all created cloud resources have been terminated.")
		return err
	default:
		return err
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
