/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package idkeydigest

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"go.uber.org/multierr"
)

// IDKeyDigests is a list of trusted digest values for the ID key.
type IDKeyDigests [][]byte

type encodedIDKeyDigests []string

// NewIDKeyDigests creates a new IDKeyDigests from a list of digests.
func NewIDKeyDigests(digests [][]byte) IDKeyDigests {
	idKeyDigests := make(IDKeyDigests, len(digests))
	copy(idKeyDigests, digests)
	return idKeyDigests
}

// MustGetDefaultIDKeyDigests returns the default ID key digests.
func MustGetDefaultIDKeyDigests() IDKeyDigests {
	digestOld, err := hex.DecodeString("57486a447ec0f1958002a22a06b7673b9fd27d11e1c6527498056054c5fa92d23c50f9de44072760fe2b6fb89740b696")
	if err != nil {
		panic(err)
	}
	digestNew, err := hex.DecodeString("0356215882a825279a85b300b0b742931d113bf7e32dde2e50ffde7ec743ca491ecdd7f336dc28a6e0b2bb57af7a44a3")
	if err != nil {
		panic(err)
	}
	return IDKeyDigests{digestOld, digestNew}
}

// ToProto converts the IDKeyDigests to a protobuf compatible format.
func (d IDKeyDigests) ToProto() [][]byte {
	digests := make([][]byte, len(d))
	copy(digests, d)
	return digests
}

// MarshalYAML implements the yaml.Marshaler interface.
func (d IDKeyDigests) MarshalYAML() (any, error) {
	encodedIDKeyDigests := []string{}
	for _, digest := range d {
		encodedIDKeyDigests = append(encodedIDKeyDigests, hex.EncodeToString(digest))
	}
	return encodedIDKeyDigests, nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (d *IDKeyDigests) UnmarshalYAML(unmarshal func(any) error) error {
	var encodedDigests encodedIDKeyDigests
	if err := unmarshal(&encodedDigests); err != nil {
		// Unmarshalling failed, IDKeyDigests might be a simple string instead of IDKeyDigests struct.
		var unmarshalledString string
		if legacyErr := unmarshal(&unmarshalledString); legacyErr != nil {
			return multierr.Append(
				err,
				fmt.Errorf("trying legacy format: %w", legacyErr),
			)
		}
		encodedDigests = append(encodedDigests, unmarshalledString)
	}
	if err := d.unmarshal(encodedDigests); err != nil {
		return fmt.Errorf("unmarshalling yaml: %w", err)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (d IDKeyDigests) MarshalJSON() ([]byte, error) {
	encodedIDKeyDigests := []string{}
	for _, digest := range d {
		encodedIDKeyDigests = append(encodedIDKeyDigests, hex.EncodeToString(digest))
	}
	return json.Marshal(encodedIDKeyDigests)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *IDKeyDigests) UnmarshalJSON(b []byte) error {
	var encodedDigests encodedIDKeyDigests
	if err := json.Unmarshal(b, &encodedDigests); err != nil {
		// Unmarshalling failed, IDKeyDigests might be a simple string instead of IDKeyDigests struct.
		var unmarshalledString string
		if legacyErr := json.Unmarshal(b, &unmarshalledString); legacyErr != nil {
			return multierr.Append(
				err,
				fmt.Errorf("trying legacy format: %w", legacyErr),
			)
		}
		encodedDigests = []string{unmarshalledString}
	}
	if err := d.unmarshal(encodedDigests); err != nil {
		return fmt.Errorf("unmarshalling json: %w", err)
	}
	return nil
}

// unmarshal is a helper function for unmarshalling encodedIDKeyDigests into IDKeyDigests.
func (d *IDKeyDigests) unmarshal(encodedDigests encodedIDKeyDigests) error {
	for _, encodedDigest := range encodedDigests {
		digest, err := hex.DecodeString(encodedDigest)
		if err != nil {
			return fmt.Errorf("decoding digest: %w", err)
		}
		*d = append(*d, digest)
	}
	return nil
}
