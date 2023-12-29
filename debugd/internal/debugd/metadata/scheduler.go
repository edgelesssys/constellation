/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package metadata

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd"
)

// Fetcher retrieves other debugd IPs from cloud provider metadata.
type Fetcher interface {
	DiscoverDebugdIPs(ctx context.Context) ([]string, error)
	DiscoverLoadBalancerIP(ctx context.Context) (string, error)
}

// Scheduler schedules fetching of metadata using timers.
type Scheduler struct {
	log            *slog.Logger
	fetcher        Fetcher
	downloader     downloader
	deploymentDone bool
	infoDone       bool
	interval       time.Duration
}

// NewScheduler returns a new scheduler.
func NewScheduler(log *slog.Logger, fetcher Fetcher, downloader downloader) *Scheduler {
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
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}

			ips, err := s.fetcher.DiscoverDebugdIPs(ctx)
			if err != nil {
				s.log.With(slog.Any("error", err)).Warn("Discovering debugd IPs failed")
			}

			lbip, err := s.fetcher.DiscoverLoadBalancerIP(ctx)
			if err != nil {
				s.log.With(slog.Any("error", err)).Warn("Discovering load balancer IP failed")
			} else {
				ips = append(ips, lbip)
			}

			if len(ips) == 0 {
				s.log.With(slog.Any("error", err)).Warn("No debugd IPs discovered")
				continue
			}

			s.log.With(slog.Any("ips", ips)).Info("Discovered instances")
			s.download(ctx, ips)
			if s.deploymentDone && s.infoDone {
				return
			}

		}
	}()
}

// download tries to download deployment from a list of ips and logs errors encountered.
func (s *Scheduler) download(ctx context.Context, ips []string) {
	for _, ip := range ips {
		if !s.deploymentDone {
			if err := s.downloader.DownloadDeployment(ctx, ip); err != nil {
				s.log.With(slog.Any("error", err), slog.String("peer", ip)).
					Warn("Downloading deployment from %s: %s", ip, err)
			} else {
				s.deploymentDone = true
			}
		}

		if !s.infoDone {
			if err := s.downloader.DownloadInfo(ctx, ip); err != nil {
				s.log.With(slog.Any("error", err), slog.String("peer", ip)).
					Warn("Downloading info from %s: %s", ip, err)
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
