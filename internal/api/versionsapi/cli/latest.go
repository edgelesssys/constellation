/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/cobra"
)

func newLatestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest",
		Short: "Find latest version",
		Long:  "Find latest version of a ref/stream. The returned version is in short format, if --json flag is not set.",
		RunE:  runLatest,
		Args:  cobra.ExactArgs(0),
	}

	cmd.Flags().String("ref", "-", "Ref to query")
	cmd.Flags().String("stream", "stable", "Stream to query")
	cmd.Flags().Bool("json", false, "Whether to output the result as JSON")

	return cmd
}

func runLatest(cmd *cobra.Command, _ []string) (retErr error) {
	flags, err := parseLatestFlags(cmd)
	if err != nil {
		return err
	}
	log := logger.NewTextLogger(flags.logLevel)
	log.Debug("Using flags", "ref", flags.ref, "stream", flags.stream, "json", flags.json)

	log.Debug("Validating flags")
	if err := flags.validate(); err != nil {
		return err
	}

	log.Debug("Creating versions API client")
	client, clientClose, err := versionsapi.NewReadOnlyClient(cmd.Context(), flags.region, flags.bucket, flags.distributionID, log)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	defer func() {
		err := clientClose(cmd.Context())
		if err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("failed to invalidate cache: %w", err))
		}
	}()

	log.Debug("Requesting latest version")
	latest := versionsapi.Latest{
		Ref:    flags.ref,
		Stream: flags.stream,
		Kind:   versionsapi.VersionKindImage,
	}
	latest, err = client.FetchVersionLatest(cmd.Context(), latest)
	if err != nil {
		return fmt.Errorf("fetching latest version: %w", err)
	}

	if flags.json {
		out, err := json.MarshalIndent(latest, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling JSON: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), latest.ShortPath())
	return nil
}

type latestFlags struct {
	ref            string
	stream         string
	json           bool
	region         string
	bucket         string
	distributionID string
	logLevel       slog.Level
}

func (l *latestFlags) validate() error {
	if err := versionsapi.ValidateRef(l.ref); err != nil {
		return fmt.Errorf("invalid ref: %w", err)
	}
	if err := versionsapi.ValidateStream(l.ref, l.stream); err != nil {
		return fmt.Errorf("invalid stream: %w", err)
	}

	return nil
}

func parseLatestFlags(cmd *cobra.Command) (latestFlags, error) {
	ref, err := cmd.Flags().GetString("ref")
	if err != nil {
		return latestFlags{}, err
	}
	ref = versionsapi.CanonicalizeRef(ref)
	stream, err := cmd.Flags().GetString("stream")
	if err != nil {
		return latestFlags{}, err
	}
	json, err := cmd.Flags().GetBool("json")
	if err != nil {
		return latestFlags{}, err
	}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return latestFlags{}, err
	}
	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return latestFlags{}, err
	}
	distributionID, err := cmd.Flags().GetString("distribution-id")
	if err != nil {
		return latestFlags{}, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return latestFlags{}, err
	}
	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}

	return latestFlags{
		ref:            ref,
		stream:         stream,
		json:           json,
		region:         region,
		bucket:         bucket,
		distributionID: distributionID,
		logLevel:       logLevel,
	}, nil
}
