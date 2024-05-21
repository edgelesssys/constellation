/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package dnsmasq

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDHCPLeases(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	fs := afero.NewMemMapFs()
	leasesFileName := "dnsmasq.leases"
	leasesFile, err := fs.Create(leasesFileName)
	require.NoError(err)
	_, err = leasesFile.WriteString("1716219737 52:54:af:a1:98:9f 10.42.2.1 worker0 ff:c2:72:f6:09:00:02:00:00:ab:11:18:fc:48:85:40:3f:bc:41\n")
	require.NoError(err)
	_, err = leasesFile.WriteString("1716219735 52:54:7f:8f:ba:91 10.42.1.1 controlplane0 ff:c2:72:f6:09:00:02:00:00:ab:11:21:7c:b5:14:ec:43:b7:43\n")
	require.NoError(err)
	leasesFile.Close()

	d := DNSMasq{leasesFileName: leasesFileName, fs: &afero.Afero{Fs: fs}}
	leases, err := d.GetDHCPLeases()
	require.NoError(err)

	assert.Len(leases, 2)
	assert.Equal("10.42.2.1", leases[0].IPaddr)
	assert.Equal("worker0", leases[0].Hostname)
	assert.Equal("10.42.1.1", leases[1].IPaddr)
	assert.Equal("controlplane0", leases[1].Hostname)
}
