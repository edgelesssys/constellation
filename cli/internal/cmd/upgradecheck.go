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

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi/fetcher"
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

	cmd.Flags().BoolP("write-config", "w", false, "update the specified config file with the suggested versions")
	cmd.Flags().String("ref", versionsapi.ReleaseRef, "the reference to use for querying new versions")
	cmd.Flags().String("stream", "stable", "the stream to use for querying new versions")

	return cmd
}

func runUpgradeCheck(cmd *cobra.Command, args []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	fileHandler := file.NewHandler(afero.NewOsFs())
	flags, err := parseUpgradeCheckFlags(cmd)
	if err != nil {
		return err
	}
	checker, err := cloudcmd.NewUpgrader(cmd.OutOrStdout(), log)
	if err != nil {
		return err
	}
	versionListFetcher := fetcher.NewFetcher()
	rekor, err := sigstore.NewRekor()
	if err != nil {
		return fmt.Errorf("constructing Rekor client: %w", err)
	}
	up := &upgradeCheckCmd{
		collect: &versionCollector{
			writer:         cmd.OutOrStderr(),
			checker:        checker,
			verListFetcher: versionListFetcher,
			fileHandler:    fileHandler,
			client:         http.DefaultClient,
			rekor:          rekor,
			flags:          flags,
			cliVersion:     compatibility.EnsurePrefixV(constants.VersionInfo),
			log:            log,
		},
		log: log,
	}

	return up.upgradeCheck(cmd, fileHandler, flags)
}

func parseUpgradeCheckFlags(cmd *cobra.Command) (upgradeCheckFlags, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return upgradeCheckFlags{}, err
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return upgradeCheckFlags{}, err
	}
	writeConfig, err := cmd.Flags().GetBool("write-config")
	if err != nil {
		return upgradeCheckFlags{}, err
	}
	ref, err := cmd.Flags().GetString("ref")
	if err != nil {
		return upgradeCheckFlags{}, err
	}
	stream, err := cmd.Flags().GetString("stream")
	if err != nil {
		return upgradeCheckFlags{}, err
	}
	return upgradeCheckFlags{
		configPath:   configPath,
		force:        force,
		writeConfig:  writeConfig,
		ref:          ref,
		stream:       stream,
		cosignPubKey: constants.CosignPublicKey,
	}, nil
}

type upgradeCheckCmd struct {
	collect collector
	log     debugLog
}

