/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package metadata

import (
	"context"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap"
)

// Fetcher retrieves other debugd IPs from cloud provider metadata.
type Fetcher interface {
	DiscoverDebugdIPs(ctx context.Context) ([]string, error)
}

// Scheduler schedules fetching of metadata using timers.
type Scheduler struct {
	log            *logger.Logger
	fetcher        Fetcher
	downloader     downloader
	deploymentDone bool
	infoDone       bool
	interval       time.Duration
}

// NewScheduler returns a new scheduler.
func NewScheduler(log *logger.Logger, fetcher Fetcher, downloader downloader) *Scheduler {
	return &Scheduler{
		log:        log,
		fetcher:    fetcher,
		downloader: downloader,
		interval:   debugd.DiscoverDebugdInterval,
	}
}

// Start the loops for discovering debugd endpoints.
func (s *Scheduler) Start(ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(s.interval)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer ticker.Stop()

		for {
			ips, err := s.fetcher.DiscoverDebugdIPs(ctx)
			if err != nil {
				s.log.With(zap.Error(err)).Warnf("Discovering debugd IPs failed")
				continue
			} else {
				s.log.With(zap.Strings("ips", ips)).Infof("Discovered instances")
				s.download(ctx, ips)
				if s.deploymentDone && s.infoDone {
					return
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// download tries to download deployment from a list of ips and logs errors encountered.
func (s *Scheduler) download(ctx context.Context, ips []string) {
	for _, ip := range ips {
		if !s.deploymentDone {
			if err := s.downloader.DownloadDeployment(ctx, ip); err != nil {
				s.log.With(zap.Error(err), zap.String("peer", ip)).
					Warnf("Downloading deployment from %s: %s", ip, err)
			} else {
				s.deploymentDone = true
			}
		}

		if !s.infoDone {
			if err := s.downloader.DownloadInfo(ctx, ip); err != nil {
				s.log.With(zap.Error(err), zap.String("peer", ip)).
					Warnf("Downloading info from %s: %s", ip, err)
			} else {
				s.infoDone = true
			}
		}

		if s.deploymentDone && s.infoDone {
			return
		}
	}
}

type downloader interface {
	DownloadDeployment(ctx context.Context, ip string) error
	DownloadInfo(ctx context.Context, ip string) error
}
