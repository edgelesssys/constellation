/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestWrapKeyAES(t *testing.T) {
	assert := assert.New(t)

	testKEK := []byte{0xD6, 0x8A, 0xED, 0xF5, 0xDB, 0x89, 0x95, 0x66, 0xA9, 0xFF, 0xD9, 0x31, 0x27, 0x4E, 0x30, 0x2D, 0x21, 0xA9, 0x46, 0x21, 0x16, 0x6C, 0x16, 0x17, 0xD1, 0x96, 0x5D, 0xB2, 0xE9, 0x0E, 0x96, 0xD1}
	testDEK := []byte{0xCB, 0x6E, 0x4B, 0x05, 0x92, 0x6C, 0xE7, 0x38, 0x0C, 0x46, 0x47, 0x06, 0x83, 0xDE, 0x20, 0xFB, 0x73, 0xAA, 0x87, 0xC1, 0x97, 0xE3, 0x7C, 0xE4, 0xF4, 0x0B, 0x96, 0x8D, 0xC5, 0x88, 0xB6, 0xDF}
	wantWrap := []byte{0x14, 0x48, 0xC4, 0xEA, 0x4B, 0x4B, 0xCA, 0xE4, 0x5A, 0xD4, 0xCC, 0xE3, 0xF7, 0xDD, 0xD5, 0x78, 0xA5, 0xA9, 0xEF, 0x9A, 0x93, 0x36, 0x09, 0xD6, 0x23, 0x01, 0xF5, 0x5F, 0xE1, 0x20, 0xDD, 0xFC, 0xBC, 0xF3, 0xA9, 0x67, 0x8B, 0x89, 0x54, 0x96}
	res, err := WrapAES(testDEK, testKEK)
	assert.NoError(err)
	assert.Equal(wantWrap, res)

	// Decrypt the Key
	res, err = UnwrapAES(res, testKEK)
	assert.NoError(err)
	assert.Equal(testDEK, res)

	// Target key length is enforced to be at least 128 bit
	smallKey := []byte{0x46, 0x6f, 0x72, 0x50, 0x61, 0x73, 0x69}
	_, err = WrapAES(smallKey, testKEK)
	assert.Error(err)

	// Wrapping key length is enforced to be 128 or 256 bit
	key192 := []byte{0x58, 0x40, 0xdf, 0x6e, 0x29, 0xb0, 0x2a, 0xf1, 0xab, 0x49, 0x3b, 0x70, 0x5b, 0xf1, 0x6e, 0xa1, 0xae, 0x83, 0x38, 0xf4, 0xdc, 0xc1, 0x76, 0xa8}
	_, err = WrapAES(testDEK, key192)
	assert.Error(err)

	// Make sure we can wrap large keys. For example AES-XTS uses 512 bit keys
	largeKey := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	_, err = WrapAES(largeKey, testKEK)
	assert.NoError(err)
}

func TestParsePEM(t *testing.T) {
	assert := assert.New(t)
	testKeyRSA := `-----BEGIN PUBLIC KEY-----
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

	notAKey := []byte(`-----BEGIN FOO-----
dGVzdA==
-----END FOO-----`)
	ecKey := []byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQx3ShpceYTZD5lnCEMExflsyMZRa
vCYPhiEFmekMeuHsjC2HnRPA7++9Rq4+IwqKdh6+Ok9kADkyAqtckTj6lg==
-----END PUBLIC KEY-----`)

	_, err := ParsePEMtoPublicKeyRSA(nil)
	assert.Error(err)

	_, err = ParsePEMtoPublicKeyRSA(notAKey)
	assert.Error(err)

	_, err = ParsePEMtoPublicKeyRSA(ecKey)
	assert.Error(err)

	_, err = ParsePEMtoPublicKeyRSA([]byte(testKeyRSA))
	assert.NoError(err)
}
