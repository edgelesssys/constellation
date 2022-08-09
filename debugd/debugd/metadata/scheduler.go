package metadata

import (
	"context"
	"errors"
	"io/fs"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/debugd/debugd"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/logger"
	"go.uber.org/zap"
)

// Fetcher retrieves other debugd IPs and SSH keys from cloud provider metadata.
type Fetcher interface {
	DiscoverDebugdIPs(ctx context.Context) ([]string, error)
	FetchSSHKeys(ctx context.Context) ([]ssh.UserKey, error)
	DiscoverLoadbalancerIP(ctx context.Context) (string, error)
}

// Scheduler schedules fetching of metadata using timers.
type Scheduler struct {
	log        *logger.Logger
	fetcher    Fetcher
	ssh        sshDeployer
	downloader downloader
}

// NewScheduler returns a new scheduler.
func NewScheduler(log *logger.Logger, fetcher Fetcher, ssh sshDeployer, downloader downloader) *Scheduler {
	return &Scheduler{
		log:        log,
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
		s.log.With(zap.Error(err)).Errorf("Discovering debugd IPs failed")
	} else {
		if s.downloadDeployment(ctx, ips) {
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
				s.log.With(zap.Error(err)).Errorf("Discovering debugd IPs failed")
				continue
			}
			s.log.With(zap.Strings("ips", ips)).Infof("Discovered instances")
			if s.downloadDeployment(ctx, ips) {
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

	ticker := time.NewTicker(debugd.SSHCheckInterval)
	defer ticker.Stop()
	for {
		keys, err := s.fetcher.FetchSSHKeys(ctx)
		if err != nil {
			s.log.With(zap.Error(err)).Errorf("Fetching SSH keys failed")
		} else {
			s.deploySSHKeys(ctx, keys)
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

// downloadDeployment tries to download deployment from a list of ips and logs errors encountered.
func (s *Scheduler) downloadDeployment(ctx context.Context, ips []string) (success bool) {
	for _, ip := range ips {
		keys, err := s.downloader.DownloadDeployment(ctx, ip)
		if err == nil {
			s.deploySSHKeys(ctx, keys)
			return true
		}
		if errors.Is(err, fs.ErrExist) {
			// bootstrapper was already uploaded
			s.log.Infof("Bootstrapper was already uploaded.")
			return true
		}
		s.log.With(zap.Error(err), zap.String("peer", ip)).Errorf("Downloading deployment from peer failed")
	}
	return false
}

// deploySSHKeys tries to deploy a list of SSH keys and logs errors encountered.
func (s *Scheduler) deploySSHKeys(ctx context.Context, keys []ssh.UserKey) {
	for _, key := range keys {
		err := s.ssh.DeployAuthorizedKey(ctx, key)
		if err != nil {
			s.log.With(zap.Error(err), zap.Any("key", key)).Errorf("Deploying SSH key failed")
			continue
		}
	}
}

type downloader interface {
	DownloadDeployment(ctx context.Context, ip string) ([]ssh.UserKey, error)
}

type sshDeployer interface {
	DeployAuthorizedKey(ctx context.Context, sshKey ssh.UserKey) error
}
