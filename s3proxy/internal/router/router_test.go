/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateContentMD5(t *testing.T) {
	tests := map[string]struct {
		contentMD5     string
		body           []byte
		expectedErrMsg string
	}{
		"empty content-md5": {
			contentMD5: "",
			body:       []byte("hello, world"),
		},
		// https://datatracker.ietf.org/doc/html/rfc1864#section-2
		"valid content-md5": {
			contentMD5: "Q2hlY2sgSW50ZWdyaXR5IQ==",
			body:       []byte("Check Integrity!"),
		},
		"invalid content-md5": {
			contentMD5:     "invalid base64",
			body:           []byte("hello, world"),
			expectedErrMsg: "decoding base64",
		},
	}

	// Iterate over the test cases
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Call the validateContentMD5 function
			err := validateContentMD5(tc.contentMD5, tc.body)

			// Check the result against the expected value
			if tc.expectedErrMsg != "" {
				assert.ErrorContains(t, err, tc.expectedErrMsg)
			}
		})
	}
}

func TestByteSliceToByteArray(t *testing.T) {
	tests := map[string]struct {
		input   []byte
		output  [32]byte
		wantErr bool
	}{
		"empty input": {
			input:  []byte{},
			output: [32]byte{},
		},
		"successful input": {
			input:  []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
			output: [32]byte{0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41},
		},
		"input too short": {
			input:   []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
			output:  [32]byte{0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41},
			wantErr: true,
		},
		"input too long": {
			input:   []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
			output:  [32]byte{0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := byteSliceToByteArray(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, tc.output, result)
		})
	}
}
