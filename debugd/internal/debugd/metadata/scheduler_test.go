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
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestSchedulerStart(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		fetcher                 stubFetcher
		downloader              stubDownloader
		wantDiscoverCount       int
		wantDeploymentDownloads []string
		wantInfoDownloads       []string
	}{
		"no errors occur": {
			fetcher:                 stubFetcher{ips: []string{"192.0.2.1", "192.0.2.2"}},
			downloader:              stubDownloader{},
			wantDiscoverCount:       1,
			wantDeploymentDownloads: []string{"192.0.2.1"},
			wantInfoDownloads:       []string{"192.0.2.1"},
		},
		"download deployment fails": {
			fetcher:                 stubFetcher{ips: []string{"192.0.2.1", "192.0.2.2"}},
			downloader:              stubDownloader{downloadDeploymentErrs: []error{someErr, someErr}},
			wantDiscoverCount:       2,
			wantDeploymentDownloads: []string{"192.0.2.1", "192.0.2.2", "192.0.2.1"},
			wantInfoDownloads:       []string{"192.0.2.1"},
		},
		"download info fails": {
			fetcher:                 stubFetcher{ips: []string{"192.0.2.1", "192.0.2.2"}},
			downloader:              stubDownloader{downloadInfoErrs: []error{someErr, someErr}},
			wantDiscoverCount:       2,
			wantDeploymentDownloads: []string{"192.0.2.1"},
			wantInfoDownloads:       []string{"192.0.2.1", "192.0.2.2", "192.0.2.1"},
		},
		"endpoint discovery fails": {
			fetcher: stubFetcher{
				discoverErrs: []error{someErr, someErr, someErr},
				ips:          []string{"192.0.2.1", "192.0.2.2"},
			},
			wantDiscoverCount:       4,
			wantDeploymentDownloads: []string{"192.0.2.1"},
			wantInfoDownloads:       []string{"192.0.2.1"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			scheduler := Scheduler{
				log:        logger.NewTest(t),
				fetcher:    &tc.fetcher,
				downloader: &tc.downloader,
				interval:   20 * time.Millisecond,
			}

			wg := &sync.WaitGroup{}
			scheduler.Start(context.Background(), wg)
			wg.Wait()

			assert.Equal(tc.wantDeploymentDownloads, tc.downloader.downloadDeploymentIPs)
			assert.Equal(tc.wantInfoDownloads, tc.downloader.downloadInfoIPs)
			assert.Equal(tc.wantDiscoverCount, tc.fetcher.discoverCalls)
		})
	}
}

type stubFetcher struct {
	ips            []string
	discoverErrs   []error
	discoverErrIdx int
	discoverCalls  int
}

func (s *stubFetcher) DiscoverDebugdIPs(_ context.Context) ([]string, error) {
	s.discoverCalls++
	var err error
	if s.discoverErrIdx < len(s.discoverErrs) {
		err = s.discoverErrs[s.discoverErrIdx]
		s.discoverErrIdx++
		return nil, err
	}
	return s.ips, nil
}

type stubDownloader struct {
	downloadDeploymentErrs   []error
	downloadDeploymentErrIdx int
	downloadDeploymentIPs    []string
	downloadInfoErrs         []error
	downloadInfoErrIdx       int
	downloadInfoIPs          []string
}

func (s *stubDownloader) DownloadDeployment(_ context.Context, ip string) error {
	s.downloadDeploymentIPs = append(s.downloadDeploymentIPs, ip)
	var err error
	if s.downloadDeploymentErrIdx < len(s.downloadDeploymentErrs) {
		err = s.downloadDeploymentErrs[s.downloadDeploymentErrIdx]
		s.downloadDeploymentErrIdx++
	}
	return err
}

func (s *stubDownloader) DownloadInfo(_ context.Context, ip string) error {
	s.downloadInfoIPs = append(s.downloadInfoIPs, ip)
	var err error
	if s.downloadInfoErrIdx < len(s.downloadInfoErrs) {
		err = s.downloadInfoErrs[s.downloadInfoErrIdx]
		s.downloadInfoErrIdx++
	}
	return err
}