// upgradePlan plans an upgrade of a Constellation cluster.
func (u *upgradeCheckCmd) upgradeCheck(cmd *cobra.Command, fileHandler file.Handler, flags upgradeCheckFlags) error {
	conf, err := config.New(fileHandler, flags.configPath, flags.force)
	if err != nil {
		return config.DisplayValidationErrors(cmd.ErrOrStderr(), err)
	}
	u.log.Debugf("Read configuration from %q", flags.configPath)
	// get current image version of the cluster
	csp := conf.GetProvider()
	u.log.Debugf("Using provider %s", csp.String())

	currentServices, currentImage, currentK8s, err := u.collect.currentVersions(cmd.Context())
	if err != nil {
		return err
	}

	supportedServices, supportedImages, supportedK8s, err := u.collect.supportedVersions(cmd.Context(), currentImage, csp)
	if err != nil {
		return err
	}
	u.log.Debugf("Current service version: %s", currentServices)
	u.log.Debugf("Supported service version: %s", supportedServices)
	u.log.Debugf("Current k8s version: %s", currentK8s)
	u.log.Debugf("Supported k8s version: %s", supportedK8s)

	// Filter versions to only include upgrades
	newServices := supportedServices
	if err := compatibility.IsValidUpgrade(currentServices, supportedServices); err != nil {
		newServices = ""
	}

	newKubernetes := filterK8sUpgrades(currentK8s, supportedK8s)
	sort.Strings(newKubernetes)

	supportedImages = filterImageUpgrades(currentImage, supportedImages)
	newImages, err := u.collect.newMeasurementes(cmd.Context(), csp, supportedImages)
	if err != nil {
		return err
	}

	upgrade := versionUpgrade{
		newServices:       newServices,
		newImages:         newImages,
		newKubernetes:     newKubernetes,
		currentServices:   currentServices,
		currentImage:      currentImage,
		currentKubernetes: currentK8s,
	}

	updateMsg, err := upgrade.buildString()
	if err != nil {
		return err
	}
	// Using Print over Println as buildString already includes a trailing newline where necessary.
	cmd.Print(updateMsg)

	if flags.writeConfig {
		if err := upgrade.writeConfig(conf, fileHandler, flags.configPath); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
		cmd.Println("Wrote config successfully.")
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
	currentVersions(ctx context.Context) (serviceVersions string, imageVersion string, k8sVersion string, err error)
	supportedVersions(ctx context.Context, version string, csp cloudprovider.Provider) (serviceVersions string, imageVersions []versionsapi.Version, k8sVersions []string, err error)
	newImages(ctx context.Context, version string, csp cloudprovider.Provider) ([]versionsapi.Version, error)
	newMeasurementes(ctx context.Context, csp cloudprovider.Provider, images []versionsapi.Version) (map[string]measurements.M, error)
	newerVersions(ctx context.Context, currentVersion string, allowedVersions []string) ([]versionsapi.Version, error)
}

type versionCollector struct {
	writer         io.Writer
	checker        upgradeChecker
	verListFetcher versionListFetcher
	fileHandler    file.Handler
	client         *http.Client
	rekor          rekorVerifier
	flags          upgradeCheckFlags
	cliVersion     string
	log            debugLog
}

func (v *versionCollector) newMeasurementes(ctx context.Context, csp cloudprovider.Provider, images []versionsapi.Version) (map[string]measurements.M, error) {
	// get expected measurements for each image
	upgrades, err := getCompatibleImageMeasurements(ctx, v.writer, v.client, v.rekor, []byte(v.flags.cosignPubKey), csp, images, v.log)
	if err != nil {
		return nil, fmt.Errorf("fetching measurements for compatible images: %w", err)
	}
	v.log.Debugf("Compatible image measurements are %v", upgrades)

	return upgrades, nil
}

func (v *versionCollector) currentVersions(ctx context.Context) (serviceVersion string, imageVersion string, k8sVersion string, err error) {
	helmClient, err := helm.NewClient(kubectl.New(), constants.AdminConfFilename, constants.HelmNamespace, v.log)
	if err != nil {
		return "", "", "", fmt.Errorf("setting up helm client: %w", err)
	}

	serviceVersion, err = helmClient.Versions()
	if err != nil {
		return "", "", "", fmt.Errorf("getting service versions: %w", err)
	}

	imageVersion, err = getCurrentImageVersion(ctx, v.checker)
	if err != nil {
		return "", "", "", fmt.Errorf("getting image version: %w", err)
	}

	k8sVersion, err = getCurrentKubernetesVersion(ctx, v.checker)
	if err != nil {
		return "", "", "", fmt.Errorf("getting image version: %w", err)
	}

	return serviceVersion, imageVersion, k8sVersion, nil
}

// supportedVersions returns slices of supported versions.
func (v *versionCollector) supportedVersions(ctx context.Context, version string, csp cloudprovider.Provider) (serviceVersion string, imageVersions []versionsapi.Version, k8sVersions []string, err error) {
	k8sVersions = versions.SupportedK8sVersions()
	serviceVersion, err = helm.AvailableServiceVersions()
	if err != nil {
		return "", nil, nil, fmt.Errorf("loading service versions: %w", err)
	}
	imageVersions, err = v.newImages(ctx, version, csp)
	if err != nil {
		return "", nil, nil, fmt.Errorf("loading image versions: %w", err)
	}

	return serviceVersion, imageVersions, k8sVersions, nil
}

func (v *versionCollector) newImages(ctx context.Context, version string, csp cloudprovider.Provider) ([]versionsapi.Version, error) {
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

	newerImages, err := v.newerVersions(ctx, currentImageMinorVer, allowedMinorVersions)
	if err != nil {
		return nil, fmt.Errorf("newer versions: %w", err)
	}

	return newerImages, nil
}

func (v *versionCollector) newerVersions(ctx context.Context, currentVersion string, allowedVersions []string) ([]versionsapi.Version, error) {
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
	currentServices   string
	currentImage      string
	currentKubernetes string
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

	result.WriteString("No upgrades available with this CLI.\nNewer versions may be available at: https://github.com/edgelesssys/constellation/releases\n")

	return result.String(), nil
}

func (v *versionUpgrade) writeConfig(conf *config.Config, fileHandler file.Handler, configPath string) error {
	// can't sort image map because maps are unsorted. services is only one string, k8s versions are sorted.

	if v.newServices != "" {
		conf.MicroserviceVersion = v.newServices
	}
	if len(v.newServices) > 0 {
		conf.KubernetesVersion = v.newKubernetes[0]
	}
	if len(v.newImages) > 0 {
		imageUpgrade := sortedMapKeys(v.newImages)[0]
		conf.Image = imageUpgrade
		conf.UpdateMeasurements(v.newImages[imageUpgrade])
	}

	if err := fileHandler.WriteYAML(configPath, conf, file.OptOverwrite); err != nil {
		return err
	}

	return nil
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
func getCompatibleImageMeasurements(ctx context.Context, writer io.Writer, client *http.Client, rekor rekorVerifier, pubK []byte,
	csp cloudprovider.Provider, versions []versionsapi.Version, log debugLog,
) (map[string]measurements.M, error) {
	upgrades := make(map[string]measurements.M)
	for _, version := range versions {
		log.Debugf("Fetching measurements for image: %s", version)
		shortPath := version.ShortPath()
		measurementsURL, signatureURL, err := version.MeasurementURL(csp)
		if err != nil {
			return nil, err
		}

		var fetchedMeasurements measurements.M
		log.Debugf("Fetching for measurement url: %s", measurementsURL)
		hash, err := fetchedMeasurements.FetchAndVerify(
			ctx, client,
			measurementsURL,
			signatureURL,
			pubK,
			measurements.WithMetadata{
				CSP:   csp,
				Image: shortPath,
			},
		)
		if err != nil {
			if _, err := fmt.Fprintf(writer, "Skipping compatible image %q: %s\n", shortPath, err); err != nil {
				return nil, fmt.Errorf("writing to buffer: %w", err)
			}
			continue
		}

		if err = verifyWithRekor(ctx, rekor, hash); err != nil {
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

type upgradeCheckFlags struct {
	configPath   string
	force        bool
	writeConfig  bool
	ref          string
	stream       string
	cosignPubKey string
}

type upgradeChecker interface {
	CurrentImage(ctx context.Context) (string, error)
	CurrentKubernetesVersion(ctx context.Context) (string, error)
}

type versionListFetcher interface {
	FetchVersionList(ctx context.Context, list versionsapi.List) (versionsapi.List, error)
}
