/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package idkeydigest

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

// IDKeyDigests is a list of trusted digest values for the ID key.
type IDKeyDigests [][]byte

type encodedIDKeyDigests []string

// encodedDigestLength is the length of a digest in hex encoding.
const encodedDigestLength = 2 * 48

// NewIDKeyDigests creates a new IDKeyDigests from a list of digests.
func NewIDKeyDigests(digests [][]byte) IDKeyDigests {
	idKeyDigests := make(IDKeyDigests, len(digests))
	copy(idKeyDigests, digests)
	return idKeyDigests
}

// DefaultsFor returns the default IDKeyDigests for the given cloud provider.
func DefaultsFor(csp cloudprovider.Provider) IDKeyDigests {
	switch csp {
	case cloudprovider.Azure:
		return IDKeyDigests{
			{0x57, 0x48, 0x6a, 0x44, 0x7e, 0xc0, 0xf1, 0x95, 0x80, 0x02, 0xa2, 0x2a, 0x06, 0xb7, 0x67, 0x3b, 0x9f, 0xd2, 0x7d, 0x11, 0xe1, 0xc6, 0x52, 0x74, 0x98, 0x05, 0x60, 0x54, 0xc5, 0xfa, 0x92, 0xd2, 0x3c, 0x50, 0xf9, 0xde, 0x44, 0x07, 0x27, 0x60, 0xfe, 0x2b, 0x6f, 0xb8, 0x97, 0x40, 0xb6, 0x96},
			{0x03, 0x56, 0x21, 0x58, 0x82, 0xa8, 0x25, 0x27, 0x9a, 0x85, 0xb3, 0x00, 0xb0, 0xb7, 0x42, 0x93, 0x1d, 0x11, 0x3b, 0xf7, 0xe3, 0x2d, 0xde, 0x2e, 0x50, 0xff, 0xde, 0x7e, 0xc7, 0x43, 0xca, 0x49, 0x1e, 0xcd, 0xd7, 0xf3, 0x36, 0xdc, 0x28, 0xa6, 0xe0, 0xb2, 0xbb, 0x57, 0xaf, 0x7a, 0x44, 0xa3},
			{0x93, 0x4f, 0x68, 0xbd, 0x8b, 0xa0, 0x19, 0x38, 0xee, 0xc2, 0x14, 0x75, 0xc8, 0x72, 0xe3, 0xa9, 0x42, 0xb6, 0x0c, 0x59, 0xfa, 0xfc, 0x6d, 0xf9, 0xe9, 0xa7, 0x6e, 0xe6, 0x6b, 0xc4, 0x7f, 0x2d, 0x09, 0xc6, 0x76, 0xf6, 0x1c, 0x03, 0x15, 0xc5, 0x78, 0xda, 0x26, 0x08, 0x5f, 0xb1, 0x3a, 0x71},
		}
	default:
		return nil
	}
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
			return errors.Join(
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
			return errors.Join(
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
		if len(encodedDigest) != encodedDigestLength {
			return fmt.Errorf("invalid digest length: %d", len(encodedDigest))
		}
		digest, err := hex.DecodeString(encodedDigest)
		if err != nil {
			return fmt.Errorf("decoding digest: %w", err)
		}
		*d = append(*d, digest)
	}
	return nil
}
