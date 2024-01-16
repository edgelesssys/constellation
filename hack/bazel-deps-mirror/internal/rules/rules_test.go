/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package rules

import (
	"testing"

	"github.com/bazelbuild/buildtools/build"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestRules(t *testing.T) {
	assert := assert.New(t)
	const bzlFileContents = `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")
load("@bazeldnf//:deps.bzl", "rpm")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = ["https://example.com/foo.tar.gz"],
)

http_file(
	name = "bar_file",
	sha256 = "fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9",
	urls = ["https://example.com/bar"],
)

rpm(
	name = "baz_rpm",
	sha256 = "9e7ab438597fee20e16e8e441bed0ce966bd59e0fb993fa7c94be31fb1384d88",
	urls = ["https://example.com/baz.rpm"],
)

git_repository(
	name = "qux_git",
	remote = "https://example.com/qux.git",
	commit = "1234567890abcdef",
)
`
	bf, err := build.Parse("foo.bzl", []byte(bzlFileContents))
	if err != nil {
		t.Fatal(err)
	}

	rules := Rules(bf, SupportedRules)
	assert.Len(rules, 3)
	expectedNames := []string{"foo_archive", "bar_file", "baz_rpm"}
	for i, rule := range rules {
		assert.Equal(expectedNames[i], rule.Name())
	}

	allRules := Rules(bf, nil)
	assert.Len(allRules, 4)
	expectedNames = []string{"foo_archive", "bar_file", "baz_rpm", "qux_git"}
	for i, rule := range allRules {
		assert.Equal(expectedNames[i], rule.Name())
	}
}

func TestValidatePinned(t *testing.T) {
	testCases := map[string]struct {
		rule               string
		expectedIssueCount int
	}{
		"no issues, singular url": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	url = "https://example.com/foo.tar.gz",
)
`,
			expectedIssueCount: 0,
		},
		"no issues, url list": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = ["https://example.com/foo.tar.gz"],
)
`,
			expectedIssueCount: 0,
		},
		"no issues, url list with multiple urls": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = ["https://example.com/foo.tar.gz", "https://example.com/foo2.tar.gz"],
)
`,
			expectedIssueCount: 0,
		},
		"missing name": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	url = "https://example.com/foo.tar.gz",
)
`,
			expectedIssueCount: 1,
		},
		"missing sha256 attribute": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	url = "https://example.com/foo.tar.gz",
)
`,
			expectedIssueCount: 1,
		},
		"missing url attribute": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
)
`,
			expectedIssueCount: 1,
		},
		"url and urls attribute given": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	url = "https://example.com/foo.tar.gz",
	urls = ["https://example.com/foo.tar.gz"],
)
`,
			expectedIssueCount: 1,
		},
		"empty url attribute": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	url = "",
)
`,
			expectedIssueCount: 1,
		},
		"empty urls attribute": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = [],
)
`,
			expectedIssueCount: 1,
		},
		"empty url in urls attribute": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = [""],
)
`,
			expectedIssueCount: 1,
		},
		"empty sha256 attribute": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "",
	url = "https://example.com/foo.tar.gz",
)
`,
			expectedIssueCount: 1,
		},
		"missing all attributes": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
)
`,
			expectedIssueCount: 2,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			bf, err := build.Parse("foo.bzl", []byte(tc.rule))
			if err != nil {
				t.Fatal(err)
			}

			rules := Rules(bf, SupportedRules)
			require.Len(rules, 1)

			issues := ValidatePinned(rules[0])
			if tc.expectedIssueCount == 0 {
				assert.Nil(issues)
				return
			}
			assert.Len(issues, tc.expectedIssueCount)
		})
	}
}

func TestCheckNormalize(t *testing.T) {
	testCases := map[string]struct {
		rule               string
		expectedIssueCount int
		cannotFix          bool
	}{
		"rule with single url": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	url = "https://cdn.confidential.cloud/constellation/cas/sha256/2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	type = "tar.gz",
)
`,
			expectedIssueCount: 1,
		},
		"rule with unsorted urls": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = [
		"https://example.com/a/foo.tar.gz",
		"https://example.com/b/foo.tar.gz",
		"https://cdn.confidential.cloud/constellation/cas/sha256/2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
		"https://mirror.bazel.build/example.com/a/foo.tar.gz",
	],
	type = "tar.gz",
)
`,
			expectedIssueCount: 1,
		},
		"rule that is not mirrored": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = ["https://example.com/foo.tar.gz"],
	type = "tar.gz",
)
`,
			expectedIssueCount: 1,
			cannotFix:          true,
		},
		"http_archive with no type": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = [
		"https://cdn.confidential.cloud/constellation/cas/sha256/2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
		"https://example.com/foo.tar.gz",
	],
)
`,
			expectedIssueCount: 1,
		},
		"rpm rule with urls that are not the mirror": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

