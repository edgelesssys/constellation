package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/edgelesssys/constellation/hack/pseudo-version/internal/git"
	"github.com/edgelesssys/constellation/internal/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/mod/module"
)

func main() {
	major := flag.String("major", "v0", "Optional major version")
	base := flag.String("base", "", "Optional base version")
	revisionTimestamp := flag.String("time", "", "Optional revision time")
	revision := flag.String("revision", "", "Optional revision (git commit hash)")
	flag.Parse()

	log := logger.New(logger.JSONLog, zapcore.InfoLevel)

	gitc, err := git.New()
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to initialize git client")
	}

	if *base == "" {
		_, versionTag, err := gitc.FirstParentWithVersionTag()
		if err != nil {
			log.With(zap.Error(err)).Warnf("Failed to find base version. Using default.")
			versionTag = ""
		}
		*base = versionTag
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
	fmt.Println(version)
}
