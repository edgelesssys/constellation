package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/cli/gcp"
	"github.com/edgelesssys/constellation/cli/gcp/client"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newCreateGCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcp C_COUNT N_COUNT TYPE",
		Short: "Create a Constellation of C_COUNT coordinators and N_COUNT nodes of TYPE on Google Cloud Platform.",
		Long:  "Create a Constellation of C_COUNT coordinators and N_COUNT nodes of TYPE on Google Cloud Platform.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(3),
			isIntGreaterZeroArg(0),
			isIntGreaterZeroArg(1),
			isGCPInstanceType(2),
		),
		ValidArgsFunction: createGCPCompletion,
		RunE:              runCreateGCP,
	}
	return cmd
}

// runCreateGCP runs the create command.
func runCreateGCP(cmd *cobra.Command, args []string) error {
	countCoordinators, _ := strconv.Atoi(args[0]) // err already checked in args validation
	countNodes, _ := strconv.Atoi(args[1])        // err already checked in args validation
	size := strings.ToLower(args[2])
	project := "constellation-331613" // TODO: This will be user input
	zone := "europe-west3-b"          // TODO: This will be user input
	region := "europe-west3"          // TODO: This will be user input

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	if len(name) > constellationNameLength {
		return fmt.Errorf("name for constellation too long, maximum length is %d got %d: %s", constellationNameLength, len(name), name)
	}

	client, err := client.NewInitialized(cmd.Context(), project, zone, region, name)
	if err != nil {
		return err
	}
	devConfigName, err := cmd.Flags().GetString("dev-config")
	if err != nil {
		return err
	}
	fileHandler := file.NewHandler(afero.NewOsFs())
	config, err := config.FromFile(fileHandler, devConfigName)
	if err != nil {
		return err
	}
	return createGCP(cmd, client, fileHandler, config, size, countCoordinators, countNodes)
}

func createGCP(cmd *cobra.Command, cl gcpclient, fileHandler file.Handler, config *config.Config, size string, countCoordinators, countNodes int) (retErr error) {
	if err := checkDirClean(fileHandler); err != nil {
		return err
	}

	createInput := client.CreateInstancesInput{
		CountNodes:        countNodes,
		CountCoordinators: countCoordinators,
		ImageId:           *config.Provider.GCP.Image,
		InstanceType:      size,
		StateDiskSizeGB:   *config.StateDiskSizeGB,
		KubeEnv:           gcp.KubeEnv,
		DisableCVM:        *config.Provider.GCP.DisableCVM,
	}

	ok, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}
	if !ok {
		// Ask user to confirm action.
		cmd.Printf("The following Constellation will be created:\n")
		cmd.Printf("%d coordinators and %d nodes of size %s will be created.\n", countCoordinators, countNodes, size)
		ok, err := askToConfirm(cmd, "Do you want to create this Constellation?")
		if err != nil {
			return err
		}
		if !ok {
			cmd.Println("The creation of the Constellation was aborted.")
			return nil
		}
	}

	// Create all gcp resources
	defer rollbackOnError(context.Background(), cmd.OutOrStdout(), &retErr, &rollbackerGCP{client: cl})
	if err := cl.CreateVPCs(cmd.Context(), *config.Provider.GCP.VPCsInput); err != nil {
		return err
	}
	if err := cl.CreateFirewall(cmd.Context(), *config.Provider.GCP.FirewallInput); err != nil {
		return err
	}
	if err := cl.CreateInstances(cmd.Context(), createInput); err != nil {
		return err
	}

	stat, err := cl.GetState()
	if err != nil {
		return err
	}

	if err := fileHandler.WriteJSON(constants.StateFilename, stat, file.OptNone); err != nil {
		return err
	}

	cmd.Println("Your Constellation was created successfully.")
	return nil
}

// createGCPCompletion handels the completion of CLI arguments. It is frequently called
// while the user types arguments of the command to suggest completion.
func createGCPCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	case 1:
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	case 2:
		return gcp.InstanceTypes, cobra.ShellCompDirectiveDefault
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}