rpm(
	name = "foo_rpm",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = [
		"https://cdn.confidential.cloud/constellation/cas/sha256/2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
		"https://example.com/foo.rpm",
	],
)
`,
			expectedIssueCount: 1,
		},
		"http_archive rule that is correct": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = ["https://cdn.confidential.cloud/constellation/cas/sha256/2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"],
	type = "tar.gz",
)
`,
			expectedIssueCount: 0,
		},
		"rpm rule that is correct": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

rpm(
	name = "foo_rpm",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = ["https://cdn.confidential.cloud/constellation/cas/sha256/2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"],
)
`,
			expectedIssueCount: 0,
		},
		"http_file rule that is correct": {
			rule: `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

http_file(
	name = "foo_file",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = ["https://cdn.confidential.cloud/constellation/cas/sha256/2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"],
)
`,
			expectedIssueCount: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			bf, err := build.Parse("foo.bzl", []byte(tc.rule))
			if err != nil {
				t.Fatal(err)
			}

			rules := Rules(bf, SupportedRules)
			require.Len(rules, 1)

			issues := Check(rules[0])
			if tc.expectedIssueCount == 0 {
				assert.Nil(issues)
				return
			}
			assert.Equal(len(issues), tc.expectedIssueCount)

			changed := Normalize(rules[0])
			if tc.expectedIssueCount > 0 && !tc.cannotFix {
				assert.True(changed)
			} else {
				assert.False(changed)
			}
			if tc.cannotFix {
				assert.NotNil(Check(rules[0]))
			} else {
				assert.Nil(Check(rules[0]))
			}
		})
	}
}

func TestAddURLs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	rule := `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
)
`
	bf, err := build.Parse("foo.bzl", []byte(rule))
	if err != nil {
		t.Fatal(err)
	}
	rules := Rules(bf, SupportedRules)
	require.Len(rules, 1)

	AddURLs(rules[0], []string{"https://example.com/a", "https://example.com/b"})
	assert.Equal([]string{"https://example.com/a", "https://example.com/b"}, GetURLs(rules[0]))
}

func TestGetHash(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	rule := `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
)

http_archive(
	name = "bar_archive",
)
`
	bf, err := build.Parse("foo.bzl", []byte(rule))
	if err != nil {
		t.Fatal(err)
	}
	rules := Rules(bf, SupportedRules)
	require.Len(rules, 2)

	hash, err := GetHash(rules[0])
	assert.NoError(err)
	assert.Equal("2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae", hash)

	_, err = GetHash(rules[1])
	assert.Error(err)
}

func TestPrepareUpgrade(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	rule := `
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
	name = "foo_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = [
		"https://cdn.confidential.cloud/constellation/cas/sha256/2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
		"https://mirror.bazel.build/example.com/foo.tar.gz",
		"https://example.com/foo.tar.gz",
	],
	type = "tar.gz",
)

http_archive(
	name = "bar_archive",
	sha256 = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	urls = [
		"https://cdn.confidential.cloud/constellation/cas/sha256/2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
		"https://mirror.bazel.build/example.com/foo.tar.gz",
	],
	type = "tar.gz",
)
`
	bf, err := build.Parse("foo.bzl", []byte(rule))
	if err != nil {
		t.Fatal(err)
	}
	rules := Rules(bf, SupportedRules)
	require.Len(rules, 2)

	changed, err := PrepareUpgrade(rules[0])
	assert.NoError(err)
	assert.True(changed)

	urls := GetURLs(rules[0])
	assert.Equal(1, len(urls))
	assert.Equal("https://example.com/foo.tar.gz", urls[0])
	hash, err := GetHash(rules[0])
	assert.Empty(hash)
	assert.Error(err)

	changed, err = PrepareUpgrade(rules[1])
	assert.ErrorIs(err, ErrNoUpstreamURL)
	assert.False(changed)
}
