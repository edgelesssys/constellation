/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/edgelesssys/constellation/v2/debugd/internal/bootstrapper"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/deploy"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/metadata"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/metadata/cloudprovider"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/metadata/fallback"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/server"
	platform "github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/v2/internal/deploy/user"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
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
	streamer := bootstrapper.NewFileStreamer(fs)
	serviceManager := deploy.NewServiceManager(log.Named("serviceManager"))
	ssh := ssh.NewAccess(log, user.NewLinuxUserManager(fs))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := deploy.DeleteUserPassword(ctx, "root"); err != nil {
		log.Errorf("root login: %w")
	}

	download := deploy.New(log.Named("download"), &net.Dialer{}, serviceManager, streamer)
	var fetcher metadata.Fetcher
	csp := os.Getenv("CONSTEL_CSP")
	switch platform.FromString(csp) {
	case platform.Azure:
		azureFetcher, err := cloudprovider.NewAzure(ctx)
		if err != nil {
			log.Fatalf("%s", err)
		}
		fetcher = azureFetcher
	case platform.GCP:
		gcpFetcher, err := cloudprovider.NewGCP(ctx)
		if err != nil {
			log.Fatalf("%s", err)
		}
		fetcher = gcpFetcher
		log.Infof("Added load balancer IP to local routing table")
	case platform.QEMU:
		fetcher = cloudprovider.NewQEMU()
	default:
		log.Errorf("Unknown / unimplemented cloud provider CONSTEL_CSP=%v. Using fallback", csp)
		fetcher = fallback.Fetcher{}
	}
	sched := metadata.NewScheduler(log.Named("scheduler"), fetcher, ssh, download)
	serv := server.New(log.Named("server"), ssh, serviceManager, streamer)
	if err := deploy.DeployDefaultServiceUnit(ctx, serviceManager); err != nil {
		log.Fatalf("%s", err)
	}

	writeDebugBanner(log)

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go sched.Start(ctx, wg)
	wg.Add(1)
	go server.Start(log, wg, serv)

	wg.Wait()
}

func writeDebugBanner(log *logger.Logger) {
	tty, err := os.OpenFile("/dev/ttyS0", os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Infof("Unable to open /dev/ttyS0 for printing banner: %v", err)
		return
	}
	defer tty.Close()
	if _, err := fmt.Fprint(tty, debugBanner); err != nil {
		log.Infof("Unable to print to /dev/ttyS0: %v", err)
	}
}
