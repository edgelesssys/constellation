/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versionsapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// ReleaseRef is the ref used for release versions.
const ReleaseRef = "-"

// Version represents a version. A version has a ref, stream, version string and kind.
//
// Notice that version is a meta object to the versions API and there isn't an
// actual corresponding object in the S3 bucket.
// Therefore, the version object doesn't have a URL or JSON path.
//
// Versions fields are private so the type can be used in other packages by
// defining private interfaces.
type Version struct {
	ref     string
	stream  string
	version string
	kind    VersionKind
}

// NewVersion creates a new Version object and validates it.
func NewVersion(ref, stream, version string, kind VersionKind) (Version, error) {
	ver := Version{
		ref:     CanonicalizeRef(ref),
		stream:  stream,
		version: version,
		kind:    kind,
	}

	if err := ver.Validate(); err != nil {
		return Version{}, err
	}

	return ver, nil
}

// NewVersionFromShortPath creates a new Version from a version short path.
func NewVersionFromShortPath(shortPath string, kind VersionKind) (Version, error) {
	ref, stream, version, err := parseShortPath(shortPath)
	if err != nil {
		return Version{}, err
	}

	ver := Version{
		ref:     ref, // Canonicalized by parseShortPath.
		stream:  stream,
		version: version,
		kind:    kind,
	}

	if err := ver.Validate(); err != nil {
		return Version{}, err
	}

	return ver, nil
}

// Ref returns the ref of the version.
func (v Version) Ref() string {
	return v.ref
}

// Stream returns the stream of the version.
func (v Version) Stream() string {
	return v.stream
}

// Version returns the version string of the version.
func (v Version) Version() string {
	return v.version
}

// Kind returns the kind of the version.
func (v Version) Kind() VersionKind {
	return v.kind
}

// ShortPath returns the short path of the version.
func (v Version) ShortPath() string {
	return shortPath(v.ref, v.stream, v.version)
}

// Validate validates the version.
func (v Version) Validate() error {
	var retErr error
	if err := ValidateRef(v.ref); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if err := ValidateStream(v.ref, v.stream); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if !semver.IsValid(v.version) {
		retErr = errors.Join(retErr, fmt.Errorf("version %q is not a valid semantic version", v.version))
	}
	if semver.Canonical(v.version) != v.version {
		retErr = errors.Join(retErr, fmt.Errorf("version %q is not a canonical semantic version", v.version))
	}
	if v.kind == VersionKindUnknown {
		retErr = errors.Join(retErr, errors.New("version kind is unknown"))
	}
	return retErr
}

// Equal returns true if the versions are equal.
func (v Version) Equal(other Version) bool {
	return v.ref == other.ref &&
		v.stream == other.stream &&
		v.version == other.version &&
		v.kind == other.kind
}

// Major returns the major version corresponding to the version.
// For example, if the version is "v1.2.3", the major version is "v1".
func (v Version) Major() string {
	return semver.Major(v.version)
}

// MajorMinor returns the major and minor version corresponding to the version.
// For example, if the version is "v1.2.3", the major and minor version is "v1.2".
func (v Version) MajorMinor() string {
	return semver.MajorMinor(v.version)
}

// WithGranularity returns the version with the given granularity.
//
// For example, if the version is "v1.2.3" and the granularity is GranularityMajor,
// the returned version is "v1".
// This is a helper function for Major() and MajorMinor() and v.version.
// In case of an unknown granularity, an empty string is returned.
func (v Version) WithGranularity(gran Granularity) string {
	switch gran {
	case GranularityMajor:
		return v.Major()
	case GranularityMinor:
		return v.MajorMinor()
	case GranularityPatch:
		return v.version
	default:
		return ""
	}
}

// ListURL returns the URL of the list with the given granularity,
// this version is listed in.
// For example, assing GranularityMajor returns the URL of the versions list
// that maps the major version of this version to its minor version.
// In case of an unknown granularity, an empty string is returned.
func (v Version) ListURL(gran Granularity) string {
	if gran == GranularityUnknown || gran == GranularityPatch {
		return ""
	}
	return constants.CDNRepositoryURL + "/" + v.ListPath(gran)
}

