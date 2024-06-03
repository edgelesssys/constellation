/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package dnsmasq

import (
	"bufio"
	"strings"

	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/dhcp"
	"github.com/spf13/afero"
)

// DNSMasq is a DHCP lease getter for dnsmasq.
type DNSMasq struct {
	leasesFileName string
	fs             *afero.Afero
}

// New creates a new DNSMasq.
func New(leasesFileName string) *DNSMasq {
	return &DNSMasq{
		leasesFileName: leasesFileName,
		fs:             &afero.Afero{Fs: afero.NewOsFs()},
	}
}

// GetDHCPLeases returns the underlying DHCP leases.
func (d *DNSMasq) GetDHCPLeases() ([]dhcp.NetworkDHCPLease, error) {
	file, err := d.fs.Open(d.leasesFileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// read file
	var leases []dhcp.NetworkDHCPLease
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// split by whitespace
		fields := strings.Fields(line)
		leases = append(leases, dhcp.NetworkDHCPLease{
			IPaddr:   fields[2],
			Hostname: fields[3],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return leases, nil
}
