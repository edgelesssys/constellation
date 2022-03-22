/*
Package file provides functions that combine file handling, JSON marshaling
and file system abstraction.
*/
package file

import (
	"encoding/json"
	"io"
	"io/fs"
	"os"

	"github.com/spf13/afero"
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
	file, err := h.fs.OpenFile(name, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

// Write writes the data bytes into the file with the given name.
// If a file already exists at path and overwrite is true, the file will be
// overwritten. Otherwise, an error is returned.
func (h *Handler) Write(name string, data []byte, overwrite bool) error {
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if overwrite {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	file, err := h.fs.OpenFile(name, flags, 0o644)
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
func (h *Handler) ReadJSON(name string, content interface{}) error {
	data, err := h.Read(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, content)
}

// WriteJSON marshals the content interface to JSON and writes it to the path with the given name.
// If a file already exists and overwrite is true, the file will be
// overwritten. Otherwise, an error is returned.
func (h *Handler) WriteJSON(name string, content interface{}, overwrite bool) error {
	jsonData, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		return err
	}
	return h.Write(name, jsonData, overwrite)
}

// Remove deletes the file with the given name.
func (h *Handler) Remove(name string) error {
	return h.fs.Remove(name)
}

// Stat returns a FileInfo describing the named file, or an error, if any
// happens.
func (h *Handler) Stat(name string) (fs.FileInfo, error) {
	return h.fs.Stat(name)
}
