/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi/fetcher"
	"github.com/manifoldco/promptui"
	"github.com/siderolabs/talos/pkg/machinery/config/encoder"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func newUpgradePlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Plan an upgrade of a Constellation cluster",
		Long:  "Plan an upgrade of a Constellation cluster by fetching compatible image versions and their measurements.",
		Args:  cobra.NoArgs,
		RunE:  runUpgradePlan,
	}

	cmd.Flags().StringP("file", "f", "", "path to output file, or '-' for stdout (omit for interactive mode)")

	return cmd
}

type upgradePlanCmd struct {
	log debugLog
}

func runUpgradePlan(cmd *cobra.Command, args []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	fileHandler := file.NewHandler(afero.NewOsFs())
	flags, err := parseUpgradePlanFlags(cmd)
	if err != nil {
		return err
	}
	planner, err := cloudcmd.NewUpgrader(cmd.OutOrStdout(), log)
	if err != nil {
		return err
	}
	versionListFetcher := fetcher.NewFetcher()
	rekor, err := sigstore.NewRekor()
	if err != nil {
		return fmt.Errorf("constructing Rekor client: %w", err)
	}
	cliVersion := getCurrentCLIVersion()
	up := &upgradePlanCmd{log: log}

	return up.upgradePlan(cmd, planner, versionListFetcher, fileHandler, http.DefaultClient, rekor, flags, cliVersion)
}

// upgradePlan plans an upgrade of a Constellation cluster.
func (up *upgradePlanCmd) upgradePlan(cmd *cobra.Command, planner upgradePlanner, verListFetcher versionListFetcher,
	fileHandler file.Handler, client *http.Client, rekor rekorVerifier, flags upgradePlanFlags,
	cliVersion string,
) error {
	conf, err := config.New(fileHandler, flags.configPath)
	if err != nil {
		return displayConfigValidationErrors(cmd.ErrOrStderr(), err)
	}
	up.log.Debugf("Read config from %s", flags.configPath)
	// get current image version of the cluster
	csp := conf.GetProvider()
	up.log.Debugf("Using provider %s", csp.String())

	version, err := getCurrentImageVersion(cmd.Context(), planner)
	if err != nil {
		return fmt.Errorf("checking current image version: %w", err)
	}
	up.log.Debugf("Using image version %s", version)

	// find compatible images
	// image updates should always be possible for the current minor version of the cluster
	// (e.g. 0.1.0 -> 0.1.1, 0.1.2, 0.1.3, etc.)
	// additionally, we allow updates to the next minor version (e.g. 0.1.0 -> 0.2.0)
	// if the CLI minor version is newer than the cluster minor version
	currentImageMinorVer := semver.MajorMinor(version)
	currentCLIMinorVer := semver.MajorMinor(cliVersion)
	nextImageMinorVer, err := nextMinorVersion(currentImageMinorVer)
	if err != nil {
		return fmt.Errorf("calculating next image minor version: %w", err)
	}
	up.log.Debugf("Current image minor version is %s", currentImageMinorVer)
	up.log.Debugf("Current CLI minor version is %s", currentCLIMinorVer)
	up.log.Debugf("Next image minor version is %s", nextImageMinorVer)
	var allowedMinorVersions []string

	cliImageCompare := semver.Compare(currentCLIMinorVer, currentImageMinorVer)

	switch {
	case cliImageCompare < 0:
		cmd.PrintErrln("Warning: CLI version is older than cluster image version. This is not supported.")
	case cliImageCompare == 0:
		allowedMinorVersions = []string{currentImageMinorVer}
	case cliImageCompare > 0:
		allowedMinorVersions = []string{currentImageMinorVer, nextImageMinorVer}
	}
	up.log.Debugf("Allowed minor versions are %#v", allowedMinorVersions)

	var updateCandidates []string
	for _, minorVer := range allowedMinorVersions {
		patchList := versionsapi.List{
			Ref:         versionsapi.ReleaseRef,
			Stream:      "stable",
			Base:        minorVer,
			Granularity: versionsapi.GranularityMinor,
			Kind:        versionsapi.VersionKindImage,
		}
		patchList, err = verListFetcher.FetchVersionList(cmd.Context(), patchList)
		if err == nil {
			updateCandidates = append(updateCandidates, patchList.Versions...)
		}
	}
	up.log.Debugf("Update candidates are %v", updateCandidates)

	// filter out versions that are not compatible with the current cluster
	compatibleImages := getCompatibleImages(version, updateCandidates)
	up.log.Debugf("Of those images, these ones are compaitble %v", compatibleImages)

	// get expected measurements for each image
	upgrades, err := getCompatibleImageMeasurements(cmd.Context(), cmd, client, rekor, []byte(flags.cosignPubKey), csp, compatibleImages)
	if err != nil {
		return fmt.Errorf("fetching measurements for compatible images: %w", err)
	}
	up.log.Debugf("Compatible image measurements are %v", upgrades)

	if len(upgrades) == 0 {
		cmd.PrintErrln("No compatible images found to upgrade to.")
		return nil
	}

	// interactive mode
	if flags.filePath == "" {
		up.log.Debugf("Writing upgrade plan in interactive mode")
		cmd.Printf("Current version: %s\n", version)
		return upgradePlanInteractive(
			&nopWriteCloser{cmd.OutOrStdout()},
			io.NopCloser(cmd.InOrStdin()),
			flags.configPath, conf, fileHandler,
			upgrades,
		)
	}

	// write upgrade plan to stdout
	if flags.filePath == "-" {
		up.log.Debugf("Writing upgrade plan to stdout")
		content, err := encoder.NewEncoder(upgrades).Encode()
		if err != nil {
			return fmt.Errorf("encoding compatible images: %w", err)
		}
		_, err = cmd.OutOrStdout().Write(content)
		return err
	}

	// write upgrade plan to file
	up.log.Debugf("Writing upgrade plan to file")
	return fileHandler.WriteYAML(flags.filePath, upgrades)
}

