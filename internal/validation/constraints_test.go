/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package validation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIPAddress(t *testing.T) {
	testCases := map[string]struct {
		ip      string
		wantErr bool
	}{
		"valid ipv4": {
			ip: "127.0.0.1",
		},
		"valid ipv6": {
			ip: "2001:db8::68",
		},
		"invalid": {
			ip:      "invalid",
			wantErr: true,
		},
		"empty": {
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := IPAddress(tc.ip).Satisfied()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCIDR(t *testing.T) {
	testCases := map[string]struct {
		cidr    string
		wantErr bool
	}{
		"valid ipv4": {
			cidr: "192.0.2.0/24",
		},
		"valid ipv6": {
			cidr: "2001:db8::/32",
		},
		"invalid": {
			cidr:    "invalid",
			wantErr: true,
		},
		"empty": {
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := CIDR(tc.cidr).Satisfied()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDNSName(t *testing.T) {
	testCases := map[string]struct {
		dnsName string
		wantErr bool
	}{
		"valid": {
			dnsName: "example.com",
		},
		"invalid": {
			dnsName: "invalid",
			wantErr: true,
		},
		"empty": {
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := DNSName(tc.dnsName).Satisfied()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEmptySlice(t *testing.T) {
	testCases := map[string]struct {
		s       []any
		wantErr bool
	}{
		"valid": {
			s: []any{},
		},
		"nil": {
			s: nil,
		},
		"invalid": {
			s:       []any{1},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := EmptySlice(tc.s).Satisfied()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAll(t *testing.T) {
	c := func(i int, s string) *Constraint {
		return Equal(s, "abc")
	}
	testCases := map[string]struct {
		s       []string
		wantErr bool
	}{
		"valid": {
			s: []string{"abc", "abc", "abc"},
		},
		"nil": {
			s: nil,
		},
		"empty": {
			s: []string{},
		},
		"all are invalid": {
			s:       []string{"def", "lol"},
			wantErr: true,
		},
		"one is invalid": {
			s:       []string{"abc", "def"},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := All(tc.s, c).Satisfied()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
