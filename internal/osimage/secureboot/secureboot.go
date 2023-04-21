/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package secureboot holds secure boot configuration for image uploads.
package secureboot

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"

	"github.com/spf13/afero"
)

// Database holds the secure boot database that cloud providers should
// use when enabling secure boot for a Constellation OS image.
type Database struct {
	// PK is the platform key.
	PK []byte
	// Keks are trusted key-exchange-keys
	Keks [][]byte
	// DBs are entries of the signature database.
	DBs [][]byte
}

// DatabaseFromFiles creates the secure boot database from individual files.
func DatabaseFromFiles(fs afero.Fs, pk string, keks []string, dbs []string) (Database, error) {
	rawPK, err := afero.ReadFile(fs, pk)
	if err != nil {
		return Database{}, fmt.Errorf("loading PK %s: %w", pk, err)
	}
	rawKEKs := make([][]byte, len(keks))
	for i, kek := range keks {
		rawKEK, err := afero.ReadFile(fs, kek)
		if err != nil {
			return Database{}, fmt.Errorf("loading KEK %s: %w", kek, err)
		}
		rawKEKs[i] = rawKEK
	}
	rawDBs := make([][]byte, len(dbs))
	for i, db := range dbs {
		rawDB, err := afero.ReadFile(fs, db)
		if err != nil {
			return Database{}, fmt.Errorf("loading DB %s: %w", db, err)
		}
		rawDBs[i] = rawDB
	}
	return Database{
		PK:   rawPK,
		Keks: rawKEKs,
		DBs:  rawDBs,
	}, nil
}

// UEFIVarStore is a UEFI variable store.
// It is a collection of UEFIVar structs.
// This is an abstract var store that can convert to a concrete var store
// for a specific CSP.
type UEFIVarStore []UEFIVar

// VarStoreFromFiles creates the UEFI variable store
// from "EFI Signature List" (esl) files.
func VarStoreFromFiles(fs afero.Fs, pk, kek, db, dbx string) (UEFIVarStore, error) {
	vars := UEFIVarStore{}
	pkF, err := fs.OpenFile(pk, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return UEFIVarStore{}, fmt.Errorf("opening PK ESL %s: %w", pk, err)
	}
	defer pkF.Close()
	pkVar, err := ReadVar(pkF, "PK", globalEFIGUID)
	if err != nil {
		return UEFIVarStore{}, fmt.Errorf("reading PK ESL %s: %w", pk, err)
	}
	vars = append(vars, pkVar)
	kekF, err := fs.OpenFile(kek, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return UEFIVarStore{}, fmt.Errorf("opening KEK ESL %s: %w", kek, err)
	}
	defer kekF.Close()
	kekVar, err := ReadVar(kekF, "KEK", globalEFIGUID)
	if err != nil {
		return UEFIVarStore{}, fmt.Errorf("reading KEK ESL %s: %w", kek, err)
	}
	vars = append(vars, kekVar)
	dbF, err := fs.OpenFile(db, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return UEFIVarStore{}, fmt.Errorf("opening DB ESL %s: %w", db, err)
	}
	defer dbF.Close()
	dbVar, err := ReadVar(dbF, "db", secureDatabaseGUID)
	if err != nil {
		return UEFIVarStore{}, fmt.Errorf("reading DB ESL %s: %w", db, err)
	}
	vars = append(vars, dbVar)
	if len(dbx) == 0 {
		return vars, nil
	}
	dbxF, err := fs.OpenFile(dbx, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return UEFIVarStore{}, fmt.Errorf("opening DBX ESL %s: %w", dbx, err)
	}
	defer dbxF.Close()
	dbxVar, err := ReadVar(dbxF, "dbx", secureDatabaseGUID)
	if err != nil {
		return UEFIVarStore{}, fmt.Errorf("reading DBX ESL %s: %w", dbx, err)
	}
	vars = append(vars, dbxVar)
	return vars, nil
}

