/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package issues can store and report issues found during the bazel-deps-mirror process.
package issues

import (
	"errors"
	"fmt"
	"io"
	"sort"
)

// Map is a map of issues ordered by file path and rule name.
type Map map[string]map[string]error

// New creates a new Map.
func New() Map {
	return make(map[string]map[string]error)
}

// Add adds an issue to the Map.
func (m Map) Add(file string, rule string, issue error) {
	if m[file] == nil {
		m[file] = make(map[string]error)
	}
	m[file][rule] = errors.Join(m[file][rule], issue)
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
			fmt.Fprintf(w, "  Rule %s (%d issues total):\n", rule, m.IssuesPerRule(file, rule))
			for _, issue := range getErrorsFromMultiErr(m[file][rule]) {
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
		sum += countMultiErr(ruleIssues)
	}
	return sum
}

// IssuesPerRule returns the number of issues for a rule.
func (m Map) IssuesPerRule(file string, rule string) int {
	return countMultiErr(m[file][rule])
}

func getErrorsFromMultiErr(issues error) []error {
	if issues == nil {
		return nil
	}

	if me, ok := issues.(multiErr); ok {
		errs := me.Unwrap()
		var subErrs []error
		for _, err := range errs {
			subErrs = append(subErrs, getErrorsFromMultiErr(err)...)
		}
		return subErrs
	}

	return []error{issues}
}

func countMultiErr(issues error) int {
	return len(getErrorsFromMultiErr(issues))
}

type multiErr interface {
	Unwrap() []error
}
