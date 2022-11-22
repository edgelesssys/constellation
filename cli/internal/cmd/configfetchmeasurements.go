/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/image"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newConfigFetchMeasurementsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch-measurements",
		Short: "Fetch measurements for configured cloud provider and image",
		Long:  "Fetch measurements for configured cloud provider and image. A config needs to be generated first!",
		Args:  cobra.ExactArgs(0),
		RunE:  runConfigFetchMeasurements,
	}
	cmd.Flags().StringP("url", "u", "", "alternative URL to fetch measurements from")
	cmd.Flags().StringP("signature-url", "s", "", "alternative URL to fetch measurements' signature from")

	return cmd
}

type fetchMeasurementsFlags struct {
	measurementsURL *url.URL
	signatureURL    *url.URL
	configPath      string
}

func runConfigFetchMeasurements(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	rekor, err := sigstore.NewRekor()
	if err != nil {
		return fmt.Errorf("constructing Rekor client: %w", err)
	}
	return configFetchMeasurements(cmd, rekor, fileHandler, http.DefaultClient, image.New())
}

func configFetchMeasurements(cmd *cobra.Command, verifier rekorVerifier, fileHandler file.Handler, client *http.Client, img imageFetcher) error {
	flags, err := parseFetchMeasurementsFlags(cmd)
	if err != nil {
		return err
	}

	conf, err := config.New(fileHandler, flags.configPath)
	if err != nil {
		return displayConfigValidationErrors(cmd.ErrOrStderr(), err)
	}

	if !conf.IsReleaseImage() {
		cmd.PrintErrln("Configured image doesn't look like a released production image. Double check image before deploying to production.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := flags.updateURLs(ctx, conf, img); err != nil {
		return err
	}

	var fetchedMeasurements measurements.M
	hash, err := fetchedMeasurements.FetchAndVerify(ctx, client, flags.measurementsURL, flags.signatureURL, []byte(constants.CosignPublicKey))
	if err != nil {
		return err
	}

	if err := verifyWithRekor(cmd.Context(), verifier, hash); err != nil {
		cmd.PrintErrf("Ignoring Rekor related error: %v\n", err)
		cmd.PrintErrln("Make sure the downloaded measurements are trustworthy!")
	}

	conf.UpdateMeasurements(fetchedMeasurements)
	if err := fileHandler.WriteYAML(flags.configPath, conf, file.OptOverwrite); err != nil {
		return err
	}

	return nil
}

// parseURLFlag checks that flag can be parsed as URL.
// If no value was provided for flag, nil is returned.
func parseURLFlag(cmd *cobra.Command, flag string) (*url.URL, error) {
	rawURL, err := cmd.Flags().GetString(flag)
	if err != nil {
		return nil, fmt.Errorf("parsing config generate flags '%s': %w", flag, err)
	}
	if rawURL != "" {
		return url.Parse(rawURL)
	}
	return nil, nil
}

func parseFetchMeasurementsFlags(cmd *cobra.Command) (*fetchMeasurementsFlags, error) {
	measurementsURL, err := parseURLFlag(cmd, "url")
	if err != nil {
		return &fetchMeasurementsFlags{}, err
	}

	measurementsSignatureURL, err := parseURLFlag(cmd, "signature-url")
	if err != nil {
		return &fetchMeasurementsFlags{}, err
	}

	config, err := cmd.Flags().GetString("config")
	if err != nil {
		return &fetchMeasurementsFlags{}, fmt.Errorf("parsing config path argument: %w", err)
	}

	return &fetchMeasurementsFlags{
		measurementsURL: measurementsURL,
		signatureURL:    measurementsSignatureURL,
		configPath:      config,
	}, nil
}

func (f *fetchMeasurementsFlags) updateURLs(ctx context.Context, conf *config.Config, img imageFetcher) error {
	imageRef, err := img.FetchReference(ctx, conf)
	if err != nil {
		return err
	}

	if f.measurementsURL == nil {
		// TODO(AB#2644): resolve image version to reference
		parsedURL, err := url.Parse(constants.S3PublicBucket + imageRef + "/measurements.yaml")
		if err != nil {
			return err
		}
		f.measurementsURL = parsedURL
	}

	if f.signatureURL == nil {
		parsedURL, err := url.Parse(constants.S3PublicBucket + imageRef + "/measurements.yaml.sig")
		if err != nil {
			return err
		}
		f.signatureURL = parsedURL
	}
	return nil
}

type imageFetcher interface {
	FetchReference(ctx context.Context, config *config.Config) (string, error)
}
