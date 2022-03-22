//go:build !aws
// +build !aws

package aws

import (
	"errors"
	"time"
)

var errInvalidBuild = errors.New("not built with \"-tags aws\"")

func NsmGetAttestationDoc(userData []byte, nonce []byte) ([]byte, error) {
	return nil, errInvalidBuild
}

func NaAdGetVerifiedPayloadAsJson(adBlob []byte, rootCertDer []byte, ts time.Time) (string, error) {
	return "", errInvalidBuild
}
