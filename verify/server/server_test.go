/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/verify/verifyproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestRun(t *testing.T) {
	assert := assert.New(t)
	closedErr := errors.New("closed")

	var err error
	var wg sync.WaitGroup
	s := &Server{
    log:    slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
		issuer: stubIssuer{attestation: []byte("quote")},
	}

	httpListener, grpcListener := setUpTestListeners()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = s.Run(httpListener, grpcListener)
	}()
	assert.NoError(httpListener.Close())
	wg.Wait()
	assert.Equal(err, closedErr)

	httpListener, grpcListener = setUpTestListeners()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = s.Run(httpListener, grpcListener)
	}()
	assert.NoError(grpcListener.Close())
	wg.Wait()
	assert.Equal(err, closedErr)

	httpListener, grpcListener = setUpTestListeners()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = s.Run(httpListener, grpcListener)
	}()
	go assert.NoError(grpcListener.Close())
	go assert.NoError(httpListener.Close())
	wg.Wait()
	assert.Equal(err, closedErr)
}

func TestGetAttestationGRPC(t *testing.T) {
	testCases := map[string]struct {
		issuer  stubIssuer
		request *verifyproto.GetAttestationRequest
		wantErr bool
	}{
		"success": {
			issuer: stubIssuer{attestation: []byte("quote")},
			request: &verifyproto.GetAttestationRequest{
				Nonce: []byte("nonce"),
			},
		},
		"issuer fails": {
			issuer: stubIssuer{issueErr: errors.New("issuer error")},
			request: &verifyproto.GetAttestationRequest{
				Nonce: []byte("nonce"),
			},
			wantErr: true,
		},
		"no nonce": {
			issuer:  stubIssuer{attestation: []byte("quote")},
			request: &verifyproto.GetAttestationRequest{},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			server := &Server{
        log:    slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				issuer: tc.issuer,
			}

			resp, err := server.GetAttestation(context.Background(), tc.request)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.issuer.attestation, resp.Attestation)
			}
		})
	}
}

func TestGetAttestationHTTP(t *testing.T) {
	testCases := map[string]struct {
		request string
		issuer  stubIssuer
		wantErr bool
	}{
		"success": {
			request: "?nonce=" + base64.URLEncoding.EncodeToString([]byte("nonce")),
			issuer:  stubIssuer{attestation: []byte("quote")},
		},
		"invalid nonce in query": {
			request: "?nonce=not-base-64",
			issuer:  stubIssuer{attestation: []byte("quote")},
			wantErr: true,
		},
		"no nonce in query": {
			request: "?foo=bar",
			issuer:  stubIssuer{attestation: []byte("quote")},
			wantErr: true,
		},
		"empty nonce in query": {
			request: "?nonce=",
			issuer:  stubIssuer{attestation: []byte("quote")},
			wantErr: true,
		},
		"issuer fails": {
			request: "?nonce=" + base64.URLEncoding.EncodeToString([]byte("nonce")),
			issuer:  stubIssuer{issueErr: errors.New("errors")},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			server := &Server{
        log:    slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				issuer: tc.issuer,
			}

			httpServer := httptest.NewServer(http.HandlerFunc(server.getAttestationHTTP))
			defer httpServer.Close()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, httpServer.URL+tc.request, nil)
			require.NoError(err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(err)
			defer resp.Body.Close()

			if tc.wantErr {
				assert.NotEqual(http.StatusOK, resp.StatusCode)
				return
			}
			assert.Equal(http.StatusOK, resp.StatusCode)
			quote, err := io.ReadAll(resp.Body)
			require.NoError(err)

			var rawQuote attestation
			require.NoError(json.Unmarshal(quote, &rawQuote))

			assert.Equal(tc.issuer.attestation, rawQuote.Data)
		})
	}
}

func setUpTestListeners() (net.Listener, net.Listener) {
	httpListener := testdialer.NewBufconnDialer().GetListener(net.JoinHostPort("192.0.2.1", "8080"))
	grpcListener := testdialer.NewBufconnDialer().GetListener(net.JoinHostPort("192.0.2.1", "8081"))
	return httpListener, grpcListener
}

type stubIssuer struct {
	attestation []byte
	issueErr    error
}

func (i stubIssuer) Issue(_ context.Context, _ []byte, _ []byte) ([]byte, error) {
	return i.attestation, i.issueErr
}
