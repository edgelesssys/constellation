package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/edgelesssys/constellation/cli/ec2"
	"github.com/edgelesssys/constellation/cli/ec2/client"
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/internal/config"
)

func newCreateAWSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "aws C_COUNT N_COUNT TYPE",
		Short:   "Create a Constellation of C_COUNT coordinators and N_COUNT nodes of TYPE on AWS.",
		Long:    "Create a Constellation of C_COUNT coordinators and N_COUNT nodes of TYPE on AWS.",
		Example: "aws 1 4 2xlarge",
		Args: cobra.MatchAll(
			cobra.ExactArgs(3),
			isValidAWSCoordinatorCount(0),
			isIntGreaterZeroArg(1),
			isEC2InstanceType(2),
		),
		ValidArgsFunction: createAWSCompletion,
		RunE:              runCreateAWS,
	}
	return cmd
}

// runCreateAWS runs the create command.
func runCreateAWS(cmd *cobra.Command, args []string) error {
	count, _ := strconv.Atoi(args[1]) // err already checked in args validation
	count++                           // single coordinator
	size := strings.ToLower(args[2])

	name, err := cmd.Flags().GetString("name")
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

	client, err := client.NewFromDefault(cmd.Context())
	if err != nil {
		return err
	}
	return createAWS(cmd, client, fileHandler, config, size, name, count)
}

// createAWS uses the given client to create 'count' instances of 'size'.
// After the instances are running, they are tagged with the default tags.
// On success, the state of the client is saved to the state file.
func createAWS(cmd *cobra.Command, cl ec2client, fileHandler file.Handler, config *config.Config, size, name string, count int) (retErr error) {
	if err := checkDirClean(fileHandler, config); err != nil {
		return err
	}

	const maxLength = 255
	if len(name) > maxLength {
		return fmt.Errorf("name for constellation too long, maximum length is %d: %s", maxLength, name)
	}
	ec2Tags := append([]ec2.Tag{}, *config.Provider.EC2.Tags...)
	ec2Tags = append(ec2Tags, ec2.Tag{Key: "Name", Value: name})

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

	defer rollbackOnError(context.Background(), cmd.OutOrStdout(), &retErr, &rollbackerAWS{client: cl})
	if err := cl.CreateSecurityGroup(cmd.Context(), *config.Provider.EC2.SecurityGroupInput); err != nil {
		return err
	}

	createInput := client.CreateInput{
		ImageId:      *config.Provider.EC2.Image,
		InstanceType: size,
		Count:        count,
		Tags:         ec2Tags,
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

// createAWSCompletion handels the completion of CLI arguments. It is frequently called
// while the user types arguments of the command to suggest completion.
func createAWSCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	case 1:
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	case 2:
		return []string{
			"4xlarge",
			"8xlarge",
			"12xlarge",
			"16xlarge",
			"24xlarge",
		}, cobra.ShellCompDirectiveDefault
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}
