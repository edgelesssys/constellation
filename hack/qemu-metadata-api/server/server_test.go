/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/dhcp"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAll(t *testing.T) {
	testCases := map[string]struct {
		wantErr         bool
		stubLeaseGetter *stubLeaseGetter
	}{
		"success": {
			stubLeaseGetter: &stubLeaseGetter{
				leases: []dhcp.NetworkDHCPLease{
					{
						IPaddr:   "192.0.100.1",
						Hostname: "control-plane-0",
					},
					{
						IPaddr:   "192.0.100.2",
						Hostname: "control-plane-1",
					},
					{
						IPaddr:   "192.0.200.1",
						Hostname: "worker-0",
					},
				},
			},
		},
		"GetDHCPLeases error": {
			stubLeaseGetter: &stubLeaseGetter{
				getErr: assert.AnError,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			server := New(logger.NewTest(t), "test", "initSecretHash", tc.stubLeaseGetter)

			res, err := server.listAll()

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Len(tc.stubLeaseGetter.leases, len(res))
		})
	}
}

func TestListSelf(t *testing.T) {
	testCases := map[string]struct {
		remoteAddr      string
		stubLeaseGetter *stubLeaseGetter
		wantErr         bool
	}{
		"success": {
			remoteAddr: "192.0.100.1:1234",
			stubLeaseGetter: &stubLeaseGetter{
				leases: []dhcp.NetworkDHCPLease{
					{
						IPaddr:   "192.0.100.1",
						Hostname: "control-plane-0",
					},
				},
			},
		},
		"listAll error": {
			remoteAddr: "192.0.100.1:1234",
			stubLeaseGetter: &stubLeaseGetter{
				getErr: assert.AnError,
			},
			wantErr: true,
		},
		"remoteAddr error": {
			remoteAddr: "",
			stubLeaseGetter: &stubLeaseGetter{
				leases: []dhcp.NetworkDHCPLease{
					{
						IPaddr:   "192.0.100.1",
						Hostname: "control-plane-0",
					},
				},
			},
			wantErr: true,
		},
		"peer not found": {
			remoteAddr: "192.0.200.1:1234",
			stubLeaseGetter: &stubLeaseGetter{
				leases: []dhcp.NetworkDHCPLease{
					{
						IPaddr:   "192.0.100.1",
						Hostname: "control-plane-0",
					},
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			server := New(logger.NewTest(t), "test", "initSecretHash", tc.stubLeaseGetter)

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://192.0.0.1/self", nil)
			require.NoError(err)
			req.RemoteAddr = tc.remoteAddr

			w := httptest.NewRecorder()
			server.listSelf(w, req)

			if tc.wantErr {
				assert.NotEqual(http.StatusOK, w.Code)
				return
			}
			assert.Equal(http.StatusOK, w.Code)
			metadataRaw, err := io.ReadAll(w.Body)
			require.NoError(err)

			var metadata metadata.InstanceMetadata
			require.NoError(json.Unmarshal(metadataRaw, &metadata))
			assert.Equal(tc.stubLeaseGetter.leases[0].Hostname, metadata.Name)
			assert.Equal(tc.stubLeaseGetter.leases[0].IPaddr, metadata.VPCIP)
		})
	}
}

func TestListPeers(t *testing.T) {
	testCases := map[string]struct {
		remoteAddr        string
		stubNetworkGetter *stubLeaseGetter
		wantErr           bool
	}{
		"success": {
			remoteAddr: "192.0.100.1:1234",
			stubNetworkGetter: &stubLeaseGetter{
				leases: []dhcp.NetworkDHCPLease{
					{
						IPaddr:   "192.0.100.1",
						Hostname: "control-plane-0",
					},
					{
						IPaddr:   "192.0.200.1",
						Hostname: "worker-0",
					},
				},
			},
		},
		"listAll error": {
			remoteAddr: "192.0.100.1:1234",
			stubNetworkGetter: &stubLeaseGetter{
				getErr: assert.AnError,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			server := New(logger.NewTest(t), "test", "initSecretHash", tc.stubNetworkGetter)

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://192.0.0.1/peers", nil)
			require.NoError(err)
			req.RemoteAddr = tc.remoteAddr

			w := httptest.NewRecorder()
			server.listPeers(w, req)

			if tc.wantErr {
				assert.NotEqual(http.StatusOK, w.Code)
				return
			}
			assert.Equal(http.StatusOK, w.Code)
			metadataRaw, err := io.ReadAll(w.Body)
			require.NoError(err)

			var metadata []metadata.InstanceMetadata
			require.NoError(json.Unmarshal(metadataRaw, &metadata))
			assert.Len(metadata, len(tc.stubNetworkGetter.leases))
		})
	}
}

func TestInitSecretHash(t *testing.T) {
	defaultConnect := &stubLeaseGetter{
		leases: []dhcp.NetworkDHCPLease{
			{
				IPaddr:   "192.0.100.1",
				Hostname: "control-plane-0",
			},
		},
	}

	testCases := map[string]struct {
		connect  *stubLeaseGetter
		method   string
		wantHash string
		wantErr  bool
	}{
		"success": {
			connect: defaultConnect,
			method:  http.MethodGet,
		},
		"wrong method": {
			connect: defaultConnect,
			method:  http.MethodPost,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			server := New(logger.NewTest(t), "test", tc.wantHash, defaultConnect)

			req, err := http.NewRequestWithContext(context.Background(), tc.method, "http://192.0.0.1/initsecrethash", nil)
			require.NoError(err)

			w := httptest.NewRecorder()
			server.initSecretHash(w, req)

			if tc.wantErr {
				assert.NotEqual(http.StatusOK, w.Code)
				return
			}

			assert.Equal(http.StatusOK, w.Code)
			assert.Equal(tc.wantHash, w.Body.String())
		})
	}
}

type stubLeaseGetter struct {
	leases []dhcp.NetworkDHCPLease
	getErr error
}

func (c stubLeaseGetter) GetDHCPLeases() ([]dhcp.NetworkDHCPLease, error) {
	return c.leases, c.getErr
}
