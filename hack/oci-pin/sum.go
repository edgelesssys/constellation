/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/hack/oci-pin/internal/extract"
	"github.com/edgelesssys/constellation/v2/hack/oci-pin/internal/sums"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/cobra"
)

func newSumCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sum",
		Short: "Generate sha256sum file that pins an OCI image.",
		RunE:  runSum,
	}

	cmd.Flags().String("oci-path", "", "Path to the OCI image to pin.")
	cmd.Flags().String("output", "-", "Output file. If not set, the output is written to stdout.")
	cmd.Flags().String("image-name", "", "Short name (suffix) of the OCI image to pin.")
	cmd.Flags().String("repoimage-tag-file", "", "Tag file of the OCI image to pin.")
	must(cmd.MarkFlagRequired("repoimage-tag-file"))
	must(cmd.MarkFlagRequired("oci-path"))

	return cmd
}

func runSum(cmd *cobra.Command, _ []string) error {
	flags, err := parseSumFlags(cmd)
	if err != nil {
		return err
	}
	log := logger.NewTextLogger(flags.logLevel)
	log.Debug(fmt.Sprintf(
`Parsed flags:
  image-repo-tag: %q
  oci-path: %q`
flags.imageRepoTag, flags.ociPath))

	registry, prefix, name, tag, err := splitRepoTag(flags.imageRepoTag)
	if err != nil {
		return fmt.Errorf("splitting repo tag: %w", err)
	}

	log.Debug(fmt.Sprintf("Generating sum file for OCI image %q.", name))

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

	log.Debug(fmt.Sprintf("OCI image digest: %q", digest))

	refs := []sums.PinnedImageReference{
		{
			Registry: registry,
			Prefix:   prefix,
			Name:     name,
			Tag:      tag,
			Digest:   digest,
		},
	}

	if err := sums.Create(refs, out); err != nil {
		return fmt.Errorf("creating sum file: %w", err)
	}

	log.Debug(fmt.Sprintf("Sum file created at %q 🤖", flags.output))
	return nil
}

type sumFlags struct {
	ociPath      string
	output       string
	imageRepoTag string
	logLevel     slog.Level
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

	imageTagFile, err := cmd.Flags().GetString("repoimage-tag-file")
	if err != nil {
		return sumFlags{}, err
	}
	tag, err := os.ReadFile(imageTagFile)
	if err != nil {
		return sumFlags{}, fmt.Errorf("reading image repotag file %q: %w", imageTagFile, err)
	}
	imageRepoTag := strings.TrimSpace(string(tag))

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return sumFlags{}, err
	}
	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}

	return sumFlags{
		ociPath:      ociPath,
		output:       output,
		imageRepoTag: imageRepoTag,
		logLevel:     logLevel,
	}, nil
}
