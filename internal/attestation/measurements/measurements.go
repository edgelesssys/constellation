/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# Measurements

Defines default expected measurements for the current release, as well as functions for comparing, updating and marshalling measurements.

This package should not include TPM specific code.
*/
package measurements

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/google/go-tpm/tpmutil"
	"github.com/siderolabs/talos/pkg/machinery/config/encoder"
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

	// TDXIndexClusterID is the measurement used to mark the node as initialized.
	// The value is the index of the RTMR + 1, since index 0 of the TDX measurements is reserved for MRTD.
	TDXIndexClusterID = RTMRIndexClusterID + 1
	// RTMRIndexClusterID is the RTMR we extend to mark the node as initialized.
	RTMRIndexClusterID = 2

	// PCRMeasurementLength holds the length for valid PCR measurements (SHA256).
	PCRMeasurementLength = 32
	// TDXMeasurementLength holds the length for valid TDX measurements (SHA384).
	TDXMeasurementLength = 48
)

// M are Platform Configuration Register (PCR) values that make up the Measurements.
type M map[uint32]Measurement

// WithMetadata is a struct supposed to provide CSP & image metadata next to measurements.
type WithMetadata struct {
	CSP          cloudprovider.Provider `json:"csp" yaml:"csp"`
	Image        string                 `json:"image" yaml:"image"`
	Measurements M                      `json:"measurements" yaml:"measurements"`
}

// MarshalYAML returns the YAML encoding of m.
func (m M) MarshalYAML() (any, error) {
	// cast to prevent infinite recursion
	node, err := encoder.NewEncoder(map[uint32]Measurement(m)).Marshal()
	if err != nil {
		return nil, err
	}

	// sort keys numerically
	sort.Sort(mYamlContent(node.Content))

	return node, nil
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
			return "", errors.Join(
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
		if !bytes.Equal(v.Expected, otherExpected) {
			return false
		}
		if v.ValidationOpt != other[k].ValidationOpt {
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
		if measurement.ValidationOpt == Enforce {
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
			Expected:      measurement.Expected,
			ValidationOpt: WarnOnly,
		}
	}

	// set enforced measurements from list
	for _, idx := range enforced {
		measurement, ok := newM[idx]
		if !ok {
			return fmt.Errorf("measurement %d not in list, but set to enforced", idx)
		}
		measurement.ValidationOpt = Enforce
		newM[idx] = measurement
	}

	*m = newM
	return nil
}

// UnmarshalJSON unmarshals measurements from json.
// This function enforces all measurements to be of equal length.
func (m *M) UnmarshalJSON(b []byte) error {
	newM := make(map[uint32]Measurement)
	if err := json.Unmarshal(b, &newM); err != nil {
		return err
	}

	// check if all measurements are of equal length
	if err := checkLength(newM); err != nil {
		return err
	}

	*m = newM
	return nil
}

// UnmarshalYAML unmarshals measurements from yaml.
// This function enforces all measurements to be of equal length.
func (m *M) UnmarshalYAML(unmarshal func(any) error) error {
	newM := make(map[uint32]Measurement)
	if err := unmarshal(&newM); err != nil {
		return err
	}

	// check if all measurements are of equal length
	if err := checkLength(newM); err != nil {
		return err
	}

	*m = newM
	return nil
}

// Measurement wraps expected PCR value and whether it is enforced.
type Measurement struct {
	// Expected measurement value.
	// 32 bytes for vTPM attestation, 48 for TDX.
	Expected []byte `json:"expected" yaml:"expected"`
	// ValidationOpt indicates how measurement mismatches should be handled.
	ValidationOpt MeasurementValidationOption `json:"warnOnly" yaml:"warnOnly"`
}

// MeasurementValidationOption indicates how measurement mismatches should be handled.
type MeasurementValidationOption bool

const (
	// WarnOnly will only result in a warning in case of a mismatching measurement.
	WarnOnly MeasurementValidationOption = true
	// Enforce will result in an error in case of a mismatching measurement, and operation will be aborted.
	Enforce MeasurementValidationOption = false
)

// UnmarshalJSON reads a Measurement either as json object,
// or as a simple hex or base64 encoded string.
func (m *Measurement) UnmarshalJSON(b []byte) error {
	var eM encodedMeasurement
	if err := json.Unmarshal(b, &eM); err != nil {
		// Unmarshalling failed, Measurement might be a simple string instead of Measurement struct.
		// These values will always be enforced.
		if legacyErr := json.Unmarshal(b, &eM.Expected); legacyErr != nil {
			return errors.Join(
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
		WarnOnly: m.ValidationOpt,
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
			return errors.Join(
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
		WarnOnly: m.ValidationOpt,
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
			return errors.Join(
				fmt.Errorf("invalid measurement: not a hex string %w", hexErr),
				fmt.Errorf("not a base64 string: %w", err),
			)
		}
	}

	if len(expected) != 32 && len(expected) != 48 {
		return fmt.Errorf("invalid measurement: invalid length: %d", len(expected))
	}

	m.Expected = expected
	m.ValidationOpt = eM.WarnOnly

	return nil
}

// WithAllBytes returns a measurement value where all 32 bytes are set to b.
func WithAllBytes(b byte, validationOpt MeasurementValidationOption) Measurement {
	return Measurement{
		Expected:      bytes.Repeat([]byte{b}, 32),
		ValidationOpt: validationOpt,
	}
}

// PlaceHolderMeasurement returns a measurement with placeholder values for Expected.
func PlaceHolderMeasurement() Measurement {
	return Measurement{
		Expected:      bytes.Repeat([]byte{0x12, 0x34}, 16),
		ValidationOpt: Enforce,
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

func checkLength(m map[uint32]Measurement) error {
	var length int
	for idx, measurement := range m {
		if length == 0 {
			length = len(measurement.Expected)
		} else if len(measurement.Expected) != length {
			return fmt.Errorf("inconsistent measurement length: index %d: expected %d, got %d", idx, length, len(measurement.Expected))
		}
	}
	return nil
}

type encodedMeasurement struct {
	Expected string                      `json:"expected" yaml:"expected"`
	WarnOnly MeasurementValidationOption `json:"warnOnly" yaml:"warnOnly"`
}

// mYamlContent is the Content of a yaml.Node encoding of an M. It implements sort.Interface.
// The slice is filled like {key1, value1, key2, value2, ...}.
type mYamlContent []*yaml.Node

func (c mYamlContent) Len() int {
	return len(c) / 2
}

func (c mYamlContent) Less(i, j int) bool {
	lhs, err := strconv.Atoi(c[2*i].Value)
	if err != nil {
		panic(err)
	}
	rhs, err := strconv.Atoi(c[2*j].Value)
	if err != nil {
		panic(err)
	}
	return lhs < rhs
}

func (c mYamlContent) Swap(i, j int) {
	// The slice is filled like {key1, value1, key2, value2, ...}.
	// We need to swap both key and value.
	c[2*i], c[2*j] = c[2*j], c[2*i]
	c[2*i+1], c[2*j+1] = c[2*j+1], c[2*i+1]
}
