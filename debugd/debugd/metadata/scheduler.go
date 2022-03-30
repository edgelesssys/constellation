package metadata

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/debugd/debugd"
	"github.com/edgelesssys/constellation/debugd/ssh"
)

// Fetcher retrieves other debugd IPs and SSH keys from cloud provider metadata.
type Fetcher interface {
	DiscoverDebugdIPs(ctx context.Context) ([]string, error)
	FetchSSHKeys(ctx context.Context) ([]ssh.SSHKey, error)
}

// Scheduler schedules fetching of metadata using timers.
type Scheduler struct {
	fetcher    Fetcher
	ssh        sshDeployer
	downloader downloader
}

// NewScheduler returns a new scheduler.
func NewScheduler(fetcher Fetcher, ssh sshDeployer, downloader downloader) *Scheduler {
	return &Scheduler{
		fetcher:    fetcher,
		ssh:        ssh,
		downloader: downloader,
	}
}

// Start will start the loops for discovering debugd endpoints and ssh keys.
func (s *Scheduler) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	wg.Add(2)
	go s.discoveryLoop(ctx, wg)
	go s.sshLoop(ctx, wg)
}

// discoveryLoop discovers new debugd endpoints from cloud-provider metadata periodically.
func (s *Scheduler) discoveryLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	// execute debugd discovery once at the start to skip wait for first tick
	ips, err := s.fetcher.DiscoverDebugdIPs(ctx)
	if err != nil {
		log.Printf("error occurred while discovering debugd IPs: %v\n", err)
	} else {
		if s.downloadCoordinator(ctx, ips) {
			return
		}
	}

	ticker := time.NewTicker(debugd.DiscoverDebugdInterval)
	defer ticker.Stop()
	for {
		var err error
		select {
		case <-ticker.C:
			ips, err = s.fetcher.DiscoverDebugdIPs(ctx)
			if err != nil {
				log.Printf("error occurred while discovering debugd IPs: %v\n", err)
				continue
			}
			log.Printf("discovered instances: %v\n", ips)
			if s.downloadCoordinator(ctx, ips) {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// sshLoop discovers new ssh keys from cloud provider metadata periodically.
func (s *Scheduler) sshLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	// execute ssh key search once at the start to skip wait for first tick
	keys, err := s.fetcher.FetchSSHKeys(ctx)
	if err != nil {
		log.Printf("error occurred while fetching SSH keys: %v\n", err)
	} else {
		s.deploySSHKeys(ctx, keys)
	}
	ticker := time.NewTicker(debugd.SSHCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			keys, err := s.fetcher.FetchSSHKeys(ctx)
			if err != nil {
				log.Printf("error occurred while fetching ssh keys: %v\n", err)
				continue
			}
			s.deploySSHKeys(ctx, keys)
		case <-ctx.Done():
			return
		}
	}
}

// downloadCoordinator tries to download coordinator from a list of ips and logs errors encountered.
func (s *Scheduler) downloadCoordinator(ctx context.Context, ips []string) (success bool) {
	for _, ip := range ips {
		err := s.downloader.DownloadCoordinator(ctx, ip)
		if err == nil {
			// early exit with success since coordinator should only be downloaded once
			return true
		}
		if errors.Is(err, fs.ErrExist) {
			// coordinator was already uploaded
			return true
		}
		log.Printf("error occurred while downloading coordinator from %v: %v\n", ip, err)
	}
	return false
}

// deploySSHKeys tries to deploy a list of SSH keys and logs errors encountered.
func (s *Scheduler) deploySSHKeys(ctx context.Context, keys []ssh.SSHKey) {
	for _, key := range keys {
		err := s.ssh.DeploySSHAuthorizedKey(ctx, key)
		if err != nil {
			log.Printf("error occurred while deploying ssh key %v: %v\n", key, err)
			continue
		}
	}
}

type downloader interface {
	DownloadCoordinator(ctx context.Context, ip string) error
}

type sshDeployer interface {
	DeploySSHAuthorizedKey(ctx context.Context, sshKey ssh.SSHKey) error
}
