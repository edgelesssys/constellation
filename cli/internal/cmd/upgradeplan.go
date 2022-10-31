/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/manifoldco/promptui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/talos-systems/talos/pkg/machinery/config/encoder"
	"golang.org/x/mod/semver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const imageReleaseURL = "https://github.com/edgelesssys/constellation/releases/latest/download/versions-manifest.json"

var (
	azureCVMRxp = regexp.MustCompile(`^\/CommunityGalleries\/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df\/Images\/constellation\/Versions\/[\d]+.[\d]+.[\d]+$`)
	gcpCVMRxp   = regexp.MustCompile(`^projects\/constellation-images\/global\/images\/constellation-(v[\d]+-[\d]+-[\d]+)$`)
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

func runUpgradePlan(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	flags, err := parseUpgradePlanFlags(cmd)
	if err != nil {
		return err
	}
	planner, err := cloudcmd.NewUpgrader(cmd.OutOrStdout())
	if err != nil {
		return err
	}
	rekor, err := sigstore.NewRekor()
	if err != nil {
		return fmt.Errorf("constructing Rekor client: %w", err)
	}

	return upgradePlan(cmd, planner, fileHandler, http.DefaultClient, rekor, flags)
}

// upgradePlan plans an upgrade of a Constellation cluster.
func upgradePlan(cmd *cobra.Command, planner upgradePlanner,
	fileHandler file.Handler, client *http.Client, rekor rekorVerifier, flags upgradePlanFlags,
) error {
	config, err := config.FromFile(fileHandler, flags.configPath)
	if err != nil {
		return err
	}

	// get current image version of the cluster
	csp := config.GetProvider()

	version, err := getCurrentImageVersion(cmd.Context(), planner, csp)
	if err != nil {
		return fmt.Errorf("checking current image version: %w", err)
	}

	// fetch images definitions from GitHub and filter to only compatible images
	images, err := fetchImages(cmd.Context(), client)
	if err != nil {
		return fmt.Errorf("fetching available images: %w", err)
	}
	compatibleImages := getCompatibleImages(csp, version, images)
	if len(compatibleImages) == 0 {
		cmd.Println("No compatible images found to upgrade to.")
		return nil
	}

	// get expected measurements for each image
	if err := getCompatibleImageMeasurements(cmd.Context(), client, rekor, []byte(flags.cosignPubKey), compatibleImages); err != nil {
		return fmt.Errorf("fetching measurements for compatible images: %w", err)
	}

	// interactive mode
	if flags.filePath == "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Current version: %s\n", version)
		return upgradePlanInteractive(
			&nopWriteCloser{cmd.OutOrStdout()},
			io.NopCloser(cmd.InOrStdin()),
			flags.configPath, config, fileHandler,
			compatibleImages,
		)
	}

	// write upgrade plan to stdout
	if flags.filePath == "-" {
		content, err := encoder.NewEncoder(compatibleImages).Encode()
		if err != nil {
			return fmt.Errorf("encoding compatible images: %w", err)
		}
		_, err = cmd.OutOrStdout().Write(content)
		return err
	}

	// write upgrade plan to file
	return fileHandler.WriteYAML(flags.filePath, compatibleImages)
}

// fetchImages retrieves a list of the latest Constellation node images from GitHub.
func fetchImages(ctx context.Context, client *http.Client) (map[string]imageManifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageReleaseURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	imagesJSON, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	images := make(map[string]imageManifest)
	if err := json.Unmarshal(imagesJSON, &images); err != nil {
		return nil, err
	}

	return images, nil
}

// getCompatibleImages trims the list of images to only ones compatible with the current cluster.
func getCompatibleImages(csp cloudprovider.Provider, currentVersion string, images map[string]imageManifest) map[string]config.UpgradeConfig {
	compatibleImages := make(map[string]config.UpgradeConfig)

	switch csp {
	case cloudprovider.Azure:
		for imgVersion, image := range images {
			if semver.Compare(currentVersion, imgVersion) < 0 {
				compatibleImages[imgVersion] = config.UpgradeConfig{Image: image.AzureImage}
			}
		}

	case cloudprovider.GCP:
		for imgVersion, image := range images {
			if semver.Compare(currentVersion, imgVersion) < 0 {
				compatibleImages[imgVersion] = config.UpgradeConfig{Image: image.GCPImage}
			}
		}
	}

	return compatibleImages
}