// ListPath returns the path of the list with the given granularity,
// this version is listed in.
// For example, assing GranularityMajor returns the path of the versions list
// that maps the major version of this version to its minor version.
// In case of an unknown granularity, an empty string is returned.
func (v Version) ListPath(gran Granularity) string {
	if gran == GranularityUnknown || gran == GranularityPatch {
		return ""
	}
	return path.Join(
		constants.CDNAPIPrefix,
		"ref", v.ref,
		"stream", v.stream,
		"versions",
		gran.String(), v.WithGranularity(gran),
		v.kind.String()+".json",
	)
}

// ArtifactsURL returns the URL to the artifacts stored for this version.
// The URL points to a directory.
func (v Version) ArtifactsURL(apiVer apiVersion) string {
	return constants.CDNRepositoryURL + "/" + v.ArtifactPath(apiVer)
}

// ArtifactPath returns the path to the artifacts stored for this version.
// The path points to a directory.
func (v Version) ArtifactPath(apiVer apiVersion) string {
	return path.Join(
		constants.CDNAPIBase,
		apiVer.String(),
		"ref", v.ref,
		"stream", v.stream,
		v.version,
	)
}

// VersionKind represents the kind of resource the version versions.
type VersionKind int

const (
	// VersionKindUnknown is the default value for VersionKind.
	VersionKindUnknown VersionKind = iota
	// VersionKindImage is the kind for image versions.
	VersionKindImage
	// VersionKindCLI is the kind for CLI versions.
	VersionKindCLI
)

// MarshalJSON marshals the VersionKind to JSON.
func (k VersionKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

// UnmarshalJSON unmarshals the VersionKind from JSON.
func (k *VersionKind) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*k = VersionKindFromString(s)
	return nil
}

// String returns the string representation of the VersionKind.
func (k VersionKind) String() string {
	switch k {
	case VersionKindImage:
		return "image"
	case VersionKindCLI:
		return "cli"
	default:
		return "unknown"
	}
}

// VersionKindFromString returns the VersionKind for the given string.
func VersionKindFromString(s string) VersionKind {
	switch strings.ToLower(s) {
	case "image":
		return VersionKindImage
	case "cli":
		return VersionKindCLI
	default:
		return VersionKindUnknown
	}
}

// Granularity represents the granularity of a semantic version.
type Granularity int

const (
	// GranularityUnknown is the default granularity.
	GranularityUnknown Granularity = iota
	// GranularityMajor is the granularity of a major version, e.g. "v1".
	// Lists with granularity "major" map from a major version to a list of minor versions.
	GranularityMajor
	// GranularityMinor is the granularity of a minor version, e.g. "v1.0".
	// Lists with granularity "minor" map from a minor version to a list of patch versions.
	GranularityMinor
	// GranularityPatch is the granularity of a patch version, e.g. "v1.0.0".
	// There are no lists with granularity "patch".
	GranularityPatch
)

// MarshalJSON marshals the granularity to JSON.
func (g Granularity) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.String())
}

// UnmarshalJSON unmarshals the granularity from JSON.
func (g *Granularity) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*g = GranularityFromString(s)
	return nil
}

// String returns the string representation of the granularity.
func (g Granularity) String() string {
	switch g {
	case GranularityMajor:
		return "major"
	case GranularityMinor:
		return "minor"
	case GranularityPatch:
		return "patch"
	default:
		return "unknown"
	}
}

// GranularityFromString returns the granularity for the given string.
func GranularityFromString(s string) Granularity {
	switch strings.ToLower(s) {
	case "major":
		return GranularityMajor
	case "minor":
		return GranularityMinor
	case "patch":
		return GranularityPatch
	default:
		return GranularityUnknown
	}
}

var notAZ09Regexp = regexp.MustCompile("[^a-zA-Z0-9-]")

// CanonicalizeRef returns the canonicalized ref for the given ref.
func CanonicalizeRef(ref string) string {
	if ref == ReleaseRef {
		return ref
	}

	canRef := notAZ09Regexp.ReplaceAllString(ref, "-")

	if canRef == ReleaseRef {
		return "" // No ref should be canonicalized to the release ref.
	}

	return canRef
}

