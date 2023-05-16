/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"sort"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/variant"
)

const (
	AttestationTypeAWSNitroTPM        AttestationType = "aws-nitro-tpm"        // AttestationTypeAWSNitroTPM
	AttestationTypeAzureSEVSNP        AttestationType = "azure-sev-snp"        // AttestationTypeAzureSEVSNP
	AttestationTypeAzureTrustedLaunch AttestationType = "azure-trusted-launch" // AttestationTypeAzureTrustedLaunch
	AttestationTypeGCPSEVES           AttestationType = "gcp-sev-es"           // AttestationTypeGCPSEVES
	AttestationTypeQEMUVTPM           AttestationType = "qemu-vtpm"            // AttestationTypeQEMUVTPM
	AttestationTypeDefault            AttestationType = "default"              // AttestationTypeDefault is the default attestation type for the cloud provider
)

var providerAttestationMapping map[cloudprovider.Provider][]AttestationType = map[cloudprovider.Provider][]AttestationType{
	cloudprovider.AWS:       {AttestationTypeAWSNitroTPM},
	cloudprovider.Azure:     {AttestationTypeAzureSEVSNP, AttestationTypeAzureTrustedLaunch},
	cloudprovider.GCP:       {AttestationTypeGCPSEVES},
	cloudprovider.QEMU:      {AttestationTypeQEMUVTPM},
	cloudprovider.OpenStack: {AttestationTypeQEMUVTPM},
}

// GetDefaultAttestationType returns the default attestation type for the given provider.
func GetDefaultAttestationType(provider cloudprovider.Provider) AttestationType {
	res, ok := providerAttestationMapping[provider]
	if ok {
		return res[0]
	}
	return AttestationType("")
}

// GetAvailableAttestationTypes returns the available attestation types.
func GetAvailableAttestationTypes() []AttestationType {
	var res []AttestationType

	// assumes that cloudprovider.Provider is a uint32 to sort the providers and get a consistent order
	var keys []cloudprovider.Provider
	for k := range providerAttestationMapping {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return uint(keys[i]) < uint(keys[j])
	})

	for _, k := range keys {
		res = append(res, providerAttestationMapping[k]...)
	}
	return res
}

// AttestationType is the type of attestation.
type AttestationType (string)

// ValidProvider returns true if the attestation type is valid for the given provider.
func (a AttestationType) ValidProvider(provider cloudprovider.Provider) bool {
	validTypes, ok := providerAttestationMapping[provider]
	if ok {
		for _, aType := range validTypes {
			if a == aType {
				return true
			}
		}
	}
	return false
}

// AttestationCfg is the common interface for passing attestation configs.
type AttestationCfg interface {
	// GetMeasurements returns the measurements that should be used for attestation.
	GetMeasurements() measurements.M
	// SetMeasurements updates a config's measurements using the given measurements.
	SetMeasurements(m measurements.M)
	// GetVariant returns the variant of the attestation config.
	GetVariant() variant.Variant
	// NewerThan returns true if the config is equal to the given config.
	EqualTo(AttestationCfg) (bool, error)
}

// UnmarshalAttestationConfig unmarshals the config file into the correct type.
func UnmarshalAttestationConfig(data []byte, attestVariant variant.Variant) (AttestationCfg, error) {
	switch attestVariant {
	case variant.AWSNitroTPM{}:
		return unmarshalTypedConfig[*AWSNitroTPM](data)
	case variant.AzureSEVSNP{}:
		return unmarshalTypedConfig[*AzureSEVSNP](data)
	case variant.AzureTrustedLaunch{}:
		return unmarshalTypedConfig[*AzureTrustedLaunch](data)
	case variant.GCPSEVES{}:
		return unmarshalTypedConfig[*GCPSEVES](data)
	case variant.QEMUVTPM{}:
		return unmarshalTypedConfig[*QEMUVTPM](data)
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

// MarshalJSON marshals the certificate to PEM.
func (c Certificate) MarshalJSON() ([]byte, error) {
	pem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
	return json.Marshal(string(pem))
}

// MarshalYAML marshals the certificate to PEM.
func (c Certificate) MarshalYAML() (any, error) {
	pem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
	return string(pem), nil
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
