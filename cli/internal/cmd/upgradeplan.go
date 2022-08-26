package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/talos-systems/talos/pkg/machinery/config/encoder"
	"golang.org/x/mod/semver"
)

func newUpgradePlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Plan an upgrade of a Constellation cluster",
		Long:  "Plan an upgrade of a Constellation cluster by fetching compatible image versions and their measurements.",
		Args:  cobra.NoArgs,
		RunE:  runUpgradePlan,
	}

	cmd.Flags().StringP("file", "f", "upgrade-plan.yaml", "path to output file, or '-' for stdout")
	cmd.Flags().StringP("url", "u", constants.S3PublicBucket, "alternative base URL to fetch measurements from")
	cmd.Flags().StringP("signature-url", "s", constants.S3PublicBucket, "alternative base URL to fetch measurements' signature from")

	return cmd
}

func runUpgradePlan(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	flags, err := parseUpgradePlanFlags(cmd)
	if err != nil {
		return err
	}

	return upgradePlan(cmd, fileHandler, http.DefaultClient, flags)
}

// upgradePlan plans an upgrade of a Constellation cluster.
func upgradePlan(cmd *cobra.Command, fileHandler file.Handler,
	client *http.Client, flags upgradePlanFlags,
) error {
	config, err := config.FromFile(fileHandler, flags.configPath)
	if err != nil {
		return err
	}

	// get current image version of the cluster
	// if an update was already performed, use the updated image version
	csp := config.GetProvider()
	version := getCurrentImageVersion(csp, config)
	if version == "" || !semver.IsValid(version) {
		return errors.New("unable to determine valid version for current image")
	}

	// fetch images definitions from GitHub and filter to only compatible images
	images, err := fetchImages(cmd.Context(), client)
	if err != nil {
		return fmt.Errorf("fetching available images: %w", err)
	}
	compatibleImages := getCompatibleImages(csp, version, images)

	// get expected measurements for each image
	if err := getMeasurements(cmd.Context(), client, flags, compatibleImages); err != nil {
		return fmt.Errorf("fetching measurements for compatible images: %w", err)
	}

	// write upgrade plan to file
	if flags.filePath == "-" {
		content, err := encoder.NewEncoder(compatibleImages).Encode()
		if err != nil {
			return fmt.Errorf("encoding compatible images: %w", err)
		}
		_, err = cmd.OutOrStdout().Write(content)
		return err
	}
	return fileHandler.WriteYAML(flags.filePath, compatibleImages)
}

// fetchImages retrieves a list of the latest Constellation node images from GitHub.
func fetchImages(ctx context.Context, imageFetcher imageFetcher) (map[string]imageManifest, error) {
	const releaseURL = "https://github.com/edgelesssys/constellation/releases/latest/download/image-manifest.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	res, err := imageFetcher.Do(req)
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

// getMeasurements retrieves the expected measurements for each image.
func getMeasurements(ctx context.Context, client *http.Client, flags upgradePlanFlags, images map[string]config.UpgradeConfig) error {
	for idx, img := range images {
		parsedMeasurementsURL, err := url.Parse(flags.measurementsURL + img.Image + "/measurements.yaml")
		if err != nil {
			return err
		}
		parsedSignatureURL, err := url.Parse(constants.S3PublicBucket + img.Image + "/measurements.yaml.sig")
		if err != nil {
			return err
		}

		if err := img.Measurements.FetchAndVerify(ctx, client, parsedMeasurementsURL, parsedSignatureURL, []byte(constants.CosignPublicKey)); err != nil {
			return err
		}
		images[idx] = img
	}

	return nil
}

func getCurrentImageVersion(csp cloudprovider.Provider, config *config.Config) string {
	switch csp {
	case cloudprovider.Azure:
		image := config.Provider.Azure.Image
		if config.Upgrade.Image != "" {
			// override image if upgrade image is specified
			image = config.Upgrade.Image
		}
		version := regexp.MustCompile(`constellation/versions/[\d]+.[\d]+.[\d]+`).FindString(image)
		return strings.TrimPrefix(version, "constellation/versions/")
	case cloudprovider.GCP:
		image := config.Provider.GCP.Image
		if config.Upgrade.Image != "" {
			// override image if upgrade image is specified
			image = config.Upgrade.Image
		}
		version := regexp.MustCompile(`v[\d]+-[\d]+-[\d]+$`).FindString(image)
		return strings.ReplaceAll(version, "-", ".")
	default:
		return ""
	}
}

func parseUpgradePlanFlags(cmd *cobra.Command) (upgradePlanFlags, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return upgradePlanFlags{}, err
	}
	filePath, err := cmd.Flags().GetString("upgrade-plan")
	if err != nil {
		return upgradePlanFlags{}, err
	}
	measurementsURL, err := cmd.Flags().GetString("url")
	if err != nil {
		return upgradePlanFlags{}, err
	}

	measurementsSignatureURL, err := cmd.Flags().GetString("signature-url")
	if err != nil {
		return upgradePlanFlags{}, err
	}

	return upgradePlanFlags{
		configPath:      configPath,
		filePath:        filePath,
		measurementsURL: measurementsURL,
		signatureURL:    measurementsSignatureURL,
	}, nil
}

type upgradePlanFlags struct {
	configPath      string
	filePath        string
	measurementsURL string
	signatureURL    string
}

type imageManifest struct {
	AzureImage string `json:"AzureCoreOSImage"`
	GCPImage   string `json:"GCPCoreOSImage"`
}

type imageFetcher interface {
	Do(req *http.Request) (*http.Response, error)
}
