package validator

import (
	"bytes"
	"context"
	"encoding/asn1"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpdateableValidator(t *testing.T) {
	testCases := map[string]struct {
		provider  string
		writeFile bool
		wantErr   bool
	}{
		"azure": {
			provider:  "azure",
			writeFile: true,
		},
		"gcp": {
			provider:  "gcp",
			writeFile: true,
		},
		"qemu": {
			provider:  "qemu",
			writeFile: true,
		},
		"no file": {
			provider:  "azure",
			writeFile: false,
			wantErr:   true,
		},
		"invalid provider": {
			provider:  "invalid",
			writeFile: true,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			handler := file.NewHandler(afero.NewMemMapFs())
			if tc.writeFile {
				require.NoError(handler.WriteJSON(
					filepath.Join(constants.ActivationBasePath, constants.ActivationMeasurementsFilename),
					map[uint32][]byte{
						11: {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
					},
					file.OptNone,
				))
			}

			_, err := New(tc.provider, handler)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	oid := fakeOID{1, 3, 9900, 1}
	newValidator := func(m map[uint32][]byte) atls.Validator {
		return fakeValidator{fakeOID: oid}
	}
	handler := file.NewHandler(afero.NewMemMapFs())

	// create server
	validator := &Updatable{newValidator: newValidator, fileHandler: handler}

	// Update should fail if the file does not exist
	assert.Error(validator.Update())

	// write measurement config
	require.NoError(handler.WriteJSON(
		filepath.Join(constants.ActivationBasePath, constants.ActivationMeasurementsFilename),
		map[uint32][]byte{
			11: {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		file.OptNone,
	))

	// call update once to initialize the server's validator
	require.NoError(validator.Update())

	// create tls config and start the server
	serverConfig, err := atls.CreateAttestationServerTLSConfig(nil, []atls.Validator{validator})
	require.NoError(err)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "hello")
	}))
	server.TLS = serverConfig
	server.StartTLS()
	defer server.Close()

	// test connection to server
	clientOID := fakeOID{1, 3, 9900, 1}
	resp, err := testConnection(require, server.URL, clientOID)
	require.NoError(err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(err)
	assert.EqualValues("hello", body)

	// update the server's validator
	oid = fakeOID{1, 3, 9900, 2}
	require.NoError(validator.Update())

	// client connection should fail now, since the server's validator expects a different OID from the client
	_, err = testConnection(require, server.URL, clientOID)
	assert.Error(err)
}

func TestUpdateConcurrency(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	handler := file.NewHandler(afero.NewMemMapFs())
	validator := &Updatable{
		fileHandler: handler,
		newValidator: func(m map[uint32][]byte) atls.Validator {
			return fakeValidator{fakeOID: fakeOID{1, 3, 9900, 1}}
		},
	}
	require.NoError(handler.WriteJSON(
		filepath.Join(constants.ActivationBasePath, constants.ActivationMeasurementsFilename),
		map[uint32][]byte{
			11: {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		file.OptNone,
	))

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NoError(validator.Update())
		}()
	}

	wg.Wait()
}

func testConnection(require *require.Assertions, url string, oid fakeOID) (*http.Response, error) {
	clientConfig, err := atls.CreateAttestationClientTLSConfig(fakeIssuer{fakeOID: oid}, nil)
	require.NoError(err)
	client := http.Client{Transport: &http.Transport{TLSClientConfig: clientConfig}}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	require.NoError(err)
	return client.Do(req)
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
