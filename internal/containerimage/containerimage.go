/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
This package provides container image names, registry info and digests.

It should only be used by the CLI and never be imported by any package that
ends up in a container image to avoid circular dependencies.
*/
package containerimage

import (
	"errors"
	"path"
	"regexp"
)

// Image is a container image reference.
// It has the following format:
// <registry>/<prefix>/<name>:<tag>@<digest>
// where <registry> is the registry where the image is located,
// <prefix> is the (optional) prefix of the image name,
// <name> is the name of the image,
// <tag> is the (optional) tag of the image,
// <digest> is the digest of the image.
type Image struct {
	// Registry is the registry where the image is located.
	Registry string
	// Prefix is the prefix of the image name.
	Prefix string
	// Name is the name of the image.
	Name string
	// Tag is the tag of the image.
	Tag string
	// Digest is the digest of the image.
	Digest string
}

// Validate validates the image.
func (i Image) Validate() error {
	if i.Registry == "" {
		return errors.New("image registry is empty")
	}
	if i.Name == "" {
		return errors.New("image name is empty")
	}
	if i.Digest == "" {
		return errors.New("image digest is empty")
	}
	if matched := digestRegexp.MatchString(i.Digest); !matched {
		return errors.New("image digest is not valid")
	}

	return nil
}

// String returns the image as a string.
// The format is <registry>/<prefix>/<name>:<tag>@<digest>
// or a shorter version if prefix or tag are empty.
func (i Image) String() string {
	var base string
	if i.Prefix == "" {
		base = path.Join(i.Registry, i.Name)
	} else {
		base = path.Join(i.Registry, i.Prefix, i.Name)
	}
	var tag string
	if i.Tag != "" {
		tag = ":" + i.Tag
	}
	return base + tag + "@" + i.Digest
}

// Builder is a builder for container images.
type Builder struct {
	Default  Image
	Registry string
	Prefix   string
}

// NewBuilder creates a new builder for container images.
func NewBuilder(def Image, registry, prefix string) *Builder {
	return &Builder{
		Default:  def,
		Registry: registry,
		Prefix:   prefix,
	}
}

// Build builds a container image.
func (b *Builder) Build() Image {
	img := Image{}
	if b.Registry == "" {
		img.Registry = b.Default.Registry
	} else {
		img.Registry = b.Registry
	}
	if b.Prefix == "" {
		img.Prefix = b.Default.Prefix
	} else {
		img.Prefix = b.Prefix
	}
	img.Name = b.Default.Name
	img.Tag = b.Default.Tag
	img.Digest = b.Default.Digest
	return img
}

var digestRegexp = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
