/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/featureset"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/cli/internal/upgrade"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/imagefetcher"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
	conSemver "github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/siderolabs/talos/pkg/machinery/config/encoder"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

func newUpgradeCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check for possible upgrades",
		Long:  "Check which upgrades can be applied to your Constellation Cluster.",
		Args:  cobra.NoArgs,
		RunE:  runUpgradeCheck,
	}

	cmd.Flags().BoolP("update-config", "u", false, "update the specified config file with the suggested versions")
	cmd.Flags().String("ref", versionsapi.ReleaseRef, "the reference to use for querying new versions")
	cmd.Flags().String("stream", "stable", "the stream to use for querying new versions")

	return cmd
}

func runUpgradeCheck(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	flags, err := parseUpgradeCheckFlags(cmd)
	if err != nil {
		return err
	}
	fileHandler := file.NewHandler(afero.NewOsFs())
	upgradeID := kubernetes.NewUpgradeID(kubernetes.UpgradeCmdKindCheck)
	checker, err := kubernetes.NewUpgrader(cmd.Context(), cmd.OutOrStdout(), fileHandler, log, upgradeID)
	if err != nil {
		return err
	}
	versionfetcher := versionsapi.NewFetcher()
	rekor, err := sigstore.NewRekor()
	if err != nil {
		return fmt.Errorf("constructing Rekor client: %w", err)
	}
	iamMigrateCmd, err := upgrade.NewIAMMigrateCmd(cmd.Context(), upgradeID.String(), cloudprovider.AWS, terraform.LogLevelDebug)
	if err != nil {
		return fmt.Errorf("setting up IAM migration command: %w", err)
	}
	up := &upgradeCheckCmd{
		canUpgradeCheck: featureset.CanUpgradeCheck,
		collect: &versionCollector{
			writer:         cmd.OutOrStderr(),
			checker:        checker,
			verListFetcher: versionfetcher,
			fileHandler:    fileHandler,
			client:         http.DefaultClient,
			cosign:         sigstore.CosignVerifier{},
			rekor:          rekor,
			flags:          flags,
			cliVersion:     compatibility.EnsurePrefixV(constants.VersionInfo()),
			log:            log,
			versionsapi:    versionfetcher,
		},
		checker:       checker,
		imagefetcher:  imagefetcher.New(),
		log:           log,
		iamMigrateCmd: iamMigrateCmd,
		planExecutor:  &migrationCmdExecutor{log},
	}

	return up.upgradeCheck(cmd, fileHandler, attestationconfigapi.NewFetcher(), flags)
}

func parseUpgradeCheckFlags(cmd *cobra.Command) (upgradeCheckFlags, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return upgradeCheckFlags{}, fmt.Errorf("parsing config string: %w", err)
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return upgradeCheckFlags{}, fmt.Errorf("parsing force bool: %w", err)
	}
	updateConfig, err := cmd.Flags().GetBool("update-config")
	if err != nil {
		return upgradeCheckFlags{}, fmt.Errorf("parsing update-config bool: %w", err)
	}
	ref, err := cmd.Flags().GetString("ref")
	if err != nil {
		return upgradeCheckFlags{}, fmt.Errorf("parsing ref string: %w", err)
	}
	stream, err := cmd.Flags().GetString("stream")
	if err != nil {
		return upgradeCheckFlags{}, fmt.Errorf("parsing stream string: %w", err)
	}

	logLevelString, err := cmd.Flags().GetString("tf-log")
	if err != nil {
		return upgradeCheckFlags{}, fmt.Errorf("parsing tf-log string: %w", err)
	}
	logLevel, err := terraform.ParseLogLevel(logLevelString)
	if err != nil {
		return upgradeCheckFlags{}, fmt.Errorf("parsing Terraform log level %s: %w", logLevelString, err)
	}

	return upgradeCheckFlags{
		configPath:        configPath,
		force:             force,
		updateConfig:      updateConfig,
		ref:               ref,
		stream:            stream,
		terraformLogLevel: logLevel,
	}, nil
}

