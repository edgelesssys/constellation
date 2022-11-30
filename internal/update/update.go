/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"golang.org/x/mod/semver"
)

// VersionsList represents a list of versions for a kind of resource.
// It has a granularity of either "major" or "minor".
//
// For example, a VersionsList with granularity "major" could contain
// the base version "v1" and a list of minor versions "v1.0", "v1.1", "v1.2" etc.
// A VersionsList with granularity "minor" could contain the base version
// "v1.0" and a list of patch versions "v1.0.0", "v1.0.1", "v1.0.2" etc.
type VersionsList struct {
	// Stream is the update stream of the list.
	// Currently, only "stable" is supported.
	Stream string `json:"stream"`
	// Granularity is the granularity of the base version of this list.
	// It can be either "major" or "minor".
	Granularity string `json:"granularity"`
	// Base is the base version of the list.
	// Every version in the list is a finer-grained version of this base version.
	Base string `json:"base"`
	// Kind is the kind of resource this list is for.
	Kind string `json:"kind"`
	// Versions is a list of all versions in this list.
	Versions []string `json:"versions"`
}

// Validate checks if the list is valid.
// This performs the following checks:
// - The stream is supported.
// - The granularity is "major" or "minor".
// - The kind is supported.
// - The base version is a valid semantic version that matches the granularity.
// - All versions in the list are valid semantic versions that are finer-grained than the base version.
func (l *VersionsList) Validate() error {
	var issues []string
	if l.Stream != "stable" {
		issues = append(issues, fmt.Sprintf("stream %q is not supported", l.Stream))
	}
	if l.Granularity != "major" && l.Granularity != "minor" {
		issues = append(issues, fmt.Sprintf("granularity %q is not supported", l.Granularity))
	}
	if l.Kind != "image" {
		issues = append(issues, fmt.Sprintf("kind %q is not supported", l.Kind))
	}
	if !semver.IsValid(l.Base) {
		issues = append(issues, fmt.Sprintf("base version %q is not a valid semantic version", l.Base))
	}
	var normalizeFunc func(string) string
	switch l.Granularity {
	case "major":
		normalizeFunc = semver.Major
	case "minor":
		normalizeFunc = semver.MajorMinor
	default:
		normalizeFunc = func(s string) string { return s }
	}
	if normalizeFunc(l.Base) != l.Base {
		issues = append(issues, fmt.Sprintf("base version %q is not a %v version", l.Base, l.Granularity))
	}
	for _, ver := range l.Versions {
		if !semver.IsValid(ver) {
			issues = append(issues, fmt.Sprintf("version %q in list is not a valid semantic version", ver))
		}
		if normalizeFunc(ver) != l.Base {
			issues = append(issues, fmt.Sprintf("version %q in list is not a finer-grained version of base version %q", ver, l.Base))
		}
	}
	if len(issues) > 0 {
		return fmt.Errorf("version list is invalid:\n%s", strings.Join(issues, "\n"))
	}
	return nil
}

// Contains returns true if the list contains the given version.
func (l *VersionsList) Contains(version string) bool {
	for _, v := range l.Versions {
		if v == version {
			return true
		}
	}
	return false
}

// VersionsFetcher fetches a list of versions.
type VersionsFetcher struct {
	httpc httpc
}

// New returns a new VersionsFetcher.
func New() *VersionsFetcher {
	return &VersionsFetcher{
		httpc: http.DefaultClient,
	}
}

// MinorVersionsOf fetches the list of minor versions for a given stream, major version and kind.
func (f *VersionsFetcher) MinorVersionsOf(ctx context.Context, stream, major, kind string) (*VersionsList, error) {
	return f.list(ctx, stream, "major", major, kind)
}

// PatchVersionsOf fetches the list of patch versions for a given stream, minor version and kind.
func (f *VersionsFetcher) PatchVersionsOf(ctx context.Context, stream, minor, kind string) (*VersionsList, error) {
	return f.list(ctx, stream, "minor", minor, kind)
}

// list fetches the list of versions for a given stream, granularity, base and kind.
func (f *VersionsFetcher) list(ctx context.Context, stream, granularity, base, kind string) (*VersionsList, error) {
	raw, err := getFromURL(ctx, f.httpc, stream, granularity, base, kind)
	if err != nil {
		return nil, fmt.Errorf("fetching versions list: %w", err)
	}
	list := &VersionsList{}
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("decoding versions list: %w", err)
	}
	if err := list.Validate(); err != nil {
		return nil, fmt.Errorf("validating versions list: %w", err)
	}
	if !f.listMatchesRequest(list, stream, granularity, base, kind) {
		return nil, fmt.Errorf("versions list does not match request")
	}
	return list, nil
}

func (f *VersionsFetcher) listMatchesRequest(list *VersionsList, stream, granularity, base, kind string) bool {
	return list.Stream == stream && list.Granularity == granularity && list.Base == base && list.Kind == kind
}

// getFromURL fetches the versions list from a URL.
func getFromURL(ctx context.Context, client httpc, stream, granularity, base, kind string) ([]byte, error) {
	url, err := url.Parse(constants.CDNRepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("parsing image version repository URL: %w", err)
	}
	kindFilename := path.Base(kind) + ".json"
	url.Path = path.Join(constants.CDNVersionsPath, "stream", stream, granularity, base, kindFilename)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return nil, fmt.Errorf("versions list %q does not exist", url.String())
		default:
			return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
		}
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

type httpc interface {
	Do(req *http.Request) (*http.Response, error)
}