// getCompatibleImageMeasurements retrieves the expected measurements for each image.
func getCompatibleImageMeasurements(ctx context.Context, client *http.Client, rekor rekorVerifier, pubK []byte, images map[string]config.UpgradeConfig) error {
	for idx, img := range images {
		measurementsURL, err := url.Parse(constants.S3PublicBucket + strings.ToLower(img.Image) + "/measurements.yaml")
		if err != nil {
			return err
		}

		signatureURL, err := url.Parse(constants.S3PublicBucket + strings.ToLower(img.Image) + "/measurements.yaml.sig")
		if err != nil {
			return err
		}

		hash, err := img.Measurements.FetchAndVerify(ctx, client, measurementsURL, signatureURL, pubK)
		if err != nil {
			return err
		}

		if err = verifyWithRekor(ctx, rekor, hash); err != nil {
			fmt.Printf("Warning: Unable to verify '%s' in Rekor.\n", hash)
			fmt.Printf("Make sure measurements are correct.\n")
		}

		images[idx] = img
	}

	return nil
}

// getCurrentImageVersion retrieves the semantic version of the image currently installed in the cluster.
// If the cluster is not using a release image, an error is returned.
func getCurrentImageVersion(ctx context.Context, planner upgradePlanner, csp cloudprovider.Provider) (string, error) {
	_, image, err := planner.GetCurrentImage(ctx)
	if err != nil {
		return "", err
	}

	var version string
	switch csp {
	case cloudprovider.Azure:
		if !azureCVMRxp.MatchString(image) {
			return "", fmt.Errorf("image %q does not look like a released production image for Azure", image)
		}
		versionRxp := regexp.MustCompile(`[\d]+.[\d]+.[\d]+$`)
		version = "v" + versionRxp.FindString(image)
	case cloudprovider.GCP:
		gcpVersion := gcpCVMRxp.FindStringSubmatch(image)
		if len(gcpVersion) != 2 {
			return "", fmt.Errorf("image %q does not look like a released production image for GCP", image)
		}
		version = strings.ReplaceAll(gcpVersion[1], "-", ".")
	default:
		return "", fmt.Errorf("unsupported cloud provider: %s", csp.String())
	}

	if !semver.IsValid(version) {
		return "", fmt.Errorf("image %q has no valid semantic version", image)
	}
	return version, nil
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
	compatibleImages map[string]config.UpgradeConfig,
) error {
	var imageVersions []string
	for k := range compatibleImages {
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

	fmt.Fprintf(out, "Image: %s\n", compatibleImages[res].Image)
	fmt.Fprintln(out, "Measurements:")
	content, err := encoder.NewEncoder(compatibleImages[res].Measurements).Encode()
	if err != nil {
		return fmt.Errorf("encoding measurements: %w", err)
	}
	measurements := strings.TrimSuffix(strings.Replace("\t"+string(content), "\n", "\n\t", -1), "\n\t")
	fmt.Fprintln(out, measurements)

	config.Upgrade = compatibleImages[res]
	return fileHandler.WriteYAML(configPath, config, file.OptOverwrite)
}

type upgradePlanFlags struct {
	configPath   string
	filePath     string
	cosignPubKey string
}

type imageManifest struct {
	AzureImage string `json:"AzureOSImage"`
	GCPImage   string `json:"GCPOSImage"`
}

type nopWriteCloser struct {
	io.Writer
}

func (c *nopWriteCloser) Close() error { return nil }

type upgradePlanner interface {
	GetCurrentImage(ctx context.Context) (*unstructured.Unstructured, string, error)
}
