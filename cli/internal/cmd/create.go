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
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/cmd/pathprefix"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// NewCreateCmd returns a new cobra.Command for the create command.
func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create instances on a cloud platform for your Constellation cluster",
		Long:  "Create instances on a cloud platform for your Constellation cluster.",
		Args:  cobra.ExactArgs(0),
		RunE:  runCreate,
	}
	cmd.Flags().BoolP("yes", "y", false, "create the cluster without further confirmation")
	return cmd
}

type createCmd struct {
	log debugLog
	pf  pathprefix.PathPrefixer
}

func runCreate(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	spinner, err := newSpinnerOrStderr(cmd)
	if err != nil {
		return fmt.Errorf("creating spinner: %w", err)
	}
	defer spinner.Stop()

	fileHandler := file.NewHandler(afero.NewOsFs())
	creator := cloudcmd.NewCreator(spinner)
	c := &createCmd{log: log}
	fetcher := attestationconfigapi.NewFetcher()
	return c.create(cmd, creator, fileHandler, spinner, fetcher)
}

func (c *createCmd) create(cmd *cobra.Command, creator cloudCreator, fileHandler file.Handler, spinner spinnerInterf, fetcher attestationconfigapi.Fetcher) (retErr error) {
	flags, err := c.parseCreateFlags(cmd)
	if err != nil {
		return err
	}
	c.log.Debugf("Using flags: %+v", flags)
	if err := c.checkDirClean(fileHandler); err != nil {
		return err
	}

	c.log.Debugf("Loading configuration file from %q", c.pf.PrefixPrintablePath(constants.ConfigFilename))
	conf, err := config.New(fileHandler, constants.ConfigFilename, fetcher, flags.force)
	c.log.Debugf("Configuration file loaded: %+v", conf)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}
	if !flags.force {
		if err := validateCLIandConstellationVersionAreEqual(constants.BinaryVersion(), conf.Image, conf.MicroserviceVersion); err != nil {
			return err
		}
	}

	c.log.Debugf("Checking configuration for warnings")
	var printedAWarning bool
	if !conf.IsReleaseImage() {
		cmd.PrintErrln("Configured image doesn't look like a released production image. Double check image before deploying to production.")
		printedAWarning = true
	}

	if conf.IsNamedLikeDebugImage() && !conf.IsDebugCluster() {
		cmd.PrintErrln("WARNING: A debug image is used but debugCluster is false.")
		printedAWarning = true
	}

	if conf.IsDebugCluster() {
		cmd.PrintErrln("WARNING: Creating a debug cluster. This cluster is not secure and should only be used for debugging purposes.")
		cmd.PrintErrln("DO NOT USE THIS CLUSTER IN PRODUCTION.")
		printedAWarning = true
	}

	if conf.GetAttestationConfig().GetVariant().Equal(variant.AzureTrustedLaunch{}) {
		cmd.PrintErrln("Disabling Confidential VMs is insecure. Use only for evaluation purposes.")
		printedAWarning = true
	}

	// Print an extra new line later to separate warnings from the prompt message of the create command
	if printedAWarning {
		cmd.PrintErrln("")
	}

	provider := conf.GetProvider()

	controlPlaneGroup, ok := conf.NodeGroups[constants.DefaultControlPlaneGroupName]
	if !ok {
		return fmt.Errorf("default control-plane node group %q not found in configuration", constants.DefaultControlPlaneGroupName)
	}
	workerGroup, ok := conf.NodeGroups[constants.DefaultWorkerGroupName]
	if !ok {
		return fmt.Errorf("default worker node group %q not found in configuration", constants.DefaultWorkerGroupName)
	}
	otherGroupNames := make([]string, 0, len(conf.NodeGroups)-2)
	for groupName := range conf.NodeGroups {
		if groupName != constants.DefaultControlPlaneGroupName && groupName != constants.DefaultWorkerGroupName {
			otherGroupNames = append(otherGroupNames, groupName)
		}
	}
	if len(otherGroupNames) > 0 {
		c.log.Debugf("Creating %d additional node groups: %v", len(otherGroupNames), otherGroupNames)
	}

	if !flags.yes {
		// Ask user to confirm action.
		cmd.Printf("The following Constellation cluster will be created:\n")
		cmd.Printf("  %d control-plane node%s of type %s will be created.\n", controlPlaneGroup.InitialCount, isPlural(controlPlaneGroup.InitialCount), controlPlaneGroup.InstanceType)
		cmd.Printf("  %d worker node%s of type %s will be created.\n", workerGroup.InitialCount, isPlural(workerGroup.InitialCount), workerGroup.InstanceType)
		for _, groupName := range otherGroupNames {
			group := conf.NodeGroups[groupName]
			cmd.Printf("  group %s with %d node%s of type %s will be created.\n", groupName, group.InitialCount, isPlural(group.InitialCount), group.InstanceType)
		}
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
	opts := cloudcmd.CreateOptions{
		Provider:    provider,
		Config:      conf,
		TFLogLevel:  flags.tfLogLevel,
		TFWorkspace: constants.TerraformWorkingDir,
	}
	infraState, err := creator.Create(cmd.Context(), opts)
	spinner.Stop()
	if err != nil {
		return translateCreateErrors(cmd, c.pf, err)
	}
	c.log.Debugf("Successfully created the cloud resources for the cluster")

	// TODO(msanft): Remove IDFile as per AB#3425
	idFile := convertToIDFile(infraState, provider)
	if err := fileHandler.WriteJSON(constants.ClusterIDsFilename, idFile, file.OptNone); err != nil {
		return err
	}

	state := state.New().SetInfrastructure(infraState)
	if err := state.WriteToFile(fileHandler, constants.StateFilename); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	cmd.Println("Your Constellation cluster was created successfully.")
	return nil
}

