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
				require.Nil(t, err)
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
				require.Nil(t, err)
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
				require.Nil(t, err)
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
				require.Nil(t, err)
			}
		})
	}
}

func TestNotEmptySlice(t *testing.T) {
	testCases := map[string]struct {
		s       []any
		wantErr bool
	}{
		"valid": {
			s: []any{1},
		},
		"invalid": {
			s:       []any{},
			wantErr: true,
		},
		"nil": {
			s:       nil,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := NotEmptySlice(tc.s).Satisfied()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
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
				require.Nil(t, err)
			}
		})
	}
}

func TestNotEqual(t *testing.T) {
	testCases := map[string]struct {
		a       any
		b       any
		wantErr bool
	}{
		"valid": {
			a: "abc",
			b: "def",
		},
		"invalid": {
			a:       "abc",
			b:       "abc",
			wantErr: true,
		},
		"empty": {
			wantErr: true,
		},
		"one empty": {
			a: "abc",
			b: "",
		},
		"one nil": {
			a: "abc",
			b: nil,
		},
		"both nil": {
			a:       nil,
			b:       nil,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := NotEqual(tc.a, tc.b).Satisfied()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestIfNotNil(t *testing.T) {
	testCases := map[string]struct {
		a       *int
		c       func() *Constraint
		wantErr bool
	}{
		"valid": {
			a: new(int),
			c: func() *Constraint {
				return &Constraint{
					Satisfied: func() *TreeError {
						return nil
					},
				}
			},
		},
		"nil": {
			a: nil,
			c: func() *Constraint {
				return &Constraint{
					Satisfied: func() *TreeError {
						t.Fatal("should not be called")
						return nil
					},
				}
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := IfNotNil(tc.a, tc.c).Satisfied()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}
