package cloudprovider

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

func TestDiscoverDebugIPs(t *testing.T) {
	err := errors.New("some err")

	testCases := map[string]struct {
		meta    stubMetadata
		wantIPs []string
		wantErr bool
	}{
		"disovery works": {
			meta: stubMetadata{
				listRes: []cloudtypes.Instance{
					{
						PrivateIPs: []string{"192.0.2.0"},
					},
					{
						PrivateIPs: []string{"192.0.2.1"},
					},
					{
						PrivateIPs: []string{"192.0.2.2"},
					},
				},
			},
			wantIPs: []string{
				"192.0.2.1", "192.0.2.2",
			},
		},
		"retrieve fails": {
			meta: stubMetadata{
				listErr: err,
			},
			wantErr: true,
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

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantIPs, ips)
		})
	}
}

func TestFetchSSHKeys(t *testing.T) {
	err := errors.New("some err")

	testCases := map[string]struct {
		meta     stubMetadata
		wantKeys []ssh.UserKey
		wantErr  bool
	}{
		"fetch works": {
			meta: stubMetadata{
				selfRes: cloudtypes.Instance{
					Name:       "name",
					ProviderID: "provider-id",
					SSHKeys:    map[string][]string{"bob": {"ssh-rsa bobskey"}},
				},
			},
			wantKeys: []ssh.UserKey{
				{
					Username:  "bob",
					PublicKey: "ssh-rsa bobskey",
				},
			},
		},
		"retrieve fails": {
			meta: stubMetadata{
				selfErr: err,
			},
			wantErr: true,
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

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantKeys, keys)
		})
	}
}

type stubMetadata struct {
	listRes        []cloudtypes.Instance
	listErr        error
	selfRes        cloudtypes.Instance
	selfErr        error
	getInstanceRes cloudtypes.Instance
	getInstanceErr error
	supportedRes   bool
}

func (m *stubMetadata) List(ctx context.Context) ([]cloudtypes.Instance, error) {
	return m.listRes, m.listErr
}

func (m *stubMetadata) Self(ctx context.Context) (cloudtypes.Instance, error) {
	return m.selfRes, m.selfErr
}

func (m *stubMetadata) GetInstance(ctx context.Context, providerID string) (cloudtypes.Instance, error) {
	return m.getInstanceRes, m.getInstanceErr
}

func (m *stubMetadata) Supported() bool {
	return m.supportedRes
}
