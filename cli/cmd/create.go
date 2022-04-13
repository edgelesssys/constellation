package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

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
		Use:   "create {aws|gcp|azure} C_COUNT N_COUNT TYPE",
		Short: "Create instances on a cloud platform for your Constellation.",
		Long: `Create instances on a cloud platform for your Constellation.
A Constellation with C_COUNT Coordinators and N_COUNT Nodes is created.
TYPE is the instance type used for all instances.`,
		Args: cobra.MatchAll(
			cobra.ExactArgs(4),
			isIntGreaterZeroArg(1),
			isIntGreaterZeroArg(2),
			isInstanceTypeForProvider(3, 0),
			warnAWS(0),
		),
		ValidArgsFunction: createCompletion,
		RunE:              runCreate,
	}
	cmd.Flags().String("name", "constell", "Set this flag to create the Constellation with the specified name.")
	cmd.Flags().BoolP("yes", "y", false, "Set this flag to create the Constellation without further confirmation.")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	provider := cloudprovider.FromString(args[0])
	countCoord, _ := strconv.Atoi(args[1]) // err checked in args validation
	countNode, _ := strconv.Atoi(args[2])  // err checked in args validation
	insType := strings.ToLower(args[3])
	fileHandler := file.NewHandler(afero.NewOsFs())
	creator := cloudcmd.NewCreator(cmd.OutOrStdout())

	return create(cmd, creator, fileHandler, countCoord, countNode, provider, insType)
}

func create(cmd *cobra.Command, creator cloudCreator, fileHandler file.Handler,
	countCoord, countNode int, provider cloudprovider.CloudProvider, insType string,
) (retErr error) {
	flags, err := parseCreateFlags(cmd)
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
		cmd.Printf("The following Constellation will be created:\n")
		cmd.Printf("%d coordinators of type %s will be created.\n", countCoord, insType)
		cmd.Printf("%d nodes of type %s will be created.\n", countNode, insType)
		ok, err := askToConfirm(cmd, "Do you want to create this Constellation?")
		if err != nil {
			return err
		}
		if !ok {
			cmd.Println("The creation of the Constellation was aborted.")
			return nil
		}
	}

	state, err := creator.Create(cmd.Context(), provider, config, flags.name, insType, countCoord, countNode)
	if err != nil {
		return err
	}

	if err := fileHandler.WriteJSON(constants.StateFilename, state, file.OptNone); err != nil {
		return err
	}

	cmd.Println("Your Constellation was created successfully.")
	return nil
}

// parseCreateFlags parses the flags of the create command.
func parseCreateFlags(cmd *cobra.Command) (createFlags, error) {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return createFlags{}, err
	}
	if len(name) > constellationNameLength {
		return createFlags{}, fmt.Errorf(
			"name for constellation too long, maximum length is %d, got %d: %s",
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
		name:          name,
		devConfigPath: devConfigPath,
		yes:           yes,
	}, nil
}

// createFlags contains the parsed flags of the create command.
type createFlags struct {
	name          string
	devConfigPath string
	yes           bool
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
	case 1:
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	case 2:
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	case 3:
		var instanceTypeList []string
		switch args[0] {
		case "aws":
			instanceTypeList = []string{
				"4xlarge",
				"8xlarge",
				"12xlarge",
				"16xlarge",
				"24xlarge",
			}
		case "gcp":
			instanceTypeList = gcp.InstanceTypes
		case "azure":
			instanceTypeList = azure.InstanceTypes
		}
		return instanceTypeList, cobra.ShellCompDirectiveNoFileComp
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}
