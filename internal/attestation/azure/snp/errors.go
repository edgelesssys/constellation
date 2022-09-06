/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"errors"
	"fmt"
)

type errSignature struct {
	innerError error
}

func (e *errSignature) Unwrap() error {
	return e.innerError
}

func (e *errSignature) Error() string {
	return fmt.Sprintf("signature validation failed: %v", e.innerError)
}

type errASK struct {
	innerError error
}

func (e *errASK) Unwrap() error {
	return e.innerError
}

func (e *errASK) Error() string {
	return fmt.Sprintf("validating ASK: %v", e.innerError)
}

type errVCEK struct {
	innerError error
}

func (e *errVCEK) Unwrap() error {
	return e.innerError
}

func (e *errVCEK) Error() string {
	return fmt.Sprintf("validating VCEK: %v", e.innerError)
}

type errIDKey struct {
	expectedValue []byte
}

func (e *errIDKey) Unwrap() error {
	return nil
}

func (e *errIDKey) Error() string {
	return fmt.Sprintf("configured idkeydigest does not match reported idkeydigest: %x", e.expectedValue)
}

type errVersion struct {
	expectedType     string
	excpectedVersion tcbVersion
}

func (e *errVersion) Unwrap() error {
	return nil
}

func (e *errVersion) Error() string {
	return fmt.Sprintf("invalid %s version: %x", e.expectedType, e.excpectedVersion)
}

var errDebugEnabled = errors.New("SNP report indicates debugging, expected no debugging")
