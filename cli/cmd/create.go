package cmd

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/edgelesssys/constellation/cli/azure"
	"github.com/edgelesssys/constellation/cli/cloud/cloudcmd"
	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/cli/gcp"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create {aws|gcp|azure}",
		Short: "Create instances on a cloud platform for your Constellation cluster",
		Long:  "Create instances on a cloud platform for your Constellation cluster.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			isCloudProvider(0),
			warnAWS(0),
		),
		ValidArgsFunction: createCompletion,
		RunE:              runCreate,
	}
	cmd.Flags().String("name", "constell", "create the cluster with the specified name")
	cmd.Flags().BoolP("yes", "y", false, "create the cluster without further confirmation")
	cmd.Flags().IntP("control-plane-nodes", "c", 0, "number of control-plane nodes (required)")
	must(cobra.MarkFlagRequired(cmd.Flags(), "control-plane-nodes"))
	cmd.Flags().IntP("worker-nodes", "w", 0, "number of worker nodes (required)")
	must(cobra.MarkFlagRequired(cmd.Flags(), "worker-nodes"))
	cmd.Flags().StringP("instance-type", "t", "", "instance type of cluster nodes")
	must(cmd.RegisterFlagCompletionFunc("instance-type", instanceTypeCompletion))
	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	provider := cloudprovider.FromString(args[0])
	fileHandler := file.NewHandler(afero.NewOsFs())
	creator := cloudcmd.NewCreator(cmd.OutOrStdout())

	return create(cmd, creator, fileHandler, provider)
}

func create(cmd *cobra.Command, creator cloudCreator, fileHandler file.Handler, provider cloudprovider.Provider,
) (retErr error) {
	flags, err := parseCreateFlags(cmd, provider)
	if err != nil {
		return err
	}

	if err := checkDirClean(fileHandler); err != nil {
		return err
	}

	config, err := config.FromFile(fileHandler, flags.devConfigPath)
	if err != nil {
		return err
	}

	if !flags.yes {
		// Ask user to confirm action.
		cmd.Printf("The following Constellation cluster will be created:\n")
		cmd.Printf("%d control-planes nodes of type %s will be created.\n", flags.controllerCount, flags.insType)
		cmd.Printf("%d worker nodes of type %s will be created.\n", flags.workerCount, flags.insType)
		ok, err := askToConfirm(cmd, "Do you want to create this cluster?")
		if err != nil {
			return err
		}
		if !ok {
			cmd.Println("The creation of the cluster was aborted.")
			return nil
		}
	}

	state, err := creator.Create(cmd.Context(), provider, config, flags.name, flags.insType, flags.controllerCount, flags.workerCount)
	if err != nil {
		return err
	}

	if err := fileHandler.WriteJSON(constants.StateFilename, state, file.OptNone); err != nil {
		return err
	}

	cmd.Println("Your Constellation cluster was created successfully.")
	return nil
}

// parseCreateFlags parses the flags of the create command.
func parseCreateFlags(cmd *cobra.Command, provider cloudprovider.Provider) (createFlags, error) {
	controllerCount, err := cmd.Flags().GetInt("control-plane-nodes")
	if err != nil {
		return createFlags{}, err
	}
	if controllerCount < constants.MinControllerCount {
		return createFlags{}, fmt.Errorf("number of control-plane nodes must be at least %d", constants.MinControllerCount)
	}

	workerCount, err := cmd.Flags().GetInt("worker-nodes")
	if err != nil {
		return createFlags{}, err
	}
	if workerCount < constants.MinWorkerCount {
		return createFlags{}, fmt.Errorf("number of worker nodes must be at least %d", constants.MinWorkerCount)
	}

	insType, err := cmd.Flags().GetString("instance-type")
	if err != nil {
		return createFlags{}, err
	}
	if insType == "" {
		insType = defaultInstanceType(provider)
	}
	if err := validInstanceTypeForProvider(insType, provider); err != nil {
		return createFlags{}, err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return createFlags{}, err
	}
	if len(name) > constellationNameLength {
		return createFlags{}, fmt.Errorf(
			"name for Constellation cluster too long, maximum length is %d, got %d: %s",
			constellationNameLength, len(name), name,
		)
	}

	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return createFlags{}, err
	}

	devConfigPath, err := cmd.Flags().GetString("dev-config")
	if err != nil {
		return createFlags{}, err
	}

	return createFlags{
		controllerCount: controllerCount,
		workerCount:     workerCount,
		insType:         insType,
		name:            name,
		devConfigPath:   devConfigPath,
		yes:             yes,
	}, nil
}

// createFlags contains the parsed flags of the create command.
type createFlags struct {
	controllerCount int
	workerCount     int
	insType         string
	name            string
	devConfigPath   string
	yes             bool
}

// defaultInstanceType returns the default instance type for the given provider.
func defaultInstanceType(provider cloudprovider.Provider) string {
	switch provider {
	case cloudprovider.GCP:
		return gcp.InstanceTypes[0]
	case cloudprovider.Azure:
		return azure.InstanceTypes[0]
	default:
		return ""
	}
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
		return fmt.Errorf("file '%s' already exists in working directory, clean it up first", constants.MasterSecretFilename)
	}

	return nil
}

// createCompletion handles the completion of the create command. It is frequently called
// while the user types arguments of the command to suggest completion.
func createCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return []string{"aws", "gcp", "azure"}, cobra.ShellCompDirectiveNoFileComp
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}

func instanceTypeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 1 {
		return []string{}, cobra.ShellCompDirectiveError
	}
	switch args[0] {
	case "gcp":
		return gcp.InstanceTypes, cobra.ShellCompDirectiveNoFileComp
	case "azure":
		return azure.InstanceTypes, cobra.ShellCompDirectiveNoFileComp
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}
