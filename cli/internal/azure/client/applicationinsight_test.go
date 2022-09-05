/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateApplicationInsight(t *testing.T) {
	testCases := map[string]struct {
		applicationInsightsAPI applicationInsightsAPI
		wantErr                bool
	}{
		"successful create": {
			applicationInsightsAPI: &stubApplicationInsightsAPI{
				err: nil,
			},
		},
		"failed create": {
			applicationInsightsAPI: &stubApplicationInsightsAPI{
				err: errors.New("some error"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := Client{
				applicationInsightsAPI: tc.applicationInsightsAPI,
			}

			err := client.CreateApplicationInsight(context.Background())

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}
