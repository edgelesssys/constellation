/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package watcher

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

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

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
					filepath.Join(constants.ServiceBasePath, constants.MeasurementsFilename),
					map[uint32][]byte{
						11: {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
					},
				))
				require.NoError(handler.WriteJSON(
					filepath.Join(constants.ServiceBasePath, constants.EnforcedPCRsFilename),
					[]uint32{11},
				))
				require.NoError(handler.Write(
					filepath.Join(constants.ServiceBasePath, constants.IDKeyDigestFilename),
					[]byte{},
				))
				require.NoError(handler.Write(
					filepath.Join(constants.ServiceBasePath, constants.EnforceIDKeyDigestFilename),
					[]byte("false"),
				))
				require.NoError(handler.Write(
					filepath.Join(constants.ServiceBasePath, constants.AzureCVM),
					[]byte("true"),
				))
			}

			_, err := NewValidator(
				logger.NewTest(t),
				tc.provider,
				handler,
				false,
			)
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

	// we need safe access for overwriting the fake validator OID
	oid := fakeOID{1, 3, 9900, 1}
	var oidLock sync.Mutex
	updatedOID := func(newOID fakeOID) {
		oidLock.Lock()
		defer oidLock.Unlock()
		oid = newOID
	}
	newValidator := func(m measurements.M, digest idkeydigest.IDKeyDigests, enforceIdKeyDigest bool, _ *logger.Logger) atls.Validator {
		oidLock.Lock()
		defer oidLock.Unlock()
		return fakeValidator{fakeOID: oid}
	}
	handler := file.NewHandler(afero.NewMemMapFs())

	// create server
	validator := &Updatable{
		log:          logger.NewTest(t),
		newValidator: newValidator,
		fileHandler:  handler,
	}

	// Update should fail if the file does not exist
	assert.Error(validator.Update())

	// write measurement config
	require.NoError(handler.WriteJSON(
		filepath.Join(constants.ServiceBasePath, constants.MeasurementsFilename),
		measurements.M{11: measurements.WithAllBytes(0x00, false, measurements.PCRMeasurementLength)},
	))
	require.NoError(handler.Write(
		filepath.Join(constants.ServiceBasePath, constants.IDKeyDigestFilename),
		[]byte{},
	))
	require.NoError(handler.Write(
		filepath.Join(constants.ServiceBasePath, constants.EnforceIDKeyDigestFilename),
		[]byte("false"),
	))
	require.NoError(handler.Write(
		filepath.Join(constants.ServiceBasePath, constants.AzureCVM),
		[]byte("true"),
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
	updatedOID(fakeOID{1, 3, 9900, 2})
	require.NoError(validator.Update())

	// client connection should fail now, since the server's validator expects a different OID from the client
	resp, err = testConnection(require, server.URL, clientOID)
	if err == nil {
		defer resp.Body.Close()
	}
	assert.Error(err)

	// update should work for legacy measurement format
	// TODO: remove with v2.4.0
	require.NoError(handler.WriteJSON(
		filepath.Join(constants.ServiceBasePath, constants.MeasurementsFilename),
		map[uint32][]byte{
			11: bytes.Repeat([]byte{0x0}, 32),
			12: bytes.Repeat([]byte{0x1}, 32),
		},
		file.OptOverwrite,
	))
	require.NoError(handler.WriteJSON(
		filepath.Join(constants.ServiceBasePath, constants.EnforcedPCRsFilename),
		[]uint32{11},
	))

	assert.NoError(validator.Update())
}

func TestOIDConcurrency(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	handler := file.NewHandler(afero.NewMemMapFs())
	require.NoError(handler.WriteJSON(
		filepath.Join(constants.ServiceBasePath, constants.MeasurementsFilename),
		measurements.M{11: measurements.WithAllBytes(0x00, false, measurements.PCRMeasurementLength)},
	))
	require.NoError(handler.Write(
		filepath.Join(constants.ServiceBasePath, constants.IDKeyDigestFilename),
		[]byte{},
	))

	newValidator := func(m measurements.M, digest idkeydigest.IDKeyDigests, enforceIdKeyDigest bool, _ *logger.Logger) atls.Validator {
		return fakeValidator{fakeOID: fakeOID{1, 3, 9900, 1}}
	}
	// create server
	validator := &Updatable{
		log:          logger.NewTest(t),
		newValidator: newValidator,
		fileHandler:  handler,
	}

	// call update once to initialize the server's validator
	require.NoError(validator.Update())

	var wg sync.WaitGroup
	wg.Add(2 * 20)
	for i := 0; i < 20; i++ {
		go func() {
			defer wg.Done()
			assert.NoError(validator.Update())
		}()
		go func() {
			defer wg.Done()
			validator.OID()
		}()
	}
	wg.Wait()
}

func TestUpdateConcurrency(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	handler := file.NewHandler(afero.NewMemMapFs())
	validator := &Updatable{
		log:         logger.NewTest(t),
		fileHandler: handler,
		newValidator: func(m measurements.M, digest idkeydigest.IDKeyDigests, enforceIdKeyDigest bool, _ *logger.Logger) atls.Validator {
			return fakeValidator{fakeOID: fakeOID{1, 3, 9900, 1}}
		},
	}
	require.NoError(handler.WriteJSON(
		filepath.Join(constants.ServiceBasePath, constants.MeasurementsFilename),
		map[uint32][]byte{
			11: {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		file.OptNone,
	))
	require.NoError(handler.WriteJSON(
		filepath.Join(constants.ServiceBasePath, constants.EnforcedPCRsFilename),
		[]uint32{11},
	))
	require.NoError(handler.Write(
		filepath.Join(constants.ServiceBasePath, constants.IDKeyDigestFilename),
		[]byte{},
	))
	require.NoError(handler.Write(
		filepath.Join(constants.ServiceBasePath, constants.EnforceIDKeyDigestFilename),
		[]byte("false"),
	))
	require.NoError(handler.Write(
		filepath.Join(constants.ServiceBasePath, constants.AzureCVM),
		[]byte("true"),
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
