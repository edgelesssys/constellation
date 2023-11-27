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

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/featureset"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
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

type verifyFetcher interface {
	FetchAndVerifyMeasurements(ctx context.Context,
		image string, csp cloudprovider.Provider, attestationVariant variant.Variant,
		noVerify bool,
	) (measurements.M, error)
}

type configFetchMeasurementsCmd struct {
	flags                fetchMeasurementsFlags
	canFetchMeasurements bool
	log                  debugLog
	verifyFetcher        verifyFetcher
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

	verifyFetcher := measurements.NewVerifyFetcher(sigstore.NewCosignVerifier, rekor, http.DefaultClient)
	cfm := &configFetchMeasurementsCmd{log: log, canFetchMeasurements: featureset.CanFetchMeasurements, verifyFetcher: verifyFetcher}
	if err := cfm.flags.parse(cmd.Flags()); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}
	cfm.log.Debugf("Using flags %+v", cfm.flags)

	fetcher := attestationconfigapi.NewFetcherWithClient(http.DefaultClient, constants.CDNRepositoryURL)
	return cfm.configFetchMeasurements(cmd, fileHandler, fetcher)
}

func (cfm *configFetchMeasurementsCmd) configFetchMeasurements(
	cmd *cobra.Command,
	fileHandler file.Handler, fetcher attestationconfigapi.Fetcher,
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
	fetchedMeasurements, err := cfm.verifyFetcher.FetchAndVerifyMeasurements(ctx, conf.Image, conf.GetProvider(),
		conf.GetAttestationConfig().GetVariant(), cfm.flags.insecure)
	if err != nil {
		if errors.Is(err, measurements.ErrRekor) {
			cmd.PrintErrf("Ignoring Rekor related error: %v\n", err)
			cmd.PrintErrln("Make sure the downloaded measurements are trustworthy!")
		} else {
			return fmt.Errorf("fetching and verifying measurements: %w", err)
		}
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
