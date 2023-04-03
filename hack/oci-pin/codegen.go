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
	"github.com/edgelesssys/constellation/v2/hack/oci-pin/internal/inject"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
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
	cmd.Flags().String("image-registry", "", "Registry where the image is stored.")
	cmd.Flags().String("image-prefix", "", "Prefix of the image name. Optional.")
	cmd.Flags().String("image-name", "", "Short name of the OCI image to pin.")
	cmd.Flags().String("image-tag", "", "Tag of the OCI image to pin. Optional.")
	cmd.Flags().String("image-tag-file", "", "Tag file of the OCI image to pin. Optional.")
	cmd.MarkFlagsMutuallyExclusive("image-tag", "image-tag-file")
	must(cmd.MarkFlagRequired("oci-path"))
	must(cmd.MarkFlagRequired("package"))
	must(cmd.MarkFlagRequired("identifier"))
	must(cmd.MarkFlagRequired("image-registry"))
	must(cmd.MarkFlagRequired("image-name"))

	return cmd
}

func runCodegen(cmd *cobra.Command, _ []string) error {
	flags, err := parseCodegenFlags(cmd)
	if err != nil {
		return err
	}
	log := logger.New(logger.PlainLog, flags.logLevel)
	log.Debugf("Parsed flags: %+v", flags)

	log.Debugf("Generating Go code for OCI image %s.", flags.imageName)

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

	log.Debugf("OCI image digest: %s", digest)

	if err := inject.Render(out, inject.PinningValues{
		Package:  flags.pkg,
		Ident:    flags.identifier,
		Registry: flags.imageRegistry,
		Prefix:   flags.imagePrefix,
		Name:     flags.imageName,
		Tag:      flags.imageTag,
		Digest:   digest,
	}); err != nil {
		return fmt.Errorf("rendering Go code: %w", err)
	}

	log.Debugf("Go code created at %q 🤖", flags.output)
	return nil
}

type codegenFlags struct {
	ociPath       string
	output        string
	pkg           string
	identifier    string
	imageRegistry string
	imagePrefix   string
	imageName     string
	imageTag      string
	logLevel      zapcore.Level
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
	imageRegistry, err := cmd.Flags().GetString("image-registry")
	if err != nil {
		return codegenFlags{}, err
	}
	imagePrefix, err := cmd.Flags().GetString("image-prefix")
	if err != nil {
		return codegenFlags{}, err
	}
	imageName, err := cmd.Flags().GetString("image-name")
	if err != nil {
		return codegenFlags{}, err
	}
	imageTag, err := cmd.Flags().GetString("image-tag")
	if err != nil {
		return codegenFlags{}, err
	}
	imageTagFile, err := cmd.Flags().GetString("image-tag-file")
	if err != nil {
		return codegenFlags{}, err
	}
	if imageTagFile != "" {
		tag, err := os.ReadFile(imageTagFile)
		if err != nil {
			return codegenFlags{}, fmt.Errorf("reading image tag file %q: %w", imageTagFile, err)
		}
		imageTag = strings.TrimSpace(string(tag))
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return codegenFlags{}, err
	}
	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
	}

	return codegenFlags{
		ociPath:       ociPath,
		output:        output,
		pkg:           pkg,
		identifier:    identifier,
		imageRegistry: imageRegistry,
		imagePrefix:   imagePrefix,
		imageName:     imageName,
		imageTag:      imageTag,
		logLevel:      logLevel,
	}, nil
}
