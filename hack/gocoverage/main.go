/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
gocoverage parses 'go test -cover' output and generates a simple coverage report in JSON format.
It can also be used to create a diff of two reports, filter for packages that were touched in
a PR, and print a markdown table with the diff.
*/
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

type packageCoverage struct {
	Coverage float64
	Notest   bool
	Nostmt   bool
}

func (c *packageCoverage) String() string {
	switch {
	case c == nil: // is nil if pkg was removed/added
		return fmt.Sprintf("%.2f%%", 0.0)
	case c.Notest:
		return "[no test files]"
	case c.Nostmt:
		return "[no statements]"
	default:
		return fmt.Sprintf("%.2f%%", c.Coverage)
	}
}

type metadata struct {
	Created string
}

type report struct {
	Metadate metadata
	Coverage map[string]*packageCoverage
}

func main() {
	dodiff := flag.Bool("diff", false, "diff reports")
	touched := flag.String("touched", "", "list of touched packages, comma separated")
	flag.Parse()
	switch {
	case *dodiff:
		if err := diff(flag.Arg(0), flag.Arg(1), *touched); err != nil {
			log.Fatal(err)
		}
	case len(flag.Args()) == 0:
		if err := parseStreaming(os.Stdin, os.Stdout); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Fprintln(os.Stderr, "gocoverage - generate coverage report from go test output and diff reports")
		fmt.Fprintln(os.Stderr, "usage:")
		fmt.Fprintln(os.Stderr, "  go test -coverage ./... | gocoverage > report.json")
		fmt.Fprintln(os.Stderr, "  gocoverage -touched pkgs/foo,pkg/bar -diff old.json new.json > diff.md")
	}
}

func parseStreaming(in io.Reader, out io.Writer) error {
	rep, err := parseTestOutput(in)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(rep)
}

func parseTestOutput(r io.Reader) (report, error) {
	rep := report{
		Metadate: metadata{Created: time.Now().UTC().Format(time.RFC3339)},
		Coverage: make(map[string]*packageCoverage),
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		pkg, cov, err := parseTestLine(scanner.Text())
		if err != nil {
			return report{}, err
		}
		rep.Coverage[pkg] = &cov
	}

	return rep, nil
}

func parseTestLine(line string) (string, packageCoverage, error) {
	fields := strings.Fields(line)
	coverage := packageCoverage{}
	var pkg string
	for i := 0; i < len(fields); i++ {
		f := fields[i]
		switch {
		case f == "FAIL":
			return "", packageCoverage{}, errors.New("test failed")

		case strings.HasPrefix(f, "github.com/"):
			pkg = f

		case f == "coverage:" && fields[i+1] != "[no":
			covStr := strings.TrimSuffix(fields[i+1], "%")
			cov, err := strconv.ParseFloat(covStr, 64)
			if err != nil {
				return "", packageCoverage{}, fmt.Errorf("parsing coverage: %v", err)
			}
			coverage.Coverage = cov

		case f == "[no":
			switch fields[i+1] {
			case "test": // [no test files]
				coverage.Notest = true
			case "statements]": // [no statements]
				coverage.Nostmt = true
				// ignoring [no tests to run], as it also prints coverage 0%
			}

		default:
			continue
		}
	}
	return pkg, coverage, nil
}

func diff(a, b, touched string) error {
	af, err := os.Open(a)
	if err != nil {
		return err
	}
	defer af.Close()
	var aReport, bReport report
	if err := json.NewDecoder(af).Decode(&aReport); err != nil {
		return err
	}
	bf, err := os.Open(b)
	if err != nil {
		return err
	}
	defer bf.Close()
	if err := json.NewDecoder(bf).Decode(&bReport); err != nil {
		return err
	}
	diffs, err := diffCoverage(aReport, bReport)
	if err != nil {
		return err
	}
	touchedPkgs := strings.Split(touched, ",")
	return diffsToMd(diffs, os.Stdout, touchedPkgs)
}

type coverageDiff struct {
	old *packageCoverage
	new *packageCoverage
}

func diffCoverage(a, b report) (map[string]coverageDiff, error) {
	allPkgs := make(map[string]struct{})
	for pkg := range a.Coverage {
		allPkgs[pkg] = struct{}{}
	}
	for pkg := range b.Coverage {
		allPkgs[pkg] = struct{}{}
	}
	diffs := make(map[string]coverageDiff)
	for pkg := range allPkgs {
		diffs[pkg] = coverageDiff{old: a.Coverage[pkg], new: b.Coverage[pkg]}
		if diffs[pkg].old == nil && diffs[pkg].new == nil {
			return nil, fmt.Errorf("both old and new coverage are nil for pkg %s", pkg)
		}
	}
	return diffs, nil
}

const pkgPrefix = "github.com/edgelesssys/constellation/v2/"

func diffsToMd(diffs map[string]coverageDiff, out io.Writer, touchedPkgs []string) error {
	pkgs := make([]string, 0, len(diffs))
	for pkg := range diffs {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)

	fmt.Fprintln(out, "### Coverage report")
	fmt.Fprintln(out, "| Package | Old  | New  | Trend |")
	fmt.Fprintln(out, "| ------- | ---- | ---- | ----- |")

	for _, pkg := range pkgs {
		diff := diffs[pkg]
		var tendency string
		switch {
		case len(touchedPkgs) != 0 && !slices.Contains(touchedPkgs, strings.TrimPrefix(pkg, pkgPrefix)):
			continue // pkg was not touched
		case diff.old == nil && (diff.new.Notest || diff.new.Nostmt):
			tendency = ":rotating_light:" // new pkg, no tests
		case diff.old == nil:
			tendency = ":new:" // pkg is new
		case diff.new == nil:
			continue // pkg was removed
		case (diff.new.Nostmt || diff.new.Notest) && (diff.old.Notest || diff.old.Nostmt):
			tendency = ":construction:" // this is bad, but not worse than before
		case (diff.new.Nostmt || diff.new.Notest) && !(diff.old.Coverage > 0.0):
			tendency = ":rotating_light:" // tests were removed
		case diff.old.Coverage == diff.new.Coverage && diff.new.Coverage == 0.0:
			tendency = ":construction:" // no tests before, no tests now - do something about it!
		case diff.old.Coverage == diff.new.Coverage && diff.new.Coverage < 50.0:
			tendency = ":construction:" // heuristically still bad - you touch it, you own it.
		case diff.old.Coverage < diff.new.Coverage:
			tendency = ":arrow_upper_right:"
		case diff.old.Coverage > diff.new.Coverage:
			tendency = ":arrow_lower_right:"
		case diff.old.Coverage == diff.new.Coverage:
			tendency = ":left_right_arrow:"
		default:
			return fmt.Errorf("unexpected case: old: %v, new: %v", diff.old, diff.new)
		}
		oldCov := diff.old.String()
		newCov := diff.new.String()
		fmt.Fprintf(out, "| %s | %s | %s | %s |\n", strings.TrimPrefix(pkg, pkgPrefix), oldCov, newCov, tendency)
	}
	return nil
}
