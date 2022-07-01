package main

import (
	"flag"
	"net"
	"os"
	"sync"

	"github.com/edgelesssys/constellation/debugd/coordinator"
	"github.com/edgelesssys/constellation/debugd/debugd/deploy"
	"github.com/edgelesssys/constellation/debugd/debugd/metadata"
	"github.com/edgelesssys/constellation/debugd/debugd/metadata/cloudprovider"
	"github.com/edgelesssys/constellation/debugd/debugd/metadata/fallback"
	"github.com/edgelesssys/constellation/debugd/debugd/server"
	platform "github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/deploy/user"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/spf13/afero"
	"golang.org/x/net/context"
)

func main() {
	wg := &sync.WaitGroup{}
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)

	flag.Parse()
	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity))
	fs := afero.NewOsFs()
	streamer := coordinator.NewFileStreamer(fs)
	serviceManager := deploy.NewServiceManager(log.Named("serviceManager"))
	ssh := ssh.NewAccess(log, user.NewLinuxUserManager(fs))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	download := deploy.New(log.Named("download"), &net.Dialer{}, serviceManager, streamer)
	var fetcher metadata.Fetcher
	csp := os.Getenv("CONSTEL_CSP")
	switch platform.FromString(csp) {
	case platform.Azure:
		azureFetcher, err := cloudprovider.NewAzure(ctx)
		if err != nil {
			panic(err)
		}
		fetcher = azureFetcher
	case platform.GCP:
		gcpFetcher, err := cloudprovider.NewGCP(ctx)
		if err != nil {
			panic(err)
		}
		fetcher = gcpFetcher
	case platform.QEMU:
		fetcher = cloudprovider.NewQEMU()
	default:
		log.Errorf("Unknown / unimplemented cloud provider CONSTEL_CSP=%v. Using fallback", csp)
		fetcher = fallback.Fetcher{}
	}
	sched := metadata.NewScheduler(log.Named("scheduler"), fetcher, ssh, download)
	serv := server.New(log.Named("server"), ssh, serviceManager, streamer)
	if err := deploy.DeployDefaultServiceUnit(ctx, serviceManager); err != nil {
		panic(err)
	}

	wg.Add(1)
	go sched.Start(ctx, wg)
	wg.Add(1)
	go server.Start(log, wg, serv)

	wg.Wait()
}
