/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package rules is used find and modify Bazel rules in WORKSPACE and bzl files.
package rules

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/bazelbuild/buildtools/build"
	"golang.org/x/exp/slices"
)

// Rules is used to find and modify Bazel rules of a set of rule kinds in WORKSPACE and .bzl files.
// Filter is a list of rule kinds to consider.
// If filter is empty, all rules are considered.
func Rules(file *build.File, filter []string) (rules []*build.Rule) {
	allRules := file.Rules("")
	if len(filter) == 0 {
		return allRules
	}
ruleLoop:
	for _, rule := range allRules {
		for _, ruleKind := range filter {
			if rule.Kind() == ruleKind {
				rules = append(rules, rule)
				continue ruleLoop
			}
		}
	}
	return
}

// ValidatePinned checks if the given rule is a pinned dependency rule.
// That is, if it has a name, either a url or urls attribute, and a sha256 attribute.
// For archive rules, it also checks if the type attribute is set.
func ValidatePinned(rule *build.Rule) (validationErrs error) {
	if rule.Name() == "" {
		validationErrs = errors.Join(validationErrs, fmt.Errorf("rule %v has no name", rule))
	}

	hasURL := rule.Attr("url") != nil
	hasURLs := rule.Attr("urls") != nil
	if !hasURL && !hasURLs {
		validationErrs = errors.Join(validationErrs, fmt.Errorf("rule %s has no url or urls attribute", rule.Name()))
	}
	if hasURL && hasURLs {
		validationErrs = errors.Join(validationErrs, fmt.Errorf("rule %s has both url and urls attribute", rule.Name()))
	}
	if hasURL {
		url := rule.AttrString("url")
		if url == "" {
			validationErrs = errors.Join(validationErrs, fmt.Errorf("rule %s has empty url attribute", rule.Name()))
		}
	}
	if hasURLs {
		urls := rule.AttrStrings("urls")
		if len(urls) == 0 {
			validationErrs = errors.Join(validationErrs, fmt.Errorf("rule %s has empty urls list attribute", rule.Name()))
		} else {
			for _, url := range urls {
				if url == "" {
					validationErrs = errors.Join(validationErrs, fmt.Errorf("rule %s has empty url in urls attribute", rule.Name()))
				}
			}
		}
	}
	if rule.Attr("sha256") == nil {
		validationErrs = errors.Join(validationErrs, fmt.Errorf("rule %s has no sha256 attribute", rule.Name()))
	} else {
		sha256 := rule.AttrString("sha256")
		if sha256 == "" {
			validationErrs = errors.Join(validationErrs, fmt.Errorf("rule %s has empty sha256 attribute", rule.Name()))
		}
	}
	return
}

// Check checks if a dependency rule is normalized and contains a mirror url.
// All errors reported by this function can be fixed by calling AddURLs and Normalize.
func Check(rule *build.Rule) (validationErrs error) {
	hasURL := rule.Attr("url") != nil
	if hasURL {
		validationErrs = errors.Join(validationErrs, fmt.Errorf("rule %s has url (singular) attribute", rule.Name()))
	}
	urls := rule.AttrStrings("urls")
	sorted := make([]string, len(urls))
	copy(sorted, urls)
	sortURLs(sorted)
	for i, url := range urls {
		if url != sorted[i] {
			validationErrs = errors.Join(validationErrs, fmt.Errorf("rule %s has unsorted urls attributes: unsorted[%d] = %q, sorted[%d] = %q", rule.Name(), i, url, i, sorted[i]))
			break
		}
	}
	if !HasMirrorURL(rule) {
		validationErrs = errors.Join(validationErrs, fmt.Errorf("rule %s is not mirrored", rule.Name()))
	}
	if rule.Kind() == "http_archive" && rule.Attr("type") == nil {
		validationErrs = errors.Join(validationErrs, fmt.Errorf("http_archive rule %s has no type attribute", rule.Name()))
	}
	if rule.Kind() == "rpm" && len(urls) != 1 {
		validationErrs = errors.Join(validationErrs, fmt.Errorf("rpm rule %s has unstable urls that are not the edgeless mirror", rule.Name()))
	}
	return
}

// Normalize normalizes a rule and returns true if the rule was changed.
func Normalize(rule *build.Rule) (changed bool) {
	changed = addTypeAttribute(rule)
	urls := GetURLs(rule)
	normalizedURLS := append([]string{}, urls...)
	// rpm rules must have exactly one url (the edgeless mirror)
	if mirrorU, err := mirrorURL(rule); rule.Kind() == "rpm" && err == nil {
		normalizedURLS = []string{mirrorU}
	}
	sortURLs(normalizedURLS)
	normalizedURLS = deduplicateURLs(normalizedURLS)
	if slices.Equal(urls, normalizedURLS) && rule.Attr("url") == nil {
		return
	}
	setURLs(rule, normalizedURLS)
	changed = true
	return
}

// AddURLs adds a url to a rule.
func AddURLs(rule *build.Rule, urls []string) {
	existingURLs := GetURLs(rule)
	existingURLs = append(existingURLs, urls...)
	sortURLs(existingURLs)
	deduplicatedURLs := deduplicateURLs(existingURLs)
	setURLs(rule, deduplicatedURLs)
}

