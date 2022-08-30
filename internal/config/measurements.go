package config

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/sigstore"
	"gopkg.in/yaml.v2"
)

type Measurements map[uint32][]byte

var (
	// gcpPCRs are the PCR values for a GCP Constellation node that are initially set in a generated config file.
	gcpPCRs = Measurements{
		0:                              {0x0F, 0x35, 0xC2, 0x14, 0x60, 0x8D, 0x93, 0xC7, 0xA6, 0xE6, 0x8A, 0xE7, 0x35, 0x9B, 0x4A, 0x8B, 0xE5, 0xA0, 0xE9, 0x9E, 0xEA, 0x91, 0x07, 0xEC, 0xE4, 0x27, 0xC4, 0xDE, 0xA4, 0xE4, 0x39, 0xCF},
		uint32(vtpm.PCRIndexOwnerID):   {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		uint32(vtpm.PCRIndexClusterID): {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}

	// azurePCRs are the PCR values for an Azure Constellation node that are initially set in a generated config file.
	azurePCRs = Measurements{
		uint32(vtpm.PCRIndexOwnerID):   {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		uint32(vtpm.PCRIndexClusterID): {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}

	qemuPCRs = Measurements{
		uint32(vtpm.PCRIndexOwnerID):   {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		uint32(vtpm.PCRIndexClusterID): {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
)

// FetchAndVerify fetches measurement and signature files via provided URLs,
// using client for download. The publicKey is used to verify the measurements.
func (m *Measurements) FetchAndVerify(ctx context.Context, client *http.Client, measurementsURL *url.URL, signatureURL *url.URL, publicKey []byte) error {
	measurements, err := getFromURL(ctx, client, measurementsURL)
	if err != nil {
		return err
	}
	signature, err := getFromURL(ctx, client, signatureURL)
	if err != nil {
		return err
	}
	if err := sigstore.VerifySignature(measurements, signature, publicKey); err != nil {
		return err
	}
	if err := yaml.NewDecoder(bytes.NewReader(measurements)).Decode(&m); err != nil {
		return err
	}
	return nil
}

// CopyFrom copies over all values from other. Overwriting existing values,
// but keeping not specified values untouched.
func (m Measurements) CopyFrom(other Measurements) {
	for idx := range other {
		m[idx] = other[idx]
	}
}

// MarshalYAML overwrites the default behaviour of writing out []byte not as
// single bytes, but as a single base64 encoded string.
func (m Measurements) MarshalYAML() (interface{}, error) {
	base64Map := make(map[uint32]string)

	for key, value := range m {
		base64Map[key] = base64.StdEncoding.EncodeToString(value[:])
	}

	return base64Map, nil
}

// UnmarshalYAML overwrites the default behaviour of reading []byte not as
// single bytes, but as a single base64 encoded string.
func (m *Measurements) UnmarshalYAML(unmarshal func(interface{}) error) error {
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
