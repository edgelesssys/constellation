/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

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
	config          string
}

func runConfigFetchMeasurements(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	return configFetchMeasurements(cmd, fileHandler, http.DefaultClient)
}

func configFetchMeasurements(cmd *cobra.Command, fileHandler file.Handler, client *http.Client) error {
	flags, err := parseFetchMeasurementsFlags(cmd)
	if err != nil {
		return err
	}

	conf, err := config.FromFile(fileHandler, flags.config)
	if err != nil {
		return err
	}

	if conf.IsDebugImage() {
		cmd.Println("Configured image doesn't look like a released production image. Double check image before deploying to production.")
	}

	if err := flags.updateURLs(conf); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var fetchedMeasurements config.Measurements
	hash, err := fetchedMeasurements.FetchAndVerify(ctx, client, flags.measurementsURL, flags.signatureURL, []byte(constants.CosignPublicKey))
	if err != nil {
		return err
	}

	if err := verifyWithRekor(cmd, hash); err != nil {
		cmd.Printf("Ignoring Rekor related error: %v\n", err)
		cmd.Println("Make sure the downloaded measurements are trustworthy!")
	}

	conf.UpdateMeasurements(fetchedMeasurements)
	if err := fileHandler.WriteYAML(flags.config, conf, file.OptOverwrite); err != nil {
		return err
	}

	return nil
}

func verifyWithRekor(cmd *cobra.Command, hash string) error {
	rekor, err := sigstore.NewRekor()
	if err != nil {
		return fmt.Errorf("constructing Rekor client: %w", err)
	}

	uuids, err := rekor.SearchByHash(cmd.Context(), hash)
	if err != nil {
		return fmt.Errorf("searching Rekor for hash: %w", err)
	}

	if len(uuids) == 0 {
		return fmt.Errorf("no matching entries in Rekor")
	}

	// We expect the first entry in Rekor to be our original entry.
	// SHA256 should ensure there is no entry with the same hash.
	// Any subsequent hashes are treated as potential attacks and are ignored.
	artifactUUID := uuids[0]

	entry, err := rekor.GetAndVerifyEntry(cmd.Context(), artifactUUID)
	if err != nil {
		return fmt.Errorf("verifying Rekor entry: %w", err)
	}

	rekord, err := sigstore.HashedRekordFromEntry(entry)
	if err != nil {
		return fmt.Errorf("extracting rekord from Rekor entry: %w", err)
	}

	cosignPubBase64 := base64.StdEncoding.EncodeToString([]byte(constants.CosignPublicKey))
	if !sigstore.IsEntrySignedBy(rekord, cosignPubBase64) {
		return errors.New("rekord signed by unknown key")
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
		config:          config,
	}, nil
}

func (f *fetchMeasurementsFlags) updateURLs(conf *config.Config) error {
	if f.measurementsURL == nil {
		parsedURL, err := url.Parse(constants.S3PublicBucket + conf.Image() + "/measurements.yaml")
		if err != nil {
			return err
		}
		f.measurementsURL = parsedURL
	}

	if f.signatureURL == nil {
		parsedURL, err := url.Parse(constants.S3PublicBucket + conf.Image() + "/measurements.yaml.sig")
		if err != nil {
			return err
		}
		f.signatureURL = parsedURL
	}
	return nil
}
