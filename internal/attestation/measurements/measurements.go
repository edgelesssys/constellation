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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/google/go-tpm/tpmutil"
	"gopkg.in/yaml.v2"
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

// Measurement wraps expected PCR value and whether it is enforced.
type Measurement struct {
	// Expected measurement value.
	Expected [32]byte `yaml:"expected"`
	// WarnOnly if set to true, a mismatching measurement will only result in a warning.
	WarnOnly bool `yaml:"warnOnly"`
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
			8:                         WithAllBytes(0x00, false),
			11:                        WithAllBytes(0x00, false),
			13:                        WithAllBytes(0x00, false),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, false),
		}
	case cloudprovider.Azure:
		return M{
			8:                         WithAllBytes(0x00, false),
			11:                        WithAllBytes(0x00, false),
			13:                        WithAllBytes(0x00, false),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, false),
		}
	case cloudprovider.GCP:
		return M{
			0: {
				Expected: [32]byte{0x0F, 0x35, 0xC2, 0x14, 0x60, 0x8D, 0x93, 0xC7, 0xA6, 0xE6, 0x8A, 0xE7, 0x35, 0x9B, 0x4A, 0x8B, 0xE5, 0xA0, 0xE9, 0x9E, 0xEA, 0x91, 0x07, 0xEC, 0xE4, 0x27, 0xC4, 0xDE, 0xA4, 0xE4, 0x39, 0xCF},
				WarnOnly: false,
			},
			8:                         WithAllBytes(0x00, false),
			11:                        WithAllBytes(0x00, false),
			13:                        WithAllBytes(0x00, false),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, false),
		}
	case cloudprovider.QEMU:
		return M{
			8:                         WithAllBytes(0x00, false),
			11:                        WithAllBytes(0x00, false),
			13:                        WithAllBytes(0x00, false),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, false),
		}
	default:
		return nil
	}
}

// FetchAndVerify fetches measurement and signature files via provided URLs,
// using client for download. The publicKey is used to verify the measurements.
// The hash of the fetched measurements is returned.
func (m *M) FetchAndVerify(ctx context.Context, client *http.Client, measurementsURL *url.URL, signatureURL *url.URL, publicKey []byte) (string, error) {
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
func (m M) CopyFrom(other M) {
	for idx := range other {
		m[idx] = other[idx]
	}
}

// EqualTo tests whether the provided other Measurements are equal to these
// measurements.
func (m M) EqualTo(other M) bool {
	if len(m) != len(other) {
		return false
	}
	for k, v := range m {
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

// MarshalYAML overwrites the default behaviour of writing out [32]byte not as
// single bytes, but as a single base64 encoded string.
func (m M) MarshalYAML() (any, error) {
	base64Map := make(map[uint32]b64Measurement)

	for key, value := range m {
		base64Map[key] = b64Measurement{
			Expected: base64.StdEncoding.EncodeToString(value.Expected[:]),
			WarnOnly: value.WarnOnly,
		}
	}

	return base64Map, nil
}

// UnmarshalYAML overwrites the default behaviour of reading [32]byte not as
// single bytes, but as a single base64 encoded string.
func (m *M) UnmarshalYAML(unmarshal func(any) error) error {
	base64Map := make(map[uint32]b64Measurement)
	err := unmarshal(base64Map)
	if err != nil {
		// If the unmarshal fails, try to unmarshal as a map of uint32 to string.
		typeErr := &yaml.TypeError{}
		if errors.As(err, &typeErr) {
			base64SimpleMap := make(map[uint32]string)
			err = unmarshal(base64SimpleMap)
			if err != nil {
				return err
			}

			for key, value := range base64SimpleMap {
				base64Map[key] = b64Measurement{
					Expected: value,
					WarnOnly: false,
				}
			}
		} else {
			return err
		}
	}

	*m = make(M)
	for key, value := range base64Map {
		measurement, err := base64.StdEncoding.DecodeString(value.Expected)
		if err != nil {
			return err
		}

		if len(measurement) != 32 {
			return fmt.Errorf("invalid measurement at key %d: invalid length: %d", key, len(measurement))
		}

		(*m)[key] = Measurement{
			Expected: *(*[32]byte)(measurement),
			WarnOnly: value.WarnOnly,
		}
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

type b64Measurement struct {
	Expected string `yaml:"expected"`
	WarnOnly bool   `yaml:"warnOnly"`
}
