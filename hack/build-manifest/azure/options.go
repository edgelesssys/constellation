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
	DefaultResourceGroupName = "CONSTELLATION-IMAGES"
	DefaultGalleryName       = "Constellation_CVM"
	DefaultImageDefinition   = "constellation"
)

type Options struct {
	SubscriptionID    string
	ResourceGroupName string
	GalleryName       string
	ImageDefinition   string
}

func DefaultOptions() Options {
	return Options{
		SubscriptionID:    "",
		ResourceGroupName: DefaultResourceGroupName,
		GalleryName:       DefaultGalleryName,
		ImageDefinition:   DefaultImageDefinition,
	}
}

func (o *Options) SetSubscription(sub string) error {
	if _, err := uuid.Parse(sub); err != nil {
		return fmt.Errorf("unable to set subscription: %w", err)
	}
	o.SubscriptionID = sub
	return nil
}
