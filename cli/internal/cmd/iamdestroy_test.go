/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/
package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDestroyIAMUser(t *testing.T) {
	someError := errors.New("failed")

	testCases := map[string]struct {
		iamDestroyer iamDestroyer
		wantErr      bool
	}{
		"success": {
			iamDestroyer: &stubIAMDestroyer{},
		},
		"failure": {
			iamDestroyer: &stubIAMDestroyer{destroyErr: someError},
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := destroyIAMUser(context.Background(), &nopSpinner{}, tc.iamDestroyer)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
