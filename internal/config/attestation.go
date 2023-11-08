/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
)

// arkPEM is the PEM encoded AMD root key. Received from the AMD Key Distribution System API (KDS).
const arkPEM = `-----BEGIN CERTIFICATE-----\nMIIGYzCCBBKgAwIBAgIDAQAAMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAIC\nBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBMHsxFDAS\nBgNVBAsMC0VuZ2luZWVyaW5nMQswCQYDVQQGEwJVUzEUMBIGA1UEBwwLU2FudGEg\nQ2xhcmExCzAJBgNVBAgMAkNBMR8wHQYDVQQKDBZBZHZhbmNlZCBNaWNybyBEZXZp\nY2VzMRIwEAYDVQQDDAlBUkstTWlsYW4wHhcNMjAxMDIyMTcyMzA1WhcNNDUxMDIy\nMTcyMzA1WjB7MRQwEgYDVQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDAS\nBgNVBAcMC1NhbnRhIENsYXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5j\nZWQgTWljcm8gRGV2aWNlczESMBAGA1UEAwwJQVJLLU1pbGFuMIICIjANBgkqhkiG\n9w0BAQEFAAOCAg8AMIICCgKCAgEA0Ld52RJOdeiJlqK2JdsVmD7FktuotWwX1fNg\nW41XY9Xz1HEhSUmhLz9Cu9DHRlvgJSNxbeYYsnJfvyjx1MfU0V5tkKiU1EesNFta\n1kTA0szNisdYc9isqk7mXT5+KfGRbfc4V/9zRIcE8jlHN61S1ju8X93+6dxDUrG2\nSzxqJ4BhqyYmUDruPXJSX4vUc01P7j98MpqOS95rORdGHeI52Naz5m2B+O+vjsC0\n60d37jY9LFeuOP4Meri8qgfi2S5kKqg/aF6aPtuAZQVR7u3KFYXP59XmJgtcog05\ngmI0T/OitLhuzVvpZcLph0odh/1IPXqx3+MnjD97A7fXpqGd/y8KxX7jksTEzAOg\nbKAeam3lm+3yKIcTYMlsRMXPcjNbIvmsBykD//xSniusuHBkgnlENEWx1UcbQQrs\n+gVDkuVPhsnzIRNgYvM48Y+7LGiJYnrmE8xcrexekBxrva2V9TJQqnN3Q53kt5vi\nQi3+gCfmkwC0F0tirIZbLkXPrPwzZ0M9eNxhIySb2npJfgnqz55I0u33wh4r0ZNQ\neTGfw03MBUtyuzGesGkcw+loqMaq1qR4tjGbPYxCvpCq7+OgpCCoMNit2uLo9M18\nfHz10lOMT8nWAUvRZFzteXCm+7PHdYPlmQwUw3LvenJ/ILXoQPHfbkH0CyPfhl1j\nWhJFZasCAwEAAaN+MHwwDgYDVR0PAQH/BAQDAgEGMB0GA1UdDgQWBBSFrBrRQ/fI\nrFXUxR1BSKvVeErUUzAPBgNVHRMBAf8EBTADAQH/MDoGA1UdHwQzMDEwL6AtoCuG\nKWh0dHBzOi8va2RzaW50Zi5hbWQuY29tL3ZjZWsvdjEvTWlsYW4vY3JsMEYGCSqG\nSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAICBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZI\nAWUDBAICBQCiAwIBMKMDAgEBA4ICAQC6m0kDp6zv4Ojfgy+zleehsx6ol0ocgVel\nETobpx+EuCsqVFRPK1jZ1sp/lyd9+0fQ0r66n7kagRk4Ca39g66WGTJMeJdqYriw\nSTjjDCKVPSesWXYPVAyDhmP5n2v+BYipZWhpvqpaiO+EGK5IBP+578QeW/sSokrK\ndHaLAxG2LhZxj9aF73fqC7OAJZ5aPonw4RE299FVarh1Tx2eT3wSgkDgutCTB1Yq\nzT5DuwvAe+co2CIVIzMDamYuSFjPN0BCgojl7V+bTou7dMsqIu/TW/rPCX9/EUcp\nKGKqPQ3P+N9r1hjEFY1plBg93t53OOo49GNI+V1zvXPLI6xIFVsh+mto2RtgEX/e\npmMKTNN6psW88qg7c1hTWtN6MbRuQ0vm+O+/2tKBF2h8THb94OvvHHoFDpbCELlq\nHnIYhxy0YKXGyaW1NjfULxrrmxVW4wcn5E8GddmvNa6yYm8scJagEi13mhGu4Jqh\n3QU3sf8iUSUr09xQDwHtOQUVIqx4maBZPBtSMf+qUDtjXSSq8lfWcd8bLr9mdsUn\nJZJ0+tuPMKmBnSH860llKk+VpVQsgqbzDIvOLvD6W1Umq25boxCYJ+TuBoa4s+HH\nCViAvgT9kf/rBq1d+ivj6skkHxuzcxbk1xv6ZGxrteJxVH7KlX7YRdZ6eARKwLe4\nAFZEAwoKCQ==\n-----END CERTIFICATE-----\n`