// ToAWS converts the UEFI variable store to the AWS UEFI vars v0 format.
// The format is documented here:
// https://github.com/awslabs/python-uefivars
// It is structured as follows:
// Header:
// - 4 bytes: magic number
// - 4 bytes: crc32 of the rest of the file
// - 4 bytes: version number
//
// Body is zlib compressed stream of:
// 8 bytes number of entries
// for each entry:
// - name (variable length field, utf8)
// - data (variable length field)
// - guid (16 bytes)
// - attr (int32 in little endian)
// OPTIONAL (if attr has EFI_VARIABLE_TIME_BASED_AUTHENTICATED_WRITE_ACCESS set):
// - timestamp (16 bytes)
// - digest (variable length field).
func (s UEFIVarStore) ToAWS() (string, error) {
	payload := bytes.Buffer{}
	// Write the number of entries.
	if err := binary.Write(&payload, binary.LittleEndian, uint64(len(s))); err != nil {
		return "", fmt.Errorf("writing number of entries: %w", err)
	}
	// Write the entries.
	for _, entry := range s {
		rawEntry, err := entry.AWSEntry()
		if err != nil {
			return "", fmt.Errorf("serializing entry: %w", err)
		}
		if _, err := payload.Write(rawEntry); err != nil {
			return "", fmt.Errorf("writing entry: %w", err)
		}
	}
	// Compress the payload.
	compressed := bytes.Buffer{}
	zlibW, err := zlib.NewWriterLevelDict(&compressed, zlib.BestCompression, zlibDict)
	if err != nil {
		return "", fmt.Errorf("creating compressor: %w", err)
	}
	if _, err := zlibW.Write(payload.Bytes()); err != nil {
		return "", fmt.Errorf("compressing payload: %w", err)
	}
	if err := zlibW.Close(); err != nil {
		return "", fmt.Errorf("closing compressor: %w", err)
	}
	compressedData := compressed.Bytes()
	// Calculate the CRC32 (Castagnoli) of the version + compressed payload.
	crcData := append(awsVersion, compressedData...)
	crc := crc32.Checksum(crcData, crc32.MakeTable(crc32.Castagnoli))
	out := bytes.Buffer{}
	// Write the header.
	if _, err := out.Write(awsMagic); err != nil {
		return "", fmt.Errorf("writing magic: %w", err)
	}
	if err := binary.Write(&out, binary.LittleEndian, crc); err != nil {
		return "", fmt.Errorf("writing crc: %w", err)
	}
	// Write the version + compressed payload.
	if _, err := out.Write(crcData); err != nil {
		return "", fmt.Errorf("writing compressed payload: %w", err)
	}
	return base64.StdEncoding.EncodeToString(out.Bytes()), nil
}

// UEFIVar is a UEFI variable.
type UEFIVar struct {
	Name      string
	Data      []byte
	GUID      []byte
	Attr      uint32
	Timestamp []byte
	Digest    []byte
}

// ReadVar reads a UEFI variable from an ESL file.
func ReadVar(reader io.Reader, name string, guid []byte) (UEFIVar, error) {
	attr := uint32(
		EFIVariableNonVolatile |
			EFIVariableBootServiceAccess |
			EFIVariableRuntimeAccess |
			EFIVariableTimeBasedAuthenticatedWriteAccess,
	)
	data, err := io.ReadAll(reader)
	if err != nil {
		return UEFIVar{}, err
	}
	return UEFIVar{
		Name: name,
		Data: data,
		GUID: guid,
		Attr: attr,
		Timestamp: []byte{
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		Digest: []byte{
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
	}, nil
}

// AWSEntry returns the AWS format entry for the UEFI variable.
func (v UEFIVar) AWSEntry() ([]byte, error) {
	var buf bytes.Buffer
	if err := appendVariableLengthField(&buf, []byte(v.Name)); err != nil {
		return nil, err
	}
	if err := appendVariableLengthField(&buf, v.Data); err != nil {
		return nil, err
	}
	if _, err := buf.Write(v.GUID); err != nil {
		return nil, err
	}
	if err := appendAttr(&buf, v.Attr); err != nil {
		return nil, err
	}
	if v.Attr&EFIVariableTimeBasedAuthenticatedWriteAccess == 0 {
		return buf.Bytes(), nil
	}
	if _, err := buf.Write(v.Timestamp); err != nil {
		return nil, err
	}
	if err := appendVariableLengthField(&buf, v.Digest); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func appendVariableLengthField(w io.Writer, data []byte) error {
	// variable length is encoded as unsigned, 64 bit little endian
	// followed by the data
	if err := binary.Write(w, binary.LittleEndian, uint64(len(data))); err != nil {
		return err
	}
	_, err := w.Write(data)
	return err
}

func appendAttr(w io.Writer, attr uint32) error {
	return binary.Write(w, binary.LittleEndian, attr)
}

// EFI constants.
const (
	EFIVariableNonVolatile                       = 0x00000001
	EFIVariableBootServiceAccess                 = 0x00000002
	EFIVariableRuntimeAccess                     = 0x00000004
	EFIVariableTimeBasedAuthenticatedWriteAccess = 0x00000020
)

var (
	awsMagic      = []byte("AMZNUEFI")
	awsVersion    = []byte{0, 0, 0, 0}
	globalEFIGUID = []byte{
		0x61, 0xdf, 0xe4, 0x8b, 0xca, 0x93, 0xd2, 0x11,
		0xaa, 0x0d, 0x00, 0xe0, 0x98, 0x03, 0x2b, 0x8c,
	}
	secureDatabaseGUID = []byte{
		0xcb, 0xb2, 0x19, 0xd7, 0x3a, 0x3d, 0x96, 0x45,
		0xa3, 0xbc, 0xda, 0xd0, 0x0e, 0x67, 0x65, 0x6f,
	}
)
