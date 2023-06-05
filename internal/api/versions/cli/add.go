/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"errors"
	"fmt"

	apiclient "github.com/edgelesssys/constellation/v2/internal/api/client"
	versionsapi "github.com/edgelesssys/constellation/v2/internal/api/versions"
	verclient "github.com/edgelesssys/constellation/v2/internal/api/versions/client"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	"golang.org/x/mod/semver"
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new version",
		Long: `Add a new version to the versions API.

Developers should not use this command directly. It is invoked by the CI/CD pipeline.
If you've build a local image, use a local override instead of adding a new version.

❗ If you use the command nevertheless, you better know what you do.
`,
		RunE: runAdd,
	}

	cmd.Flags().String("ref", "", "Ref of the version to add")
	cmd.Flags().String("stream", "", "Stream of the version to add")
	cmd.Flags().String("version", "", "Version to add (format: \"v1.2.3\")")
	cmd.Flags().String("kind", "", "Version kind to add (e.g. image, cli)")
	cmd.Flags().Bool("latest", false, "Whether the version is the latest version of the ref/stream")
	cmd.Flags().Bool("release", false, "Whether the version is a release version")
	cmd.Flags().Bool("dryrun", false, "Whether to run in dry-run mode (no changes are made)")

	cmd.MarkFlagsMutuallyExclusive("ref", "release")
	must(cmd.MarkFlagRequired("version"))

	return cmd
}

func runAdd(cmd *cobra.Command, _ []string) (retErr error) {
	flags, err := parseAddFlags(cmd)
	if err != nil {
		return err
	}
	log := logger.New(logger.PlainLog, flags.logLevel)
	log.Debugf("Parsed flags: %+v", flags)

	log.Debugf("Validating flags")
	if err := flags.validate(log); err != nil {
		return err
	}

	log.Debugf("Creating version struct")
	ver := versionsapi.Version{
		Ref:     flags.ref,
		Stream:  flags.stream,
		Version: flags.version,
		Kind:    flags.kind,
	}
	if err := ver.Validate(); err != nil {
		return err
	}

	log.Debugf("Creating versions API client")
	client, clientClose, err := verclient.NewClient(cmd.Context(), flags.region, flags.bucket, flags.distributionID, flags.dryRun, log)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	defer func(retErr *error) {
		log.Infof("Invalidating cache. This may take some time")
		if err := clientClose(cmd.Context()); err != nil && retErr == nil {
			*retErr = fmt.Errorf("invalidating cache: %w", err)
		}
	}(&retErr)

	log.Infof("Adding version")
	if err := ensureVersion(cmd.Context(), client, flags.kind, ver, versionsapi.GranularityMajor, log); err != nil {
		return err
	}

	if err := ensureVersion(cmd.Context(), client, flags.kind, ver, versionsapi.GranularityMinor, log); err != nil {
		return err
	}

	if flags.latest {
		if err := updateLatest(cmd.Context(), client, flags.kind, ver, log); err != nil {
			return fmt.Errorf("setting latest version: %w", err)
		}
	}

	log.Infof("List major->minor URL: %s", ver.ListURL(versionsapi.GranularityMajor))
	log.Infof("List minor->patch URL: %s", ver.ListURL(versionsapi.GranularityMinor))

	return nil
}

func ensureVersion(ctx context.Context, client *verclient.VersionsClient, kind versionsapi.VersionKind, ver versionsapi.Version, gran versionsapi.Granularity,
	log *logger.Logger,
) error {
	verListReq := versionsapi.List{
		Ref:         ver.Ref,
		Stream:      ver.Stream,
		Granularity: gran,
		Base:        ver.WithGranularity(gran),
		Kind:        kind,
	}
	verList, err := client.FetchVersionList(ctx, verListReq)
	var notFoundErr *apiclient.NotFoundError
	if errors.As(err, &notFoundErr) {
		log.Infof("Version list for %s versions under %q does not exist. Creating new list", gran.String(), ver.Major())
		verList = verListReq
	} else if err != nil {
		return fmt.Errorf("failed to list minor versions: %w", err)
	}
	log.Debugf("%s version list: %v", gran.String(), verList)

	insertGran := gran + 1
	insertVersion := ver.WithGranularity(insertGran)

	if verList.Contains(insertVersion) {
		log.Infof("Version %q already exists in list %v", insertVersion, verList.Versions)
		return nil
	}
	log.Infof("Inserting %s version %q into list", insertGran.String(), insertVersion)

	verList.Versions = append(verList.Versions, insertVersion)
	log.Debugf("New %s version list: %v", gran.String(), verList)

	if err := client.UpdateVersionList(ctx, verList); err != nil {
		return fmt.Errorf("failed to add %s version: %w", gran.String(), err)
	}

	log.Infof("Added %q to list", insertVersion)
	return nil
}

