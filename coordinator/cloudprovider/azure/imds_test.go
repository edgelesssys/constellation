package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/test/bufconn"
)

func TestRetrieve(t *testing.T) {
	response := metadataResponse{
		Compute: struct {
			ResourceID string `json:"resourceId,omitempty"`
		}{
			ResourceID: "resource-id",
		},
	}
	testCases := map[string]struct {
		server           httpBufconnServer
		expectErr        bool
		expectedResponse metadataResponse
	}{
		"metadata response parsed": {
			server:           newHTTPBufconnServerWithMetadataResponse(response),
			expectedResponse: response,
		},
		"invalid imds response detected": {
			server: newHTTPBufconnServer(func(writer http.ResponseWriter, request *http.Request) {
				fmt.Fprintln(writer, "invalid-result")
			}),
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			defer tc.server.Close()

			hClient := http.Client{
				Transport: &http.Transport{
					DialContext:    tc.server.DialContext,
					Dial:           tc.server.Dial,
					DialTLSContext: tc.server.DialContext,
					DialTLS:        tc.server.Dial,
				},
			}
			iClient := imdsClient{
				client: &hClient,
			}
			resp, err := iClient.Retrieve(context.Background())

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedResponse, resp)
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
