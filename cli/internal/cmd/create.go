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
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/state"
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
	spinner, writer := newSpinner(cmd, cmd.OutOrStdout())
	defer spinner.Stop()
	creator := cloudcmd.NewCreator(writer)

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

	config, err := readConfig(cmd.OutOrStdout(), fileHandler, flags.configPath)
	if err != nil {
		return fmt.Errorf("reading and validating config: %w", err)
	}

	var printedAWarning bool
	if config.IsDebugImage() {
		cmd.Println("Configured image doesn't look like a released production image. Double check image before deploying to production.")
		printedAWarning = true
	}

	if config.IsDebugCluster() {
		cmd.Println("WARNING: Creating a debug cluster. This cluster is not secure and should only be used for debugging purposes.")
		cmd.Println("DO NOT USE THIS CLUSTER IN PRODUCTION.")
		printedAWarning = true
	}

	if config.IsAzureNonCVM() {
		cmd.Println("Disabling Confidential VMs is insecure. Use only for evaluation purposes.")
		printedAWarning = true
		if config.EnforcesIDKeyDigest() {
			cmd.Println("Your config asks for enforcing the idkeydigest. This is only available on Confidential VMs. It will not be enforced.")
		}
	}

	// Print an extra new line later to separate warnings from the prompt message of the create command
	if printedAWarning {
		cmd.Println("")
	}

	provider := config.GetProvider()
	var instanceType string
	switch provider {
	case cloudprovider.AWS:
		instanceType = config.Provider.AWS.InstanceType
	case cloudprovider.Azure:
		instanceType = config.Provider.Azure.InstanceType
	case cloudprovider.GCP:
		instanceType = config.Provider.GCP.InstanceType
	case cloudprovider.QEMU:
		cpus := config.Provider.QEMU.VCPUs
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
	state, err := creator.Create(cmd.Context(), provider, config, flags.name, instanceType, flags.controllerCount, flags.workerCount)
	spinner.Stop()
	if err != nil {
		return err
	}

	if err := fileHandler.WriteJSON(constants.StateFilename, state, file.OptNone); err != nil {
		return err
	}

	if err := writeIPtoIDFile(fileHandler, state); err != nil {
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
	if _, err := fileHandler.Stat(constants.StateFilename); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("file '%s' already exists in working directory, run 'constellation terminate' before creating a new one", constants.StateFilename)
	}
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

func writeIPtoIDFile(fileHandler file.Handler, state state.ConstellationState) error {
	ip := state.LoadBalancerIP
	if ip == "" {
		return fmt.Errorf("bootstrapper ip not found")
	}
	idFile := clusterIDsFile{IP: ip}
	return fileHandler.WriteJSON(constants.ClusterIDsFileName, idFile, file.OptNone)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