type planExecutor interface {
	planMigration(cmd *cobra.Command, file file.Handler, migrateCmd upgrade.MigrationCmd) (hasDiff bool, err error)
}

type upgradeCheckCmd struct {
	canUpgradeCheck bool
	collect         collector
	checker         upgradeChecker
	imagefetcher    imageFetcher
	log             debugLog
	iamMigrateCmd   upgrade.MigrationCmd
	planExecutor    planExecutor
}

// upgradePlan plans an upgrade of a Constellation cluster.
func (u *upgradeCheckCmd) upgradeCheck(cmd *cobra.Command, fileHandler file.Handler, fetcher attestationconfigapi.Fetcher, flags upgradeCheckFlags) error {
	conf, err := config.New(fileHandler, flags.configPath, fetcher, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}
	u.log.Debugf("Read configuration from %q", flags.configPath)

	if !u.canUpgradeCheck {
		cmd.PrintErrln("Planning Constellation upgrades automatically is not supported in the OSS build of the Constellation CLI. Consult the documentation for instructions on where to download the enterprise version.")
		return errors.New("upgrade check is not supported")
	}

	// get current image version of the cluster
	csp := conf.GetProvider()
	attestationVariant := conf.GetAttestationConfig().GetVariant()
	u.log.Debugf("Using provider %s with attestation variant %s", csp.String(), attestationVariant.String())

	current, err := u.collect.currentVersions(cmd.Context())
	if err != nil {
		return err
	}

	supported, err := u.collect.supportedVersions(cmd.Context(), current.image, current.k8s)
	if err != nil {
		return err
	}
	u.log.Debugf("Current cli version: %s", current.cli)
	u.log.Debugf("Supported cli version(s): %s", supported.cli)
	u.log.Debugf("Current service version: %s", current.service)
	u.log.Debugf("Supported service version: %s", supported.service)
	u.log.Debugf("Current k8s version: %s", current.k8s)
	u.log.Debugf("Supported k8s version(s): %s", supported.k8s)

	// Filter versions to only include upgrades
	newServices := supported.service
	if err := compatibility.IsValidUpgrade(current.service, supported.service); err != nil {
		newServices = ""
	}

	newKubernetes := filterK8sUpgrades(current.k8s, supported.k8s)
	semver.Sort(newKubernetes)

	supported.image = filterImageUpgrades(current.image, supported.image)
	newImages, err := u.collect.newMeasurements(cmd.Context(), csp, attestationVariant, supported.image)
	if err != nil {
		return err
	}

	u.log.Debugf("Planning IAM migrations")
	if u.iamMigrateCmd != nil {
		hasIAMDiff, err := u.planExecutor.planMigration(cmd, fileHandler, u.iamMigrateCmd)
		if err != nil {
			return fmt.Errorf("planning IAM migration: %w", err)
		}
		if !hasIAMDiff {
			cmd.Println("  No IAM migrations are available.")
		}
	}

	u.log.Debugf("Planning Terraform migrations")
	if err := u.checker.CheckTerraformMigrations(); err != nil {
		return fmt.Errorf("checking workspace: %w", err)
	}

	// TODO(AB#3248): Remove this migration after we can assume that all existing clusters have been migrated.
	var awsZone string
	if csp == cloudprovider.AWS {
		awsZone = conf.Provider.AWS.Zone
	}
	manualMigrations := terraformMigrationAWSNodeGroups(csp, awsZone)
	for _, migration := range manualMigrations {
		u.log.Debugf("Adding manual Terraform migration: %s", migration.DisplayName)
		u.checker.AddManualStateMigration(migration)
	}

	vars, err := parseTerraformUpgradeVars(cmd, conf, u.imagefetcher)
	if err != nil {
		return fmt.Errorf("parsing upgrade variables: %w", err)
	}
	u.log.Debugf("Using Terraform variables:\n%v", vars)

	opts := upgrade.TerraformUpgradeOptions{
		LogLevel: flags.terraformLogLevel,
		CSP:      conf.GetProvider(),
		Vars:     vars,
	}

	cmd.Println("The following Terraform migrations are available with this CLI:")

	// Check if there are any Terraform migrations
	hasDiff, err := u.checker.PlanTerraformMigrations(cmd.Context(), opts)
	if err != nil {
		return fmt.Errorf("planning terraform migrations: %w", err)
	}
	defer func() {
		if err := u.checker.CleanUpTerraformMigrations(); err != nil {
			u.log.Debugf("Failed to clean up Terraform migrations: %v", err)
		}
	}()

	if !hasDiff {
		cmd.Println("  No Terraform migrations are available.")
	}

	upgrade := versionUpgrade{
		newServices:       newServices,
		newImages:         newImages,
		newKubernetes:     newKubernetes,
		newCLI:            supported.cli,
		newCompatibleCLI:  supported.compatibleCLI,
		currentServices:   current.service,
		currentImage:      current.image,
		currentKubernetes: current.k8s,
		currentCLI:        current.cli,
	}

	updateMsg, err := upgrade.buildString()
	if err != nil {
		return err
	}
	// Using Print over Println as buildString already includes a trailing newline where necessary.
	cmd.Print(updateMsg)

	if flags.updateConfig {
		if err := upgrade.writeConfig(conf, fileHandler, flags.configPath); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
		cmd.Println("Config updated successfully.")
	}

	return nil
}

