/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package uplosi implements uploading os images using uplosi.
package uplosi

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/osimage"
)

//go:embed uplosi.conf.in
var uplosiConfigTemplate string

const timestampFormat = "20060102150405"

// Uploader can upload os images using uplosi.
type Uploader struct {
	uplosiPath string

	log *logger.Logger
}

// New creates a new Uploader.
func New(uplosiPath string, log *logger.Logger) *Uploader {
	return &Uploader{
		uplosiPath: uplosiPath,
		log:        log,
	}
}

// Upload uploads the given os image using uplosi.
func (u *Uploader) Upload(ctx context.Context, req *osimage.UploadRequest) ([]versionsapi.ImageInfoEntry, error) {
	config, err := prepareUplosiConfig(req)
	if err != nil {
		return nil, err
	}

	workspace, err := prepareWorkspace(config)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(workspace)

	uplosiOutput, err := runUplosi(ctx, u.uplosiPath, workspace, req.ImagePath)
	if err != nil {
		return nil, err
	}

	return parseUplosiOutput(uplosiOutput, req.Provider, req.AttestationVariant)
}

func prepareUplosiConfig(req *osimage.UploadRequest) ([]byte, error) {
	var config map[string]any
	if _, err := toml.Decode(uplosiConfigTemplate, &config); err != nil {
		return nil, err
	}

	imageVersionStr, err := imageVersion(req.Provider, req.Version, req.Timestamp)
	if err != nil {
		return nil, err
	}
	baseConfig := config["base"].(map[string]any)
	awsConfig := baseConfig["aws"].(map[string]any)
	azureConfig := baseConfig["azure"].(map[string]any)
	gcpConfig := baseConfig["gcp"].(map[string]any)

	baseConfig["imageVersion"] = imageVersionStr
	baseConfig["provider"] = strings.ToLower(req.Provider.String())
	extendAWSConfig(awsConfig, req.Version, req.AttestationVariant, req.Timestamp)
	extendAzureConfig(azureConfig, req.Version, req.AttestationVariant, req.Timestamp)
	extendGCPConfig(gcpConfig, req.Version, req.AttestationVariant)

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(config); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func prepareWorkspace(config []byte) (string, error) {
	workspace, err := os.MkdirTemp("", "uplosi-")
	if err != nil {
		return "", err
	}
	// write config to workspace
	configPath := filepath.Join(workspace, "uplosi.conf")
	if err := os.WriteFile(configPath, config, 0o644); err != nil {
		return "", err
	}
	return workspace, nil
}

func runUplosi(ctx context.Context, uplosiPath string, workspace string, rawImage string) ([]byte, error) {
	imagePath, err := filepath.Abs(rawImage)
	if err != nil {
		return nil, err
	}

	uplosiCmd := exec.CommandContext(ctx, uplosiPath, "upload", imagePath)
	uplosiCmd.Dir = workspace
	uplosiCmd.Stderr = os.Stderr
	return uplosiCmd.Output()
}

func parseUplosiOutput(output []byte, csp cloudprovider.Provider, attestationVariant string) ([]versionsapi.ImageInfoEntry, error) {
	lines := strings.Split(string(output), "\n")
	var imageReferences []versionsapi.ImageInfoEntry
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var region, reference string
		if csp == cloudprovider.AWS {
			var err error
			region, reference, err = awsParseAMIARN(line)
			if err != nil {
				return nil, err
			}
		} else {
			reference = line
		}
		imageReferences = append(imageReferences, versionsapi.ImageInfoEntry{
			CSP:                strings.ToLower(csp.String()),
			AttestationVariant: attestationVariant,
			Reference:          reference,
			Region:             region,
		})
	}
	return imageReferences, nil
}

func imageVersion(csp cloudprovider.Provider, version versionsapi.Version, timestamp time.Time) (string, error) {
	cleanSemver := strings.TrimPrefix(regexp.MustCompile(`^v\d+\.\d+\.\d+`).FindString(version.Version()), "v")
	if csp != cloudprovider.Azure {
		return cleanSemver, nil
	}

	switch {
	case version.Stream() == "stable":
		fallthrough
	case version.Stream() == "debug" && version.Ref() == "-":
		return cleanSemver, nil
	}

	formattedTime := timestamp.Format(timestampFormat)
	if len(formattedTime) != len(timestampFormat) {
		return "", errors.New("invalid timestamp")
	}
	// <year>.<month><day>.<time>
	return formattedTime[:4] + "." + formattedTime[4:8] + "." + formattedTime[8:], nil
}

