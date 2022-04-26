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
		issuer     Issuer
		validators []Validator
		wantErr    bool
	}{
		"basic": {
			issuer:     fakeIssuer{fakeOID: oid1},
			validators: []Validator{fakeValidator{fakeOID: oid1}},
		},
		"multiple validators": {
			issuer:     fakeIssuer{fakeOID: oid2},
			validators: []Validator{fakeValidator{fakeOID: oid1}, fakeValidator{fakeOID: oid2}},
		},
		"validate error": {
			issuer:     fakeIssuer{fakeOID: oid1},
			validators: []Validator{fakeValidator{fakeOID: oid1, err: errors.New("failed")}},
			wantErr:    true,
		},
		"unknown oid": {
			issuer:     fakeIssuer{fakeOID: oid1},
			validators: []Validator{fakeValidator{fakeOID: oid2}},
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			//
			// Create server
			//

			serverConfig, err := CreateAttestationServerTLSConfig(tc.issuer)
			require.NoError(err)

			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, "hello")
			}))
			server.TLS = serverConfig

			//
			// Create client
			//

			clientConfig, err := CreateAttestationClientTLSConfig(tc.validators)
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
