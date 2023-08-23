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
	"path/filepath"
	"sort"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/featureset"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubecmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
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
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
	consemver "github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/internal/sigstore/keyselect"
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
	upgradeID := generateUpgradeID(upgradeCmdKindCheck)

	upgradeDir := filepath.Join(constants.UpgradeDir, upgradeID)
	tfClient, err := cloudcmd.NewClusterUpgrader(
		cmd.Context(),
		constants.TerraformWorkingDir,
		upgradeDir,
		flags.terraformLogLevel,
		fileHandler,
	)
	if err != nil {
		return fmt.Errorf("setting up Terraform upgrader: %w", err)
	}

	kubeChecker, err := kubecmd.New(cmd.OutOrStdout(), constants.AdminConfFilename, log)
	if err != nil {
		return fmt.Errorf("setting up Kubernetes upgrader: %w", err)
	}

	versionfetcher := versionsapi.NewFetcher()
	rekor, err := sigstore.NewRekor()
	if err != nil {
		return fmt.Errorf("constructing Rekor client: %w", err)
	}
	up := &upgradeCheckCmd{
		canUpgradeCheck: featureset.CanUpgradeCheck,
		collect: &versionCollector{
			writer:         cmd.OutOrStderr(),
			kubeChecker:    kubeChecker,
			verListFetcher: versionfetcher,
			fileHandler:    fileHandler,
			client:         http.DefaultClient,
			rekor:          rekor,
			flags:          flags,
			cliVersion:     constants.BinaryVersion(),
			log:            log,
			versionsapi:    versionfetcher,
		},
		terraformChecker: tfClient,
		fileHandler:      fileHandler,
		log:              log,
	}

	return up.upgradeCheck(cmd, attestationconfigapi.NewFetcher(), upgradeDir, flags)
}

func parseUpgradeCheckFlags(cmd *cobra.Command) (upgradeCheckFlags, error) {
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
		force:             force,
		updateConfig:      updateConfig,
		ref:               ref,
		stream:            stream,
		terraformLogLevel: logLevel,
	}, nil
}

type upgradeCheckCmd struct {
	canUpgradeCheck  bool
	collect          collector
	terraformChecker terraformChecker
	fileHandler      file.Handler
	log              debugLog
}

