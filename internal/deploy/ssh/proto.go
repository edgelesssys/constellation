package ssh

import (
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
)

// FromProtoSlice converts a SSH UserKey definition from pubproto to the Go flavor.
func FromProtoSlice(input []*pubproto.SSHUserKey) []UserKey {
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
func ToProtoSlice(input []*UserKey) []*pubproto.SSHUserKey {
	if input == nil {
		return nil
	}

	output := make([]*pubproto.SSHUserKey, 0)
	for _, pair := range input {
		singlePair := pubproto.SSHUserKey{
			Username:  pair.Username,
			PublicKey: pair.PublicKey,
		}

		output = append(output, &singlePair)
	}

	return output
}
