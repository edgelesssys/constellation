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
	osProfile := struct {
		ComputerName string `json:"computerName,omitempty"`
	}{
		ComputerName: "computer-name",
	}

	response := metadataResponse{
		Compute: metadataResponseCompute{
			ResourceID:     "resource-id",
			SubscriptionID: "subscription-id",
			ResourceGroup:  "resource-group",
			Tags:           uidTags,
			OSProfile:      osProfile,
		},
	}
	responseWithoutID := metadataResponse{
		Compute: metadataResponseCompute{
			ResourceGroup:  "resource-group",
			SubscriptionID: "subscription-id",
			Tags:           uidTags,
			OSProfile:      osProfile,
		},
	}
	responseWithoutGroup := metadataResponse{
		Compute: metadataResponseCompute{
			ResourceID:     "resource-id",
			SubscriptionID: "subscription-id",
			Tags:           uidTags,
			OSProfile:      osProfile,
		},
	}
	responseWithoutUID := metadataResponse{
		Compute: metadataResponseCompute{
			ResourceID:     "resource-id",
			SubscriptionID: "subscription-id",
			ResourceGroup:  "resource-group",
			Tags:           []metadataTag{{Name: cloud.TagRole, Value: "worker"}},
			OSProfile:      osProfile,
		},
	}
	responseWithoutRole := metadataResponse{
		Compute: metadataResponseCompute{
			ResourceID:     "resource-id",
			SubscriptionID: "subscription-id",
			ResourceGroup:  "resource-group",
			Tags:           []metadataTag{{Name: cloud.TagUID, Value: "uid"}},
			OSProfile:      osProfile,
		},
	}
	responseWithoutName := metadataResponse{
		Compute: metadataResponseCompute{
			ResourceID:     "resource-id",
			SubscriptionID: "subscription-id",
			ResourceGroup:  "resource-group",
			Tags:           uidTags,
		},
	}
	responseWithoutSubscriptionID := metadataResponse{
		Compute: metadataResponseCompute{
			ResourceID:    "resource-id",
			ResourceGroup: "resource-group",
			Tags:          uidTags,
			OSProfile:     osProfile,
		},
	}

	defaultWantTags := map[string]string{
		cloud.TagUID:  "uid",
		cloud.TagRole: "worker",
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
		wantNameErr          bool
		wantName             string
		wantSubscriptionErr  bool
		wantSubscriptionID   string
		wantTagsErr          bool
		wantTags             map[string]string
	}{
		"metadata response parsed": {
			server:             newHTTPBufconnServerWithMetadataResponse(response),
			wantProviderID:     "resource-id",
			wantResourceGroup:  "resource-group",
			wantUID:            "uid",
			wantRole:           role.Worker,
			wantName:           "computer-name",
			wantSubscriptionID: "subscription-id",
			wantTags:           defaultWantTags,
		},
		"metadata response without resource ID": {
			server:             newHTTPBufconnServerWithMetadataResponse(responseWithoutID),
			wantProviderIDErr:  true,
			wantResourceGroup:  "resource-group",
			wantUID:            "uid",
			wantRole:           role.Worker,
			wantName:           "computer-name",
			wantSubscriptionID: "subscription-id",
			wantTags:           defaultWantTags,
		},
		"metadata response without UID tag": {
			server:             newHTTPBufconnServerWithMetadataResponse(responseWithoutUID),
			wantProviderID:     "resource-id",
			wantResourceGroup:  "resource-group",
			wantUIDErr:         true,
			wantRole:           role.Worker,
			wantName:           "computer-name",
			wantSubscriptionID: "subscription-id",
			wantTags:           map[string]string{cloud.TagRole: "worker"},
		},
		"metadata response without role tag": {
			server:             newHTTPBufconnServerWithMetadataResponse(responseWithoutRole),
			wantProviderID:     "resource-id",
			wantResourceGroup:  "resource-group",
			wantUID:            "uid",
			wantRoleErr:        true,
			wantName:           "computer-name",
			wantSubscriptionID: "subscription-id",
			wantTags:           map[string]string{cloud.TagUID: "uid"},
		},
		"metadata response without resource group": {
			server:               newHTTPBufconnServerWithMetadataResponse(responseWithoutGroup),
			wantProviderID:       "resource-id",
			wantResourceGroupErr: true,
			wantUID:              "uid",
			wantRole:             role.Worker,
			wantName:             "computer-name",
			wantSubscriptionID:   "subscription-id",
			wantTags:             defaultWantTags,
		},
		"metadata response without name": {
			server:             newHTTPBufconnServerWithMetadataResponse(responseWithoutName),
			wantProviderID:     "resource-id",
			wantResourceGroup:  "resource-group",
			wantUID:            "uid",
			wantRole:           role.Worker,
			wantNameErr:        true,
			wantSubscriptionID: "subscription-id",
			wantTags:           defaultWantTags,
		},
		"metadata response without subscription ID": {
			server:              newHTTPBufconnServerWithMetadataResponse(responseWithoutSubscriptionID),
			wantProviderID:      "resource-id",
			wantResourceGroup:   "resource-group",
			wantUID:             "uid",
			wantRole:            role.Worker,
			wantName:            "computer-name",
			wantSubscriptionErr: true,
			wantTags:            defaultWantTags,
		},
		"invalid imds response detected": {
			server: newHTTPBufconnServer(func(writer http.ResponseWriter, _ *http.Request) {
				fmt.Fprintln(writer, "invalid-result")
			}),
			wantProviderIDErr:    true,
			wantResourceGroupErr: true,
			wantUIDErr:           true,
			wantRoleErr:          true,
			wantNameErr:          true,
			wantSubscriptionErr:  true,
			wantTagsErr:          true,
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
			iClient := IMDSClient{client: &hClient}

			ctx := context.Background()

			id, err := iClient.providerID(ctx)
			if tc.wantProviderIDErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantProviderID, id)
			}

			group, err := iClient.resourceGroup(ctx)
			if tc.wantResourceGroupErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantResourceGroup, group)
			}

			uid, err := iClient.uid(ctx)
			if tc.wantUIDErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantUID, uid)
			}

			role, err := iClient.role(ctx)
			if tc.wantRoleErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantRole, role)
			}

			name, err := iClient.name(ctx)
			if tc.wantNameErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantName, name)
			}

			subscriptionID, err := iClient.subscriptionID(ctx)
			if tc.wantSubscriptionErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantSubscriptionID, subscriptionID)
			}

			tags, err := iClient.Tags(ctx)
			if tc.wantTagsErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantTags, tags)
			}
		})
	}
}

type httpBufconnServer struct {
	*httptest.Server
	*bufconn.Listener
}

func (s *httpBufconnServer) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	return s.Listener.DialContext(ctx)
}

func (s *httpBufconnServer) Dial(_, _ string) (net.Conn, error) {
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
