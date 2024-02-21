/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package atls

import (
	"context"
	"encoding/asn1"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestTLSConfig(t *testing.T) {
	oid1 := fakeOID{asn1.ObjectIdentifier{1, 3, 9900, 1}}
	oid2 := fakeOID{asn1.ObjectIdentifier{1, 3, 9900, 2}}

	testCases := map[string]struct {
		clientIssuer     Issuer
		clientValidators []Validator
		serverIssuer     Issuer
		serverValidators []Validator
		wantErr          bool
	}{
		"client->server basic": {
			serverIssuer:     NewFakeIssuer(oid1),
			clientValidators: []Validator{NewFakeValidator(oid1)},
		},
		"client->server multiple validators": {
			serverIssuer:     NewFakeIssuer(oid2),
			clientValidators: []Validator{NewFakeValidator(oid1), NewFakeValidator(oid2)},
		},
		"client->server validate error": {
			serverIssuer:     NewFakeIssuer(oid1),
			clientValidators: []Validator{FakeValidator{oid1, errors.New("failed")}},
			wantErr:          true,
		},
		"client->server unknown oid": {
			serverIssuer:     NewFakeIssuer(oid1),
			clientValidators: []Validator{NewFakeValidator(oid2)},
			wantErr:          true,
		},
		"client->server client cert is not verified": {
			serverIssuer:     NewFakeIssuer(oid1),
			clientValidators: []Validator{NewFakeValidator(oid1)},
		},
		"server->client basic": {
			serverValidators: []Validator{NewFakeValidator(oid1)},
			clientIssuer:     NewFakeIssuer(oid1),
		},
		"server->client multiple validators": {
			serverValidators: []Validator{NewFakeValidator(oid1), NewFakeValidator(oid2)},
			clientIssuer:     NewFakeIssuer(oid2),
		},
		"server->client validate error": {
			serverValidators: []Validator{FakeValidator{oid1, errors.New("failed")}},
			clientIssuer:     NewFakeIssuer(oid1),
			wantErr:          true,
		},
		"server->client unknown oid": {
			serverValidators: []Validator{NewFakeValidator(oid2)},
			clientIssuer:     NewFakeIssuer(oid1),
			wantErr:          true,
		},
		"mutual basic": {
			serverIssuer:     NewFakeIssuer(oid1),
			serverValidators: []Validator{NewFakeValidator(oid1)},
			clientIssuer:     NewFakeIssuer(oid1),
			clientValidators: []Validator{NewFakeValidator(oid1)},
		},
		"mutual multiple validators": {
			serverIssuer:     NewFakeIssuer(oid2),
			serverValidators: []Validator{NewFakeValidator(oid1), NewFakeValidator(oid2)},
			clientIssuer:     NewFakeIssuer(oid2),
			clientValidators: []Validator{NewFakeValidator(oid1), NewFakeValidator(oid2)},
		},
		"mutual fails if client sends no attestation": {
			serverIssuer:     NewFakeIssuer(oid1),
			serverValidators: []Validator{NewFakeValidator(oid1)},
			clientValidators: []Validator{NewFakeValidator(oid1)},
			wantErr:          true,
		},
		"mutual fails if server sends no attestation": {
			serverValidators: []Validator{NewFakeValidator(oid1)},
			clientIssuer:     NewFakeIssuer(oid1),
			clientValidators: []Validator{NewFakeValidator(oid1)},
			wantErr:          true,
		},
		"mutual validate error client side": {
			serverIssuer:     NewFakeIssuer(oid1),
			serverValidators: []Validator{NewFakeValidator(oid1)},
			clientIssuer:     NewFakeIssuer(oid1),
			clientValidators: []Validator{FakeValidator{oid1, errors.New("failed")}},
			wantErr:          true,
		},
		"mutual validate error server side": {
			serverIssuer:     NewFakeIssuer(oid1),
			serverValidators: []Validator{FakeValidator{oid1, errors.New("failed")}},
			clientIssuer:     NewFakeIssuer(oid1),
			clientValidators: []Validator{NewFakeValidator(oid1)},
			wantErr:          true,
		},
		"mutual unknown oid from client": {
			serverIssuer:     NewFakeIssuer(oid1),
			serverValidators: []Validator{NewFakeValidator(oid1)},
			clientIssuer:     NewFakeIssuer(oid2),
			clientValidators: []Validator{NewFakeValidator(oid1)},
			wantErr:          true,
		},
		"mutual unknown oid from server": {
			serverIssuer:     NewFakeIssuer(oid2),
			serverValidators: []Validator{NewFakeValidator(oid1)},
			clientIssuer:     NewFakeIssuer(oid1),
			clientValidators: []Validator{NewFakeValidator(oid1)},
			wantErr:          true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			//
			// Create server
			//

			serverConfig, err := CreateAttestationServerTLSConfig(tc.serverIssuer, tc.serverValidators)
			require.NoError(err)

			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, "hello")
			}))
			server.TLS = serverConfig

			//
			// Create client
			//

			clientConfig, err := CreateAttestationClientTLSConfig(tc.clientIssuer, tc.clientValidators)
			require.NoError(err)
			client := http.Client{Transport: &http.Transport{TLSClientConfig: clientConfig}}

			//
			// Test connection
			//

			server.StartTLS()
			defer server.Close()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, http.NoBody)
			require.NoError(err)
			resp, err := client.Do(req)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(err)
			assert.EqualValues("hello", body)
		})
	}
}

