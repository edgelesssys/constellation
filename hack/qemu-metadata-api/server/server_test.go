/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/virtwrapper"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAll(t *testing.T) {
	someErr := errors.New("error")

	testCases := map[string]struct {
		wantErr bool
		connect *stubConnect
	}{
		"success": {
			connect: &stubConnect{
				network: newStubNetwork([]virtwrapper.NetworkDHCPLease{
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
				}, nil),
			},
		},
		"LookupNetworkByName error": {
			connect: &stubConnect{
				getNetworkErr: someErr,
			},
			wantErr: true,
		},
		"GetDHCPLeases error": {
			connect: &stubConnect{
				network: stubNetwork{
					getLeaseErr: someErr,
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

      server := New(slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)), "test", "initSecretHash", tc.connect)

			res, err := server.listAll()

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Len(tc.connect.network.leases, len(res))
		})
	}
}

func TestListSelf(t *testing.T) {
	someErr := errors.New("error")

	testCases := map[string]struct {
		remoteAddr string
		connect    *stubConnect
		wantErr    bool
	}{
		"success": {
			remoteAddr: "192.0.100.1:1234",
			connect: &stubConnect{
				network: newStubNetwork([]virtwrapper.NetworkDHCPLease{
					{
						IPaddr:   "192.0.100.1",
						Hostname: "control-plane-0",
					},
				}, nil),
			},
		},
		"listAll error": {
			remoteAddr: "192.0.100.1:1234",
			connect: &stubConnect{
				getNetworkErr: someErr,
			},
			wantErr: true,
		},
		"remoteAddr error": {
			remoteAddr: "",
			connect: &stubConnect{
				network: newStubNetwork([]virtwrapper.NetworkDHCPLease{
					{
						IPaddr:   "192.0.100.1",
						Hostname: "control-plane-0",
					},
				}, nil),
			},
			wantErr: true,
		},
		"peer not found": {
			remoteAddr: "192.0.200.1:1234",
			connect: &stubConnect{
				network: newStubNetwork([]virtwrapper.NetworkDHCPLease{
					{
						IPaddr:   "192.0.100.1",
						Hostname: "control-plane-0",
					},
				}, nil),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

      server := New(slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)), "test", "initSecretHash", tc.connect)

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
			assert.Equal(tc.connect.network.leases[0].Hostname, metadata.Name)
			assert.Equal(tc.connect.network.leases[0].IPaddr, metadata.VPCIP)
		})
	}
}

func TestListPeers(t *testing.T) {
	testCases := map[string]struct {
		remoteAddr string
		connect    *stubConnect
		wantErr    bool
	}{
		"success": {
			remoteAddr: "192.0.100.1:1234",
			connect: &stubConnect{
				network: newStubNetwork([]virtwrapper.NetworkDHCPLease{
					{
						IPaddr:   "192.0.100.1",
						Hostname: "control-plane-0",
					},
					{
						IPaddr:   "192.0.200.1",
						Hostname: "worker-0",
					},
				}, nil),
			},
		},
		"listAll error": {
			remoteAddr: "192.0.100.1:1234",
			connect: &stubConnect{
				getNetworkErr: errors.New("error"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

      server := New(slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)), "test", "initSecretHash", tc.connect)

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
			assert.Len(metadata, len(tc.connect.network.leases))
		})
	}
}

func TestInitSecretHash(t *testing.T) {
	defaultConnect := &stubConnect{
		network: newStubNetwork([]virtwrapper.NetworkDHCPLease{
			{
				IPaddr:   "192.0.100.1",
				Hostname: "control-plane-0",
			},
		}, nil),
	}
	testCases := map[string]struct {
		connect  *stubConnect
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

      server := New(slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)), "test", tc.wantHash, defaultConnect)

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

type stubConnect struct {
	network       stubNetwork
	getNetworkErr error
}

func (c stubConnect) LookupNetworkByName(_ string) (*virtwrapper.Network, error) {
	return &virtwrapper.Network{Net: c.network}, c.getNetworkErr
}