func sortedMapKeys[T any](a map[string]T) []string {
	keys := []string{}
	for k := range a {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// filterImageUpgrades filters out image versions that are not valid upgrades.
func filterImageUpgrades(currentVersion string, newVersions []versionsapi.Version) []versionsapi.Version {
	newImages := []versionsapi.Version{}
	for i := range newVersions {
		if err := compatibility.IsValidUpgrade(currentVersion, newVersions[i].Version); err != nil {
			continue
		}
		newImages = append(newImages, newVersions[i])
	}
	return newImages
}

// filterK8sUpgrades filters out K8s versions that are not valid upgrades.
func filterK8sUpgrades(currentVersion string, newVersions []string) []string {
	result := []string{}
	for i := range newVersions {
		if err := compatibility.IsValidUpgrade(currentVersion, newVersions[i]); err != nil {
			continue
		}
		result = append(result, newVersions[i])
	}

	return result
}

type collector interface {
	currentVersions(ctx context.Context) (currentVersionInfo, error)
	supportedVersions(ctx context.Context, version, currentK8sVersion string) (supportedVersionInfo, error)
	newImages(ctx context.Context, version string) ([]versionsapi.Version, error)
	newMeasurements(ctx context.Context, csp cloudprovider.Provider, attestationVariant variant.Variant, images []versionsapi.Version) (map[string]measurements.M, error)
	newerVersions(ctx context.Context, allowedVersions []string) ([]versionsapi.Version, error)
	newCLIVersions(ctx context.Context) ([]string, error)
	filterCompatibleCLIVersions(ctx context.Context, cliPatchVersions []string, currentK8sVersion string) ([]string, error)
}

type versionCollector struct {
	writer         io.Writer
	checker        upgradeChecker
	verListFetcher versionListFetcher
	fileHandler    file.Handler
	client         *http.Client
	cosign         cosignVerifier
	rekor          rekorVerifier
	flags          upgradeCheckFlags
	versionsapi    versionFetcher
	cliVersion     string
	log            debugLog
}

func (v *versionCollector) newMeasurements(ctx context.Context, csp cloudprovider.Provider, attestationVariant variant.Variant, images []versionsapi.Version) (map[string]measurements.M, error) {
	// get expected measurements for each image
	upgrades, err := getCompatibleImageMeasurements(ctx, v.writer, v.client, v.cosign, v.rekor, csp, attestationVariant, images, v.log)
	if err != nil {
		return nil, fmt.Errorf("fetching measurements for compatible images: %w", err)
	}
	v.log.Debugf("Compatible image measurements are %v", upgrades)

	return upgrades, nil
}

type currentVersionInfo struct {
	service string
	image   string
	k8s     string
	cli     string
}

func (v *versionCollector) currentVersions(ctx context.Context) (currentVersionInfo, error) {
	helmClient, err := helm.NewClient(kubectl.New(), constants.AdminConfFilename, constants.HelmNamespace, v.log)
	if err != nil {
		return currentVersionInfo{}, fmt.Errorf("setting up helm client: %w", err)
	}

	serviceVersions, err := helmClient.Versions()
	if err != nil {
		return currentVersionInfo{}, fmt.Errorf("getting service versions: %w", err)
	}

	imageVersion, err := getCurrentImageVersion(ctx, v.checker)
	if err != nil {
		return currentVersionInfo{}, fmt.Errorf("getting image version: %w", err)
	}

	k8sVersion, err := getCurrentKubernetesVersion(ctx, v.checker)
	if err != nil {
		return currentVersionInfo{}, fmt.Errorf("getting Kubernetes version: %w", err)
	}

	return currentVersionInfo{
		service: serviceVersions.ConstellationServices(),
		image:   imageVersion,
		k8s:     k8sVersion,
		cli:     v.cliVersion,
	}, nil
}

type supportedVersionInfo struct {
	service string
	image   []versionsapi.Version
	k8s     []string
	// CLI versions including those incompatible with the current Kubernetes version.
	cli []string
	// CLI versions compatible with the current Kubernetes version.
	compatibleCLI []string
}

// supportedVersions returns slices of supported versions.
func (v *versionCollector) supportedVersions(ctx context.Context, version, currentK8sVersion string) (supportedVersionInfo, error) {
	k8sVersions := versions.SupportedK8sVersions()
	// Each CLI comes with a set of services that have the same version as the CLI.
	serviceVersion := compatibility.EnsurePrefixV(constants.VersionInfo())

	imageVersions, err := v.newImages(ctx, version)
	if err != nil {
		return supportedVersionInfo{}, fmt.Errorf("loading image versions: %w", err)
	}
	cliVersions, err := v.newCLIVersions(ctx)
	if err != nil {
		return supportedVersionInfo{}, fmt.Errorf("loading cli versions: %w", err)
	}
	compatibleCLIVersions, err := v.filterCompatibleCLIVersions(ctx, cliVersions, currentK8sVersion)
	if err != nil {
		return supportedVersionInfo{}, fmt.Errorf("filtering cli versions: %w", err)
	}

	return supportedVersionInfo{
		service:       serviceVersion,
		image:         imageVersions,
		k8s:           k8sVersions,
		cli:           cliVersions,
		compatibleCLI: compatibleCLIVersions,
	}, nil
}

func (v *versionCollector) newImages(ctx context.Context, version string) ([]versionsapi.Version, error) {
	// find compatible images
	// image updates should always be possible for the current minor version of the cluster
	// (e.g. 0.1.0 -> 0.1.1, 0.1.2, 0.1.3, etc.)
	// additionally, we allow updates to the next minor version (e.g. 0.1.0 -> 0.2.0)
	// if the CLI minor version is newer than the cluster minor version
	currentImageMinorVer := semver.MajorMinor(version)
	currentCLIMinorVer := semver.MajorMinor(v.cliVersion)
	nextImageMinorVer, err := compatibility.NextMinorVersion(currentImageMinorVer)
	if err != nil {
		return nil, fmt.Errorf("calculating next image minor version: %w", err)
	}
	v.log.Debugf("Current image minor version is %s", currentImageMinorVer)
	v.log.Debugf("Current CLI minor version is %s", currentCLIMinorVer)
	v.log.Debugf("Next image minor version is %s", nextImageMinorVer)

	allowedMinorVersions := []string{currentImageMinorVer, nextImageMinorVer}
	switch cliImageCompare := semver.Compare(currentCLIMinorVer, currentImageMinorVer); {
	case cliImageCompare < 0:
		if !v.flags.force {
			return nil, fmt.Errorf("cluster image version (%s) newer than CLI version (%s)", currentImageMinorVer, currentCLIMinorVer)
		}
		if _, err := fmt.Fprintln(v.writer, "WARNING: CLI version is older than cluster image version. Continuing due to force flag."); err != nil {
			return nil, fmt.Errorf("writing to buffer: %w", err)
		}
	case cliImageCompare == 0:
		allowedMinorVersions = []string{currentImageMinorVer}
	case cliImageCompare > 0:
		allowedMinorVersions = []string{currentImageMinorVer, nextImageMinorVer}
	}
	v.log.Debugf("Allowed minor versions are %#v", allowedMinorVersions)

	newerImages, err := v.newerVersions(ctx, allowedMinorVersions)
	if err != nil {
		return nil, fmt.Errorf("newer versions: %w", err)
	}

	return newerImages, nil
}

func (v *versionCollector) newerVersions(ctx context.Context, allowedVersions []string) ([]versionsapi.Version, error) {
	var updateCandidates []versionsapi.Version
	for _, minorVer := range allowedVersions {
		patchList := versionsapi.List{
			Ref:         v.flags.ref,
			Stream:      v.flags.stream,
			Base:        minorVer,
			Granularity: versionsapi.GranularityMinor,
			Kind:        versionsapi.VersionKindImage,
		}
		patchList, err := v.verListFetcher.FetchVersionList(ctx, patchList)
		var notFound *fetcher.NotFoundError
		if errors.As(err, &notFound) {
			v.log.Debugf("Skipping version: %s", err)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("fetching version list: %w", err)
		}
		updateCandidates = append(updateCandidates, patchList.StructuredVersions()...)
	}
	v.log.Debugf("Update candidates are %v", updateCandidates)

	return updateCandidates, nil
}

type versionUpgrade struct {
	newServices       string
	newImages         map[string]measurements.M
	newKubernetes     []string
	newCLI            []string
	newCompatibleCLI  []string
	currentServices   string
	currentImage      string
	currentKubernetes string
	currentCLI        string
}

func (v *versionUpgrade) buildString() (string, error) {
	upgradeMsg := strings.Builder{}

	if len(v.newKubernetes) > 0 {
		upgradeMsg.WriteString(fmt.Sprintf("  Kubernetes: %s --> %s\n", v.currentKubernetes, strings.Join(v.newKubernetes, " ")))
	}

	if len(v.newImages) > 0 {
		imageMsgs := strings.Builder{}
		newImagesSorted := sortedMapKeys(v.newImages)
		for i, image := range newImagesSorted {
			// prevent trailing newlines
			if i > 0 {
				imageMsgs.WriteString("\n")
			}
			content, err := encoder.NewEncoder(v.newImages[image]).Encode()
			contentFormated := strings.ReplaceAll(string(content), "\n", "\n      ")
			if err != nil {
				return "", fmt.Errorf("marshalling measurements: %w", err)
			}
			imageMsgs.WriteString(fmt.Sprintf("    %s --> %s\n      Includes these measurements:\n      %s", v.currentImage, image, contentFormated))
		}
		upgradeMsg.WriteString("  Images:\n")
		upgradeMsg.WriteString(imageMsgs.String())
		fmt.Fprintln(&upgradeMsg, "")
	}

	if v.newServices != "" {
		upgradeMsg.WriteString(fmt.Sprintf("  Services: %s --> %s\n", v.currentServices, v.newServices))
	}

	result := strings.Builder{}
	if upgradeMsg.Len() > 0 {
		result.WriteString("The following updates are available with this CLI:\n")
		result.WriteString(upgradeMsg.String())
		return result.String(), nil
	}

	// no upgrades available
	if v.newServices == "" && len(v.newImages) == 0 {
		if len(v.newCompatibleCLI) > 0 {
			result.WriteString(fmt.Sprintf("Newer CLI versions that are compatible with your cluster are: %s\n", strings.Join(v.newCompatibleCLI, " ")))
			return result.String(), nil
		} else if len(v.newCLI) > 0 {
			result.WriteString(fmt.Sprintf("There are newer CLIs available (%s), however, you need to upgrade your cluster's Kubernetes version first.\n", strings.Join(v.newCLI, " ")))
			return result.String(), nil
		}
	}

	result.WriteString("You are up to date.\n")
	return result.String(), nil
}

func (v *versionUpgrade) writeConfig(conf *config.Config, fileHandler file.Handler, configPath string) error {
	// can't sort image map because maps are unsorted. services is only one string, k8s versions are sorted.

	if v.newServices != "" {
		conf.MicroserviceVersion = v.newServices
	}
	if len(v.newKubernetes) > 0 {
		conf.KubernetesVersion = v.newKubernetes[0]
	}
	if len(v.newImages) > 0 {
		imageUpgrade := sortedMapKeys(v.newImages)[0]
		conf.Image = imageUpgrade
		conf.UpdateMeasurements(v.newImages[imageUpgrade])
	}

	return fileHandler.WriteYAML(configPath, conf, file.OptOverwrite)
}

// getCurrentImageVersion retrieves the semantic version of the image currently installed in the cluster.
// If the cluster is not using a release image, an error is returned.
func getCurrentImageVersion(ctx context.Context, checker upgradeChecker) (string, error) {
	imageVersion, err := checker.CurrentImage(ctx)
	if err != nil {
		return "", err
	}

	if !semver.IsValid(imageVersion) {
		return "", fmt.Errorf("current image version is not a release image version: %q", imageVersion)
	}

	return imageVersion, nil
}

// getCurrentKubernetesVersion retrieves the semantic version of Kubernetes currently installed in the cluster.
func getCurrentKubernetesVersion(ctx context.Context, checker upgradeChecker) (string, error) {
	k8sVersion, err := checker.CurrentKubernetesVersion(ctx)
	if err != nil {
		return "", err
	}

	if !semver.IsValid(k8sVersion) {
		return "", fmt.Errorf("current kubernetes version is not a valid semver string: %q", k8sVersion)
	}

	return k8sVersion, nil
}

// getCompatibleImageMeasurements retrieves the expected measurements for each image.
func getCompatibleImageMeasurements(ctx context.Context, writer io.Writer, client *http.Client, cosign cosignVerifier, rekor rekorVerifier,
	csp cloudprovider.Provider, attestationVariant variant.Variant, versions []versionsapi.Version, log debugLog,
) (map[string]measurements.M, error) {
	upgrades := make(map[string]measurements.M)
	for _, version := range versions {
		log.Debugf("Fetching measurements for image: %s", version)
		shortPath := version.ShortPath()
		measurementsURL, signatureURL, err := versionsapi.MeasurementURL(version)
		if err != nil {
			return nil, err
		}

		var fetchedMeasurements measurements.M
		log.Debugf("Fetching for measurement url: %s", measurementsURL)
		hash, err := fetchedMeasurements.FetchAndVerify(
			ctx, client, cosign,
			measurementsURL,
			signatureURL,
			version,
			csp,
			attestationVariant,
		)
		if err != nil {
			if _, err := fmt.Fprintf(writer, "Skipping compatible image %q: %s\n", shortPath, err); err != nil {
				return nil, fmt.Errorf("writing to buffer: %w", err)
			}
			continue
		}

		if err = sigstore.VerifyWithRekor(ctx, version, rekor, hash); err != nil {
			if _, err := fmt.Fprintf(writer, "Warning: Unable to verify '%s' in Rekor.\n", hash); err != nil {
				return nil, fmt.Errorf("writing to buffer: %w", err)
			}
			if _, err := fmt.Fprintf(writer, "Make sure measurements are correct.\n"); err != nil {
				return nil, fmt.Errorf("writing to buffer: %w", err)
			}
		}

		upgrades[shortPath] = fetchedMeasurements

	}

	return upgrades, nil
}

type versionFetcher interface {
	FetchVersionList(ctx context.Context, list versionsapi.List) (versionsapi.List, error)
	FetchCLIInfo(ctx context.Context, cliInfo versionsapi.CLIInfo) (versionsapi.CLIInfo, error)
}

// newCLIVersions returns a list of versions of the CLI which are a valid upgrade.
func (v *versionCollector) newCLIVersions(ctx context.Context) ([]string, error) {
	cliVersion, err := conSemver.New(constants.VersionInfo())
	if err != nil {
		return nil, fmt.Errorf("parsing current CLI version: %w", err)
	}
	list := versionsapi.List{
		Ref:         v.flags.ref,
		Stream:      v.flags.stream,
		Granularity: versionsapi.GranularityMajor,
		Base:        fmt.Sprintf("v%d", cliVersion.Major),
		Kind:        versionsapi.VersionKindCLI,
	}
	minorList, err := v.versionsapi.FetchVersionList(ctx, list)
	if err != nil {
		return nil, fmt.Errorf("listing major versions: %w", err)
	}

	var patchVersions []string
	for _, version := range minorList.Versions {
		if err := compatibility.IsValidUpgrade(v.cliVersion, version); err != nil {
			v.log.Debugf("Skipping incompatible minor version %q: %s", version, err)
			continue
		}
		list := versionsapi.List{
			Ref:         v.flags.ref,
			Stream:      v.flags.stream,
			Granularity: versionsapi.GranularityMinor,
			Base:        version,
			Kind:        versionsapi.VersionKindCLI,
		}
		patchList, err := v.versionsapi.FetchVersionList(ctx, list)
		if err != nil {
			return nil, fmt.Errorf("listing minor versions for major version %s: %w", version, err)
		}
		patchVersions = append(patchVersions, patchList.Versions...)
	}

	semver.Sort(patchVersions)

	return patchVersions, nil
}

// filterCompatibleCLIVersions filters a list of CLI versions which are compatible with the current Kubernetes version.
func (v *versionCollector) filterCompatibleCLIVersions(ctx context.Context, cliPatchVersions []string, currentK8sVersion string) ([]string, error) {
	// filter out invalid upgrades and versions which are not compatible with the current Kubernetes version
	var compatibleVersions []string
	for _, version := range cliPatchVersions {
		if err := compatibility.IsValidUpgrade(v.cliVersion, version); err != nil {
			v.log.Debugf("Skipping incompatible patch version %q: %s", version, err)
			continue
		}
		req := versionsapi.CLIInfo{
			Ref:     v.flags.ref,
			Stream:  v.flags.stream,
			Version: version,
		}
		info, err := v.versionsapi.FetchCLIInfo(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("fetching CLI info: %w", err)
		}

		for _, k8sVersion := range info.Kubernetes {
			if k8sVersion == currentK8sVersion {
				compatibleVersions = append(compatibleVersions, version)
			}
		}
	}

	semver.Sort(compatibleVersions)

	return compatibleVersions, nil
}

type upgradeCheckFlags struct {
	configPath        string
	force             bool
	updateConfig      bool
	ref               string
	stream            string
	terraformLogLevel terraform.LogLevel
}

type upgradeChecker interface {
	CurrentImage(ctx context.Context) (string, error)
	CurrentKubernetesVersion(ctx context.Context) (string, error)
	PlanTerraformMigrations(ctx context.Context, opts upgrade.TerraformUpgradeOptions) (bool, error)
	CheckTerraformMigrations() error
	CleanUpTerraformMigrations() error
	AddManualStateMigration(migration terraform.StateMigration)
}

type versionListFetcher interface {
	FetchVersionList(ctx context.Context, list versionsapi.List) (versionsapi.List, error)
}
