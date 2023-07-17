/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package clusterid

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	testCases := map[string]struct {
		current File
		other   File
		want    File
	}{
		"empty": {
			current: File{},
			other:   File{},
			want:    File{},
		},
		"current empty": {
			current: File{},
			other: File{
				ClusterID: "clusterID",
			},
			want: File{
				ClusterID: "clusterID",
			},
		},
		"other empty": {
			current: File{
				ClusterID: "clusterID",
			},
			other: File{},
			want: File{
				ClusterID: "clusterID",
			},
		},
		"both set": {
			current: File{
				ClusterID: "clusterID",
			},
			other: File{
				ClusterID: "otherClusterID",
			},
			want: File{
				ClusterID: "otherClusterID",
			},
		},
	}

	for _, tc := range testCases {
		require := require.New(t)

		ret := tc.current.Merge(tc.other)
		require.Equal(tc.want, *ret)
	}
}
