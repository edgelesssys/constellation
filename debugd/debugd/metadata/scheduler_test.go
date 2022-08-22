package metadata

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/logger"
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
		ssh                 stubSSHDeployer
		downloader          stubDownloader
		timeout             time.Duration
		wantSSHKeys         []ssh.UserKey
		wantDebugdDownloads []string
	}{
		"scheduler works and calls fetcher functions at least once": {},
		"ssh keys are fetched": {
			fetcher: stubFetcher{
				keys: []ssh.UserKey{{Username: "test", PublicKey: "testkey"}},
			},
			wantSSHKeys: []ssh.UserKey{{Username: "test", PublicKey: "testkey"}},
		},
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
		"ssh key fetch can fail": {
			fetcher: stubFetcher{fetchSSHKeysErr: someErr},
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
				ssh:        &tc.ssh,
				downloader: &tc.downloader,
			}
			wg.Add(1)
			go scheduler.Start(ctx, wg)

			wg.Wait()
			assert.Equal(tc.wantSSHKeys, tc.ssh.sshKeys)
			assert.Equal(tc.wantDebugdDownloads, tc.downloader.ips)
			assert.Greater(tc.fetcher.discoverCalls, 0)
			assert.Greater(tc.fetcher.fetchSSHKeysCalls, 0)
		})
	}
}

type stubFetcher struct {
	discoverCalls     int
	fetchSSHKeysCalls int

	ips             []string
	keys            []ssh.UserKey
	discoverErr     error
	fetchSSHKeysErr error
}

func (s *stubFetcher) Role(_ context.Context) (role.Role, error) {
	return role.Unknown, nil
}

func (s *stubFetcher) DiscoverDebugdIPs(ctx context.Context) ([]string, error) {
	s.discoverCalls++
	return s.ips, s.discoverErr
}

func (s *stubFetcher) FetchSSHKeys(ctx context.Context) ([]ssh.UserKey, error) {
	s.fetchSSHKeysCalls++
	return s.keys, s.fetchSSHKeysErr
}

func (s *stubFetcher) DiscoverLoadbalancerIP(ctx context.Context) (string, error) {
	return "", nil
}

type stubSSHDeployer struct {
	sshKeys []ssh.UserKey

	deployErr error
}

func (s *stubSSHDeployer) DeployAuthorizedKey(ctx context.Context, sshKey ssh.UserKey) error {
	s.sshKeys = append(s.sshKeys, sshKey)

	return s.deployErr
}

type stubDownloader struct {
	ips         []string
	downloadErr error
	keys        []ssh.UserKey
}

func (s *stubDownloader) DownloadDeployment(ctx context.Context, ip string) ([]ssh.UserKey, error) {
	s.ips = append(s.ips, ip)
	return s.keys, s.downloadErr
}
