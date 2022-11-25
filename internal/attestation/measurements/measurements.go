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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/google/go-tpm/tpmutil"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
)

const (
	// PCRIndexClusterID is a PCR we extend to mark the node as initialized.
	// The value used to extend is a random generated 32 Byte value.
	PCRIndexClusterID = tpmutil.Handle(15)
	// PCRIndexOwnerID is a PCR we extend to mark the node as initialized.
	// The value used to extend is derived from Constellation's master key.
	// TODO: move to stable, non-debug PCR before use.
	PCRIndexOwnerID = tpmutil.Handle(16)
)

// M are Platform Configuration Register (PCR) values that make up the Measurements.
type M map[uint32]Measurement

// WithMetadata is a struct supposed to provide CSP & image metadata next to measurements.
type WithMetadata struct {
	CSP          cloudprovider.Provider `json:"csp" yaml:"csp"`
	Image        string                 `json:"image" yaml:"image"`
	Measurements M                      `json:"measurements" yaml:"measurements"`
}

// FetchAndVerify fetches measurement and signature files via provided URLs,
// using client for download. The publicKey is used to verify the measurements.
// The hash of the fetched measurements is returned.
func (m *M) FetchAndVerify(
	ctx context.Context, client *http.Client, measurementsURL, signatureURL *url.URL,
	publicKey []byte, metadata WithMetadata,
) (string, error) {
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

	var mWithMetadata WithMetadata
	if err := json.Unmarshal(measurements, &mWithMetadata); err != nil {
		if yamlErr := yaml.Unmarshal(measurements, &mWithMetadata); yamlErr != nil {
			return "", multierr.Append(
				err,
				fmt.Errorf("trying yaml format: %w", yamlErr),
			)
		}
	}

	if mWithMetadata.CSP != metadata.CSP {
		return "", fmt.Errorf("invalid measurement metadata: CSP mismatch: expected %s, got %s", metadata.CSP, mWithMetadata.CSP)
	}
	if mWithMetadata.Image != metadata.Image {
		return "", fmt.Errorf("invalid measurement metadata: image mismatch: expected %s, got %s", metadata.Image, mWithMetadata.Image)
	}

	*m = mWithMetadata.Measurements

	shaHash := sha256.Sum256(measurements)

	return hex.EncodeToString(shaHash[:]), nil
}

// CopyFrom copies over all values from other. Overwriting existing values,
// but keeping not specified values untouched.
func (m *M) CopyFrom(other M) {
	for idx := range other {
		(*m)[idx] = other[idx]
	}
}

// EqualTo tests whether the provided other Measurements are equal to these
// measurements.
func (m *M) EqualTo(other M) bool {
	if len(*m) != len(other) {
		return false
	}
	for k, v := range *m {
		otherExpected := other[k].Expected
		if !bytes.Equal(v.Expected[:], otherExpected[:]) {
			return false
		}
		if v.WarnOnly != other[k].WarnOnly {
			return false
		}
	}
	return true
}

// GetEnforced returns a list of all enforced Measurements,
// i.e. all Measurements that are not marked as WarnOnly.
func (m *M) GetEnforced() []uint32 {
	var enforced []uint32
	for idx, measurement := range *m {
		if !measurement.WarnOnly {
			enforced = append(enforced, idx)
		}
	}
	return enforced
}

// SetEnforced sets the WarnOnly flag to true for all Measurements
// that are NOT included in the provided list of enforced measurements.
func (m *M) SetEnforced(enforced []uint32) error {
	newM := make(M)

	// set all measurements to warn only
	for idx, measurement := range *m {
		newM[idx] = Measurement{
			Expected: measurement.Expected,
			WarnOnly: true,
		}
	}

	// set enforced measurements from list
	for _, idx := range enforced {
		measurement, ok := newM[idx]
		if !ok {
			return fmt.Errorf("measurement %d not in list, but set to enforced", idx)
		}
		measurement.WarnOnly = false
		newM[idx] = measurement
	}

	*m = newM
	return nil
}

// Measurement wraps expected PCR value and whether it is enforced.
type Measurement struct {
	// Expected measurement value.
	Expected [32]byte `json:"expected" yaml:"expected"`
	// WarnOnly if set to true, a mismatching measurement will only result in a warning.
	WarnOnly bool `json:"warnOnly" yaml:"warnOnly"`
}

// UnmarshalJSON reads a Measurement either as json object,
// or as a simple hex or base64 encoded string.
func (m *Measurement) UnmarshalJSON(b []byte) error {
	var eM encodedMeasurement
	if err := json.Unmarshal(b, &eM); err != nil {
		// Unmarshalling failed, Measurement might be a simple string instead of Measurement struct.
		// These values will always be enforced.
		if legacyErr := json.Unmarshal(b, &eM.Expected); legacyErr != nil {
			return multierr.Append(
				err,
				fmt.Errorf("trying legacy format: %w", legacyErr),
			)
		}
	}

	if err := m.unmarshal(eM); err != nil {
		return fmt.Errorf("unmarshalling json: %w", err)
	}
	return nil
}

