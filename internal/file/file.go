/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package file provides functions that combine file handling, JSON marshaling
and file system abstraction.
*/
package file

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/spf13/afero"
	"github.com/talos-systems/talos/pkg/machinery/config/encoder"
	"gopkg.in/yaml.v3"
)

// Option is a bitmask of options for file operations.
type Option struct {
	uint
}

// hasOption determines if a set of options contains the given option.
func hasOption(options []Option, op Option) bool {
	for _, ops := range options {
		if ops == op {
			return true
		}
	}
	return false
}

const (
	// OptNone is a no-op.
	optNone uint = iota
	// OptOverwrite overwrites an existing file.
	optOverwrite
	// OptMkdirAll creates the path to the file.
	optMkdirAll
)

var (
	OptNone      = Option{optNone}
	OptOverwrite = Option{optOverwrite}
	OptMkdirAll  = Option{optMkdirAll}
)

// Handler handles file interaction.
type Handler struct {
	fs *afero.Afero
}

// NewHandler returns a new file handler.
func NewHandler(fs afero.Fs) Handler {
	afs := &afero.Afero{Fs: fs}
	return Handler{fs: afs}
}

// Read reads the file given name and returns the bytes read.
func (h *Handler) Read(name string) ([]byte, error) {
	file, err := h.fs.OpenFile(name, os.O_RDONLY, 0o600)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

// Write writes the data bytes into the file with the given name.
func (h *Handler) Write(name string, data []byte, options ...Option) error {
	if hasOption(options, OptMkdirAll) {
		if err := h.fs.MkdirAll(path.Dir(name), os.ModePerm); err != nil {
			return err
		}
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if hasOption(options, OptOverwrite) {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	file, err := h.fs.OpenFile(name, flags, 0o600)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if errTmp := file.Close(); errTmp != nil && err == nil {
		err = errTmp
	}
	return err
}

// ReadJSON reads a JSON file from name and unmarshals it into the content interface.
// The interface content must be a pointer to a JSON marchalable object.
func (h *Handler) ReadJSON(name string, content any) error {
	data, err := h.Read(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, content)
}

// WriteJSON marshals the content interface to JSON and writes it to the path with the given name.
func (h *Handler) WriteJSON(name string, content any, options ...Option) error {
	jsonData, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		return err
	}
	return h.Write(name, jsonData, options...)
}

// ReadYAML reads a YAML file from name and unmarshals it into the content interface.
// The interface content must be a pointer to a YAML marchalable object.
func (h *Handler) ReadYAML(name string, content any) error {
	return h.readYAML(name, content, false)
}

// ReadYAMLStrict does the same as ReadYAML, but fails if YAML contains fields
// that are not specified in content.
func (h *Handler) ReadYAMLStrict(name string, content any) error {
	return h.readYAML(name, content, true)
}

func (h *Handler) readYAML(name string, content any, strict bool) error {
	data, err := h.Read(name)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(bytes.NewBuffer(data))
	decoder.KnownFields(strict)
	return decoder.Decode(content)
}

// WriteYAML marshals the content interface to YAML and writes it to the path with the given name.
func (h *Handler) WriteYAML(name string, content any, options ...Option) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("recovered from panic")
		}
	}()
	data, err := encoder.NewEncoder(content).Encode()
	if err != nil {
		return err
	}
	return h.Write(name, data, options...)
}

// Remove deletes the file with the given name.
func (h *Handler) Remove(name string) error {
	return h.fs.Remove(name)
}

// RemoveAll deletes the file or directory with the given name.
func (h *Handler) RemoveAll(name string) error {
	return h.fs.RemoveAll(name)
}

// Stat returns a FileInfo describing the named file, or an error, if any
// happens.
func (h *Handler) Stat(name string) (fs.FileInfo, error) {
	return h.fs.Stat(name)
}
