/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
)

const fileSeparator = "<!--AUTO_GENERATED_BY_BAZEL-->\n<!--DO_NOT_EDIT-->\n"

func main() {
	filePath := flag.String("file-path", "../../docs/docs/architecture/versions.md", "path to the version file to update")
	flag.Parse()

	k8sVersionStrings := versions.SupportedK8sVersions()
	var k8sVersions []semver.Semver
	for _, k8sVersionString := range k8sVersionStrings {
		k8sVersion, err := semver.New(k8sVersionString)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid kubernetes version %q: %s", k8sVersionString, err)
			os.Exit(1)
		}
		k8sVersions = append(k8sVersions, k8sVersion)
	}

	if err := updateDocFile(*filePath, k8sVersions); err != nil {
		fmt.Fprintf(os.Stderr, "error updating versions file: %s\n", err)
		os.Exit(1)
	}
}

func updateDocFile(filePath string, supportedVersions []semver.Semver) error {
	fileHeader, err := readVersionsFile(filePath)
	if err != nil {
		return err
	}

	var versionList strings.Builder
	for _, version := range supportedVersions {
		if _, err := versionList.WriteString(
			fmt.Sprintf("* %s\n", version.String()),
		); err != nil {
			return fmt.Errorf("writing matrix doc file: %w", err)
		}
	}

	return os.WriteFile(filePath, []byte(fileHeader+fileSeparator+versionList.String()), 0o644)
}

func readVersionsFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening version info file: %w", err)
	}
	defer f.Close()

	fileContentRaw, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("reading version info file: %w", err)
	}
	fileContent := strings.Split(string(fileContentRaw), fileSeparator)
	return fileContent[0], nil
}
