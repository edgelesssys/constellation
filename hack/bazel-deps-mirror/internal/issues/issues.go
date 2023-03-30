/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package issues can store and report issues found during the bazel-deps-mirror process.
package issues

import (
	"fmt"
	"io"
	"sort"
)

// Map is a map of issues arranged by path => rulename => issues.
type Map map[string]map[string][]error

// New creates a new Map.
func New() Map {
	return make(map[string]map[string][]error)
}

// Set sets all issues for a file.
func (m Map) Set(file string, issues ByFile) {
	m[file] = issues
}

// Report prints all issues to a writer in a human-readable format.
func (m Map) Report(w io.Writer) {
	files := make([]string, 0, len(m))
	for f := range m {
		files = append(files, f)
	}
	sort.Strings(files)

	for _, file := range files {
		rules := make([]string, 0, len(m[file]))
		for r := range m[file] {
			rules = append(rules, r)
		}
		sort.Strings(rules)

		fmt.Fprintf(w, "File %s (%d issues total):\n", file, m.IssuesPerFile(file))
		for _, rule := range rules {
			ruleIssues := m[file][rule]
			if len(ruleIssues) == 0 {
				continue
			}
			fmt.Fprintf(w, "  Rule %s (%d issues total):\n", rule, m.IssuesPerRule(file, rule))
			for _, issue := range ruleIssues {
				fmt.Fprintf(w, "    %s\n", issue)
			}
		}
	}
}

// FileHasIssues returns true if the file has any issues.
func (m Map) FileHasIssues(file string) bool {
	return m[file] != nil
}

// IssuesPerFile returns the number of issues for a file.
func (m Map) IssuesPerFile(file string) int {
	sum := 0
	for _, ruleIssues := range m[file] {
		sum += len(ruleIssues)
	}
	return sum
}

// IssuesPerRule returns the number of issues for a rule.
func (m Map) IssuesPerRule(file string, rule string) int {
	return len(m[file][rule])
}

// ByFile is a map of issues belonging to one file arranged by rulename => issues.
type ByFile map[string][]error

// NewByFile creates a new ByFile.
func NewByFile() ByFile {
	return make(map[string][]error)
}

// Add adds one or more issues belonging to a rule.
func (m ByFile) Add(rule string, issues ...error) {
	m[rule] = append(m[rule], issues...)
}
