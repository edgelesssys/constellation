/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

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
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"gopkg.in/yaml.v2"
)

type Measurements map[uint32][]byte

var (
	zero = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	// gcpPCRs are the PCR values for a GCP Constellation node that are initially set in a generated config file.
	gcpPCRs = Measurements{
		0:                              {0x0F, 0x35, 0xC2, 0x14, 0x60, 0x8D, 0x93, 0xC7, 0xA6, 0xE6, 0x8A, 0xE7, 0x35, 0x9B, 0x4A, 0x8B, 0xE5, 0xA0, 0xE9, 0x9E, 0xEA, 0x91, 0x07, 0xEC, 0xE4, 0x27, 0xC4, 0xDE, 0xA4, 0xE4, 0x39, 0xCF},
		11:                             zero,
		12:                             zero,
		13:                             zero,
		uint32(vtpm.PCRIndexClusterID): zero,
	}

	// azurePCRs are the PCR values for an Azure Constellation node that are initially set in a generated config file.
	azurePCRs = Measurements{
		11:                             zero,
		12:                             zero,
		13:                             zero,
		uint32(vtpm.PCRIndexClusterID): zero,
	}

	// awsPCRs are the PCR values for an AWS Nitro Constellation node that are initially set in a generated config file.
	awsPCRs = Measurements{
		11:                             zero,
		12:                             zero,
		13:                             zero,
		uint32(vtpm.PCRIndexClusterID): zero,
	}

	qemuPCRs = Measurements{
		4:                              {0xfe, 0xf9, 0xd6, 0x0b, 0x72, 0x2b, 0xa0, 0xd3, 0x8d, 0xa6, 0x5d, 0x65, 0x10, 0xc0, 0x59, 0x4e, 0xd2, 0x1e, 0x46, 0x74, 0x94, 0x6c, 0x92, 0x81, 0x75, 0x3f, 0x4d, 0xeb, 0x8d, 0x87, 0x88, 0x64},
		8:                              {0xc0, 0xfe, 0x30, 0xe1, 0x4d, 0x1f, 0xb3, 0x3f, 0xa1, 0x04, 0x4e, 0x27, 0xcf, 0x0d, 0x0d, 0x28, 0x13, 0xdf, 0xdb, 0x99, 0x76, 0xc7, 0x11, 0x55, 0xf6, 0x4c, 0xd2, 0x65, 0xb6, 0x0c, 0xcb, 0x68},
		9:                              {0xe2, 0x20, 0x12, 0x89, 0xbd, 0x92, 0x08, 0x56, 0x7b, 0x2d, 0x95, 0xdf, 0xab, 0xac, 0x27, 0x6e, 0x18, 0x01, 0x96, 0xf3, 0x57, 0x1f, 0xab, 0x22, 0x85, 0xb7, 0xd0, 0xa8, 0x43, 0xed, 0x71, 0x37},
		11:                             zero,
		12:                             zero,
		13:                             zero,
		uint32(vtpm.PCRIndexClusterID): zero,
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
