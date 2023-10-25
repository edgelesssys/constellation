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

	"github.com/edgelesssys/constellation/v2/cli/internal/featureset"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/internal/sigstore/keyselect"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	cmd.Flags().Bool("insecure", false, "skip the measurement signature verification")
	must(cmd.Flags().MarkHidden("insecure"))

	return cmd
}

type fetchMeasurementsFlags struct {
	rootFlags
	measurementsURL *url.URL
	signatureURL    *url.URL
	insecure        bool
}

func (f *fetchMeasurementsFlags) parse(flags *pflag.FlagSet) error {
	var err error
	if err := f.rootFlags.parse(flags); err != nil {
		return err
	}

	f.measurementsURL, err = parseURLFlag(flags, "url")
	if err != nil {
		return err
	}
	f.signatureURL, err = parseURLFlag(flags, "signature-url")
	if err != nil {
		return err
	}
	f.insecure, err = flags.GetBool("insecure")
	if err != nil {
		return fmt.Errorf("getting 'insecure' flag: %w", err)
	}
	return nil
}

type configFetchMeasurementsCmd struct {
	flags                fetchMeasurementsFlags
	canFetchMeasurements bool
	log                  debugLog
}

func runConfigFetchMeasurements(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	fileHandler := file.NewHandler(afero.NewOsFs())
	rekor, err := sigstore.NewRekor()
	if err != nil {
		return fmt.Errorf("constructing Rekor client: %w", err)
	}
	cfm := &configFetchMeasurementsCmd{log: log, canFetchMeasurements: featureset.CanFetchMeasurements}
	if err := cfm.flags.parse(cmd.Flags()); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}
	cfm.log.Debugf("Using flags %+v", cfm.flags)

	fetcher := attestationconfigapi.NewFetcherWithClient(http.DefaultClient, constants.CDNRepositoryURL)
	return cfm.configFetchMeasurements(cmd, sigstore.NewCosignVerifier, rekor, fileHandler, fetcher, http.DefaultClient)
}

func (cfm *configFetchMeasurementsCmd) configFetchMeasurements(
	cmd *cobra.Command, newCosignVerifier cosignVerifierConstructor, rekor rekorVerifier,
	fileHandler file.Handler, fetcher attestationconfigapi.Fetcher, client *http.Client,
) error {
	if !cfm.canFetchMeasurements {
		cmd.PrintErrln("Fetching measurements is not supported in the OSS build of the Constellation CLI. Consult the documentation for instructions on where to download the enterprise version.")
		return errors.New("fetching measurements is not supported")
	}

	cfm.log.Debugf("Loading configuration file from %q", cfm.flags.pathPrefixer.PrefixPrintablePath(constants.ConfigFilename))

	conf, err := config.New(fileHandler, constants.ConfigFilename, fetcher, cfm.flags.force)
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
	if err := cfm.flags.updateURLs(conf); err != nil {
		return err
	}

	cfm.log.Debugf("Fetching and verifying measurements")
	imageVersion, err := versionsapi.NewVersionFromShortPath(conf.Image, versionsapi.VersionKindImage)
	if err != nil {
		return err
	}

	publicKey, err := keyselect.CosignPublicKeyForVersion(imageVersion)
	if err != nil {
		return fmt.Errorf("getting public key: %w", err)
	}
	cosign, err := newCosignVerifier(publicKey)
	if err != nil {
		return fmt.Errorf("creating cosign verifier: %w", err)
	}

	var fetchedMeasurements measurements.M
	var hash string
	if cfm.flags.insecure {
		if err := fetchedMeasurements.FetchNoVerify(
			ctx,
			client,
			cfm.flags.measurementsURL,
			imageVersion,
			conf.GetProvider(),
			conf.GetAttestationConfig().GetVariant(),
		); err != nil {
			return fmt.Errorf("fetching measurements without verification: %w", err)
		}

		cfm.log.Debugf("Fetched measurements without verification")
	} else {
		hash, err = fetchedMeasurements.FetchAndVerify(
			ctx,
			client,
			cosign,
			cfm.flags.measurementsURL,
			cfm.flags.signatureURL,
			imageVersion,
			conf.GetProvider(),
			conf.GetAttestationConfig().GetVariant(),
		)
		if err != nil {
			return fmt.Errorf("fetching and verifying measurements: %w", err)
		}
		cfm.log.Debugf("Fetched and verified measurements, hash is %s", hash)
		if err := sigstore.VerifyWithRekor(cmd.Context(), publicKey, rekor, hash); err != nil {
			cmd.PrintErrf("Ignoring Rekor related error: %v\n", err)
			cmd.PrintErrln("Make sure the downloaded measurements are trustworthy!")
		}

		cfm.log.Debugf("Verified measurements with Rekor")
	}
	cfm.log.Debugf("Measurements:\n", fetchedMeasurements)

	cfm.log.Debugf("Updating measurements in configuration")
	conf.UpdateMeasurements(fetchedMeasurements)
	if err := fileHandler.WriteYAML(constants.ConfigFilename, conf, file.OptOverwrite); err != nil {
		return err
	}
	cfm.log.Debugf("Configuration written to %s", cfm.flags.pathPrefixer.PrefixPrintablePath(constants.ConfigFilename))
	cmd.Print("Successfully fetched measurements and updated Configuration\n")
	return nil
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

// parseURLFlag checks that flag can be parsed as URL.
// If no value was provided for flag, nil is returned.
func parseURLFlag(flags *pflag.FlagSet, flag string) (*url.URL, error) {
	rawURL, err := flags.GetString(flag)
	if err != nil {
		return nil, fmt.Errorf("getting '%s' flag: %w", flag, err)
	}
	if rawURL != "" {
		return url.Parse(rawURL)
	}
	return nil, nil
}

type rekorVerifier interface {
	SearchByHash(context.Context, string) ([]string, error)
	VerifyEntry(context.Context, string, string) error
}

type cosignVerifierConstructor func([]byte) (sigstore.Verifier, error)
