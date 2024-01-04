/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
)

type s3Flags struct {
	region         string
	bucket         string
	distributionID string
	logLevel       zapcore.Level
}

func parseS3Flags(cmd *cobra.Command) (s3Flags, error) {
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return s3Flags{}, err
	}
	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return s3Flags{}, err
	}
	distributionID, err := cmd.Flags().GetString("distribution-id")
	if err != nil {
		return s3Flags{}, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return s3Flags{}, err
	}
	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
	}

	return s3Flags{
		region:         region,
		bucket:         bucket,
		distributionID: distributionID,
		logLevel:       logLevel,
	}, nil
}

type measurementsFlags struct {
	s3Flags
	measurementsPath string
	signaturePath    string
}

func parseUploadMeasurementsFlags(cmd *cobra.Command) (measurementsFlags, error) {
	s3, err := parseS3Flags(cmd)
	if err != nil {
		return measurementsFlags{}, err
	}

	measurementsPath, err := cmd.Flags().GetString("measurements")
	if err != nil {
		return measurementsFlags{}, err
	}
	signaturePath, err := cmd.Flags().GetString("signature")
	if err != nil {
		return measurementsFlags{}, err
	}

	return measurementsFlags{
		s3Flags:          s3,
		measurementsPath: measurementsPath,
		signaturePath:    signaturePath,
	}, nil
}

type mergeMeasurementsFlags struct {
	out      string
	logLevel zapcore.Level
}

func parseMergeMeasurementsFlags(cmd *cobra.Command) (mergeMeasurementsFlags, error) {
	out, err := cmd.Flags().GetString("out")
	if err != nil {
		return mergeMeasurementsFlags{}, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return mergeMeasurementsFlags{}, err
	}
	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
	}

	return mergeMeasurementsFlags{
		out:      out,
		logLevel: logLevel,
	}, nil
}

type envelopeMeasurementsFlags struct {
	version            versionsapi.Version
	csp                cloudprovider.Provider
	attestationVariant string
	in, out            string
	logLevel           zapcore.Level
}

func parseEnvelopeMeasurementsFlags(cmd *cobra.Command) (envelopeMeasurementsFlags, error) {
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return envelopeMeasurementsFlags{}, err
	}
	ver, err := versionsapi.NewVersionFromShortPath(version, versionsapi.VersionKindImage)
	if err != nil {
		return envelopeMeasurementsFlags{}, err
	}
	csp, err := cmd.Flags().GetString("csp")
	if err != nil {
		return envelopeMeasurementsFlags{}, err
	}
	provider := cloudprovider.FromString(csp)
	attestationVariant, err := cmd.Flags().GetString("attestation-variant")
	if err != nil {
		return envelopeMeasurementsFlags{}, err
	}
	if provider == cloudprovider.Unknown {
		return envelopeMeasurementsFlags{}, errors.New("unknown cloud provider")
	}
	in, err := cmd.Flags().GetString("in")
	if err != nil {
		return envelopeMeasurementsFlags{}, err
	}
	out, err := cmd.Flags().GetString("out")
	if err != nil {
		return envelopeMeasurementsFlags{}, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return envelopeMeasurementsFlags{}, err
	}
	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
	}

	return envelopeMeasurementsFlags{
		version:            ver,
		csp:                provider,
		attestationVariant: attestationVariant,
		in:                 in,
		out:                out,
		logLevel:           logLevel,
	}, nil
}

type uplosiFlags struct {
	rawImage           string
	provider           cloudprovider.Provider
	version            versionsapi.Version
	attestationVariant string
	out                string
	uplosiPath         string

	// archiver flags
	region         string
	bucket         string
	distributionID string

	logLevel zapcore.Level
}

func parseUplosiFlags(cmd *cobra.Command) (uplosiFlags, error) {
	rawImage, err := cmd.Flags().GetString("raw-image")
	if err != nil {
		return uplosiFlags{}, err
	}
	rawImage, err = filepath.Abs(rawImage)
	if err != nil {
		return uplosiFlags{}, err
	}

	// extract csp, attestation variant, and stream from the raw image name
	// format: <csp>_<attestation-variant>_<stream>/constellation.raw

	var guessedCSP, guessedAttestationVariant, guessedStream string
	pathComponents := strings.Split(filepath.ToSlash(rawImage), "/")
	if len(pathComponents) >= 2 && pathComponents[len(pathComponents)-1] == "constellation.raw" {
		imageMetadata := pathComponents[len(pathComponents)-2]
		imageMetadataComponents := strings.Split(imageMetadata, "_")
		if len(imageMetadataComponents) == 3 {
			guessedCSP = imageMetadataComponents[0]
			guessedAttestationVariant = imageMetadataComponents[1]
			guessedStream = imageMetadataComponents[2]
		}
	}

	csp, err := cmd.Flags().GetString("csp")
	if err != nil || csp == "" {
		csp = guessedCSP
	}
	if csp == "" {
		if err != nil {
			return uplosiFlags{}, err
		}
		return uplosiFlags{}, errors.New("csp is required")
	}
	provider := cloudprovider.FromString(csp)
	ref, err := cmd.Flags().GetString("ref")
	if err != nil {
		return uplosiFlags{}, err
	}
	stream, err := cmd.Flags().GetString("stream")
	if err != nil || stream == "" {
		stream = guessedStream
	}
	if stream == "" {
		if err != nil {
			return uplosiFlags{}, err
		}
		return uplosiFlags{}, errors.New("stream is required")
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return uplosiFlags{}, err
	}
	ver, err := versionsapi.NewVersion(ref, stream, version, versionsapi.VersionKindImage)
	if err != nil {
		return uplosiFlags{}, err
	}
	attestationVariant, err := cmd.Flags().GetString("attestation-variant")
	if err != nil || attestationVariant == "" {
		attestationVariant = guessedAttestationVariant
	}
	if attestationVariant == "" {
		if err != nil {
			return uplosiFlags{}, err
		}
		return uplosiFlags{}, errors.New("attestation-variant is required")
	}
	out, err := cmd.Flags().GetString("out")
	if err != nil {
		return uplosiFlags{}, err
	}
	uplosiPath, err := cmd.Flags().GetString("uplosi-path")
	if err != nil {
		return uplosiFlags{}, err
	}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return uplosiFlags{}, err
	}
	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return uplosiFlags{}, err
	}
	distributionID, err := cmd.Flags().GetString("distribution-id")
	if err != nil {
		return uplosiFlags{}, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return uplosiFlags{}, err
	}
	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
	}

	return uplosiFlags{
		rawImage:           rawImage,
		provider:           provider,
		version:            ver,
		attestationVariant: attestationVariant,
		out:                out,
		uplosiPath:         uplosiPath,
		region:             region,
		bucket:             bucket,
		distributionID:     distributionID,
		logLevel:           logLevel,
	}, nil
}
