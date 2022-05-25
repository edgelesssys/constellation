package atls

import (
	"bytes"
	"context"
	"encoding/asn1"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTLSConfig(t *testing.T) {
	oid1 := fakeOID{1, 3, 9900, 1}
	oid2 := fakeOID{1, 3, 9900, 2}

	testCases := map[string]struct {
		clientIssuer     Issuer
		clientValidators []Validator
		serverIssuer     Issuer
		serverValidators []Validator
		wantErr          bool
	}{
		"client->server basic": {
			serverIssuer:     fakeIssuer{fakeOID: oid1},
			clientValidators: []Validator{fakeValidator{fakeOID: oid1}},
		},
		"client->server multiple validators": {
			serverIssuer:     fakeIssuer{fakeOID: oid2},
			clientValidators: []Validator{fakeValidator{fakeOID: oid1}, fakeValidator{fakeOID: oid2}},
		},
		"client->server validate error": {
			serverIssuer:     fakeIssuer{fakeOID: oid1},
			clientValidators: []Validator{fakeValidator{fakeOID: oid1, err: errors.New("failed")}},
			wantErr:          true,
		},
		"client->server unknown oid": {
			serverIssuer:     fakeIssuer{fakeOID: oid1},
			clientValidators: []Validator{fakeValidator{fakeOID: oid2}},
			wantErr:          true,
		},
		"client->server client cert is not verified": {
			serverIssuer:     fakeIssuer{fakeOID: oid1},
			clientValidators: []Validator{fakeValidator{fakeOID: oid1}},
		},
		"server->client basic": {
			serverValidators: []Validator{fakeValidator{fakeOID: oid1}},
			clientIssuer:     fakeIssuer{fakeOID: oid1},
		},
		"server->client multiple validators": {
			serverValidators: []Validator{fakeValidator{fakeOID: oid1}, fakeValidator{fakeOID: oid2}},
			clientIssuer:     fakeIssuer{fakeOID: oid2},
		},
		"server->client validate error": {
			serverValidators: []Validator{fakeValidator{fakeOID: oid1, err: errors.New("failed")}},
			clientIssuer:     fakeIssuer{fakeOID: oid1},
			wantErr:          true,
		},
		"server->client unknown oid": {
			serverValidators: []Validator{fakeValidator{fakeOID: oid2}},
			clientIssuer:     fakeIssuer{fakeOID: oid1},
			wantErr:          true,
		},
		"mutual basic": {
			serverIssuer:     fakeIssuer{fakeOID: oid1},
			serverValidators: []Validator{fakeValidator{fakeOID: oid1}},
			clientIssuer:     fakeIssuer{fakeOID: oid1},
			clientValidators: []Validator{fakeValidator{fakeOID: oid1}},
		},
		"mutual multiple validators": {
			serverIssuer:     fakeIssuer{fakeOID: oid2},
			serverValidators: []Validator{fakeValidator{fakeOID: oid1}, fakeValidator{fakeOID: oid2}},
			clientIssuer:     fakeIssuer{fakeOID: oid2},
			clientValidators: []Validator{fakeValidator{fakeOID: oid1}, fakeValidator{fakeOID: oid2}},
		},
		"mutual fails if client sends no attestation": {
			serverIssuer:     fakeIssuer{fakeOID: oid1},
			serverValidators: []Validator{fakeValidator{fakeOID: oid1}},
			clientValidators: []Validator{fakeValidator{fakeOID: oid1}},
			wantErr:          true,
		},
		"mutual fails if server sends no attestation": {
			serverValidators: []Validator{fakeValidator{fakeOID: oid1}},
			clientIssuer:     fakeIssuer{fakeOID: oid1},
			clientValidators: []Validator{fakeValidator{fakeOID: oid1}},
			wantErr:          true,
		},
		"mutual validate error client side": {
			serverIssuer:     fakeIssuer{fakeOID: oid1},
			serverValidators: []Validator{fakeValidator{fakeOID: oid1}},
			clientIssuer:     fakeIssuer{fakeOID: oid1},
			clientValidators: []Validator{fakeValidator{fakeOID: oid1, err: errors.New("failed")}},
			wantErr:          true,
		},
		"mutual validate error server side": {
			serverIssuer:     fakeIssuer{fakeOID: oid1},
			serverValidators: []Validator{fakeValidator{fakeOID: oid1, err: errors.New("failed")}},
			clientIssuer:     fakeIssuer{fakeOID: oid1},
			clientValidators: []Validator{fakeValidator{fakeOID: oid1}},
			wantErr:          true,
		},
		"mutual unknown oid from client": {
			serverIssuer:     fakeIssuer{fakeOID: oid1},
			serverValidators: []Validator{fakeValidator{fakeOID: oid1}},
			clientIssuer:     fakeIssuer{fakeOID: oid2},
			clientValidators: []Validator{fakeValidator{fakeOID: oid1}},
			wantErr:          true,
		},
		"mutual unknown oid from server": {
			serverIssuer:     fakeIssuer{fakeOID: oid2},
			serverValidators: []Validator{fakeValidator{fakeOID: oid1}},
			clientIssuer:     fakeIssuer{fakeOID: oid1},
			clientValidators: []Validator{fakeValidator{fakeOID: oid1}},
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

			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

type fakeIssuer struct {
	fakeOID
}

func (fakeIssuer) Issue(userData []byte, nonce []byte) ([]byte, error) {
	return json.Marshal(fakeDoc{UserData: userData, Nonce: nonce})
}

type fakeValidator struct {
	fakeOID
	err error
}

func (v fakeValidator) Validate(attDoc []byte, nonce []byte) ([]byte, error) {
	var doc fakeDoc
	if err := json.Unmarshal(attDoc, &doc); err != nil {
		return nil, err
	}
	if !bytes.Equal(doc.Nonce, nonce) {
		return nil, errors.New("invalid nonce")
	}
	return doc.UserData, v.err
}

type fakeOID asn1.ObjectIdentifier

func (o fakeOID) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier(o)
}

type fakeDoc struct {
	UserData []byte
	Nonce    []byte
}
