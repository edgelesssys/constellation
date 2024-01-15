/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
  "os"
	"log/slog"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"

	apiclient "github.com/edgelesssys/constellation/v2/internal/api/client"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List versions",
		Long:  "List all versions of a ref/stream. The returned version are in short format, if --json flag is not set.",
		RunE:  runList,
		Args:  cobra.ExactArgs(0),
	}

	cmd.Flags().String("ref", "-", "Ref to query")
	cmd.Flags().String("stream", "stable", "Stream to query")
	cmd.Flags().String("minor-version", "", "Minor version to query (format: \"v1.2\")")
	cmd.Flags().Bool("json", false, "Whether to output the result as JSON")

	return cmd
}

func runList(cmd *cobra.Command, _ []string) (retErr error) {
	flags, err := parseListFlags(cmd)
	if err != nil {
		return err
	}
  log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: flags.logLevel}))
	log.Debug(fmt.Sprintf("Parsed flags: %+v", flags))

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

	var minorVersions []string
	if flags.minorVersion != "" {
		minorVersions = []string{flags.minorVersion}
	} else {
		log.Debug("Getting minor versions")
		minorVersions, err = listMinorVersions(cmd.Context(), client, flags.ref, flags.stream)
		var errNotFound *apiclient.NotFoundError
		if err != nil && errors.As(err, &errNotFound) {
			log.Info(fmt.Sprintf("No minor versions found for ref %q and stream %q.", flags.ref, flags.stream))
			return nil
		} else if err != nil {
			return err
		}
	}

	log.Debug("Getting patch versions")
	patchVersions, err := listPatchVersions(cmd.Context(), client, flags.ref, flags.stream, minorVersions)
	var errNotFound *apiclient.NotFoundError
	if err != nil && errors.As(err, &errNotFound) {
		log.Info(fmt.Sprintf("No patch versions found for ref %q, stream %q and minor versions %v.", flags.ref, flags.stream, minorVersions))
		return nil
	} else if err != nil {
		return err
	}

	if flags.json {
		log.Debug("Printing versions as JSON")
		var vers []string
		for _, v := range patchVersions {
			vers = append(vers, v.Version())
		}
		raw, err := json.MarshalIndent(vers, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling versions: %w", err)
		}
		fmt.Println(string(raw))
		return nil
	}

	log.Debug("Printing versions")
	for _, v := range patchVersions {
		fmt.Println(v.ShortPath())
	}

	return nil
}

func listMinorVersions(ctx context.Context, client *versionsapi.Client, ref string, stream string) ([]string, error) {
	list := versionsapi.List{
		Ref:         ref,
		Stream:      stream,
		Granularity: versionsapi.GranularityMajor,
		Base:        "v2",
		Kind:        versionsapi.VersionKindImage,
	}
	list, err := client.FetchVersionList(ctx, list)
	if err != nil {
		return nil, fmt.Errorf("listing minor versions: %w", err)
	}

	return list.Versions, nil
}

func listPatchVersions(ctx context.Context, client *versionsapi.Client, ref string, stream string, minorVer []string,
) ([]versionsapi.Version, error) {
	var patchVers []versionsapi.Version

	list := versionsapi.List{
		Ref:         ref,
		Stream:      stream,
		Granularity: versionsapi.GranularityMinor,
		Kind:        versionsapi.VersionKindImage,
	}

	for _, ver := range minorVer {
		list.Base = ver
		list, err := client.FetchVersionList(ctx, list)
		if err != nil {
			return nil, fmt.Errorf("listing patch versions: %w", err)
		}

		patchVers = append(patchVers, list.StructuredVersions()...)
	}

	return patchVers, nil
}

type listFlags struct {
	ref            string
	stream         string
	minorVersion   string
	region         string
	bucket         string
	distributionID string
	json           bool
	logLevel       slog.Level
}

func (l *listFlags) validate() error {
	if err := versionsapi.ValidateRef(l.ref); err != nil {
		return fmt.Errorf("invalid ref: %w", err)
	}
	if err := versionsapi.ValidateStream(l.ref, l.stream); err != nil {
		return fmt.Errorf("invalid stream: %w", err)
	}
	if l.minorVersion != "" {
		if !semver.IsValid(l.minorVersion) || semver.MajorMinor(l.minorVersion) != l.minorVersion {
			return fmt.Errorf("invalid minor version: %q", l.minorVersion)
		}
	}

	return nil
}

func parseListFlags(cmd *cobra.Command) (listFlags, error) {
	ref, err := cmd.Flags().GetString("ref")
	if err != nil {
		return listFlags{}, err
	}
	ref = versionsapi.CanonicalizeRef(ref)
	stream, err := cmd.Flags().GetString("stream")
	if err != nil {
		return listFlags{}, err
	}
	minorVersion, err := cmd.Flags().GetString("minor-version")
	if err != nil {
		return listFlags{}, err
	}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return listFlags{}, err
	}
	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return listFlags{}, err
	}
	distributionID, err := cmd.Flags().GetString("distribution-id")
	if err != nil {
		return listFlags{}, err
	}
	json, err := cmd.Flags().GetBool("json")
	if err != nil {
		return listFlags{}, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return listFlags{}, err
	}
	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}

	return listFlags{
		ref:            ref,
		stream:         stream,
		minorVersion:   minorVersion,
		region:         region,
		bucket:         bucket,
		distributionID: distributionID,
		json:           json,
		logLevel:       logLevel,
	}, nil
}