// ValidateRef checks if the given ref is a valid ref.
// Canonicalize the ref before checking if it is valid.
func ValidateRef(ref string) error {
	if ref == "" {
		return errors.New("ref must not be empty")
	}

	if notAZ09Regexp.FindString(ref) != "" {
		return errors.New("ref must only contain alphanumeric characters and dashes")
	}

	if strings.HasPrefix(ref, "refs-heads") {
		return errors.New("ref must not start with 'refs-heads'")
	}

	return nil
}

// ValidateStream checks if the given stream is a valid stream for the given ref.
func ValidateStream(ref, stream string) error {
	validReleaseStreams := []string{"stable", "console", "debug"}
	validStreams := []string{"nightly", "console", "debug"}

	if ref == ReleaseRef {
		validStreams = validReleaseStreams
	}

	for _, validStream := range validStreams {
		if stream == validStream {
			return nil
		}
	}

	return fmt.Errorf("stream %q is unknown or not supported on ref %q", stream, ref)
}

// MeasurementURL builds the measurement and signature URLs for the given version.
func MeasurementURL(version Version) (measurementURL, signatureURL *url.URL, err error) {
	if version.kind != VersionKindImage {
		return &url.URL{}, &url.URL{}, fmt.Errorf("kind %q is not supported", version.kind)
	}

	measurementPath, err := url.JoinPath(version.ArtifactsURL(APIV2), "image", constants.CDNMeasurementsFile)
	if err != nil {
		return &url.URL{}, &url.URL{}, fmt.Errorf("joining path for measurement: %w", err)
	}
	signaturePath, err := url.JoinPath(version.ArtifactsURL(APIV2), "image", constants.CDNMeasurementsSignature)
	if err != nil {
		return &url.URL{}, &url.URL{}, fmt.Errorf("joining path for signature: %w", err)
	}

	measurementURL, err = url.Parse(measurementPath)
	if err != nil {
		return &url.URL{}, &url.URL{}, fmt.Errorf("parsing path for measurement: %w", err)
	}

	signatureURL, err = url.Parse(signaturePath)
	if err != nil {
		return &url.URL{}, &url.URL{}, fmt.Errorf("parsing path for signature: %w", err)
	}
	return measurementURL, signatureURL, nil
}

var (
	shortPathRegex        = regexp.MustCompile(`^ref/([^/]+)/stream/([a-zA-Z0-9-]+)/([a-zA-Z0-9.-]+)$`)
	shortPathReleaseRegex = regexp.MustCompile(`^stream/([a-zA-Z0-9-]+)/([a-zA-Z0-9.-]+)$`)
)

func shortPath(ref, stream, version string) string {
	var sp string
	if ref != ReleaseRef {
		return path.Join("ref", ref, "stream", stream, version)
	}

	if stream != "stable" {
		return path.Join(sp, "stream", stream, version)
	}

	return version
}

func parseShortPath(shortPath string) (ref, stream, version string, err error) {
	if shortPathRegex.MatchString(shortPath) {
		matches := shortPathRegex.FindStringSubmatch(shortPath)
		ref := matches[1]
		ref = CanonicalizeRef(ref)
		if err := ValidateRef(ref); err != nil {
			return "", "", "", err
		}
		stream := matches[2]
		if err := ValidateStream(ref, stream); err != nil {
			return "", "", "", err
		}
		version := matches[3]
		if !semver.IsValid(version) {
			return "", "", "", fmt.Errorf("invalid version %q", version)
		}

		return ref, stream, version, nil
	}

	if shortPathReleaseRegex.MatchString(shortPath) {
		matches := shortPathReleaseRegex.FindStringSubmatch(shortPath)
		stream := matches[1]
		if err := ValidateStream(ref, stream); err != nil {
			return "", "", "", err
		}
		version := matches[2]
		if !semver.IsValid(version) {
			return "", "", "", fmt.Errorf("invalid version %q", version)
		}
		return ReleaseRef, stream, version, nil
	}

	if semver.IsValid(shortPath) {
		return ReleaseRef, "stable", shortPath, nil
	}

	return "", "", "", fmt.Errorf("invalid short path %q", shortPath)
}
