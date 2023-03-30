/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"errors"

	"github.com/bazelbuild/buildtools/build"
	"github.com/edgelesssys/constellation/v2/hack/bazel-deps-mirror/internal/bazelfiles"
	"github.com/edgelesssys/constellation/v2/hack/bazel-deps-mirror/internal/issues"
	"github.com/edgelesssys/constellation/v2/hack/bazel-deps-mirror/internal/mirror"
	"github.com/edgelesssys/constellation/v2/hack/bazel-deps-mirror/internal/rules"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
)

func newFixCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fix",
		Short: "fix all Bazel dependency rules by uploading artifacts to the mirror (if needed) and formatting the rules.",
		RunE:  runFix,
	}

	cmd.Flags().Bool("unauthenticated", false, "Doesn't require authentication to the mirror but cannot upload files.")
	cmd.Flags().Bool("dry-run", false, "Don't actually change files or upload anything.")

	return cmd
}

func runFix(cmd *cobra.Command, _ []string) error {
	flags, err := parseFixFlags(cmd)
	if err != nil {
		return err
	}
	log := logger.New(logger.PlainLog, flags.logLevel)
	log.Debugf("Parsed flags: %+v", flags)

	fileHelper, err := bazelfiles.New()
	if err != nil {
		return err
	}

	log.Debugf("Searching for Bazel files in the current WORKSPACE and all subdirectories...")
	bazelFiles, err := fileHelper.FindFiles()
	if err != nil {
		return err
	}

	var mirrorUpload mirrorUploader
	switch {
	case flags.unauthenticated:
		log.Warnf("Fixing rules without authentication for AWS S3. If artifacts are not yet mirrored, this will fail.")
		mirrorUpload = mirror.NewUnauthenticated(flags.mirrorBaseURL, flags.dryRun, log)
	default:
		log.Debugf("Fixing rules with authentication for AWS S3.")
		mirrorUpload, err = mirror.New(cmd.Context(), flags.region, flags.bucket, flags.mirrorBaseURL, flags.dryRun, log)
		if err != nil {
			return err
		}
	}

	issues := issues.New()
	for _, bazelFile := range bazelFiles {
		fileIssues, err := fixBazelFile(cmd.Context(), fileHelper, mirrorUpload, bazelFile, flags.dryRun, log)
		if err != nil {
			return err
		}
		if len(fileIssues) > 0 {
			issues.Set(bazelFile.AbsPath, fileIssues)
		}
	}
	if len(issues) > 0 {
		log.Warnf("Found %d unfixable issues in rules", len(issues))
		issues.Report(cmd.OutOrStdout())
		return errors.New("found issues in rules")
	}

	log.Infof("No unfixable issues found")
	return nil
}

func fixBazelFile(ctx context.Context, fileHelper *bazelfiles.Helper, mirrorUpload mirrorUploader, bazelFile bazelfiles.BazelFile, dryRun bool, log *logger.Logger) (iss issues.ByFile, err error) {
	iss = issues.NewByFile()
	var changed bool // true if any rule in this file was changed
	log.Infof("Checking file: %s", bazelFile.RelPath)
	buildfile, err := fileHelper.LoadFile(bazelFile)
	if err != nil {
		return iss, err
	}
	found := rules.Rules(buildfile, rules.SupportedRules)
	if len(found) == 0 {
		log.Debugf("No rules found in file: %s", bazelFile.RelPath)
		return iss, nil
	}
	log.Debugf("Found %d rules in file: %s", len(found), bazelFile.RelPath)
	for _, rule := range found {
		changedRule, ruleIssues := fixRule(ctx, mirrorUpload, rule, log)
		if len(ruleIssues) > 0 {
			iss.Add(rule.Name(), ruleIssues...)
		}
		changed = changed || changedRule
	}

	if len(iss) > 0 {
		log.Warnf("File %s has issues. Not saving!", bazelFile.RelPath)
		return iss, nil
	}
	if !changed {
		log.Debugf("No changes to file: %s", bazelFile.RelPath)
		return iss, nil
	}
	if dryRun {
		diff, err := fileHelper.Diff(bazelFile, buildfile)
		if err != nil {
			return iss, err
		}
		log.Infof("Dry run: would save updated file %s with diff:\n%s", bazelFile.RelPath, diff)
		return iss, nil
	}
	log.Infof("Saving updated file: %s", bazelFile.RelPath)
	if err := fileHelper.WriteFile(bazelFile, buildfile); err != nil {
		return iss, err
	}

	return iss, nil
}

func fixRule(ctx context.Context, mirrorUpload mirrorUploader, rule *build.Rule, log *logger.Logger) (changed bool, iss []error) {
	log.Debugf("Fixing rule: %s", rule.Name())
	// check if the rule is a valid pinned dependency rule (has all required attributes)
	issue := rules.ValidatePinned(rule)
	if issue != nil {
		// don't try to fix the rule if it's invalid
		iss = append(iss, issue...)
		return
	}

	// check if the referenced CAS object exists in the mirror and is consistent
	expectedHash, expectedHashErr := rules.GetHash(rule)
	if expectedHashErr != nil {
		// don't try to fix the rule if the hash is missing
		iss = append(iss,
			errors.New("hash attribute is missing. unable to check if the artifact is already mirrored or upload it"))
		return
	}

	if rules.HasMirrorURL(rule) {
		changed = rules.Normalize(rule)
		return
	}

	log.Infof("Artifact %s with hash %s is not yet mirrored. Uploading...", rule.Name(), expectedHash)
	if uploadErr := mirrorUpload.Mirror(ctx, expectedHash, rules.GetURLs(rule)); uploadErr != nil {
		// don't try to fix the rule if the upload failed
		iss = append(iss, uploadErr)
		return
	}
	// now the artifact is mirrored (if it wasn't already) and we can fix the rule
	mirrorURL, err := mirrorUpload.MirrorURL(expectedHash)
	if err != nil {
		iss = append(iss, err)
		return
	}
	rules.AddURLs(rule, []string{mirrorURL})

	// normalize the rule
	rules.Normalize(rule)
	return true, iss
}

type fixFlags struct {
	unauthenticated bool
	dryRun          bool
	region          string
	bucket          string
	mirrorBaseURL   string
	logLevel        zapcore.Level
}

func parseFixFlags(cmd *cobra.Command) (fixFlags, error) {
	unauthenticated, err := cmd.Flags().GetBool("unauthenticated")
	if err != nil {
		return fixFlags{}, err
	}
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return fixFlags{}, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fixFlags{}, err
	}
	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
	}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return fixFlags{}, err
	}
	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return fixFlags{}, err
	}
	mirrorBaseURL, err := cmd.Flags().GetString("mirror-base-url")
	if err != nil {
		return fixFlags{}, err
	}

	return fixFlags{
		unauthenticated: unauthenticated,
		dryRun:          dryRun,
		region:          region,
		bucket:          bucket,
		mirrorBaseURL:   mirrorBaseURL,
		logLevel:        logLevel,
	}, nil
}

type mirrorUploader interface {
	Check(ctx context.Context, expectedHash string) error
	Mirror(ctx context.Context, hash string, urls []string) error
	MirrorURL(hash string) (string, error)
}
