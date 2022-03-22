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
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newCreateGCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcp",
		Short: "Create a Constellation of NUMBER nodes of SIZE on Google Cloud Platform.",
		Long:  "Create a Constellation of NUMBER nodes of SIZE on Google Cloud Platform.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(2),
			isIntGreaterArg(0, 1),
			isGCPInstanceType(1),
		),
		ValidArgsFunction: createGCPCompletion,
		RunE:              runCreateGCP,
	}
	return cmd
}

// runCreateGCP runs the create command.
func runCreateGCP(cmd *cobra.Command, args []string) error {
	count, _ := strconv.Atoi(args[0]) // err already checked in args validation
	size := strings.ToLower(args[1])
	project := "constellation-331613" // TODO: This will be user input
	zone := "us-central1-c"           // TODO: This will be user input
	region := "us-central1"           // TODO: This will be user input

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
	return createGCP(cmd, client, fileHandler, config, size, count)
}

func createGCP(cmd *cobra.Command, cl gcpclient, fileHandler file.Handler, config *config.Config, size string, count int) (retErr error) {
	if err := checkDirClean(fileHandler, config); err != nil {
		return err
	}

	createInput := client.CreateInstancesInput{
		Count:        count,
		ImageId:      *config.Provider.GCP.Image,
		InstanceType: size,
		KubeEnv:      gcp.KubeEnv,
		DisableCVM:   *config.Provider.GCP.DisableCVM,
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

	if err := fileHandler.WriteJSON(*config.StatePath, stat, false); err != nil {
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
		return gcp.InstanceTypes, cobra.ShellCompDirectiveDefault
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}
