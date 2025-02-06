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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/google/go-tpm/tpmutil"
	"github.com/siderolabs/talos/pkg/machinery/config/encoder"
	"gopkg.in/yaml.v3"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
)

//go:generate measurement-generator

const (
	// PCRIndexClusterID is a PCR we extend to mark the node as initialized.
	// The value used to extend is a random generated 32 Byte value.
	PCRIndexClusterID = tpmutil.Handle(15)
	// PCRIndexOwnerID is a PCR we extend to mark the node as initialized.
	// The value used to extend is derived from Constellation's master key.
	// TODO(daniel-weisse): move to stable, non-debug PCR before use.
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

// ImageMeasurementsV2 is a struct to hold measurements for a specific image.
// .List contains measurements for all variants of the image.
type ImageMeasurementsV2 struct {
	Version string                     `json:"version" yaml:"version"`
	Ref     string                     `json:"ref" yaml:"ref"`
	Stream  string                     `json:"stream" yaml:"stream"`
	List    []ImageMeasurementsV2Entry `json:"list" yaml:"list"`
}

// ImageMeasurementsV2Entry is a struct to hold measurements for one variant of a specific image.
type ImageMeasurementsV2Entry struct {
	CSP                cloudprovider.Provider `json:"csp" yaml:"csp"`
	AttestationVariant string                 `json:"attestationVariant" yaml:"attestationVariant"`
	Measurements       M                      `json:"measurements" yaml:"measurements"`
}

// MergeImageMeasurementsV2 combines the image measurement entries from multiple sources into a single
// ImageMeasurementsV2 object.
func MergeImageMeasurementsV2(measurements ...ImageMeasurementsV2) (ImageMeasurementsV2, error) {
	if len(measurements) == 0 {
		return ImageMeasurementsV2{}, errors.New("no measurement objects specified")
	}
	if len(measurements) == 1 {
		return measurements[0], nil
	}
	out := ImageMeasurementsV2{
		Version: measurements[0].Version,
		Ref:     measurements[0].Ref,
		Stream:  measurements[0].Stream,
		List:    []ImageMeasurementsV2Entry{},
	}
	for _, m := range measurements {
		if m.Version != out.Version {
			return ImageMeasurementsV2{}, errors.New("version mismatch")
		}
		if m.Ref != out.Ref {
			return ImageMeasurementsV2{}, errors.New("ref mismatch")
		}
		if m.Stream != out.Stream {
			return ImageMeasurementsV2{}, errors.New("stream mismatch")
		}
		out.List = append(out.List, m.List...)
	}
	sort.SliceStable(out.List, func(i, j int) bool {
		if out.List[i].CSP != out.List[j].CSP {
			return out.List[i].CSP < out.List[j].CSP
		}
		return out.List[i].AttestationVariant < out.List[j].AttestationVariant
	})
	return out, nil
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
// using client for download.
// The hash of the fetched measurements is returned.
func (m *M) FetchAndVerify(
	ctx context.Context, client *http.Client, verifier cosignVerifier,
	measurementsURL, signatureURL *url.URL,
	version versionsapi.Version, csp cloudprovider.Provider, attestationVariant variant.Variant,
) (string, error) {
	return m.fetchAndVerify(
		ctx, client, verifier,
		measurementsURL, signatureURL,
		version, csp, attestationVariant,
	)
}

// fetchAndVerify fetches measurement and signature files via provided URLs,
// using client for download. The publicKey is used to verify the measurements.
// The hash of the fetched measurements is returned.
func (m *M) fetchAndVerify(
	ctx context.Context, client *http.Client, verifier cosignVerifier,
	measurementsURL, signatureURL *url.URL,
	version versionsapi.Version, csp cloudprovider.Provider, attestationVariant variant.Variant,
) (string, error) {
	measurementsRaw, err := getFromURL(ctx, client, measurementsURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch measurements: %w", err)
	}
	signature, err := getFromURL(ctx, client, signatureURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch signature: %w", err)
	}
	if err := verifier.VerifySignature(measurementsRaw, signature); err != nil {
		return "", err
	}

	var measurements ImageMeasurementsV2
	if err := json.Unmarshal(measurementsRaw, &measurements); err != nil {
		return "", err
	}
	if err := m.fromImageMeasurementsV2(measurements, version, csp, attestationVariant); err != nil {
		return "", err
	}
	shaHash := sha256.Sum256(measurementsRaw)
	return hex.EncodeToString(shaHash[:]), nil
}

// FetchNoVerify fetches measurement via provided URLs,
// using client for download. Measurements are not verified.
func (m *M) FetchNoVerify(ctx context.Context, client *http.Client, measurementsURL *url.URL,
	version versionsapi.Version, csp cloudprovider.Provider, attestationVariant variant.Variant,
) error {
	measurementsRaw, err := getFromURL(ctx, client, measurementsURL)
	if err != nil {
		return fmt.Errorf("failed to fetch measurements from %s: %w", measurementsURL.String(), err)
	}

	var measurements ImageMeasurementsV2
	if err := json.Unmarshal(measurementsRaw, &measurements); err != nil {
		return err
	}
	return m.fromImageMeasurementsV2(measurements, version, csp, attestationVariant)
}

// CopyFrom copies over all values from other. Overwriting existing values,
// but keeping not specified values untouched.
func (m *M) CopyFrom(other M) {
	for idx := range other {
		(*m)[idx] = other[idx]
	}
}

// Copy creates a new map with the same values as the original.
func (m *M) Copy() M {
	newM := make(M, len(*m))
	for idx := range *m {
		newM[idx] = (*m)[idx]
	}
	return newM
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

// Compare compares the expected measurements to the given list of measurements.
// It returns a list of warnings for non matching measurements for WarnOnly entries,
// and a list of errors for non matching measurements for Enforce entries.
func (m M) Compare(other map[uint32][]byte) (warnings []string, errs []error) {
	// Get list of indices in expected measurements
	var mIndices []uint32
	for idx := range m {
		mIndices = append(mIndices, idx)
	}
	sort.SliceStable(mIndices, func(i, j int) bool {
		return mIndices[i] < mIndices[j]
	})

	for _, idx := range mIndices {
		if !bytes.Equal(m[idx].Expected, other[idx]) {
			msg := fmt.Sprintf("untrusted measurement value %x at index %d", other[idx], idx)
			if len(other[idx]) == 0 {
				msg = fmt.Sprintf("missing measurement value for index %d", idx)
			}

			if m[idx].ValidationOpt == Enforce {
				errs = append(errs, errors.New(msg))
			} else {
				warnings = append(warnings, fmt.Sprintf("Encountered %s", msg))
			}
		}
	}

	return warnings, errs
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

// String returns a string representation of the measurements.
func (m M) String() string {
	var returnString string
	for i, measurement := range m {
		returnString = strings.Join([]string{returnString, fmt.Sprintf("%d: 0x%s", i, hex.EncodeToString(measurement.Expected))}, ",")
	}
	return returnString
}

func (m *M) fromImageMeasurementsV2(
	measurements ImageMeasurementsV2, wantVersion versionsapi.Version,
	csp cloudprovider.Provider, attestationVariant variant.Variant,
) error {
	gotVersion, err := versionsapi.NewVersion(measurements.Ref, measurements.Stream, measurements.Version, versionsapi.VersionKindImage)
	if err != nil {
		return fmt.Errorf("invalid metadata version: %w", err)
	}
	if !wantVersion.Equal(gotVersion) {
		return fmt.Errorf("invalid measurement metadata: version mismatch: expected %s, got %s", wantVersion.ShortPath(), gotVersion.ShortPath())
	}

	// find measurements for requested image in list
	var measurementsEntry ImageMeasurementsV2Entry
	var found bool
	for _, entry := range measurements.List {
		gotCSP := entry.CSP
		if gotCSP != csp {
			continue
		}
		gotAttestationVariant, err := variant.FromString(entry.AttestationVariant)
		if err != nil {
			continue
		}
		if gotAttestationVariant == nil || attestationVariant == nil {
			continue
		}
		if !gotAttestationVariant.Equal(attestationVariant) {
			continue
		}
		measurementsEntry = entry
		found = true
		break
	}

	if !found {
		return fmt.Errorf("invalid measurement metadata: no measurements found for csp %s, attestationVariant %s and image %s", csp.String(), attestationVariant, wantVersion.ShortPath())
	}

	*m = measurementsEntry.Measurements
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
		return fmt.Errorf("decoding measurement: %w", err)
	}

	if len(expected) != 32 && len(expected) != 48 {
		return fmt.Errorf("invalid measurement: invalid length: %d", len(expected))
	}

	m.Expected = expected
	m.ValidationOpt = eM.WarnOnly

	return nil
}

// WithAllBytes returns a measurement value where all bytes are set to b. Takes a dynamic length as input.
// Expected are either 32 bytes (PCRMeasurementLength) or 48 bytes (TDXMeasurementLength).
// Over inputs are possible in this function, but potentially rejected elsewhere.
func WithAllBytes(b byte, validationOpt MeasurementValidationOption, length int) Measurement {
	return Measurement{
		Expected:      bytes.Repeat([]byte{b}, length),
		ValidationOpt: validationOpt,
	}
}

// PlaceHolderMeasurement returns a measurement with placeholder values for Expected.
func PlaceHolderMeasurement(length int) Measurement {
	return Measurement{
		Expected:      bytes.Repeat([]byte{0x12, 0x34}, length/2),
		ValidationOpt: Enforce,
	}
}

// DefaultsFor provides the default measurements for given cloud provider.
func DefaultsFor(provider cloudprovider.Provider, attestationVariant variant.Variant) M {
	switch {
	case provider == cloudprovider.AWS && attestationVariant == variant.AWSNitroTPM{}:
		return aws_AWSNitroTPM.Copy()

	case provider == cloudprovider.AWS && attestationVariant == variant.AWSSEVSNP{}:
		return aws_AWSSEVSNP.Copy()

	case provider == cloudprovider.Azure && attestationVariant == variant.AzureSEVSNP{}:
		return azure_AzureSEVSNP.Copy()

	case provider == cloudprovider.Azure && attestationVariant == variant.AzureTDX{}:
		return azure_AzureTDX.Copy()

	case provider == cloudprovider.Azure && attestationVariant == variant.AzureTrustedLaunch{}:
		return azure_AzureTrustedLaunch.Copy()

	case provider == cloudprovider.GCP && attestationVariant == variant.GCPSEVES{}:
		return gcp_GCPSEVES.Copy()

	case provider == cloudprovider.GCP && attestationVariant == variant.GCPSEVSNP{}:
		return gcp_GCPSEVSNP.Copy()

	case provider == cloudprovider.OpenStack && attestationVariant == variant.QEMUVTPM{}:
		return openstack_QEMUVTPM.Copy()

	case provider == cloudprovider.QEMU && attestationVariant == variant.QEMUTDX{}:
		return qemu_QEMUTDX.Copy()

	case provider == cloudprovider.QEMU && attestationVariant == variant.QEMUVTPM{}:
		return qemu_QEMUVTPM.Copy()

	default:
		return nil
	}
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

// getFromURL fetches the content from the given URL and returns the content as a byte slice.
func getFromURL(ctx context.Context, client *http.Client, sourceURL *url.URL) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

type cosignVerifier interface {
	VerifySignature(content, signature []byte) error
}
