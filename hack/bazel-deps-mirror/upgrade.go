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

func newUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "upgrade all Bazel dependency rules by recalculating expected hashes, uploading artifacts to the mirror (if needed) and formatting the rules.",
		RunE:  runUpgrade,
	}

	cmd.Flags().Bool("unauthenticated", false, "Doesn't require authentication to the mirror but cannot upload files.")
	cmd.Flags().Bool("dry-run", false, "Don't actually change files or upload anything.")

	return cmd
}

func runUpgrade(cmd *cobra.Command, _ []string) error {
	flags, err := parseUpgradeFlags(cmd)
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
		log.Warnf("Upgrading rules without authentication for AWS S3. If artifacts are not yet mirrored, this will fail.")
		mirrorUpload = mirror.NewUnauthenticated(flags.mirrorBaseURL, flags.dryRun, log)
	default:
		log.Debugf("Upgrading rules with authentication for AWS S3.")
		mirrorUpload, err = mirror.New(cmd.Context(), flags.region, flags.bucket, flags.mirrorBaseURL, flags.dryRun, log)
		if err != nil {
			return err
		}
	}

	issues := issues.New()
	for _, bazelFile := range bazelFiles {
		fileIssues, err := upgradeBazelFile(cmd.Context(), fileHelper, mirrorUpload, bazelFile, flags.dryRun, log)
		if err != nil {
			return err
		}
		if len(fileIssues) > 0 {
			issues.Set(bazelFile.AbsPath, fileIssues)
		}
	}
	if len(issues) > 0 {
		log.Warnf("Found %d issues in rules", len(issues))
		issues.Report(cmd.OutOrStdout())
		return errors.New("found issues in rules")
	}

	log.Infof("No issues found")
	return nil
}

func upgradeBazelFile(ctx context.Context, fileHelper *bazelfiles.Helper, mirrorUpload mirrorUploader, bazelFile bazelfiles.BazelFile, dryRun bool, log *logger.Logger) (iss issues.ByFile, err error) {
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
		changedRule, ruleIssues := upgradeRule(ctx, mirrorUpload, rule, log)
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

func upgradeRule(ctx context.Context, mirrorUpload mirrorUploader, rule *build.Rule, log *logger.Logger) (changed bool, iss []error) {
	log.Debugf("Upgrading rule: %s", rule.Name())

	upstreamURLs, err := rules.UpstreamURLs(rule)
	if errors.Is(err, rules.ErrNoUpstreamURL) {
		log.Debugf("Rule has no upstream URL. Skipping.")
		return false, nil
	} else if err != nil {
		iss = append(iss, err)
		return false, iss
	}

	// learn the hash of the upstream artifact
	learnedHash, learnErr := mirrorUpload.Learn(ctx, upstreamURLs)
	if learnErr != nil {
		iss = append(iss, learnErr)
		return false, iss
	}

	existingHash, err := rules.GetHash(rule)
	if err == nil && learnedHash == existingHash {
		log.Debugf("Rule already upgraded. Skipping.")
		return false, nil
	}

	changed, err = rules.PrepareUpgrade(rule)
	if err != nil {
		iss = append(iss, err)
		return changed, iss
	}
	rules.SetHash(rule, learnedHash)
	changed = true

	if _, fixErr := fixRule(ctx, mirrorUpload, rule, log); fixErr != nil {
		iss = append(iss, fixErr...)
		return changed, iss
	}
	return changed, iss
}

type upgradeFlags struct {
	unauthenticated bool
	dryRun          bool
	region          string
	bucket          string
	mirrorBaseURL   string
	logLevel        zapcore.Level
}

func parseUpgradeFlags(cmd *cobra.Command) (upgradeFlags, error) {
	unauthenticated, err := cmd.Flags().GetBool("unauthenticated")
	if err != nil {
		return upgradeFlags{}, err
	}
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return upgradeFlags{}, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return upgradeFlags{}, err
	}
	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
	}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return upgradeFlags{}, err
	}
	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return upgradeFlags{}, err
	}
	mirrorBaseURL, err := cmd.Flags().GetString("mirror-base-url")
	if err != nil {
		return upgradeFlags{}, err
	}

	return upgradeFlags{
		unauthenticated: unauthenticated,
		dryRun:          dryRun,
		region:          region,
		bucket:          bucket,
		mirrorBaseURL:   mirrorBaseURL,
		logLevel:        logLevel,
	}, nil
}
