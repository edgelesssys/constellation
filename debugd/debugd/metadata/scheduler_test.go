package metadata

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/debugd/ssh"
	"github.com/stretchr/testify/assert"
)

func TestSchedulerStart(t *testing.T) {
	testCases := map[string]struct {
		fetcher                 stubFetcher
		ssh                     stubSSHDeployer
		downloader              stubDownloader
		timeout                 time.Duration
		expectedSSHKeys         []ssh.SSHKey
		expectedDebugdDownloads []string
	}{
		"scheduler works and calls fetcher functions at least once": {},
		"ssh keys are fetched": {
			fetcher: stubFetcher{
				keys: []ssh.SSHKey{
					{
						Username: "test",
						KeyValue: "testkey",
					},
				},
			},
			expectedSSHKeys: []ssh.SSHKey{
				{
					Username: "test",
					KeyValue: "testkey",
				},
			},
		},
		"download for discovered debugd ips is started": {
			fetcher: stubFetcher{
				ips: []string{"192.0.2.1", "192.0.2.2"},
			},
			downloader: stubDownloader{
				err: errors.New("download fails"),
			},

			expectedDebugdDownloads: []string{"192.0.2.1", "192.0.2.2"},
		},
		"if download is successful, second download is not attempted": {
			fetcher: stubFetcher{
				ips: []string{"192.0.2.1", "192.0.2.2"},
			},

			expectedDebugdDownloads: []string{"192.0.2.1"},
		},
		"endpoint discovery can fail": {
			fetcher: stubFetcher{
				discoverErr: errors.New("discovery error"),
			},
		},
		"ssh key fetch can fail": {
			fetcher: stubFetcher{
				fetchSSHKeysErr: errors.New("fetch error"),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			wg := &sync.WaitGroup{}
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			scheduler := Scheduler{
				fetcher:    &tc.fetcher,
				ssh:        &tc.ssh,
				downloader: &tc.downloader,
			}
			wg.Add(1)
			go scheduler.Start(ctx, wg)

			wg.Wait()
			assert.Equal(tc.expectedSSHKeys, tc.ssh.sshKeys)
			assert.Equal(tc.expectedDebugdDownloads, tc.downloader.ips)
			assert.Greater(tc.fetcher.discoverCalls, 0)
			assert.Greater(tc.fetcher.fetchSSHKeysCalls, 0)
		})
	}
}

type stubFetcher struct {
	discoverCalls     int
	fetchSSHKeysCalls int

	ips             []string
	keys            []ssh.SSHKey
	discoverErr     error
	fetchSSHKeysErr error
}

func (s *stubFetcher) DiscoverDebugdIPs(ctx context.Context) ([]string, error) {
	s.discoverCalls++
	return s.ips, s.discoverErr
}

func (s *stubFetcher) FetchSSHKeys(ctx context.Context) ([]ssh.SSHKey, error) {
	s.fetchSSHKeysCalls++
	return s.keys, s.fetchSSHKeysErr
}

type stubSSHDeployer struct {
	sshKeys []ssh.SSHKey

	deployErr error
}

func (s *stubSSHDeployer) DeploySSHAuthorizedKey(ctx context.Context, sshKey ssh.SSHKey) error {
	s.sshKeys = append(s.sshKeys, sshKey)

	return s.deployErr
}

type stubDownloader struct {
	ips []string
	err error
}

func (s *stubDownloader) DownloadCoordinator(ctx context.Context, ip string) error {
	s.ips = append(s.ips, ip)
	return s.err
}