// AttestationCfg is the common interface for passing attestation configs.
type AttestationCfg interface {
	// GetMeasurements returns the measurements that should be used for attestation.
	GetMeasurements() measurements.M
	// SetMeasurements updates a config's measurements using the given measurements.
	SetMeasurements(m measurements.M)
	// GetVariant returns the variant of the attestation config.
	GetVariant() variant.Variant
	// EqualTo returns true if the config is equal to the given config.
	// If the variant differs, an error must be returned.
	EqualTo(AttestationCfg) (bool, error)
}

// UnmarshalAttestationConfig unmarshals the config file into the correct type.
func UnmarshalAttestationConfig(data []byte, attestVariant variant.Variant) (AttestationCfg, error) {
	switch attestVariant {
	case variant.AWSNitroTPM{}:
		return unmarshalTypedConfig[*AWSNitroTPM](data)
	case variant.AWSSEVSNP{}:
		return unmarshalTypedConfig[*AWSSEVSNP](data)
	case variant.AzureSEVSNP{}:
		return unmarshalTypedConfig[*AzureSEVSNP](data)
	case variant.AzureTrustedLaunch{}:
		return unmarshalTypedConfig[*AzureTrustedLaunch](data)
	case variant.GCPSEVES{}:
		return unmarshalTypedConfig[*GCPSEVES](data)
	case variant.QEMUVTPM{}:
		return unmarshalTypedConfig[*QEMUVTPM](data)
	case variant.QEMUTDX{}:
		return unmarshalTypedConfig[*QEMUTDX](data)
	case variant.Dummy{}:
		return unmarshalTypedConfig[*DummyCfg](data)
	default:
		return nil, fmt.Errorf("unknown variant: %s", attestVariant)
	}
}

func unmarshalTypedConfig[T AttestationCfg](data []byte) (AttestationCfg, error) {
	var cfg T
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Certificate is a wrapper around x509.Certificate allowing custom marshaling.
type Certificate x509.Certificate

// Equal returns true if the embedded Raw values are equal.
func (c Certificate) Equal(other Certificate) bool {
	return bytes.Equal(c.Raw, other.Raw)
}

// MarshalJSON marshals the certificate to PEM.
func (c Certificate) MarshalJSON() ([]byte, error) {
	if len(c.Raw) == 0 {
		return json.Marshal(new(string))
	}
	pem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
	return json.Marshal(string(pem))
}

// MarshalYAML marshals the certificate to PEM.
func (c Certificate) MarshalYAML() (any, error) {
	if len(c.Raw) == 0 {
		return "", nil
	}
	pem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
	return string(pem), nil
}

// UnmarshalJSON unmarshals the certificate from PEM.
func (c *Certificate) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
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
	if pemData == "" {
		return nil
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
	if err := json.Unmarshal([]byte(jsonData), &cert); err != nil {
		panic(err)
	}
	return cert
}

// DummyCfg is a placeholder for unknown attestation configs.
type DummyCfg struct {
	// description: |
	//   The measurements that should be used for attestation.
	Measurements measurements.M `json:"measurements,omitempty"`
}

// GetMeasurements returns the configs measurements.
func (c DummyCfg) GetMeasurements() measurements.M {
	return c.Measurements
}

// GetVariant returns a dummy variant.
func (DummyCfg) GetVariant() variant.Variant {
	return variant.Dummy{}
}

// SetMeasurements sets the configs measurements.
func (c *DummyCfg) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if measurements of the configs are equal.
func (c DummyCfg) EqualTo(other AttestationCfg) (bool, error) {
	return c.Measurements.EqualTo(other.GetMeasurements()), nil
}