func extendAWSConfig(awsConfig map[string]any, version versionsapi.Version, attestationVariant string, timestamp time.Time) {
	awsConfig["amiName"] = awsAMIName(version, attestationVariant, timestamp)
	awsConfig["snapshotName"] = awsAMIName(version, attestationVariant, timestamp)
	awsConfig["blobName"] = fmt.Sprintf("image-%s-%s-%s-%d.raw", version.Stream(), version.Version(), attestationVariant, timestamp.Unix())
}

func awsAMIName(version versionsapi.Version, attestationVariant string, timestamp time.Time) string {
	if version.Stream() == "stable" {
		return fmt.Sprintf("constellation-%s-%s", version.Version(), attestationVariant)
	}
	return fmt.Sprintf("constellation-%s-%s-%s-%s", version.Stream(), version.Version(), attestationVariant, timestamp.Format(timestampFormat))
}

func awsParseAMIARN(arn string) (region string, amiID string, retErr error) {
	parts := strings.Split(arn, ":")
	if len(parts) != 6 {
		return "", "", fmt.Errorf("invalid ARN (expected 5 path components) %q", arn)
	}
	if parts[0] != "arn" {
		return "", "", fmt.Errorf("invalid ARN (prefix mismatch) %q", arn)
	}
	if parts[1] != "aws" {
		return "", "", fmt.Errorf("invalid ARN (provider mismatch) %q", arn)
	}
	if parts[2] != "ec2" {
		return "", "", fmt.Errorf("invalid ARN (service mismatch) %q", arn)
	}
	resourceParts := strings.Split(parts[5], "/")
	if len(resourceParts) != 2 {
		return "", "", fmt.Errorf("invalid ARN (expected resource type/id) %q", arn)
	}
	if resourceParts[0] != "image" {
		return "", "", fmt.Errorf("invalid ARN (resource type mismatch) %q", arn)
	}
	return parts[3], resourceParts[1], nil
}

func extendAzureConfig(azureConfig map[string]any, version versionsapi.Version, attestationVariant string, timestamp time.Time) {
	azureConfig["replicationRegions"] = azureReplicationRegions(attestationVariant)
	azureConfig["attestationVariant"] = attestationVariant
	azureConfig["sharedImageGallery"] = azureGalleryName(version, attestationVariant)
	azureConfig["imageDefinitionName"] = azureImageOffer(version)
	azureConfig["offer"] = azureImageOffer(version)
	formattedTime := timestamp.Format(timestampFormat)
	azureConfig["diskName"] = fmt.Sprintf("constellation-%s-%s-%s", version.Stream(), formattedTime, attestationVariant)
}

func azureGalleryName(version versionsapi.Version, attestationVariant string) string {
	var prefix string
	switch version.Stream() {
	case "stable":
		prefix = "Constellation"
	case "debug":
		prefix = "Constellation_Debug"
	default:
		prefix = "Constellation_Testing"
	}

	var suffix string
	switch attestationVariant {
	case "azure-tdx":
		suffix = "_TDX"
	case "azure-sev-snp":
		suffix = "_CVM"
	}
	return prefix + suffix
}

func azureImageOffer(version versionsapi.Version) string {
	switch {
	case version.Stream() == "stable":
		return "constellation"
	case version.Stream() == "debug" && version.Ref() == "-":
		return version.Version()
	}
	return version.Ref() + "-" + version.Stream()
}

func azureReplicationRegions(attestationVariant string) []string {
	switch attestationVariant {
	case "azure-tdx":
		return []string{"northeurope", "westeurope", "centralus", "eastus2"}
	case "azure-sev-snp":
		return []string{"northeurope", "westeurope", "germanywestcentral", "eastus", "westus", "southeastasia"}
	}
	return nil
}

func extendGCPConfig(gcpConfig map[string]any, version versionsapi.Version, attestationVariant string) {
	gcpConfig["imageFamily"] = gcpImageFamily(version)
	gcpConfig["imageName"] = gcpImageName(version, attestationVariant)
	gcpConfig["blobName"] = gcpImageName(version, attestationVariant) + ".tar.gz"
}

func gcpImageFamily(version versionsapi.Version) string {
	if version.Stream() == "stable" {
		return "constellation"
	}
	truncatedRef := version.Ref()
	if len(version.Ref()) > 45 {
		truncatedRef = version.Ref()[:45]
	}
	return "constellation-" + truncatedRef
}

func gcpImageName(version versionsapi.Version, attestationVariant string) string {
	return strings.ReplaceAll(version.Version(), ".", "-") + "-" + attestationVariant + "-" + version.Stream()
}
