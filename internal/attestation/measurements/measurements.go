/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measurements

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"gopkg.in/yaml.v2"
)

// Measurements are Platform Configuration Register (PCR) values.
type Measurements map[uint32][]byte

// Zero returns a PCR value with all bits set to zero.
func Zero() []byte {
	return AllBytes(0x00)
}

// One returns a PCR value with all bits set to one.
func One() []byte {
	return AllBytes(0xFF)
}

// AllBytes returns a PCR value with all bytes set to b.
func AllBytes(b byte) []byte {
	return bytes.Repeat([]byte{b}, 32)
}

// DefaultsFor provides the default measurements for given cloud provider.
func DefaultsFor(provider cloudprovider.Provider) Measurements {
	switch provider {
	case cloudprovider.AWS:
		return awsPCRs
	case cloudprovider.Azure:
		return azurePCRs
	case cloudprovider.GCP:
		return gcpPCRs
	case cloudprovider.QEMU:
		return qemuPCRs
	default:
		return nil
	}
}

var (
	// gcpPCRs are the PCR values for a GCP Constellation node that are initially set in a generated config file.
	gcpPCRs = Measurements{
		0:                              {0x0F, 0x35, 0xC2, 0x14, 0x60, 0x8D, 0x93, 0xC7, 0xA6, 0xE6, 0x8A, 0xE7, 0x35, 0x9B, 0x4A, 0x8B, 0xE5, 0xA0, 0xE9, 0x9E, 0xEA, 0x91, 0x07, 0xEC, 0xE4, 0x27, 0xC4, 0xDE, 0xA4, 0xE4, 0x39, 0xCF},
		11:                             Zero(),
		12:                             Zero(),
		13:                             Zero(),
		uint32(vtpm.PCRIndexClusterID): Zero(),
	}

	// azurePCRs are the PCR values for an Azure Constellation node that are initially set in a generated config file.
	azurePCRs = Measurements{
		11:                             Zero(),
		12:                             Zero(),
		13:                             Zero(),
		uint32(vtpm.PCRIndexClusterID): Zero(),
	}

	// awsPCRs are the PCR values for an AWS Nitro Constellation node that are initially set in a generated config file.
	awsPCRs = Measurements{
		11:                             Zero(),
		12:                             Zero(),
		13:                             Zero(),
		uint32(vtpm.PCRIndexClusterID): Zero(),
	}

	qemuPCRs = Measurements{
		11:                             Zero(),
		12:                             Zero(),
		13:                             Zero(),
		uint32(vtpm.PCRIndexClusterID): Zero(),
	}
)

// FetchAndVerify fetches measurement and signature files via provided URLs,
// using client for download. The publicKey is used to verify the measurements.
// The hash of the fetched measurements is returned.
func (m *Measurements) FetchAndVerify(ctx context.Context, client *http.Client, measurementsURL *url.URL, signatureURL *url.URL, publicKey []byte) (string, error) {
	measurements, err := getFromURL(ctx, client, measurementsURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch measurements: %w", err)
	}
	signature, err := getFromURL(ctx, client, signatureURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch signature: %w", err)
	}
	if err := sigstore.VerifySignature(measurements, signature, publicKey); err != nil {
		return "", err
	}
	if err := yaml.NewDecoder(bytes.NewReader(measurements)).Decode(&m); err != nil {
		return "", err
	}

	shaHash := sha256.Sum256(measurements)

	return hex.EncodeToString(shaHash[:]), nil
}

// CopyFrom copies over all values from other. Overwriting existing values,
// but keeping not specified values untouched.
func (m Measurements) CopyFrom(other Measurements) {
	for idx := range other {
		m[idx] = other[idx]
	}
}

// EqualTo tests whether the provided other Measurements are equal to these
// measurements.
func (m Measurements) EqualTo(other Measurements) bool {
	if len(m) != len(other) {
		return false
	}
	for k, v := range m {
		if !bytes.Equal(v, other[k]) {
			return false
		}
	}
	return true
}

// MarshalYAML overwrites the default behaviour of writing out []byte not as
// single bytes, but as a single base64 encoded string.
func (m Measurements) MarshalYAML() (any, error) {
	base64Map := make(map[uint32]string)

	for key, value := range m {
		base64Map[key] = base64.StdEncoding.EncodeToString(value[:])
	}

	return base64Map, nil
}

// UnmarshalYAML overwrites the default behaviour of reading []byte not as
// single bytes, but as a single base64 encoded string.
func (m *Measurements) UnmarshalYAML(unmarshal func(any) error) error {
	base64Map := make(map[uint32]string)
	err := unmarshal(base64Map)
	if err != nil {
		return err
	}

	*m = make(Measurements)
	for key, value := range base64Map {
		measurement, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return err
		}
		(*m)[key] = measurement
	}
	return nil
}

func getFromURL(ctx context.Context, client *http.Client, sourceURL *url.URL) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL.String(), http.NoBody)
	if err != nil {
		return []byte{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("http status code: %d", resp.StatusCode)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return content, nil
}

// Copy creates a new instance of measurements m with the same values.
func Copy(m Measurements) Measurements {
	res := make(Measurements)
	res.CopyFrom(m)
	return res
}
