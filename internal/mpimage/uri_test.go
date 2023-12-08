/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package mpimage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFromURI(t *testing.T) {
	testCases := map[string]struct {
		uri     string
		want    MarketplaceImage
		wantErr bool
	}{
		"azure valid": {
			uri: "constellation-marketplace-image://Azure?offer=constellation&publisher=edgelesssystems&sku=constellation&version=1.2.3",
			want: AzureMarketplaceImage{
				Publisher: "edgelesssystems",
				Offer:     "constellation",
				SKU:       "constellation",
				Version:   "1.2.3",
			},
		},
		"azure invalid version": {
			uri:     "constellation-marketplace-image://Azure?offer=constellation&publisher=edgelesssystems&sku=constellation&version=asdf",
			wantErr: true,
		},
		"invalid scheme": {
			uri:     "invalid://Azure?offer=constellation&publisher=edgelesssystems&sku=constellation&version=1.2.3",
			wantErr: true,
		},
		"invalid host": {
			uri:     "constellation-marketplace-image://invalid?offer=constellation&publisher=edgelesssystems&sku=constellation&version=1.2.3",
			wantErr: true,
		},
		"no uri": {
			uri:     "no uri",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			got, err := NewFromURI(tc.uri)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.want, got)
			}
		})
	}
}

func TestAzureURI(t *testing.T) {
	testCases := map[string]struct {
		image AzureMarketplaceImage
		want  string
	}{
		"valid": {
			image: AzureMarketplaceImage{
				Publisher: "foo",
				Offer:     "bar",
				SKU:       "baz",
				Version:   "1.2.3",
			},
			want: "constellation-marketplace-image://Azure?offer=bar&publisher=foo&sku=baz&version=1.2.3",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.image.URI())
		})
	}
}
