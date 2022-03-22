package cloudprovider

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/debugd/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverDebugIPs(t *testing.T) {
	err := errors.New("some err")

	testCases := map[string]struct {
		meta        stubMetadata
		expectedIPs []string
		expectErr   bool
	}{
		"disovery works": {
			meta: stubMetadata{
				listRes: []core.Instance{
					{
						IPs: []string{"192.0.2.0"},
					},
					{
						IPs: []string{"192.0.2.1"},
					},
					{
						IPs: []string{"192.0.2.2"},
					},
				},
			},
			expectedIPs: []string{
				"192.0.2.0", "192.0.2.1", "192.0.2.2",
			},
		},
		"retrieve fails": {
			meta: stubMetadata{
				listErr: err,
			},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fetcher := Fetcher{
				metaAPI: &tc.meta,
			}
			ips, err := fetcher.DiscoverDebugdIPs(context.Background())

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.expectedIPs, ips)
		})
	}
}

func TestFetchSSHKeys(t *testing.T) {
	err := errors.New("some err")

	testCases := map[string]struct {
		meta         stubMetadata
		expectedKeys []ssh.SSHKey
		expectErr    bool
	}{
		"fetch works": {
			meta: stubMetadata{
				selfRes: core.Instance{
					Name:       "name",
					ProviderID: "provider-id",
					SSHKeys:    map[string][]string{"bob": {"ssh-rsa bobskey"}},
				},
			},
			expectedKeys: []ssh.SSHKey{
				{
					Username: "bob",
					KeyValue: "ssh-rsa bobskey",
				},
			},
		},
		"retrieve fails": {
			meta: stubMetadata{
				selfErr: err,
			},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fetcher := Fetcher{
				metaAPI: &tc.meta,
			}
			keys, err := fetcher.FetchSSHKeys(context.Background())

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.expectedKeys, keys)
		})
	}
}

type stubMetadata struct {
	listRes        []core.Instance
	listErr        error
	selfRes        core.Instance
	selfErr        error
	getInstanceRes core.Instance
	getInstanceErr error
	supportedRes   bool
}

func (m *stubMetadata) List(ctx context.Context) ([]core.Instance, error) {
	return m.listRes, m.listErr
}

func (m *stubMetadata) Self(ctx context.Context) (core.Instance, error) {
	return m.selfRes, m.selfErr
}

func (m *stubMetadata) GetInstance(ctx context.Context, providerID string) (core.Instance, error) {
	return m.getInstanceRes, m.getInstanceErr
}

func (m *stubMetadata) Supported() bool {
	return m.supportedRes
}
