/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"errors"
	"fmt"

	"github.com/google/go-sev-guest/kds"
)

type signatureError struct {
	innerError error
}

func (e *signatureError) Unwrap() error {
	return e.innerError
}

func (e *signatureError) Error() string {
	return fmt.Sprintf("signature validation failed: %v", e.innerError)
}

type askError struct {
	innerError error
}

func (e *askError) Unwrap() error {
	return e.innerError
}

func (e *askError) Error() string {
	return fmt.Sprintf("validating ASK: %v", e.innerError)
}

type vcekError struct {
	innerError error
}

func (e *vcekError) Unwrap() error {
	return e.innerError
}

func (e *vcekError) Error() string {
	return fmt.Sprintf("validating VCEK: %v", e.innerError)
}

type idKeyError struct {
	encounteredValue []byte
	expectedValues   [][]byte
}

func (e *idKeyError) Unwrap() error {
	return nil
}

func (e *idKeyError) Error() string {
	return fmt.Sprintf("accepted idkeydigest list %x doesn't contain reported idkeydigest %x", e.expectedValues, e.encounteredValue)
}

type versionError struct {
	expectedType     string
	excpectedVersion *kds.TCBParts
}

func (e *versionError) Unwrap() error {
	return nil
}

func (e *versionError) Error() string {
	return fmt.Sprintf("invalid %s version: %x", e.expectedType, e.excpectedVersion)
}

var errDebugEnabled = errors.New("SNP report indicates debugging, expected no debugging")
