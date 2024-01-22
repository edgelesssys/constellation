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
	"github.com/edgelesssys/constellation/v2/hack/oci-pin/internal/inject"
	"github.com/spf13/cobra"
)

func newCodegenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codegen",
		Short: "Generate Go code that pins an OCI image.",
		RunE:  runCodegen,
	}

	cmd.Flags().String("oci-path", "", "Path to the OCI image to pin.")
	cmd.Flags().String("output", "-", "Output file. If not set, the output is written to stdout.")
	cmd.Flags().String("package", "", "Name of the Go package.")
	cmd.Flags().String("identifier", "", "Base name of the Go const identifiers.")
	cmd.Flags().String("repoimage-tag-file", "", "Tag file of the OCI image to pin.")
	must(cmd.MarkFlagRequired("oci-path"))
	must(cmd.MarkFlagRequired("package"))
	must(cmd.MarkFlagRequired("identifier"))
	must(cmd.MarkFlagRequired("repoimage-tag-file"))

	return cmd
}

func runCodegen(cmd *cobra.Command, _ []string) error {
	flags, err := parseCodegenFlags(cmd)
	if err != nil {
		return err
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: flags.logLevel}))
	log.Debug(fmt.Sprintf("Parsed flags: %+v", flags))

	registry, prefix, name, tag, err := splitRepoTag(flags.imageRepoTag)
	if err != nil {
		return fmt.Errorf("splitting OCI image reference %q: %w", flags.imageRepoTag, err)
	}

	log.Debug(fmt.Sprintf("Generating Go code for OCI image %s.", name))

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
		return err
	}

	log.Debug(fmt.Sprintf("OCI image digest: %s", digest))

	if err := inject.Render(out, inject.PinningValues{
		Package:  flags.pkg,
		Ident:    flags.identifier,
		Registry: registry,
		Prefix:   prefix,
		Name:     name,
		Tag:      tag,
		Digest:   digest,
	}); err != nil {
		return fmt.Errorf("rendering Go code: %w", err)
	}

	log.Debug(fmt.Sprintf("Go code created at %q 🤖", flags.output))
	return nil
}

type codegenFlags struct {
	ociPath      string
	output       string
	pkg          string
	identifier   string
	imageRepoTag string
	logLevel     slog.Level
}

func parseCodegenFlags(cmd *cobra.Command) (codegenFlags, error) {
	ociPath, err := cmd.Flags().GetString("oci-path")
	if err != nil {
		return codegenFlags{}, err
	}
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return codegenFlags{}, err
	}
	pkg, err := cmd.Flags().GetString("package")
	if err != nil {
		return codegenFlags{}, err
	}
	identifier, err := cmd.Flags().GetString("identifier")
	if err != nil {
		return codegenFlags{}, err
	}

	imageRepoTagFile, err := cmd.Flags().GetString("repoimage-tag-file")
	if err != nil {
		return codegenFlags{}, err
	}
	repotag, err := os.ReadFile(imageRepoTagFile)
	if err != nil {
		return codegenFlags{}, fmt.Errorf("reading image repotag file %q: %w", imageRepoTagFile, err)
	}
	imageRepoTag := strings.TrimSpace(string(repotag))

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return codegenFlags{}, err
	}
	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}

	return codegenFlags{
		ociPath:      ociPath,
		output:       output,
		pkg:          pkg,
		identifier:   identifier,
		imageRepoTag: imageRepoTag,
		logLevel:     logLevel,
	}, nil
}
