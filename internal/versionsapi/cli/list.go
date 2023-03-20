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

	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
	verclient "github.com/edgelesssys/constellation/v2/internal/versionsapi/client"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	"golang.org/x/mod/semver"
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

func runList(cmd *cobra.Command, _ []string) error {
	flags, err := parseListFlags(cmd)
	if err != nil {
		return err
	}
	log := logger.New(logger.PlainLog, flags.logLevel)
	log.Debugf("Parsed flags: %+v", flags)

	log.Debugf("Validating flags")
	if err := flags.validate(); err != nil {
		return err
	}

	log.Debugf("Creating versions API client")
	client, err := verclient.NewReadOnlyClient(cmd.Context(), flags.region, flags.bucket, flags.distributionID, log)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	var minorVersions []string
	if flags.minorVersion != "" {
		minorVersions = []string{flags.minorVersion}
	} else {
		log.Debugf("Getting minor versions")
		minorVersions, err = listMinorVersions(cmd.Context(), client, flags.ref, flags.stream)
		var errNotFound *verclient.NotFoundError
		if err != nil && errors.As(err, &errNotFound) {
			log.Infof("No minor versions found for ref %q and stream %q.", flags.ref, flags.stream)
			return nil
		} else if err != nil {
			return err
		}
	}

	log.Debugf("Getting patch versions")
	patchVersions, err := listPatchVersions(cmd.Context(), client, flags.ref, flags.stream, minorVersions)
	var errNotFound *verclient.NotFoundError
	if err != nil && errors.As(err, &errNotFound) {
		log.Infof("No patch versions found for ref %q, stream %q and minor versions %v.", flags.ref, flags.stream, minorVersions)
		return nil
	} else if err != nil {
		return err
	}

	if flags.json {
		log.Debugf("Printing versions as JSON")
		var vers []string
		for _, v := range patchVersions {
			vers = append(vers, v.Version)
		}
		raw, err := json.Marshal(vers)
		if err != nil {
			return fmt.Errorf("marshaling versions: %w", err)
		}
		fmt.Println(string(raw))
		return nil
	}

	log.Debugf("Printing versions")
	for _, v := range patchVersions {
		fmt.Println(v.ShortPath())
	}

	return nil
}

func listMinorVersions(ctx context.Context, client *verclient.Client, ref string, stream string) ([]string, error) {
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

func listPatchVersions(ctx context.Context, client *verclient.Client, ref string, stream string, minorVer []string,
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

		if len(list.Versions) == 0 {
			return nil, fmt.Errorf("no versions found")
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
	logLevel       zapcore.Level
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
	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
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
