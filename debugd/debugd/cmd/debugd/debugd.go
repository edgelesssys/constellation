package main

import (
	"net"
	"os"
	"strings"
	"sync"

	"github.com/edgelesssys/constellation/debugd/coordinator"
	"github.com/edgelesssys/constellation/debugd/debugd/deploy"
	"github.com/edgelesssys/constellation/debugd/debugd/metadata"
	"github.com/edgelesssys/constellation/debugd/debugd/metadata/cloudprovider"
	"github.com/edgelesssys/constellation/debugd/debugd/metadata/fallback"
	"github.com/edgelesssys/constellation/debugd/debugd/server"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/deploy/user"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/spf13/afero"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
)

func main() {
	wg := &sync.WaitGroup{}

	log := logger.New(logger.JSONLog, zapcore.InfoLevel)
	fs := afero.NewOsFs()
	streamer := coordinator.NewFileStreamer(fs)
	serviceManager := deploy.NewServiceManager(log.Named("serviceManager"))
	ssh := ssh.NewAccess(log, user.NewLinuxUserManager(fs))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	download := deploy.New(log.Named("download"), &net.Dialer{}, serviceManager, streamer)
	var fetcher metadata.Fetcher
	constellationCSP := strings.ToLower(os.Getenv("CONSTEL_CSP"))
	switch constellationCSP {
	case "azure":
		azureFetcher, err := cloudprovider.NewAzure(ctx)
		if err != nil {
			panic(err)
		}
		fetcher = azureFetcher
	case "gcp":
		gcpFetcher, err := cloudprovider.NewGCP(ctx)
		if err != nil {
			panic(err)
		}
		fetcher = gcpFetcher
	default:
		log.Errorf("Unknown / unimplemented cloud provider CONSTEL_CSP=%v. Using fallback", constellationCSP)
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
