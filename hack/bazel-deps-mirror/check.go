/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"errors"

	"github.com/edgelesssys/constellation/v2/hack/bazel-deps-mirror/internal/bazelfiles"
	"github.com/edgelesssys/constellation/v2/hack/bazel-deps-mirror/internal/issues"
	"github.com/edgelesssys/constellation/v2/hack/bazel-deps-mirror/internal/mirror"
	"github.com/edgelesssys/constellation/v2/hack/bazel-deps-mirror/internal/rules"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
)

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check if all Bazel dependencies are mirrored and the corresponding rules are properly formatted.",
		RunE:  runCheck,
	}

	cmd.Flags().Bool("mirror", false, "Performs authenticated checks to validate if all referenced CAS objects are still consistent within the mirror.")
	cmd.Flags().Bool("mirror-unauthenticated", false, "Performs unauthenticated, slow checks to validate if all referenced CAS objects are still consistent within the mirror. Doesn't require authentication to the mirror but may be inefficient.")
	cmd.MarkFlagsMutuallyExclusive("mirror", "mirror-unauthenticated")

	return cmd
}

func runCheck(cmd *cobra.Command, _ []string) error {
	flags, err := parseCheckFlags(cmd)
	if err != nil {
		return err
	}
	log := logger.New(logger.PlainLog, flags.logLevel)
	log.Debugf("Parsed flags: %+v", flags)

	filesHelper, err := bazelfiles.New()
	if err != nil {
		return err
	}

	log.Debugf("Searching for Bazel files in the current WORKSPACE and all subdirectories...")
	bazelFiles, err := filesHelper.FindFiles()
	if err != nil {
		return err
	}

	var mirrorCheck mirrorChecker
	switch {
	case flags.mirrorUnauthenticated:
		log.Debugf("Checking consistency of all referenced CAS objects without authentication.")
		mirrorCheck = mirror.NewUnauthenticated(flags.mirrorBaseURL, mirror.Run, log)
	case flags.mirror:
		log.Debugf("Checking consistency of all referenced CAS objects using AWS S3.")
		mirrorCheck, err = mirror.New(cmd.Context(), flags.region, flags.bucket, flags.mirrorBaseURL, mirror.Run, log)
		if err != nil {
			return err
		}
	default:
		mirrorCheck = &noOpMirrorChecker{}
	}

	iss := issues.New()
	for _, bazelFile := range bazelFiles {
		issByFile, err := checkBazelFile(cmd.Context(), filesHelper, mirrorCheck, bazelFile, log)
		if err != nil {
			return err
		}
		if len(issByFile) > 0 {
			iss.Set(bazelFile.AbsPath, issByFile)
		}
	}
	if len(iss) > 0 {
		log.Infof("Found issues in rules")
		iss.Report(cmd.OutOrStdout())
		return errors.New("found issues in rules")
	}

	log.Infof("No issues found 🦭")
	return nil
}

func checkBazelFile(ctx context.Context, fileHelper *bazelfiles.Helper, mirrorCheck mirrorChecker, bazelFile bazelfiles.BazelFile, log *logger.Logger) (issByFile issues.ByFile, err error) {
	log.Debugf("Checking file: %s", bazelFile.RelPath)
	issByFile = issues.NewByFile()
	buildfile, err := fileHelper.LoadFile(bazelFile)
	if err != nil {
		return nil, err
	}
	found := rules.Rules(buildfile, rules.SupportedRules)
	if len(found) == 0 {
		log.Debugf("No rules found in file: %s", bazelFile.RelPath)
		return
	}
	log.Debugf("Found %d rules in file: %s", len(found), bazelFile.RelPath)
	for _, rule := range found {
		log.Debugf("Checking rule: %s", rule.Name())
		// check if the rule is a valid pinned dependency rule (has all required attributes)
		if issues := rules.ValidatePinned(rule); len(issues) > 0 {
			issByFile.Add(rule.Name(), issues...)
			continue
		}
		// check if the rule is a valid mirror rule
		if issues := rules.Check(rule); len(issues) > 0 {
			issByFile.Add(rule.Name(), issues...)
		}

		// check if the referenced CAS object is still consistent
		// may be a no-op if --check-consistency is not set
		expectedHash, expectedHashErr := rules.GetHash(rule)
		if expectedHashErr == nil && rules.HasMirrorURL(rule) {
			if issue := mirrorCheck.Check(ctx, expectedHash); issue != nil {
				issByFile.Add(rule.Name(), issue)
			}
		}
	}
	return
}

type checkFlags struct {
	mirrorUnauthenticated bool
	mirror                bool
	region                string
	bucket                string
	mirrorBaseURL         string
	logLevel              zapcore.Level
}

func parseCheckFlags(cmd *cobra.Command) (checkFlags, error) {
	mirrorUnauthenticated, err := cmd.Flags().GetBool("mirror-unauthenticated")
	if err != nil {
		return checkFlags{}, err
	}
	mirror, err := cmd.Flags().GetBool("mirror")
	if err != nil {
		return checkFlags{}, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return checkFlags{}, err
	}
	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
	}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return checkFlags{}, err
	}
	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return checkFlags{}, err
	}
	mirrorBaseURL, err := cmd.Flags().GetString("mirror-base-url")
	if err != nil {
		return checkFlags{}, err
	}

	return checkFlags{
		mirrorUnauthenticated: mirrorUnauthenticated,
		mirror:                mirror,
		region:                region,
		bucket:                bucket,
		mirrorBaseURL:         mirrorBaseURL,
		logLevel:              logLevel,
	}, nil
}

type mirrorChecker interface {
	Check(ctx context.Context, expectedHash string) error
}

type noOpMirrorChecker struct{}

func (m *noOpMirrorChecker) Check(ctx context.Context, expectedHash string) error {
	return nil
}
