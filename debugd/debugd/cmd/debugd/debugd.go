package main

import (
	"log"
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
	"github.com/spf13/afero"
	"golang.org/x/net/context"
)

func main() {
	wg := &sync.WaitGroup{}

	fs := afero.NewOsFs()
	streamer := coordinator.NewFileStreamer(fs)
	serviceManager := deploy.NewServiceManager()
	ssh := ssh.NewSSHAccess(user.NewLinuxUserManager(fs))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	download := deploy.New(&net.Dialer{}, serviceManager, streamer)
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
		log.Printf("Unknown / unimplemented cloud provider CONSTEL_CSP=%v\n", constellationCSP)
		fetcher = fallback.Fetcher{}
	}
	sched := metadata.NewScheduler(fetcher, ssh, download)
	serv := server.New(ssh, serviceManager, streamer)
	if err := deploy.DeployDefaultServiceUnit(ctx, serviceManager); err != nil {
		panic(err)
	}

	wg.Add(1)
	go sched.Start(ctx, wg)
	wg.Add(1)
	go server.Start(wg, serv)

	wg.Wait()
}
