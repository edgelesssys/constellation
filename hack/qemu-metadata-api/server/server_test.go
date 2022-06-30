package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/hack/qemu-metadata-api/virtwrapper"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"libvirt.org/go/libvirt"
)

func TestListAll(t *testing.T) {
	someErr := errors.New("error")

	testCases := map[string]struct {
		wantErr bool
		connect *stubConnect
	}{
		"success": {
			connect: &stubConnect{
				network: stubNetwork{
					leases: []libvirt.NetworkDHCPLease{
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

			server := New(logger.NewTest(t), tc.connect)

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
				network: stubNetwork{
					leases: []libvirt.NetworkDHCPLease{
						{
							IPaddr:   "192.0.100.1",
							Hostname: "control-plane-0",
						},
					},
				},
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
				network: stubNetwork{
					leases: []libvirt.NetworkDHCPLease{
						{
							IPaddr:   "192.0.100.1",
							Hostname: "control-plane-0",
						},
					},
				},
			},
			wantErr: true,
		},
		"peer not found": {
			remoteAddr: "192.0.200.1:1234",
			connect: &stubConnect{
				network: stubNetwork{
					leases: []libvirt.NetworkDHCPLease{
						{
							IPaddr:   "192.0.100.1",
							Hostname: "control-plane-0",
						},
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

			server := New(logger.NewTest(t), tc.connect)

			req, err := http.NewRequest(http.MethodGet, "http://192.0.0.1/self", nil)
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

			var metadata cloudtypes.Instance
			require.NoError(json.Unmarshal(metadataRaw, &metadata))
			assert.Equal(tc.connect.network.leases[0].Hostname, metadata.Name)
			assert.Equal(tc.connect.network.leases[0].IPaddr, metadata.PublicIPs[0])
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
				network: stubNetwork{
					leases: []libvirt.NetworkDHCPLease{
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

			server := New(logger.NewTest(t), tc.connect)

			req, err := http.NewRequest(http.MethodGet, "http://192.0.0.1/peers", nil)
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

			var metadata []cloudtypes.Instance
			require.NoError(json.Unmarshal(metadataRaw, &metadata))
			assert.Len(metadata, len(tc.connect.network.leases))
		})
	}
}

type stubConnect struct {
	network       stubNetwork
	getNetworkErr error
}

func (c stubConnect) LookupNetworkByName(name string) (*virtwrapper.Network, error) {
	return &virtwrapper.Network{Net: c.network}, c.getNetworkErr
}

type stubNetwork struct {
	leases      []libvirt.NetworkDHCPLease
	getLeaseErr error
}

func (n stubNetwork) GetDHCPLeases() ([]libvirt.NetworkDHCPLease, error) {
	return n.leases, n.getLeaseErr
}

func (n stubNetwork) Free() error {
	return nil
}
