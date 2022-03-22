package core

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoordinatorEndpoints(t *testing.T) {
	err := errors.New("some err")

	testCases := map[string]struct {
		metadata          stubMetadata
		expectErr         bool
		expectedEndpoints []string
	}{
		"getting coordinator endpoints works and role is checked": {
			metadata: stubMetadata{
				listRes: []Instance{
					{
						Name:       "someInstanceA",
						Role:       role.Coordinator,
						ProviderID: "provider://somePath/someInstanceA",
						IPs:        []string{"192.0.2.1"},
					},
					{
						Name:       "someInstanceB",
						Role:       role.Node,
						ProviderID: "provider://somePath/someInstanceB",
						IPs:        []string{"192.0.2.2"},
					},
				},
				supportedRes: true,
			},
			expectErr:         false,
			expectedEndpoints: []string{"192.0.2.1:9000"},
		},
		"List fails": {
			metadata: stubMetadata{
				listErr:      err,
				supportedRes: true,
			},
			expectErr: true,
		},
		"metadata API unsupported": {
			metadata:  stubMetadata{},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			endpoints, err := CoordinatorEndpoints(context.Background(), &tc.metadata)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.ElementsMatch(tc.expectedEndpoints, endpoints)
		})
	}
}

func TestPrepareInstanceForCCM(t *testing.T) {
	err := errors.New("some err")

	testCases := map[string]struct {
		metadata  stubMetadata
		vpnIP     string
		expectErr bool
	}{
		"updating role works": {
			metadata:  stubMetadata{},
			vpnIP:     "192.0.2.1",
			expectErr: false,
		},
		"setting VPN IP fails": {
			metadata: stubMetadata{
				setVPNIPErr: err,
			},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			err := PrepareInstanceForCCM(context.Background(), &tc.metadata, &CloudControllerManagerFake{}, tc.vpnIP)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

type stubMetadata struct {
	listRes        []Instance
	listErr        error
	selfRes        Instance
	selfErr        error
	getInstanceRes Instance
	getInstanceErr error
	signalRoleErr  error
	setVPNIPErr    error
	supportedRes   bool
}

func (m *stubMetadata) List(ctx context.Context) ([]Instance, error) {
	return m.listRes, m.listErr
}

func (m *stubMetadata) Self(ctx context.Context) (Instance, error) {
	return m.selfRes, m.selfErr
}

func (m *stubMetadata) GetInstance(ctx context.Context, providerID string) (Instance, error) {
	return m.getInstanceRes, m.getInstanceErr
}

func (m *stubMetadata) SignalRole(ctx context.Context, role role.Role) error {
	return m.signalRoleErr
}

func (m *stubMetadata) SetVPNIP(ctx context.Context, vpnIP string) error {
	return m.setVPNIPErr
}

func (m *stubMetadata) Supported() bool {
	return m.supportedRes
}
