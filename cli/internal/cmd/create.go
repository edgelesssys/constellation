/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

// createFlags contains the parsed flags of the create command.
type createFlags struct {
	rootFlags
	yes bool
}

// parse parses the flags of the create command.
func (f *createFlags) parse(flags *pflag.FlagSet) error {
	if err := f.rootFlags.parse(flags); err != nil {
		return err
	}

	yes, err := flags.GetBool("yes")
	if err != nil {
		return fmt.Errorf("getting 'yes' flag: %w", err)
	}
	f.yes = yes
	return nil
}

type createCmd struct {
	log   debugLog
	flags createFlags
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
	c := &createCmd{log: log}
	if err := c.flags.parse(cmd.Flags()); err != nil {
		return err
	}
	c.log.Debugf("Using flags: %+v", c.flags)

	applier, removeInstaller, err := cloudcmd.NewApplier(
		cmd.Context(),
		spinner,
		constants.TerraformWorkingDir,
		filepath.Join(constants.UpgradeDir, "create"), // Not used by create
		c.flags.tfLogLevel,
		fileHandler,
	)
	if err != nil {
		return err
	}
	defer removeInstaller()

	fetcher := attestationconfigapi.NewFetcher()
	return c.create(cmd, applier, fileHandler, spinner, fetcher)
}

func (c *createCmd) create(cmd *cobra.Command, applier cloudApplier, fileHandler file.Handler, spinner spinnerInterf, fetcher attestationconfigapi.Fetcher) (retErr error) {
	if err := c.checkDirClean(fileHandler); err != nil {
		return err
	}

	c.log.Debugf("Loading configuration file from %q", c.flags.pathPrefixer.PrefixPrintablePath(constants.ConfigFilename))
	conf, err := config.New(fileHandler, constants.ConfigFilename, fetcher, c.flags.force)
	c.log.Debugf("Configuration file loaded: %+v", conf)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}
	if !c.flags.force {
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

	if !c.flags.yes {
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
	if _, err := applier.Plan(cmd.Context(), conf); err != nil {
		return fmt.Errorf("planning infrastructure creation: %w", err)
	}
	infraState, err := applier.Apply(cmd.Context(), conf.GetProvider(), true)
	spinner.Stop()
	if err != nil {
		return err
	}
	c.log.Debugf("Successfully created the cloud resources for the cluster")

	stateFile, err := state.CreateOrRead(fileHandler, constants.StateFilename)
	if err != nil {
		return fmt.Errorf("reading state file: %w", err)
	}
	stateFile = stateFile.SetInfrastructure(infraState)
	if err := stateFile.WriteToFile(fileHandler, constants.StateFilename); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	cmd.Println("Your Constellation cluster was created successfully.")
	return nil
}

// checkDirClean checks if files of a previous Constellation are left in the current working dir.
func (c *createCmd) checkDirClean(fileHandler file.Handler) error {
	c.log.Debugf("Checking admin configuration file")
	if _, err := fileHandler.Stat(constants.AdminConfFilename); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf(
			"file '%s' already exists in working directory, run 'constellation terminate' before creating a new one",
			c.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename),
		)
	}
	c.log.Debugf("Checking master secrets file")
	if _, err := fileHandler.Stat(constants.MasterSecretFilename); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf(
			"file '%s' already exists in working directory. Constellation won't overwrite previous master secrets. Move it somewhere or delete it before creating a new cluster",
			c.flags.pathPrefixer.PrefixPrintablePath(constants.MasterSecretFilename),
		)
	}
	c.log.Debugf("Checking terraform working directory")
	if clean, err := fileHandler.IsEmpty(constants.TerraformWorkingDir); err != nil {
		return fmt.Errorf("checking if terraform working directory is empty: %w", err)
	} else if !clean {
		return fmt.Errorf(
			"directory '%s' already exists and is not empty, run 'constellation terminate' before creating a new one",
			c.flags.pathPrefixer.PrefixPrintablePath(constants.TerraformWorkingDir),
		)
	}

	return nil
}

func isPlural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
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
