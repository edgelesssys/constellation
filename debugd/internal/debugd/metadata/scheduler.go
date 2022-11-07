/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package metadata

import (
	"context"
	"errors"
	"io/fs"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"go.uber.org/zap"
)

// Fetcher retrieves other debugd IPs from cloud provider metadata.
type Fetcher interface {
	Role(ctx context.Context) (role.Role, error)
	DiscoverDebugdIPs(ctx context.Context) ([]string, error)
	DiscoverLoadbalancerIP(ctx context.Context) (string, error)
}

// Scheduler schedules fetching of metadata using timers.
type Scheduler struct {
	log        *logger.Logger
	fetcher    Fetcher
	downloader downloader
}

// NewScheduler returns a new scheduler.
func NewScheduler(log *logger.Logger, fetcher Fetcher, downloader downloader) *Scheduler {
	return &Scheduler{
		log:        log,
		fetcher:    fetcher,
		downloader: downloader,
	}
}

// Start the loops for discovering debugd endpoints.
func (s *Scheduler) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	wg.Add(1)
	go s.discoveryLoop(ctx, wg)
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

// downloadDeployment tries to download deployment from a list of ips and logs errors encountered.
func (s *Scheduler) downloadDeployment(ctx context.Context, ips []string) (success bool) {
	for _, ip := range ips {
		err := s.downloader.DownloadDeployment(ctx, ip)
		if err == nil {
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

type downloader interface {
	DownloadDeployment(ctx context.Context, ip string) error
}
