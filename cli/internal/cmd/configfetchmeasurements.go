/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newConfigFetchMeasurementsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch-measurements",
		Short: "Fetch measurements for configured cloud provider and image",
		Long: "Fetch measurements for configured cloud provider and image.\n\n" +
			"A config needs to be generated first.",
		Args: cobra.ExactArgs(0),
		RunE: runConfigFetchMeasurements,
	}
	cmd.Flags().StringP("url", "u", "", "alternative URL to fetch measurements from")
	cmd.Flags().StringP("signature-url", "s", "", "alternative URL to fetch measurements' signature from")

	return cmd
}

type fetchMeasurementsFlags struct {
	measurementsURL *url.URL
	signatureURL    *url.URL
	configPath      string
	force           bool
}

type configFetchMeasurementsCmd struct {
	log debugLog
}

func runConfigFetchMeasurements(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	fileHandler := file.NewHandler(afero.NewOsFs())
	rekor, err := sigstore.NewRekor()
	if err != nil {
		return fmt.Errorf("constructing Rekor client: %w", err)
	}
	cfm := &configFetchMeasurementsCmd{log: log}

	return cfm.configFetchMeasurements(cmd, rekor, []byte(constants.CosignPublicKey), fileHandler, http.DefaultClient)
}

func (cfm *configFetchMeasurementsCmd) configFetchMeasurements(
	cmd *cobra.Command, verifier rekorVerifier, cosignPublicKey []byte,
	fileHandler file.Handler, client *http.Client,
) error {
	flags, err := cfm.parseFetchMeasurementsFlags(cmd)
	if err != nil {
		return err
	}
	cfm.log.Debugf("Using flags %v", flags)

	cfm.log.Debugf("Loading configuration file from %q", flags.configPath)
	conf, err := config.NewWithClient(fileHandler, flags.configPath, client, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}

	if !conf.IsReleaseImage() {
		cmd.PrintErrln("Configured image doesn't look like a released production image. Double check image before deploying to production.")
	}

	cfm.log.Debugf("Creating context")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cfm.log.Debugf("Updating URLs")
	if err := flags.updateURLs(conf); err != nil {
		return err
	}

	cfm.log.Debugf("Fetching and verifying measurements")
	imageVersion, err := versionsapi.NewVersionFromShortPath(conf.Image, versionsapi.VersionKindImage)
	if err != nil {
		return err
	}
	var fetchedMeasurements measurements.M
	hash, err := fetchedMeasurements.FetchAndVerify(
		ctx, client,
		flags.measurementsURL,
		flags.signatureURL,
		cosignPublicKey,
		imageVersion,
		conf.GetProvider(),
		conf.GetAttestationConfig().GetVariant(),
	)
	if err != nil {
		return err
	}

	cfm.log.Debugf("Fetched and verified measurements, hash is %s", hash)
	if err := verifyWithRekor(cmd.Context(), verifier, hash); err != nil {
		cmd.PrintErrf("Ignoring Rekor related error: %v\n", err)
		cmd.PrintErrln("Make sure the downloaded measurements are trustworthy!")
	}

	cfm.log.Debugf("Verified measurements with Rekor, updating measurements in configuration")
	conf.UpdateMeasurements(fetchedMeasurements)
	if err := fileHandler.WriteYAML(flags.configPath, conf, file.OptOverwrite); err != nil {
		return err
	}
	cfm.log.Debugf("Configuration written to %s", flags.configPath)
	return nil
}

// parseURLFlag checks that flag can be parsed as URL.
// If no value was provided for flag, nil is returned.
func (cfm *configFetchMeasurementsCmd) parseURLFlag(cmd *cobra.Command, flag string) (*url.URL, error) {
	rawURL, err := cmd.Flags().GetString(flag)
	if err != nil {
		return nil, fmt.Errorf("parsing config generate flags '%s': %w", flag, err)
	}
	cfm.log.Debugf("Flag %s has raw URL %q", flag, rawURL)
	if rawURL != "" {
		cfm.log.Debugf("Parsing raw URL")
		return url.Parse(rawURL)
	}
	return nil, nil
}

func (cfm *configFetchMeasurementsCmd) parseFetchMeasurementsFlags(cmd *cobra.Command) (*fetchMeasurementsFlags, error) {
	measurementsURL, err := cfm.parseURLFlag(cmd, "url")
	if err != nil {
		return &fetchMeasurementsFlags{}, err
	}
	cfm.log.Debugf("Parsed measurements URL as %v", measurementsURL)

	measurementsSignatureURL, err := cfm.parseURLFlag(cmd, "signature-url")
	if err != nil {
		return &fetchMeasurementsFlags{}, err
	}
	cfm.log.Debugf("Parsed measurements signature URL as %v", measurementsSignatureURL)

	config, err := cmd.Flags().GetString("config")
	if err != nil {
		return &fetchMeasurementsFlags{}, fmt.Errorf("parsing config path argument: %w", err)
	}
	cfm.log.Debugf("Configuration path is %q", config)

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return &fetchMeasurementsFlags{}, fmt.Errorf("parsing force argument: %w", err)
	}

	return &fetchMeasurementsFlags{
		measurementsURL: measurementsURL,
		signatureURL:    measurementsSignatureURL,
		configPath:      config,
		force:           force,
	}, nil
}

func (f *fetchMeasurementsFlags) updateURLs(conf *config.Config) error {
	ver, err := versionsapi.NewVersionFromShortPath(conf.Image, versionsapi.VersionKindImage)
	if err != nil {
		return fmt.Errorf("creating version from image name: %w", err)
	}
	measurementsURL, signatureURL, err := versionsapi.MeasurementURL(ver)
	if err != nil {
		return err
	}

	if f.measurementsURL == nil {
		f.measurementsURL = measurementsURL
	}

	if f.signatureURL == nil {
		f.signatureURL = signatureURL
	}
	return nil
}