// getCompatibleImages trims the list of images to only ones compatible with the current cluster.
func getCompatibleImages(currentImageVersion string, images []string) []string {
	var compatibleImages []string

	for _, image := range images {
		// check if image is newer than current version
		if semver.Compare(image, currentImageVersion) <= 0 {
			continue
		}
		compatibleImages = append(compatibleImages, image)
	}
	return compatibleImages
}

// getCompatibleImageMeasurements retrieves the expected measurements for each image.
func getCompatibleImageMeasurements(ctx context.Context, cmd *cobra.Command, client *http.Client, rekor rekorVerifier, pubK []byte,
	csp cloudprovider.Provider, images []string,
) (map[string]config.UpgradeConfig, error) {
	upgrades := make(map[string]config.UpgradeConfig)
	for _, img := range images {
		measurementsURL, err := measurementURL(csp, img, "measurements.json")
		if err != nil {
			return nil, err
		}

		signatureURL, err := measurementURL(csp, img, "measurements.json.sig")
		if err != nil {
			return nil, err
		}

		var fetchedMeasurements measurements.M
		hash, err := fetchedMeasurements.FetchAndVerify(
			ctx, client,
			measurementsURL,
			signatureURL,
			pubK,
			measurements.WithMetadata{
				CSP:   csp,
				Image: img,
			},
		)
		if err != nil {
			cmd.PrintErrf("Skipping image %q: %s\n", img, err)
			continue
		}

		if err = verifyWithRekor(ctx, rekor, hash); err != nil {
			cmd.PrintErrf("Warning: Unable to verify '%s' in Rekor.\n", hash)
			cmd.PrintErrf("Make sure measurements are correct.\n")
		}

		upgrades[img] = config.UpgradeConfig{
			Image:        img,
			Measurements: fetchedMeasurements,
			CSP:          csp,
		}

	}

	return upgrades, nil
}

// getCurrentImageVersion retrieves the semantic version of the image currently installed in the cluster.
// If the cluster is not using a release image, an error is returned.
func getCurrentImageVersion(ctx context.Context, planner upgradePlanner) (string, error) {
	_, imageVersion, err := planner.GetCurrentImage(ctx)
	if err != nil {
		return "", err
	}

	if !semver.IsValid(imageVersion) {
		return "", fmt.Errorf("current image version is not a release image version: %q", imageVersion)
	}

	return imageVersion, nil
}

func getCurrentCLIVersion() string {
	return "v" + constants.VersionInfo
}

func parseUpgradePlanFlags(cmd *cobra.Command) (upgradePlanFlags, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return upgradePlanFlags{}, err
	}
	filePath, err := cmd.Flags().GetString("file")
	if err != nil {
		return upgradePlanFlags{}, err
	}

	return upgradePlanFlags{
		configPath:   configPath,
		filePath:     filePath,
		cosignPubKey: constants.CosignPublicKey,
	}, nil
}

func upgradePlanInteractive(out io.WriteCloser, in io.ReadCloser,
	configPath string, config *config.Config, fileHandler file.Handler,
	compatibleUpgrades map[string]config.UpgradeConfig,
) error {
	var imageVersions []string
	for k := range compatibleUpgrades {
		imageVersions = append(imageVersions, k)
	}
	semver.Sort(imageVersions)

	prompt := promptui.Select{
		Label: "Select an image version to upgrade to",
		Items: imageVersions,
		Searcher: func(input string, index int) bool {
			version := imageVersions[index]
			trimmedVersion := strings.TrimPrefix(strings.Replace(version, ".", "", -1), "v")
			input = strings.TrimPrefix(strings.Replace(input, ".", "", -1), "v")
			return strings.Contains(trimmedVersion, input)
		},
		Size:   10,
		Stdin:  in,
		Stdout: out,
	}

	_, res, err := prompt.Run()
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "Updating config to the following:")

	fmt.Fprintf(out, "Image: %s\n", compatibleUpgrades[res].Image)
	fmt.Fprintln(out, "Measurements:")
	content, err := encoder.NewEncoder(compatibleUpgrades[res].Measurements).Encode()
	if err != nil {
		return fmt.Errorf("encoding measurements: %w", err)
	}
	measurements := strings.TrimSuffix(strings.Replace("\t"+string(content), "\n", "\n\t", -1), "\n\t")
	fmt.Fprintln(out, measurements)

	config.Upgrade = compatibleUpgrades[res]
	return fileHandler.WriteYAML(configPath, config, file.OptOverwrite)
}

func nextMinorVersion(version string) (string, error) {
	major, minor, _, err := parseCanonicalSemver(version)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("v%d.%d", major, minor+1), nil
}

func parseCanonicalSemver(version string) (major int, minor int, patch int, err error) {
	version = semver.Canonical(version) // ensure version is in canonical form (vX.Y.Z)
	num, err := fmt.Sscanf(version, "v%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parsing version: %w", err)
	}
	if num != 3 {
		return 0, 0, 0, fmt.Errorf("parsing version: expected 3 numbers, got %d", num)
	}

	return major, minor, patch, nil
}

type upgradePlanFlags struct {
	configPath   string
	filePath     string
	cosignPubKey string
}

type nopWriteCloser struct {
	io.Writer
}

func (c *nopWriteCloser) Close() error { return nil }

type upgradePlanner interface {
	GetCurrentImage(ctx context.Context) (*unstructured.Unstructured, string, error)
}

type versionListFetcher interface {
	FetchVersionList(ctx context.Context, list versionsapi.List) (versionsapi.List, error)
}
