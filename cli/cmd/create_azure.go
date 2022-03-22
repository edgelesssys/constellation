package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/cli/azure"
	"github.com/edgelesssys/constellation/cli/azure/client"
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newCreateAzureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azure",
		Short: "Create a Constellation of NUMBER nodes of SIZE on Azure.",
		Long:  "Create a Constellation of NUMBER nodes of SIZE on Azure.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(2),
			isIntGreaterArg(0, 1),
			isAzureInstanceType(1),
		),
		ValidArgsFunction: createAzureCompletion,
		RunE:              runCreateAzure,
	}
	return cmd
}

// runCreateAzure runs the create command.
func runCreateAzure(cmd *cobra.Command, args []string) error {
	count, _ := strconv.Atoi(args[0]) // err already checked in args validation
	size := strings.ToLower(args[1])
	subscriptionID := "0d202bbb-4fa7-4af8-8125-58c269a05435" // TODO: This will be user input
	tenantID := "adb650a8-5da3-4b15-b4b0-3daf65ff7626"       // TODO: This will be user input
	location := "North Europe"                               // TODO: This will be user input

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	if len(name) > constellationNameLength {
		return fmt.Errorf("name for constellation too long, maximum length is %d got %d: %s", constellationNameLength, len(name), name)
	}

	client, err := client.NewInitialized(
		subscriptionID,
		tenantID,
		name,
		location,
	)
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

	return createAzure(cmd, client, fileHandler, config, size, count)
}

func createAzure(cmd *cobra.Command, cl azureclient, fileHandler file.Handler, config *config.Config, size string, count int) (retErr error) {
	if err := checkDirClean(fileHandler, config); err != nil {
		return err
	}

	ok, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}
	if !ok {
		// Ask user to confirm action.
		cmd.Printf("The following Constellation will be created:\n")
		cmd.Printf("%d nodes of size %s will be created.\n", count, size)
		ok, err := askToConfirm(cmd, "Do you want to create this Constellation?")
		if err != nil {
			return err
		}
		if !ok {
			cmd.Println("The creation of the Constellation was aborted.")
			return nil
		}
	}

	// Create all azure resources
	defer rollbackOnError(context.Background(), cmd.OutOrStdout(), &retErr, &rollbackerAzure{client: cl})
	if err := cl.CreateResourceGroup(cmd.Context()); err != nil {
		return err
	}
	if err := cl.CreateVirtualNetwork(cmd.Context()); err != nil {
		return err
	}
	if err := cl.CreateSecurityGroup(cmd.Context(), *config.Provider.Azure.NetworkSecurityGroupInput); err != nil {
		return err
	}
	if err := cl.CreateInstances(cmd.Context(), client.CreateInstancesInput{
		Count:                count,
		InstanceType:         size,
		Image:                *config.Provider.Azure.Image,
		UserAssingedIdentity: *config.Provider.Azure.UserAssignedIdentity,
	}); err != nil {
		return err
	}

	stat, err := cl.GetState()
	if err != nil {
		return err
	}
	if err := fileHandler.WriteJSON(*config.StatePath, stat, false); err != nil {
		return err
	}

	cmd.Println("Your Constellation was created successfully.")
	return nil
}

// createAzureCompletion handels the completion of CLI arguments. It is frequently called
// while the user types arguments of the command to suggest completion.
func createAzureCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	case 1:
		return azure.InstanceTypes, cobra.ShellCompDirectiveDefault
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}