// GetHash returns the sha256 hash of a rule.
func GetHash(rule *build.Rule) (string, error) {
	hash := rule.AttrString("sha256")
	if hash == "" {
		return "", fmt.Errorf("rule %s has empty or missing sha256 attribute", rule.Name())
	}
	return hash, nil
}

// GetURLs returns the sorted urls of a rule.
func GetURLs(rule *build.Rule) []string {
	urls := rule.AttrStrings("urls")
	url := rule.AttrString("url")
	if url != "" {
		urls = append(urls, url)
	}
	return urls
}

// HasMirrorURL returns true if the rule has a url from the Edgeless mirror.
func HasMirrorURL(rule *build.Rule) bool {
	_, err := mirrorURL(rule)
	return err == nil
}

func deduplicateURLs(urls []string) (deduplicated []string) {
	seen := make(map[string]bool)
	for _, url := range urls {
		if !seen[url] {
			deduplicated = append(deduplicated, url)
			seen[url] = true
		}
	}
	return
}

// addTypeAttribute adds the type attribute to http_archive rules if it is missing.
// it returns true if the rule was changed.
// it returns an error if the rule does not have enough information to add the type attribute.
func addTypeAttribute(rule *build.Rule) bool {
	// only http_archive rules have a type attribute
	if rule.Kind() != "http_archive" {
		return false
	}
	if rule.Attr("type") != nil {
		return false
	}
	// iterate over all URLs and check if they have a known archive type
	var archiveType string
urlLoop:
	for _, url := range GetURLs(rule) {
		switch {
		case strings.HasSuffix(url, ".aar"):
			archiveType = "aar"
			break urlLoop
		case strings.HasSuffix(url, ".ar"):
			archiveType = "ar"
			break urlLoop
		case strings.HasSuffix(url, ".deb"):
			archiveType = "deb"
			break urlLoop
		case strings.HasSuffix(url, ".jar"):
			archiveType = "jar"
			break urlLoop
		case strings.HasSuffix(url, ".tar.bz2"):
			archiveType = "tar.bz2"
			break urlLoop
		case strings.HasSuffix(url, ".tar.gz"):
			archiveType = "tar.gz"
			break urlLoop
		case strings.HasSuffix(url, ".tar.xz"):
			archiveType = "tar.xz"
			break urlLoop
		case strings.HasSuffix(url, ".tar.zst"):
			archiveType = "tar.zst"
			break urlLoop
		case strings.HasSuffix(url, ".tar"):
			archiveType = "tar"
			break urlLoop
		case strings.HasSuffix(url, ".tgz"):
			archiveType = "tgz"
			break urlLoop
		case strings.HasSuffix(url, ".txz"):
			archiveType = "txz"
			break urlLoop
		case strings.HasSuffix(url, ".tzst"):
			archiveType = "tzst"
			break urlLoop
		case strings.HasSuffix(url, ".war"):
			archiveType = "war"
			break urlLoop
		case strings.HasSuffix(url, ".zip"):
			archiveType = "zip"
			break urlLoop
		}
	}
	if archiveType == "" {
		return false
	}
	rule.SetAttr("type", &build.StringExpr{Value: archiveType})
	return true
}

// mirrorURL returns the first mirror URL for a rule.
func mirrorURL(rule *build.Rule) (string, error) {
	urls := GetURLs(rule)
	for _, url := range urls {
		if strings.HasPrefix(url, edgelessMirrorPrefix) {
			return url, nil
		}
	}
	return "", fmt.Errorf("rule %s has no mirror url", rule.Name())
}

func setURLs(rule *build.Rule, urls []string) {
	// delete single url attribute if it exists
	rule.DelAttr("url")
	urlsAttr := []build.Expr{}
	for _, url := range urls {
		urlsAttr = append(urlsAttr, &build.StringExpr{Value: url})
	}
	rule.SetAttr("urls", &build.ListExpr{List: urlsAttr, ForceMultiLine: true})
}

func sortURLs(urls []string) {
	// Bazel mirror should be first
	// edgeless mirror should be second
	// other urls should be last
	// if there are multiple urls from the same mirror, they should be sorted alphabetically
	sort.Slice(urls, func(i, j int) bool {
		rank := func(url string) int {
			if strings.HasPrefix(url, bazelMirrorPrefix) {
				return 0
			}
			if strings.HasPrefix(url, edgelessMirrorPrefix) {
				return 1
			}
			return 2
		}
		if rank(urls[i]) != rank(urls[j]) {
			return rank(urls[i]) < rank(urls[j])
		}
		return urls[i] < urls[j]
	})
}

// SupportedRules is a list of all rules that can be mirrored.
var SupportedRules = []string{
	"http_archive",
	"http_file",
	"rpm",
}

const (
	bazelMirrorPrefix    = "https://mirror.bazel.build/"
	edgelessMirrorPrefix = "https://cdn.confidential.cloud/constellation/cas/sha256/"
)
