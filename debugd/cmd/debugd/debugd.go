/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/spf13/afero"

	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/deploy"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/info"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/logcollector"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/metadata"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/metadata/cloudprovider"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/metadata/fallback"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/server"
	"github.com/edgelesssys/constellation/v2/debugd/internal/filetransfer"
	"github.com/edgelesssys/constellation/v2/debugd/internal/filetransfer/streamer"
	awscloud "github.com/edgelesssys/constellation/v2/internal/cloud/aws"
	azurecloud "github.com/edgelesssys/constellation/v2/internal/cloud/azure"
	platform "github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	gcpcloud "github.com/edgelesssys/constellation/v2/internal/cloud/gcp"
	openstackcloud "github.com/edgelesssys/constellation/v2/internal/cloud/openstack"
	qemucloud "github.com/edgelesssys/constellation/v2/internal/cloud/qemu"
	"github.com/edgelesssys/constellation/v2/internal/logger"
)

const debugBanner = `
*****************************************
    THIS IS A CONSTELLATION DEBUG IMAGE.
    DO NOT USE IN PRODUCTION.
*****************************************
`

func main() {
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)
	flag.Parse()

	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity))
	fs := afero.NewOsFs()
	streamer := streamer.New(fs)
	filetransferer := filetransfer.New(log.Grouped("filetransfer"), streamer, filetransfer.DontShowProgress)
	serviceManager := deploy.NewServiceManager(log.Grouped("serviceManager"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}

	csp := os.Getenv("CONSTEL_CSP")

	var fetcher *cloudprovider.Fetcher
	switch platform.FromString(csp) {
	case platform.AWS:
		meta, err := awscloud.New(ctx)
		if err != nil {
			log.With(slog.Any("error", err)).Fatalf("Failed to initialize AWS metadata")
		}
		fetcher = cloudprovider.New(meta)

	case platform.Azure:
		meta, err := azurecloud.New(ctx)
		if err != nil {
			log.With(slog.Any("error", err)).Fatalf("Failed to initialize Azure metadata")
		}
		fetcher = cloudprovider.New(meta)

	case platform.GCP:
		meta, err := gcpcloud.New(ctx)
		if err != nil {
			log.With(slog.Any("error", err)).Fatalf("Failed to initialize GCP metadata")
		}
		defer meta.Close()
		fetcher = cloudprovider.New(meta)

	case platform.OpenStack:
		meta, err := openstackcloud.New(ctx)
		if err != nil {
			log.With(slog.Any("error", err)).Fatalf("Failed to initialize OpenStack metadata")
		}
		fetcher = cloudprovider.New(meta)
	case platform.QEMU:
		fetcher = cloudprovider.New(qemucloud.New())

	default:
		log.Errorf("Unknown / unimplemented cloud provider CONSTEL_CSP=%v. Using fallback", csp)
		fetcher = fallback.NewFallbackFetcher()
	}

	infoMap := info.NewMap()
	infoMap.RegisterOnReceiveTrigger(
		logcollector.NewStartTrigger(ctx, wg, platform.FromString(csp), fetcher, log.Grouped("logcollector")),
	)

	download := deploy.New(log.Grouped("download"), &net.Dialer{}, serviceManager, filetransferer, infoMap)

	sched := metadata.NewScheduler(log.Grouped("scheduler"), fetcher, download)
	serv := server.New(log.Grouped("server"), serviceManager, filetransferer, infoMap)

	writeDebugBanner(log)

	sched.Start(ctx, wg)
	server.Start(log, wg, serv)
	wg.Wait()
}

func writeDebugBanner(log *logger.Logger) {
	tty, err := os.OpenFile("/dev/ttyS0", os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.With(slog.Any("error", err)).Errorf("Unable to open /dev/ttyS0 for printing banner")
		return
	}
	defer tty.Close()
	if _, err := fmt.Fprint(tty, debugBanner); err != nil {
		log.With(slog.Any("error", err)).Errorf("Unable to print to /dev/ttyS0")
	}
}
