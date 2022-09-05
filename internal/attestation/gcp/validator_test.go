/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

func TestGceNonHostInfoEvent(t *testing.T) {
	testCases := map[string]struct {
		attDoc  vtpm.AttestationDocument
		wantErr bool
	}{
		"is cvm": {
			attDoc: vtpm.AttestationDocument{
				Attestation: &attest.Attestation{
					EventLog: []byte("\x00\x00\x00GCE NonHostInfo\x00\x01\x00\x00"),
				},
			},
		},
		"attestation is nil": {
			attDoc: vtpm.AttestationDocument{
				Attestation: nil,
			},
			wantErr: true,
		},
		"missing GCE Non-Host info event": {
			attDoc: vtpm.AttestationDocument{
				Attestation: &attest.Attestation{
					EventLog: []byte("No GCE Event"),
				},
			},
			wantErr: true,
		},
		"not a cvm": {
			attDoc: vtpm.AttestationDocument{
				Attestation: &attest.Attestation{
					EventLog: []byte("\x00\x00\x00GCE NonHostInfo\x00\x00\x00\x00"),
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			err := gceNonHostInfoEvent(tc.attDoc)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestTrustedKeyFromGCEAPI(t *testing.T) {
	testPubK := `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAu+OepfHCTiTi27nkTGke
dn+AIkiM1AIWWDwqfqG85aNulcj60mGQGXIYV8LoEVkyKOhYBIUmJUaVczB4ltqq
ZhR7l46RQw2vnv+XiUmfK555d4ZDInyjTusO69hE6tkuYKdXLlG1HzcrhJ254LE2
wXtE1Yf9DygOsWet+S32gmpfH2whUY1mRTdwW4zoY4c3qtmmWImhVVNr6qR8Z95X
Y49EteCoNIomQNEZH7EnMlBsh34L7doOsckh1aTvQcrJorQSrBkWKbdV6kvuBKZp
fLK0DZiOh9BwZCZANtOqgH3V+AuNk338iON8eKCFRjoiQ40YGM6xKH3E6PHVnuKt
uIO0MPvE0qdV8Lvs+nCCrvwP5sJKZuciM40ioEO1pV1y3491xIxYhx3OfN4gg2h8
cgdKob/R8qwxqTrfceO36FBFb1vXCUApsm5oy6WxmUtIUgoYhK+6JYpVWDyOJYwP
iMJhdJA65n2ZliN8NxEhsaFoMgw76BOiD0wkt/CKPmNbOm5MGS3/fiZCt6A6u3cn
Ubhn4tvjy/q5XzVqZtBeoseW2TyyrsAN53LBkSqag5tG/264CQDigQ6Y/OADOE2x
n08MyrFHIL/wFMscOvJo7c2Eo4EW1yXkEkAy5tF5PZgnfRObakj4gdqPeq18FNzc
Y+t5OxL3kL15VzY1Ob0d5cMCAwEAAQ==
-----END PUBLIC KEY-----`

	testCases := map[string]struct {
		instanceInfo []byte
		getClient    func(ctx context.Context, opts ...option.ClientOption) (gcpRestClient, error)
		wantErr      bool
	}{
		"success": {
			instanceInfo: mustMarshal(attest.GCEInstanceInfo{}, require.New(t)),
			getClient: prepareFakeClient(&computepb.ShieldedInstanceIdentity{
				SigningKey: &computepb.ShieldedInstanceIdentityEntry{
					EkPub: proto.String(testPubK),
				},
			}, nil, nil),
			wantErr: false,
		},
		"Unmarshal error": {
			instanceInfo: []byte("error"),
			getClient: prepareFakeClient(&computepb.ShieldedInstanceIdentity{
				SigningKey: &computepb.ShieldedInstanceIdentityEntry{
					EkPub: proto.String(testPubK),
				},
			}, nil, nil),
			wantErr: true,
		},
		"empty signing key": {
			instanceInfo: mustMarshal(attest.GCEInstanceInfo{}, require.New(t)),
			getClient:    prepareFakeClient(&computepb.ShieldedInstanceIdentity{}, nil, nil),
			wantErr:      true,
		},
		"new client error": {
			instanceInfo: mustMarshal(attest.GCEInstanceInfo{}, require.New(t)),
			getClient: prepareFakeClient(&computepb.ShieldedInstanceIdentity{
				SigningKey: &computepb.ShieldedInstanceIdentityEntry{
					EkPub: proto.String(testPubK),
				},
			}, errors.New("error"), nil),
			wantErr: true,
		},
		"GetShieldedInstanceIdentity error": {
			instanceInfo: mustMarshal(attest.GCEInstanceInfo{}, require.New(t)),
			getClient: prepareFakeClient(&computepb.ShieldedInstanceIdentity{
				SigningKey: &computepb.ShieldedInstanceIdentityEntry{
					EkPub: proto.String(testPubK),
				},
			}, nil, errors.New("error")),
			wantErr: true,
		},
		"Decode error": {
			instanceInfo: mustMarshal(attest.GCEInstanceInfo{}, require.New(t)),
			getClient: prepareFakeClient(&computepb.ShieldedInstanceIdentity{
				SigningKey: &computepb.ShieldedInstanceIdentityEntry{
					EkPub: proto.String("Not a public key"),
				},
			}, nil, nil),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			out, err := trustedKeyFromGCEAPI(tc.getClient)(nil, tc.instanceInfo)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				_, ok := out.(*rsa.PublicKey)
				assert.True(ok)
			}
		})
	}
}

func mustMarshal(in any, require *require.Assertions) []byte {
	out, err := json.Marshal(in)
	require.NoError(err)
	return out
}

type fakeInstanceClient struct {
	getIdentErr error
	ident       *computepb.ShieldedInstanceIdentity
}

func prepareFakeClient(ident *computepb.ShieldedInstanceIdentity, newErr, getIdentErr error) func(ctx context.Context, opts ...option.ClientOption) (gcpRestClient, error) {
	return func(ctx context.Context, opts ...option.ClientOption) (gcpRestClient, error) {
		return &fakeInstanceClient{
			getIdentErr: getIdentErr,
			ident:       ident,
		}, newErr
	}
}

func (c *fakeInstanceClient) Close() error {
	return nil
}

func (c *fakeInstanceClient) GetShieldedInstanceIdentity(ctx context.Context, req *computepb.GetShieldedInstanceIdentityInstanceRequest, opts ...gax.CallOption) (*computepb.ShieldedInstanceIdentity, error) {
	return c.ident, c.getIdentErr
}
