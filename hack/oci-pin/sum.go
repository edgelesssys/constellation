/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/hack/oci-pin/internal/extract"
	"github.com/edgelesssys/constellation/v2/hack/oci-pin/internal/sums"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
)

func newSumCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sum",
		Short: "Generate sha256sum file that pins an OCI image.",
		RunE:  runSum,
	}

	cmd.Flags().String("oci-path", "", "Path to the OCI image to pin.")
	cmd.Flags().String("output", "-", "Output file. If not set, the output is written to stdout.")
	cmd.Flags().String("registry", "", "OCI registry to use.")
	cmd.Flags().String("prefix", "", "Prefix of the OCI image to pin.")
	cmd.Flags().String("image-name", "", "Short name (suffix) of the OCI image to pin.")
	cmd.Flags().String("image-tag", "", "Tag of the OCI image to pin. Optional.")
	cmd.Flags().String("image-tag-file", "", "Tag file of the OCI image to pin. Optional.")
	cmd.MarkFlagsMutuallyExclusive("image-tag", "image-tag-file")
	must(cmd.MarkFlagRequired("registry"))
	must(cmd.MarkFlagRequired("oci-path"))
	must(cmd.MarkFlagRequired("image-name"))

	return cmd
}

func runSum(cmd *cobra.Command, _ []string) error {
	flags, err := parseSumFlags(cmd)
	if err != nil {
		return err
	}
	log := logger.New(logger.PlainLog, flags.logLevel)
	log.Debugf("Parsed flags: %+v", flags)

	log.Debugf("Generating sum file for OCI image %s.", flags.imageName)

	ociIndexPath := filepath.Join(flags.ociPath, "index.json")
	index, err := os.Open(ociIndexPath)
	if err != nil {
		return fmt.Errorf("opening OCI index at %q: %w", ociIndexPath, err)
	}
	defer index.Close()

	var out io.Writer
	if flags.output == "-" {
		out = cmd.OutOrStdout()
	} else {
		f, err := os.Create(flags.output)
		if err != nil {
			return fmt.Errorf("creating output file %q: %w", flags.output, err)
		}
		defer f.Close()
		out = f
	}

	digest, err := extract.Digest(index)
	if err != nil {
		return fmt.Errorf("extracting OCI image digest: %w", err)
	}

	log.Debugf("OCI image digest: %s", digest)

	refs := []sums.PinnedImageReference{
		{
			Registry: flags.registry,
			Prefix:   flags.prefix,
			Name:     flags.imageName,
			Tag:      flags.imageTag,
			Digest:   digest,
		},
	}

	if err := sums.Create(refs, out); err != nil {
		return fmt.Errorf("creating sum file: %w", err)
	}

	log.Debugf("Sum file created at %q 🤖", flags.output)
	return nil
}

type sumFlags struct {
	ociPath   string
	output    string
	registry  string
	prefix    string
	imageName string
	imageTag  string
	logLevel  zapcore.Level
}

func parseSumFlags(cmd *cobra.Command) (sumFlags, error) {
	ociPath, err := cmd.Flags().GetString("oci-path")
	if err != nil {
		return sumFlags{}, err
	}
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return sumFlags{}, err
	}
	registry, err := cmd.Flags().GetString("registry")
	if err != nil {
		return sumFlags{}, err
	}
	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return sumFlags{}, err
	}
	imageName, err := cmd.Flags().GetString("image-name")
	if err != nil {
		return sumFlags{}, err
	}
	imageTag, err := cmd.Flags().GetString("image-tag")
	if err != nil {
		return sumFlags{}, err
	}
	imageTagFile, err := cmd.Flags().GetString("image-tag-file")
	if err != nil {
		return sumFlags{}, err
	}
	if imageTagFile != "" {
		tag, err := os.ReadFile(imageTagFile)
		if err != nil {
			return sumFlags{}, fmt.Errorf("reading image tag file %q: %w", imageTagFile, err)
		}
		imageTag = strings.TrimSpace(string(tag))
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return sumFlags{}, err
	}
	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
	}

	return sumFlags{
		ociPath:   ociPath,
		output:    output,
		registry:  registry,
		prefix:    prefix,
		imageName: imageName,
		imageTag:  imageTag,
		logLevel:  logLevel,
	}, nil
}
