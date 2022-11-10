/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	// DefaultResourceGroupName to find Constellation images in.
	DefaultResourceGroupName = "CONSTELLATION-IMAGES"
	// DefaultGalleryName to find Constellation images in.
	DefaultGalleryName = "Constellation_CVM"
	// DefaultImageDefinition to find Constellation images in.
	DefaultImageDefinition = "constellation"
)

// Options for Azure Client to download image references.
type Options struct {
	SubscriptionID    string
	ResourceGroupName string
	GalleryName       string
	ImageDefinition   string
}

// DefaultOptions creates an Options object with good defaults.
func DefaultOptions() Options {
	return Options{
		SubscriptionID:    "",
		ResourceGroupName: DefaultResourceGroupName,
		GalleryName:       DefaultGalleryName,
		ImageDefinition:   DefaultImageDefinition,
	}
}

// SetSubscription sets subscription from string. It expects a UUID conform value.
func (o *Options) SetSubscription(sub string) error {
	if _, err := uuid.Parse(sub); err != nil {
		return fmt.Errorf("unable to set subscription: %w", err)
	}
	o.SubscriptionID = sub
	return nil
}