// upgradePlan plans an upgrade of a Constellation cluster.
func (u *upgradeCheckCmd) upgradeCheck(cmd *cobra.Command, fetcher attestationconfigapi.Fetcher, upgradeDir string, flags upgradeCheckFlags) error {
	conf, err := config.New(u.fileHandler, constants.ConfigFilename, fetcher, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}

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
	if err := supported.service.IsUpgradeTo(current.service); err != nil {
		newServices = consemver.Semver{}
		u.log.Debugf("No valid service upgrades are available from %q to %q. The minor version can only drift by 1.\n", current.service.String(), supported.service.String())
	}

	newKubernetes := filterK8sUpgrades(current.k8s, supported.k8s)
	semver.Sort(newKubernetes)

	supported.image = filterImageUpgrades(current.image, supported.image)
	newImages, err := u.collect.newMeasurements(cmd.Context(), csp, attestationVariant, supported.image)
	if err != nil {
		return err
	}

	u.log.Debugf("Planning Terraform migrations")

	// Add manual migrations here if required
	//
	// var manualMigrations []terraform.StateMigration
	// for _, migration := range manualMigrations {
	// 	  u.log.Debugf("Adding manual Terraform migration: %s", migration.DisplayName)
	// 	  u.upgrader.AddManualStateMigration(migration)
	// }

	vars, err := cloudcmd.TerraformUpgradeVars(conf)
	if err != nil {
		return fmt.Errorf("parsing upgrade variables: %w", err)
	}
	u.log.Debugf("Using Terraform variables:\n%v", vars)

	cmd.Println("The following Terraform migrations are available with this CLI:")
	hasDiff, err := u.terraformChecker.PlanClusterUpgrade(cmd.Context(), cmd.OutOrStdout(), vars, conf.GetProvider())
	if err != nil {
		return fmt.Errorf("planning terraform migrations: %w", err)
	}
	defer func() {
		// Remove the upgrade directory
		if err := u.fileHandler.RemoveAll(upgradeDir); err != nil {
			u.log.Debugf("Failed to clean up Terraform migrations: %s", err)
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
		if err := upgrade.writeConfig(conf, u.fileHandler, constants.ConfigFilename); err != nil {
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
func filterImageUpgrades(currentVersion consemver.Semver, newVersions []versionsapi.Version) []versionsapi.Version {
	newImages := []versionsapi.Version{}
	for i := range newVersions {
		if err := compatibility.IsValidUpgrade(currentVersion.String(), newVersions[i].Version()); err != nil {
			continue
		}
		newImages = append(newImages, newVersions[i])
	}
	return newImages
}

// filterK8sUpgrades filters out K8s versions that are not valid upgrades.
func filterK8sUpgrades(currentVersion consemver.Semver, newVersions []string) []string {
	result := []string{}
	for i := range newVersions {
		if err := compatibility.IsValidUpgrade(currentVersion.String(), newVersions[i]); err != nil {
			continue
		}
		result = append(result, newVersions[i])
	}

	return result
}

type collector interface {
	currentVersions(ctx context.Context) (currentVersionInfo, error)
	supportedVersions(ctx context.Context, currentImageVersion, currentK8sVersion consemver.Semver) (supportedVersionInfo, error)
	newImages(ctx context.Context, currentImageVersion consemver.Semver) ([]versionsapi.Version, error)
	newMeasurements(ctx context.Context, csp cloudprovider.Provider, attestationVariant variant.Variant, images []versionsapi.Version) (map[string]measurements.M, error)
	newerVersions(ctx context.Context, allowedVersions []string) ([]versionsapi.Version, error)
	newCLIVersions(ctx context.Context) ([]consemver.Semver, error)
	filterCompatibleCLIVersions(ctx context.Context, cliPatchVersions []consemver.Semver, currentK8sVersion consemver.Semver) ([]consemver.Semver, error)
}

type versionCollector struct {
	writer         io.Writer
	kubeChecker    kubernetesChecker
	verListFetcher versionListFetcher
	fileHandler    file.Handler
	client         *http.Client
	rekor          rekorVerifier
	flags          upgradeCheckFlags
	versionsapi    versionFetcher
	cliVersion     consemver.Semver
	log            debugLog
}

func (v *versionCollector) newMeasurements(ctx context.Context, csp cloudprovider.Provider, attestationVariant variant.Variant, versions []versionsapi.Version) (map[string]measurements.M, error) {
	// get expected measurements for each image
	upgrades := make(map[string]measurements.M)
	for _, version := range versions {
		v.log.Debugf("Fetching measurements for image: %s", version)
		shortPath := version.ShortPath()

		publicKey, err := keyselect.CosignPublicKeyForVersion(version)
		if err != nil {
			return nil, fmt.Errorf("getting public key: %w", err)
		}
		cosign, err := sigstore.NewCosignVerifier(publicKey)
		if err != nil {
			return nil, fmt.Errorf("setting public key: %w", err)
		}

		measurements, err := getCompatibleImageMeasurements(ctx, v.writer, v.client, cosign, v.rekor, csp, attestationVariant, version, v.log)
		if err != nil {
			if _, err := fmt.Fprintf(v.writer, "Skipping compatible image %q: %s\n", shortPath, err); err != nil {
				return nil, fmt.Errorf("writing to buffer: %w", err)
			}
			continue
		}
		upgrades[shortPath] = measurements
	}
	v.log.Debugf("Compatible image measurements are %v", upgrades)

	return upgrades, nil
}

type currentVersionInfo struct {
	service consemver.Semver
	image   consemver.Semver
	k8s     consemver.Semver
	cli     consemver.Semver
}

func (v *versionCollector) currentVersions(ctx context.Context) (currentVersionInfo, error) {
	helmClient, err := helm.NewUpgradeClient(kubectl.NewUninitialized(), constants.AdminConfFilename, constants.HelmNamespace, v.log)
	if err != nil {
		return currentVersionInfo{}, fmt.Errorf("setting up helm client: %w", err)
	}

	serviceVersions, err := helmClient.Versions()
	if err != nil {
		return currentVersionInfo{}, fmt.Errorf("getting service versions: %w", err)
	}

	clusterVersions, err := v.kubeChecker.GetConstellationVersion(ctx)
	if err != nil {
		return currentVersionInfo{}, fmt.Errorf("getting cluster versions: %w", err)
	}
	imageVersion, err := consemver.New(clusterVersions.ImageVersion())
	if err != nil {
		return currentVersionInfo{}, fmt.Errorf("parsing image semantic version: %w", err)
	}
	k8sVersion, err := consemver.New(clusterVersions.KubernetesVersion())
	if err != nil {
		return currentVersionInfo{}, fmt.Errorf("parsing Kubernetes semantic version: %w", err)
	}

	return currentVersionInfo{
		service: serviceVersions.ConstellationServices(),
		image:   imageVersion,
		k8s:     k8sVersion,
		cli:     v.cliVersion,
	}, nil
}

type supportedVersionInfo struct {
	service consemver.Semver
	image   []versionsapi.Version
	k8s     []string
	// CLI versions including those incompatible with the current Kubernetes version.
	cli []consemver.Semver
	// CLI versions compatible with the current Kubernetes version.
	compatibleCLI []consemver.Semver
}

// supportedVersions returns slices of supported versions.
func (v *versionCollector) supportedVersions(ctx context.Context, currentImageVersion, currentK8sVersion consemver.Semver) (supportedVersionInfo, error) {
	k8sVersions := versions.SupportedK8sVersions()

	imageVersions, err := v.newImages(ctx, currentImageVersion)
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
		// Each CLI comes with a set of services that have the same version as the CLI.
		service:       constants.BinaryVersion(),
		image:         imageVersions,
		k8s:           k8sVersions,
		cli:           cliVersions,
		compatibleCLI: compatibleCLIVersions,
	}, nil
}

func (v *versionCollector) newImages(ctx context.Context, currentImageVersion consemver.Semver) ([]versionsapi.Version, error) {
	// find compatible images
	// image updates should always be possible for the current minor version of the cluster
	// (e.g. 0.1.0 -> 0.1.1, 0.1.2, 0.1.3, etc.)
	// additionally, we allow updates to the next minor version (e.g. 0.1.0 -> 0.2.0)
	// if the CLI minor version is newer than the cluster minor version
	currentImageMinorVer := semver.MajorMinor(currentImageVersion.String())
	currentCLIMinorVer := semver.MajorMinor(v.cliVersion.String())
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
	newServices       consemver.Semver
	newImages         map[string]measurements.M
	newKubernetes     []string
	newCLI            []consemver.Semver
	newCompatibleCLI  []consemver.Semver
	currentServices   consemver.Semver
	currentImage      consemver.Semver
	currentKubernetes consemver.Semver
	currentCLI        consemver.Semver
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

	if v.newServices != (consemver.Semver{}) {
		upgradeMsg.WriteString(fmt.Sprintf("  Services: %s --> %s\n", v.currentServices, v.newServices))
	}

	result := strings.Builder{}
	if upgradeMsg.Len() > 0 {
		result.WriteString("The following updates are available with this CLI:\n")
		result.WriteString(upgradeMsg.String())
		return result.String(), nil
	}

	// no upgrades available
	if v.newServices == (consemver.Semver{}) && len(v.newImages) == 0 {
		if len(v.newCompatibleCLI) > 0 {
			result.WriteString(fmt.Sprintf("Newer CLI versions that are compatible with your cluster are: %s\n", strings.Join(consemver.ToStrings(v.newCompatibleCLI), " ")))
			return result.String(), nil
		} else if len(v.newCLI) > 0 {
			result.WriteString(fmt.Sprintf("There are newer CLIs available (%s), however, you need to upgrade your cluster's Kubernetes version first.\n", strings.Join(consemver.ToStrings(v.newCLI), " ")))
			return result.String(), nil
		}
	}

	result.WriteString("You are up to date.\n")
	return result.String(), nil
}

func (v *versionUpgrade) writeConfig(conf *config.Config, fileHandler file.Handler, configPath string) error {
	// can't sort image map because maps are unsorted. services is only one string, k8s versions are sorted.

	if v.newServices != (consemver.Semver{}) {
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

// getCompatibleImageMeasurements retrieves the expected measurements for each image.
func getCompatibleImageMeasurements(ctx context.Context, writer io.Writer, client *http.Client, cosign sigstore.Verifier, rekor rekorVerifier,
	csp cloudprovider.Provider, attestationVariant variant.Variant, version versionsapi.Version, log debugLog,
) (measurements.M, error) {
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
		return nil, fmt.Errorf("fetching measurements: %w", err)
	}

	pubkey, err := keyselect.CosignPublicKeyForVersion(version)
	if err != nil {
		return nil, fmt.Errorf("getting public key: %w", err)
	}

	if err = sigstore.VerifyWithRekor(ctx, pubkey, rekor, hash); err != nil {
		if _, err := fmt.Fprintf(writer, "Warning: Unable to verify '%s' in Rekor.\nMake sure measurements are correct.\n", hash); err != nil {
			return nil, fmt.Errorf("writing to buffer: %w", err)
		}
	}

	return fetchedMeasurements, nil
}

type versionFetcher interface {
	FetchVersionList(ctx context.Context, list versionsapi.List) (versionsapi.List, error)
	FetchCLIInfo(ctx context.Context, cliInfo versionsapi.CLIInfo) (versionsapi.CLIInfo, error)
}

// newCLIVersions returns a list of versions of the CLI which are a valid upgrade.
func (v *versionCollector) newCLIVersions(ctx context.Context) ([]consemver.Semver, error) {
	list := versionsapi.List{
		Ref:         v.flags.ref,
		Stream:      v.flags.stream,
		Granularity: versionsapi.GranularityMajor,
		Base:        fmt.Sprintf("v%d", constants.BinaryVersion().Major()),
		Kind:        versionsapi.VersionKindCLI,
	}
	minorList, err := v.versionsapi.FetchVersionList(ctx, list)
	if err != nil {
		return nil, fmt.Errorf("listing major versions: %w", err)
	}

	var patchVersions []string
	for _, version := range minorList.Versions {
		target, err := consemver.New(version)
		if err != nil {
			return nil, fmt.Errorf("parsing version %s: %w", version, err)
		}
		if err := target.IsUpgradeTo(v.cliVersion); err != nil {
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

	out, err := consemver.NewSlice(patchVersions)
	if err != nil {
		return nil, fmt.Errorf("parsing versions: %w", err)
	}

	consemver.Sort(out)

	return out, nil
}

// filterCompatibleCLIVersions filters a list of CLI versions which are compatible with the current Kubernetes version.
func (v *versionCollector) filterCompatibleCLIVersions(ctx context.Context, cliPatchVersions []consemver.Semver, currentK8sVersion consemver.Semver,
) ([]consemver.Semver, error) {
	// filter out invalid upgrades and versions which are not compatible with the current Kubernetes version
	var compatibleVersions []consemver.Semver
	for _, version := range cliPatchVersions {
		if err := version.IsUpgradeTo(v.cliVersion); err != nil {
			v.log.Debugf("Skipping incompatible patch version %q: %s", version, err)
			continue
		}
		req := versionsapi.CLIInfo{
			Ref:     v.flags.ref,
			Stream:  v.flags.stream,
			Version: version.String(),
		}
		info, err := v.versionsapi.FetchCLIInfo(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("fetching CLI info: %w", err)
		}

		for _, k8sVersion := range info.Kubernetes {
			k8sVersionSem, err := consemver.New(k8sVersion)
			if err != nil {
				return nil, fmt.Errorf("parsing Kubernetes version %s: %w", k8sVersion, err)
			}
			if k8sVersionSem.Compare(currentK8sVersion) == 0 {
				compatibleVersions = append(compatibleVersions, version)
			}
		}
	}

	consemver.Sort(compatibleVersions)

	return compatibleVersions, nil
}

type upgradeCheckFlags struct {
	force             bool
	updateConfig      bool
	ref               string
	stream            string
	terraformLogLevel terraform.LogLevel
}

type kubernetesChecker interface {
	GetConstellationVersion(ctx context.Context) (kubecmd.NodeVersion, error)
}

type terraformChecker interface {
	PlanClusterUpgrade(ctx context.Context, outWriter io.Writer, vars terraform.Variables, csp cloudprovider.Provider) (bool, error)
}

type versionListFetcher interface {
	FetchVersionList(ctx context.Context, list versionsapi.List) (versionsapi.List, error)
}