func convertToIDFile(infra state.Infrastructure, provider cloudprovider.Provider) clusterid.File {
	var file clusterid.File
	file.CloudProvider = provider
	file.IP = infra.ClusterEndpoint
	file.APIServerCertSANs = infra.APIServerCertSANs
	file.InitSecret = []byte(infra.InitSecret) // Convert string to []byte
	file.UID = infra.UID

	if infra.Azure != nil {
		file.AttestationURL = infra.Azure.AttestationURL
	}

	return file
}

// parseCreateFlags parses the flags of the create command.
func (c *createCmd) parseCreateFlags(cmd *cobra.Command) (createFlags, error) {
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return createFlags{}, fmt.Errorf("parsing yes bool: %w", err)
	}
	c.log.Debugf("Yes flag is %t", yes)

	workDir, err := cmd.Flags().GetString("workspace")
	if err != nil {
		return createFlags{}, fmt.Errorf("parsing config path argument: %w", err)
	}
	c.log.Debugf("Workspace set to %q", workDir)
	c.pf = pathprefix.New(workDir)

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return createFlags{}, fmt.Errorf("parsing force argument: %w", err)
	}
	c.log.Debugf("force flag is %t", force)

	logLevelString, err := cmd.Flags().GetString("tf-log")
	if err != nil {
		return createFlags{}, fmt.Errorf("parsing tf-log string: %w", err)
	}
	logLevel, err := terraform.ParseLogLevel(logLevelString)
	if err != nil {
		return createFlags{}, fmt.Errorf("parsing Terraform log level %s: %w", logLevelString, err)
	}
	c.log.Debugf("Terraform logs will be written into %s at level %s", c.pf.PrefixPrintablePath(constants.TerraformLogFile), logLevel.String())

	return createFlags{
		tfLogLevel: logLevel,
		force:      force,
		yes:        yes,
	}, nil
}

// createFlags contains the parsed flags of the create command.
type createFlags struct {
	tfLogLevel terraform.LogLevel
	force      bool
	yes        bool
}

// checkDirClean checks if files of a previous Constellation are left in the current working dir.
func (c *createCmd) checkDirClean(fileHandler file.Handler) error {
	c.log.Debugf("Checking admin configuration file")
	if _, err := fileHandler.Stat(constants.AdminConfFilename); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("file '%s' already exists in working directory, run 'constellation terminate' before creating a new one", c.pf.PrefixPrintablePath(constants.AdminConfFilename))
	}
	c.log.Debugf("Checking master secrets file")
	if _, err := fileHandler.Stat(constants.MasterSecretFilename); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("file '%s' already exists in working directory. Constellation won't overwrite previous master secrets. Move it somewhere or delete it before creating a new cluster", c.pf.PrefixPrintablePath(constants.MasterSecretFilename))
	}
	c.log.Debugf("Checking cluster IDs file")
	if _, err := fileHandler.Stat(constants.ClusterIDsFilename); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("file '%s' already exists in working directory. Constellation won't overwrite previous cluster IDs. Move it somewhere or delete it before creating a new cluster", c.pf.PrefixPrintablePath(constants.ClusterIDsFilename))
	}

	return nil
}

func translateCreateErrors(cmd *cobra.Command, pf pathprefix.PathPrefixer, err error) error {
	switch {
	case errors.Is(err, terraform.ErrTerraformWorkspaceDifferentFiles):
		cmd.PrintErrln("\nYour current working directory contains an existing Terraform workspace which does not match the expected state.")
		cmd.PrintErrln("This can be due to a mix up between providers, versions or an otherwise corrupted workspace.")
		cmd.PrintErrln("Before creating a new cluster, try \"constellation terminate\".")
		cmd.PrintErrf("If this does not work, either move or delete the directory %q.\n", pf.PrefixPrintablePath(constants.TerraformWorkingDir))
		cmd.PrintErrln("Please only delete the directory if you made sure that all created cloud resources have been terminated.")
		return err
	case errors.Is(err, terraform.ErrTerraformWorkspaceExistsWithDifferentVariables):
		cmd.PrintErrln("\nYour current working directory contains an existing Terraform workspace which was initiated with different input variables.")
		cmd.PrintErrln("This can be the case if you have tried to create a cluster before with different options which did not complete, or the workspace is corrupted.")
		cmd.PrintErrln("Before creating a new cluster, try \"constellation terminate\".")
		cmd.PrintErrf("If this does not work, either move or delete the directory %q.\n", pf.PrefixPrintablePath(constants.TerraformWorkingDir))
		cmd.PrintErrln("Please only delete the directory if you made sure that all created cloud resources have been terminated.")
		return err
	default:
		return err
	}
}

func isPlural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// validateCLIandConstellationVersionAreEqual checks if the image and microservice version are equal (down to patch level) to the CLI version.
func validateCLIandConstellationVersionAreEqual(cliVersion semver.Semver, imageVersion string, microserviceVersion semver.Semver) error {
	parsedImageVersion, err := versionsapi.NewVersionFromShortPath(imageVersion, versionsapi.VersionKindImage)
	if err != nil {
		return fmt.Errorf("parsing image version: %w", err)
	}

	semImage, err := semver.New(parsedImageVersion.Version())
	if err != nil {
		return fmt.Errorf("parsing image semantical version: %w", err)
	}

	if !cliVersion.MajorMinorEqual(semImage) {
		return fmt.Errorf("image version %q does not match the major and minor version of the cli version %q", semImage.String(), cliVersion.String())
	}
	if cliVersion.Compare(microserviceVersion) != 0 {
		return fmt.Errorf("cli version %q does not match microservice version %q", cliVersion.String(), microserviceVersion.String())
	}
	return nil
}
