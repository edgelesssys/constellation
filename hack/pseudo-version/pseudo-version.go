/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/hack/pseudo-version/internal/git"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/mod/module"
)

func main() {
	printSemVer := flag.Bool("semantic-version", false, "Only print semantic version")
	printTimestamp := flag.Bool("print-timestamp", false, "Only print timestamp")
	timestampFormat := flag.String("timestamp-format", "20060102150405", "Timestamp format")
	printBranch := flag.Bool("print-branch", false, "Only print branch name")
	major := flag.String("major", "v0", "Optional major version")
	base := flag.String("base", "", "Optional base version")
	revisionTimestamp := flag.String("time", "", "Optional revision time")
	revision := flag.String("revision", "", "Optional revision (git commit hash)")
	skipV := flag.Bool("skip-v", false, "Skip 'v' prefix in version")
	flag.Parse()

	log := logger.New(logger.JSONLog, zapcore.InfoLevel)

	gitc, err := git.New()
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to initialize git client")
	}

	parsedBranch, err := gitc.ParsedBranchName()
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to get parsed branch name")
	}

	if *base == "" {
		rootPath, err := gitc.Path()
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to get git root path. Using default base version.")
		} else {
			versionTag, err := os.ReadFile(filepath.Join(rootPath, "version.txt"))
			if err != nil {
				log.With(zap.Error(err)).Warnf("Failed to read version.txt. Using default base version.")
			} else {
				*base = strings.TrimSpace(string(versionTag))
			}
		}
	}

	var headRevision string
	var headTime time.Time
	if *revisionTimestamp == "" || *revision == "" {
		var err error
		headRevision, headTime, err = gitc.Revision()
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to retrieve HEAD")
		}
	}

	if *revisionTimestamp != "" {
		headTime, err := time.Parse("20060102150405", *revisionTimestamp)
		if err != nil {
			log.With(zap.Error(err)).With("time", headTime).Fatalf("Failed to parse revision timestamp")
		}
	}

	if *revision == "" {
		*revision = headRevision
	}

	version := module.PseudoVersion(*major, *base, headTime, *revision)
	if *skipV {
		version = strings.TrimPrefix(version, "v")
	}

	switch {
	case *printSemVer:
		fmt.Println(*base)
	case *printTimestamp:
		fmt.Println(headTime.Format(*timestampFormat))
	case *printBranch:
		fmt.Println(parsedBranch)
	default:
		if !strings.Contains(*base, "pre") {
			// "v2.7.0" inside the version.txt will lead to "v2.7.0" as version
			fmt.Println(*base)
		} else {
			// "2.7.0-pre" inside the version.txt will lead to "v2.7.0-pre.0.20230313121936-bab76e8a9acf" as version
			fmt.Println(version)
		}
	}
}
