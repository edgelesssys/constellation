//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sigstore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestRekorSearchByHash(t *testing.T) {
	testCases := map[string]struct {
		hash      string
		wantEmpty bool
	}{
		"Constellation CLI v2.0.0 hash": {
			hash: "40e137b9b9b8204d672642fd1e181c6d5ccb50cfc5cc7fcbb06a8c2c78f44aff",
		},
		"other hash": {
			hash:      "d9c5a43ba6284e1059b7e871bcf9b52f376d62b9198f300b1402d1c4d9b7431f",
			wantEmpty: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			rekor, err := NewRekor()
			require.NoError(err)

			uuids, err := rekor.SearchByHash(context.Background(), tc.hash)
			assert.NoError(err)

			if tc.wantEmpty {
				assert.Empty(err)
				return
			}
			assert.NotEmpty(uuids)
		})
	}
}

func TestVerifyEntry(t *testing.T) {
	testCases := map[string]struct {
		uuid      string
		pubKey    string
		wantError bool
	}{
		"Constellation CLI v2.0.0": {
			uuid:   "362f8ecba72f4326afaba7f6635b3e058888692841848e5514357315be9528474b23f5dcccb82b13",
			pubKey: "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFZjhGMWhwbXdFK1lDRlh6akd0YVFjckw2WFpWVApKbUVlNWlTTHZHMVN5UVNBZXc3V2RNS0Y2bzl0OGUyVEZ1Q2t6bE9oaGx3czJPSFdiaUZabkZXQ0Z3PT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==",
		},
		"unknown uuid": {
			uuid:      "46073a33852fc797ccc341a30323bd69119ff03936bf8d17061606e3e2e4be1fe70dccaa1b66bc34",
			pubKey:    "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFZjhGMWhwbXdFK1lDRlh6akd0YVFjckw2WFpWVApKbUVlNWlTTHZHMVN5UVNBZXc3V2RNS0Y2bzl0OGUyVEZ1Q2t6bE9oaGx3czJPSFdiaUZabkZXQ0Z3PT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==",
			wantError: true,
		},
		"broken key": {
			uuid:      "362f8ecba72f4326afaba7f6635b3e058888692841848e5514357315be9528474b23f5dcccb82b13",
			pubKey:    "d2VsbCB0aGlzIGlzIGRlZmluaXRlbHkgbm90IGEga2V5",
			wantError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			rekor, err := NewRekor()
			require.NoError(err)

			err = rekor.VerifyEntry(context.Background(), tc.uuid, tc.pubKey)
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}