// MarshalJSON writes out a Measurement with Expected encoded as a hex string.
func (m Measurement) MarshalJSON() ([]byte, error) {
	return json.Marshal(encodedMeasurement{
		Expected: hex.EncodeToString(m.Expected[:]),
		WarnOnly: m.WarnOnly,
	})
}

// UnmarshalYAML reads a Measurement either as yaml object,
// or as a simple hex or base64 encoded string.
func (m *Measurement) UnmarshalYAML(unmarshal func(any) error) error {
	var eM encodedMeasurement
	if err := unmarshal(&eM); err != nil {
		// Unmarshalling failed, Measurement might be a simple string instead of Measurement struct.
		// These values will always be enforced.
		if legacyErr := unmarshal(&eM.Expected); legacyErr != nil {
			return multierr.Append(
				err,
				fmt.Errorf("trying legacy format: %w", legacyErr),
			)
		}
	}

	if err := m.unmarshal(eM); err != nil {
		return fmt.Errorf("unmarshalling yaml: %w", err)
	}
	return nil
}

// MarshalYAML writes out a Measurement with Expected encoded as a hex string.
func (m Measurement) MarshalYAML() (any, error) {
	return encodedMeasurement{
		Expected: hex.EncodeToString(m.Expected[:]),
		WarnOnly: m.WarnOnly,
	}, nil
}

// unmarshal parses a hex or base64 encoded Measurement.
func (m *Measurement) unmarshal(eM encodedMeasurement) error {
	expected, err := hex.DecodeString(eM.Expected)
	if err != nil {
		// expected value might be in base64 legacy format
		// TODO: Remove with v2.4.0
		hexErr := err
		expected, err = base64.StdEncoding.DecodeString(eM.Expected)
		if err != nil {
			return multierr.Append(
				fmt.Errorf("invalid measurement: not a hex string %w", hexErr),
				fmt.Errorf("not a base64 string: %w", err),
			)
		}
	}

	if len(expected) != 32 {
		return fmt.Errorf("invalid measurement: invalid length: %d", len(expected))
	}

	m.Expected = *(*[32]byte)(expected)
	m.WarnOnly = eM.WarnOnly

	return nil
}

// WithAllBytes returns a measurement value where all 32 bytes are set to b.
func WithAllBytes(b byte, warnOnly bool) Measurement {
	return Measurement{
		Expected: *(*[32]byte)(bytes.Repeat([]byte{b}, 32)),
		WarnOnly: warnOnly,
	}
}

// DefaultsFor provides the default measurements for given cloud provider.
func DefaultsFor(provider cloudprovider.Provider) M {
	switch provider {
	case cloudprovider.AWS:
		return M{
			4:                         PlaceHolderMeasurement(),
			8:                         WithAllBytes(0x00, false),
			9:                         PlaceHolderMeasurement(),
			11:                        WithAllBytes(0x00, false),
			12:                        PlaceHolderMeasurement(),
			13:                        WithAllBytes(0x00, false),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, false),
		}
	case cloudprovider.Azure:
		return M{
			4:                         PlaceHolderMeasurement(),
			8:                         WithAllBytes(0x00, false),
			9:                         PlaceHolderMeasurement(),
			11:                        WithAllBytes(0x00, false),
			12:                        PlaceHolderMeasurement(),
			13:                        WithAllBytes(0x00, false),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, false),
		}
	case cloudprovider.GCP:
		return M{
			0: {
				Expected: [32]byte{0x0F, 0x35, 0xC2, 0x14, 0x60, 0x8D, 0x93, 0xC7, 0xA6, 0xE6, 0x8A, 0xE7, 0x35, 0x9B, 0x4A, 0x8B, 0xE5, 0xA0, 0xE9, 0x9E, 0xEA, 0x91, 0x07, 0xEC, 0xE4, 0x27, 0xC4, 0xDE, 0xA4, 0xE4, 0x39, 0xCF},
				WarnOnly: false,
			},
			4:                         PlaceHolderMeasurement(),
			8:                         WithAllBytes(0x00, false),
			9:                         PlaceHolderMeasurement(),
			11:                        WithAllBytes(0x00, false),
			12:                        PlaceHolderMeasurement(),
			13:                        WithAllBytes(0x00, false),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, false),
		}
	case cloudprovider.QEMU:
		return M{
			4:                         PlaceHolderMeasurement(),
			8:                         WithAllBytes(0x00, false),
			9:                         PlaceHolderMeasurement(),
			11:                        WithAllBytes(0x00, false),
			12:                        PlaceHolderMeasurement(),
			13:                        WithAllBytes(0x00, false),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, false),
		}
	default:
		return nil
	}
}

// PlaceHolderMeasurement returns a measurement with placeholder values for Expected.
func PlaceHolderMeasurement() Measurement {
	return Measurement{
		Expected: *(*[32]byte)(bytes.Repeat([]byte{0x12, 0x34}, 16)),
		WarnOnly: false,
	}
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

type encodedMeasurement struct {
	Expected string `json:"expected" yaml:"expected"`
	WarnOnly bool   `json:"warnOnly" yaml:"warnOnly"`
}