func updateLatest(ctx context.Context, client *verclient.VersionsClient, kind versionsapi.VersionKind, ver versionsapi.Version, log *logger.Logger) error {
	latest := versionsapi.Latest{
		Ref:    ver.Ref,
		Stream: ver.Stream,
		Kind:   kind,
	}
	latest, err := client.FetchVersionLatest(ctx, latest)
	var notFoundErr *apiclient.NotFoundError
	if errors.As(err, &notFoundErr) {
		log.Debugf("Latest version for ref %q and stream %q not found", ver.Ref, ver.Stream)
	} else if err != nil {
		return fmt.Errorf("fetching latest version: %w", err)
	}

	if latest.Version == ver.Version {
		log.Infof("Version %q is already latest version", ver)
		return nil
	}

	log.Infof("Setting %q as latest version", ver)
	latest = versionsapi.Latest{
		Ref:     ver.Ref,
		Stream:  ver.Stream,
		Version: ver.Version,
		Kind:    kind,
	}
	if err := client.UpdateVersionLatest(ctx, latest); err != nil {
		return fmt.Errorf("updating latest version: %w", err)
	}

	return nil
}

type addFlags struct {
	version        string
	stream         string
	ref            string
	release        bool
	latest         bool
	dryRun         bool
	region         string
	bucket         string
	distributionID string
	kind           versionsapi.VersionKind
	logLevel       zapcore.Level
}

func (f *addFlags) validate(log *logger.Logger) error {
	if !semver.IsValid(f.version) {
		return fmt.Errorf("version %q is not a valid semantic version", f.version)
	}
	if semver.Canonical(f.version) != f.version {
		return fmt.Errorf("version %q is not a canonical semantic version", f.version)
	}

	if f.ref == "" && !f.release {
		return fmt.Errorf("either --ref or --release must be set")
	}

	if f.kind == versionsapi.VersionKindUnknown {
		return fmt.Errorf("unknown version kind %q", f.kind)
	}

	if f.release {
		log.Debugf("Setting ref to %q, as release flag is set", versionsapi.ReleaseRef)
		f.ref = versionsapi.ReleaseRef
	} else {
		log.Debugf("Setting latest to true, as release flag is not set")
		f.latest = true // always set latest for non-release versions
	}

	if err := versionsapi.ValidateRef(f.ref); err != nil {
		return fmt.Errorf("invalid ref %w", err)
	}

	if err := versionsapi.ValidateStream(f.ref, f.stream); err != nil {
		return fmt.Errorf("invalid stream %w", err)
	}

	return nil
}

func parseAddFlags(cmd *cobra.Command) (addFlags, error) {
	ref, err := cmd.Flags().GetString("ref")
	if err != nil {
		return addFlags{}, err
	}
	ref = versionsapi.CanonicalizeRef(ref)
	stream, err := cmd.Flags().GetString("stream")
	if err != nil {
		return addFlags{}, err
	}
	kindFlag, err := cmd.Flags().GetString("kind")
	if err != nil {
		return addFlags{}, err
	}
	kind := versionsapi.VersionKindFromString(kindFlag)
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return addFlags{}, err
	}
	release, err := cmd.Flags().GetBool("release")
	if err != nil {
		return addFlags{}, err
	}
	latest, err := cmd.Flags().GetBool("latest")
	if err != nil {
		return addFlags{}, err
	}
	dryRun, err := cmd.Flags().GetBool("dryrun")
	if err != nil {
		return addFlags{}, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return addFlags{}, err
	}
	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
	}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return addFlags{}, err
	}
	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return addFlags{}, err
	}
	distributionID, err := cmd.Flags().GetString("distribution-id")
	if err != nil {
		return addFlags{}, err
	}

	return addFlags{
		version:        version,
		stream:         stream,
		ref:            versionsapi.CanonicalizeRef(ref),
		release:        release,
		latest:         latest,
		dryRun:         dryRun,
		region:         region,
		bucket:         bucket,
		distributionID: distributionID,
		logLevel:       logLevel,
		kind:           kind,
	}, nil
}
