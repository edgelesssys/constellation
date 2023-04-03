/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

// inject renders Go source files with injected pinning values.
package inject

import (
	"errors"
	"io"
	"regexp"
	"text/template"
)

// Render renders the source code to inject the pinned values.
func Render(out io.Writer, vals PinningValues) error {
	if err := validate(vals); err != nil {
		return err
	}
	return sourceTpl.Execute(out, vals)
}

// validate validates the values to inject.
func validate(vals PinningValues) error {
	if vals.Package == "" {
		return errors.New("package is empty")
	}
	if vals.Ident == "" {
		return errors.New("identifier is empty")
	}
	if vals.Registry == "" {
		return errors.New("registry is empty")
	}
	if vals.Name == "" {
		return errors.New("name is empty")
	}
	if vals.Digest == "" {
		return errors.New("digest is empty")
	}

	if matched := packageNameRegexp.MatchString(vals.Package); !matched {
		return errors.New("package is not valid")
	}
	if matched := identRegexp.MatchString(vals.Ident); !matched {
		return errors.New("identifier is not valid")
	}
	if matched := registryRegexp.MatchString(vals.Registry); !matched {
		return errors.New("registry is not valid")
	}
	if matched := prefixRegexp.MatchString(vals.Prefix); !matched {
		return errors.New("prefix is not valid")
	}
	if matched := nameRegexp.MatchString(vals.Name); !matched {
		return errors.New("name is not valid")
	}
	if matched := tagRegexp.MatchString(vals.Tag); vals.Tag != "" && !matched {
		return errors.New("tag is not valid")
	}
	if matched := digestRegexp.MatchString(vals.Digest); !matched {
		return errors.New("digest is not valid")
	}

	return nil
}

// PinningValues contains the values to inject into the generated source code.
type PinningValues struct {
	// Package is the name of the package to generate.
	Package string
	// Ident is the base identifier of the generated constants.
	Ident string
	// Registry string
	Registry string
	// Prefix is the prefix of the container image name.
	Prefix string
	// Name is the name of the container image.
	Name string
	// Tag is the (optional) tag of the container image.
	Tag string
	// Digest is the digest of the container image.
	Digest string
}

var (
	sourceTpl = template.Must(template.New("goSource").Parse(goSourceTpl))

	// packageNameRegexp is the regular expression for a valid package name.
	packageNameRegexp = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

	// identRegexp is the regular expression for a valid identifier.
	identRegexp = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

	// registryRegexp is the regular expression for a valid registry.
	registryRegexp = regexp.MustCompile(`^[^\s/"]+$`)

	// prefixRegexp is the regular expression for a valid prefix.
	prefixRegexp = regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9-_/]*)?$`)

	// nameRegexp is the regular expression for a valid name.
	nameRegexp = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-_]*$`)

	// tagRegexp is the regular expression for a valid tag.
	tagRegexp = regexp.MustCompile(`[\w][\w.-]{0,127}`)

	// digestRegexp is the regular expression for a valid digest.
	digestRegexp = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
)

const goSourceTpl = `package {{.Package}}

// Code generated by oci-pin. DO NOT EDIT.

const (
	// {{.Ident}}Registry is the {{.Name}} container image registry.
	{{.Ident}}Registry = "{{.Registry}}"

{{if .Prefix }}	// {{.Ident}}Prefix is the {{.Name}} container image prefix.
	{{.Ident}}Prefix = "{{.Prefix}}"

{{end}}	// {{.Ident}}Name is the {{.Name}} container image short name part.
	{{.Ident}}Name = "{{.Name}}"

{{if .Tag }}	// {{.Ident}}Tag is the tag for the {{.Name}} container image.
	{{.Ident}}Tag = "{{.Tag}}"

{{end}}	// {{.Ident}}Digest is the digest for the {{.Name}} container image.
	{{.Ident}}Digest = "{{.Digest}}"
)
`