func TestClientConnectionConcurrency(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	//
	// Create servers.
	//

	const serverCount = 15

	var urls []string

	for i := 0; i < serverCount; i++ {
		serverCfg, err := CreateAttestationServerTLSConfig(NewFakeIssuer(variant.Dummy{}), NewFakeValidators(variant.Dummy{}))
		require.NoError(err)

		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, "hello")
		}))
		server.TLS = serverCfg

		server.StartTLS()
		defer server.Close()

		urls = append(urls, server.URL)
	}

	//
	// Create client.
	//

	clientConfig, err := CreateAttestationClientTLSConfig(NewFakeIssuer(variant.Dummy{}), NewFakeValidators(variant.Dummy{}))
	require.NoError(err)
	client := http.Client{Transport: &http.Transport{TLSClientConfig: clientConfig}}

	//
	// Prepare a request for each server.
	//

	var reqs []*http.Request
	for _, url := range urls {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
		require.NoError(err)
		reqs = append(reqs, req)
	}

	//
	// Do the request concurrently and collect the errors in a channel.
	// The config of the client is reused, so the nonce isn't fresh.
	// This explicitly checks for data races on the clientConnection.
	//

	errChan := make(chan error, serverCount)

	for _, req := range reqs {
		go func(req *http.Request) {
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
			}
			errChan <- err
		}(req)
	}

	//
	// Wait for the requests to finish and check the errors.
	//

	for i := 0; i < serverCount; i++ {
		assert.NoError(<-errChan)
	}
}

func TestServerConnectionConcurrency(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	//
	// Create servers.
	// The serverCfg is reused.
	//

	const serverCount = 10

	var urls []string

	serverCfg, err := CreateAttestationServerTLSConfig(NewFakeIssuer(variant.Dummy{}), NewFakeValidators(variant.Dummy{}))
	require.NoError(err)

	for i := 0; i < serverCount; i++ {
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, "hello")
		}))
		server.TLS = serverCfg

		server.StartTLS()
		defer server.Close()

		urls = append(urls, server.URL)
	}

	//
	// Create client.
	//

	clientConfig, err := CreateAttestationClientTLSConfig(NewFakeIssuer(variant.Dummy{}), NewFakeValidators(variant.Dummy{}))
	require.NoError(err)
	client := http.Client{Transport: &http.Transport{TLSClientConfig: clientConfig}}

	//
	// Prepare a request for each server.
	//

	var reqs []*http.Request
	for _, url := range urls {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
		require.NoError(err)
		reqs = append(reqs, req)
	}

	//
	// Do the request concurrently and collect the errors in a channel.
	//

	errChan := make(chan error, serverCount)

	for _, req := range reqs {
		go func(req *http.Request) {
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
			}
			errChan <- err
		}(req)
	}

	//
	// Wait for the requests to finish and check the errors.
	//

	for i := 0; i < serverCount; i++ {
		assert.NoError(<-errChan)
	}
}
