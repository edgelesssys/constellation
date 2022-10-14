/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/test/bufconn"
)

func TestIMDSClient(t *testing.T) {
	uidTags := []metadataTag{
		{Name: cloud.TagUID, Value: "uid"},
		{Name: cloud.TagRole, Value: "worker"},
	}
	response := metadataResponse{
		Compute: metadataResponseCompute{
			ResourceID:    "resource-id",
			ResourceGroup: "resource-group",
			Tags:          uidTags,
		},
	}
	responseWithoutID := metadataResponse{
		Compute: metadataResponseCompute{
			ResourceGroup: "resource-group",
			Tags:          uidTags,
		},
	}
	responseWithoutGroup := metadataResponse{
		Compute: metadataResponseCompute{
			ResourceID: "resource-id",
			Tags:       uidTags,
		},
	}
	responseWithoutUID := metadataResponse{
		Compute: metadataResponseCompute{
			ResourceID:    "resource-id",
			ResourceGroup: "resource-group",
			Tags:          []metadataTag{{Name: cloud.TagRole, Value: "worker"}},
		},
	}
	responseWithoutRole := metadataResponse{
		Compute: metadataResponseCompute{
			ResourceID:    "resource-id",
			ResourceGroup: "resource-group",
			Tags:          []metadataTag{{Name: cloud.TagUID, Value: "uid"}},
		},
	}

	testCases := map[string]struct {
		server               httpBufconnServer
		wantProviderIDErr    bool
		wantProviderID       string
		wantResourceGroupErr bool
		wantResourceGroup    string
		wantUIDErr           bool
		wantUID              string
		wantRoleErr          bool
		wantRole             role.Role
	}{
		"metadata response parsed": {
			server:            newHTTPBufconnServerWithMetadataResponse(response),
			wantProviderID:    "resource-id",
			wantResourceGroup: "resource-group",
			wantUID:           "uid",
			wantRole:          role.Worker,
		},
		"metadata response without resource ID": {
			server:            newHTTPBufconnServerWithMetadataResponse(responseWithoutID),
			wantProviderIDErr: true,
			wantResourceGroup: "resource-group",
			wantUID:           "uid",
			wantRole:          role.Worker,
		},
		"metadata response without UID tag": {
			server:            newHTTPBufconnServerWithMetadataResponse(responseWithoutUID),
			wantProviderID:    "resource-id",
			wantResourceGroup: "resource-group",
			wantUIDErr:        true,
			wantRole:          role.Worker,
		},
		"metadata response without role tag": {
			server:            newHTTPBufconnServerWithMetadataResponse(responseWithoutRole),
			wantProviderID:    "resource-id",
			wantResourceGroup: "resource-group",
			wantUID:           "uid",
			wantRoleErr:       true,
		},
		"metadata response without resource group": {
			server:               newHTTPBufconnServerWithMetadataResponse(responseWithoutGroup),
			wantProviderID:       "resource-id",
			wantResourceGroupErr: true,
			wantUID:              "uid",
			wantRole:             role.Worker,
		},
		"invalid imds response detected": {
			server: newHTTPBufconnServer(func(writer http.ResponseWriter, request *http.Request) {
				fmt.Fprintln(writer, "invalid-result")
			}),
			wantProviderIDErr:    true,
			wantResourceGroupErr: true,
			wantUIDErr:           true,
			wantRoleErr:          true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			defer tc.server.Close()

			hClient := http.Client{
				Transport: &http.Transport{
					DialContext:    tc.server.DialContext,
					Dial:           tc.server.Dial,
					DialTLSContext: tc.server.DialContext,
					DialTLS:        tc.server.Dial,
				},
			}
			iClient := imdsClient{client: &hClient}

			ctx := context.Background()

			id, err := iClient.ProviderID(ctx)
			if tc.wantProviderIDErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantProviderID, id)
			}

			group, err := iClient.ResourceGroup(ctx)
			if tc.wantResourceGroupErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantResourceGroup, group)
			}

			uid, err := iClient.UID(ctx)
			if tc.wantUIDErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantUID, uid)
			}

			role, err := iClient.Role(ctx)
			if tc.wantRoleErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantRole, role)
			}
		})
	}
}

type httpBufconnServer struct {
	*httptest.Server
	*bufconn.Listener
}

func (s *httpBufconnServer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return s.Listener.DialContext(ctx)
}

func (s *httpBufconnServer) Dial(network, addr string) (net.Conn, error) {
	return s.Listener.Dial()
}

func (s *httpBufconnServer) Close() {
	s.Server.Close()
	s.Listener.Close()
}

func newHTTPBufconnServer(handlerFunc http.HandlerFunc) httpBufconnServer {
	server := httptest.NewUnstartedServer(handlerFunc)
	listener := bufconn.Listen(1024)
	server.Listener = listener
	server.Start()
	return httpBufconnServer{
		Server:   server,
		Listener: listener,
	}
}

func newHTTPBufconnServerWithMetadataResponse(res metadataResponse) httpBufconnServer {
	return newHTTPBufconnServer(func(writer http.ResponseWriter, request *http.Request) {
		if request.Host != "169.254.169.254" || request.Header.Get("Metadata") != "True" || request.URL.Query().Get("format") != "json" || request.URL.Query().Get("api-version") != imdsAPIVersion {
			writer.WriteHeader(http.StatusInternalServerError)
			_, err := writer.Write([]byte("error"))
			if err != nil {
				panic(err)
			}
			return
		}
		rawResp, err := json.Marshal(res)
		if err != nil {
			panic(err)
		}
		_, err = writer.Write(rawResp)
		if err != nil {
			panic(err)
		}
	})
}
