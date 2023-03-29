/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package config defines types and interfaces to use for configuring attestation Issuers and Validators.
package config

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/variant"
)

// AttestationConfig is the common interface for passing attestation configs.
type AttestationConfig interface {
	// GetMeasurements returns the measurements that should be used for attestation.
	GetMeasurements() measurements.M
	// GetVariant returns the variant of the attestation config. TODO: decide if this is needed.
	GetVariant() variant.Variant
}

// Certificate is a wrapper around x509.Certificate allowing custom marshaling.
type Certificate x509.Certificate

// MarshalJSON marshals the certificate to PEM.
func (c Certificate) MarshalJSON() ([]byte, error) {
	pem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
	return json.Marshal(string(pem))
}

// MarshalYAML marshals the certificate to PEM.
func (c Certificate) MarshalYAML() (any, error) {
	pemData := &bytes.Buffer{}
	if err := pem.Encode(pemData, &pem.Block{Type: "CERTIFICATE", Bytes: c.Raw}); err != nil {
		return nil, err
	}
	return pemData.String(), nil
}

// UnmarshalJSON unmarshals the certificate from PEM.
func (c *Certificate) UnmarshalJSON(data []byte) error {
	return c.unmarshal(func(val any) error {
		return json.Unmarshal(data, val)
	})
}

// UnmarshalYAML unmarshals the certificate from PEM.
func (c *Certificate) UnmarshalYAML(unmarshal func(any) error) error {
	return c.unmarshal(unmarshal)
}

func (c *Certificate) unmarshal(unmarshalFunc func(any) error) error {
	var pemData string
	if err := unmarshalFunc(&pemData); err != nil {
		return err
	}
	block, _ := pem.Decode([]byte(pemData))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}
	*c = Certificate(*cert)
	return nil
}

func mustParsePEM(data string) Certificate {
	jsonData := fmt.Sprintf("\"%s\"", data)
	var cert Certificate
	err := json.Unmarshal([]byte(jsonData), &cert)
	if err != nil {
		panic(err)
	}
	return cert
}
