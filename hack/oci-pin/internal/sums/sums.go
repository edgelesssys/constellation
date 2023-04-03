/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

// sums creates and combines sha256sums files.
package sums

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"
)

var sumRegexp = regexp.MustCompile(`^[a-f0-9]{64}$`)

const (
	sumLength     = 64
	minLineLength = sumLength + 2 // 64 hex + 2 space
	notFound      = -1            // index not found
)

// Create creates a sha256sums file from a list of digests.
// The digests going in are expected to be in the oci format "sha256:hex".
func Create(refs []PinnedImageReference, out io.Writer) error {
	coalescedRefs, err := coalesceRefs(refs)
	if err != nil {
		return err
	}
	for _, ref := range coalescedRefs {
		if _, err := fmt.Fprintf(out, "%s  %s\n", ref.Sum(), ref.Reference()); err != nil {
			return err
		}
	}
	return nil
}

// Merge merges multiple sha256sums files into one.
func Merge(refs [][]PinnedImageReference, out io.Writer) error {
	mergedRefs := make([]PinnedImageReference, 0)
	for _, refs := range refs {
		mergedRefs = append(mergedRefs, refs...)
	}
	return Create(mergedRefs, out)
}

// Parse parses a sha256sums file.
func Parse(in io.Reader) ([]PinnedImageReference, error) {
	scanner := newLineScanner(in)
	return parseLines(scanner)
}

// PinnedImageReference contains the components of a pinned image reference.
type PinnedImageReference struct {
	Registry string
	Prefix   string
	Name     string
	Tag      string
	Digest   string
}

// Reference returns the string representation of the reference (without the digest).
func (r PinnedImageReference) Reference() string {
	var base string
	if r.Prefix == "" {
		base = path.Join(r.Registry, r.Name)
	} else {
		base = path.Join(r.Registry, r.Prefix, r.Name)
	}
	var tag string
	if r.Tag != "" {
		tag = ":" + r.Tag
	}
	return base + tag
}

// Sum returns the string representation of the digest.
// The digest is expected to be in the oci format "sha256:hex".
// The resulting Sum is only the hex part.
func (r PinnedImageReference) Sum() string {
	return r.Digest[len("sha256:"):]
}

// coalesceRefs coalesces the image references.
// It sorts the references and removes duplicates.
// If conflicting digests are found, an error is returned.
func coalesceRefs(refs []PinnedImageReference) ([]PinnedImageReference, error) {
	sortRefs(refs)
	uniqueRefs := make([]PinnedImageReference, 0, len(refs))
	var prev PinnedImageReference
	for _, ref := range refs {
		equal, err := compareRefs(prev, ref)
		if err != nil {
			return nil, err
		}
		if !equal {
			uniqueRefs = append(uniqueRefs, ref)
			prev = ref
		}
	}
	return uniqueRefs, nil
}

func sortRefs(refs []PinnedImageReference) {
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Registry != refs[j].Registry {
			return refs[i].Registry < refs[j].Registry
		}
		if refs[i].Prefix != refs[j].Prefix {
			return refs[i].Prefix < refs[j].Prefix
		}
		if refs[i].Name != refs[j].Name {
			return refs[i].Name < refs[j].Name
		}
		if refs[i].Tag != refs[j].Tag {
			return refs[i].Tag < refs[j].Tag
		}
		return refs[i].Digest < refs[j].Digest
	})
}

// compareRefs compares two references.
// references are equal if they have the same registry, prefix, name, tag and digest.
// If refs only differ in digest, they are inconsistent and an error is returned.
func compareRefs(a, b PinnedImageReference) (equal bool, err error) {
	if a.Registry != b.Registry {
		return false, nil
	}
	if a.Prefix != b.Prefix {
		return false, nil
	}
	if a.Name != b.Name {
		return false, nil
	}
	if a.Tag != b.Tag {
		return false, nil
	}
	if a.Digest != b.Digest {
		return false, errors.New("conflicting digests")
	}
	return true, nil
}

func parseLines(scanner scanner) ([]PinnedImageReference, error) {
	var refs []PinnedImageReference
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		ref, err := parseLine(line)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

// parseLine parses a line from a sha256sums file.
// The line is expected to be in the format "hex  reference".
func parseLine(line string) (PinnedImageReference, error) {
	parts := strings.Split(line, "  ")
	if len(parts) != 2 {
		return PinnedImageReference{}, fmt.Errorf("line does not have exactly 2 parts separated by two spaces: %q", line)
	}
	return refFromSumAndRef(parts[0], parts[1])
}

func refFromSumAndRef(sum, ref string) (PinnedImageReference, error) {
	if !sumRegexp.MatchString(sum) {
		return PinnedImageReference{}, fmt.Errorf("invalid sum: %q", sum)
	}

	// last colon is separator between name and tag
	var base, tag string
	tagSep := strings.LastIndexByte(ref, ':')
	if tagSep == notFound {
		base = ref
	} else {
		base = ref[:tagSep]
		tag = ref[tagSep+1:]
	}

	// first slash is separator between registry and full name
	registrySep := strings.IndexByte(base, '/')
	if registrySep == notFound {
		return PinnedImageReference{}, fmt.Errorf("invalid reference: missing registry %q", ref)
	}

	registry := base[:registrySep]
	fullName := base[registrySep+1:]

	// last slash is separator between prefix and short name
	var prefix, name string
	nameSep := strings.LastIndexByte(fullName, '/')
	if nameSep == notFound {
		name = fullName
	} else {
		prefix = fullName[:nameSep]
		name = fullName[nameSep+1:]
	}

	return PinnedImageReference{
		Registry: registry,
		Prefix:   prefix,
		Name:     name,
		Tag:      tag,
		Digest:   "sha256:" + sum,
	}, nil
}

func newLineScanner(r io.Reader) scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	return scanner
}

type scanner interface {
	Scan() bool
	Text() string
}
