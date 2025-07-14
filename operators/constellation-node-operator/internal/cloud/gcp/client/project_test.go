/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package client

import (
	"errors"
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestCanonicalProjectID(t *testing.T) {
	testCases := map[string]struct {
		projectID     string
		project       *computepb.Project
		getProjectErr error
		wantProjectID string
		wantErr       bool
	}{
		"already canonical": {
			projectID:     "project-12345",
			wantProjectID: "project-12345",
		},
		"numeric project id": {
			projectID:     "12345",
			wantProjectID: "project-12345",
			project:       &computepb.Project{Name: proto.String("project-12345")},
		},
		"numeric project id with error": {
			projectID:     "12345",
			wantProjectID: "project-12345",
			getProjectErr: errors.New("get error"),
			wantErr:       true,
		},
		"numeric project id with nil project": {
			projectID: "12345",
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				projectAPI: &stubProjectAPI{
					project: tc.project,
					getErr:  tc.getProjectErr,
				},
			}
			gotID, err := client.canonicalProjectID(t.Context(), tc.projectID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantProjectID, gotID)
		})
	}
}
