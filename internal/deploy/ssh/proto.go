/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package ssh

import (
	"github.com/edgelesssys/constellation/bootstrapper/initproto"
)

// FromProtoSlice converts a SSH UserKey definition from pubproto to the Go flavor.
func FromProtoSlice(input []*initproto.SSHUserKey) []UserKey {
	if input == nil {
		return nil
	}

	output := make([]UserKey, 0)

	for _, pair := range input {
		singlePair := UserKey{
			Username:  pair.Username,
			PublicKey: pair.PublicKey,
		}

		output = append(output, singlePair)
	}

	return output
}

// ToProtoSlice converts a SSH UserKey definition from Go to pubproto flavor.
func ToProtoSlice(input []*UserKey) []*initproto.SSHUserKey {
	if input == nil {
		return nil
	}

	output := make([]*initproto.SSHUserKey, 0)
	for _, pair := range input {
		singlePair := initproto.SSHUserKey{
			Username:  pair.Username,
			PublicKey: pair.PublicKey,
		}

		output = append(output, &singlePair)
	}

	return output
}
