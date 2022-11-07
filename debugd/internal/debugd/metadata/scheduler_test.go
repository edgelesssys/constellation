/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package metadata

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestSchedulerStart(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		fetcher             stubFetcher
		downloader          stubDownloader
		timeout             time.Duration
		wantDebugdDownloads []string
	}{
		"scheduler works and calls fetcher functions at least once": {},
		"download for discovered debugd ips is started": {
			fetcher: stubFetcher{
				ips: []string{"192.0.2.1", "192.0.2.2"},
			},
			downloader:          stubDownloader{downloadErr: someErr},
			wantDebugdDownloads: []string{"192.0.2.1", "192.0.2.2"},
		},
		"if download is successful, second download is not attempted": {
			fetcher: stubFetcher{
				ips: []string{"192.0.2.1", "192.0.2.2"},
			},
			wantDebugdDownloads: []string{"192.0.2.1"},
		},
		"endpoint discovery can fail": {
			fetcher: stubFetcher{discoverErr: someErr},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			wg := &sync.WaitGroup{}
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			scheduler := Scheduler{
				log:        logger.NewTest(t),
				fetcher:    &tc.fetcher,
				downloader: &tc.downloader,
			}
			wg.Add(1)
			go scheduler.Start(ctx, wg)

			wg.Wait()
			assert.Equal(tc.wantDebugdDownloads, tc.downloader.ips)
			assert.Greater(tc.fetcher.discoverCalls, 0)
		})
	}
}

type stubFetcher struct {
	discoverCalls int

	ips         []string
	discoverErr error
}

func (s *stubFetcher) Role(_ context.Context) (role.Role, error) {
	return role.Unknown, nil
}

func (s *stubFetcher) DiscoverDebugdIPs(ctx context.Context) ([]string, error) {
	s.discoverCalls++
	return s.ips, s.discoverErr
}

func (s *stubFetcher) DiscoverLoadbalancerIP(ctx context.Context) (string, error) {
	return "", nil
}

type stubDownloader struct {
	ips         []string
	downloadErr error
}

func (s *stubDownloader) DownloadDeployment(ctx context.Context, ip string) error {
	s.ips = append(s.ips, ip)
	return s.downloadErr
}
