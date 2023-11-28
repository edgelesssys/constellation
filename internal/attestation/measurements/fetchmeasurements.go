/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measurements

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/internal/sigstore/keyselect"
)

// RekorError is returned when verifying measurements with Rekor fails.
type RekorError struct {
	err error
}

// Error returns the error message.
func (e *RekorError) Error() string {
	return fmt.Sprintf("verifying measurements with Rekor failed: %s", e.err)
}

// Unwrap returns the wrapped error.
func (e *RekorError) Unwrap() error {
	return e.err
}

// VerifyFetcher is a high-level fetcher that fetches measurements and verifies them.
type VerifyFetcher struct {
	client            *http.Client
	newCosignVerifier cosignVerifierConstructor
	rekor             rekorVerifier
}

// NewVerifyFetcher creates a new MeasurementFetcher.
func NewVerifyFetcher(newCosignVerifier func([]byte) (sigstore.Verifier, error), rekor rekorVerifier, client *http.Client) *VerifyFetcher {
	return &VerifyFetcher{
		newCosignVerifier: newCosignVerifier,
		rekor:             rekor,
		client:            client,
	}
}

// FetchAndVerifyMeasurements fetches and verifies measurements for the given version and attestation variant.
func (m *VerifyFetcher) FetchAndVerifyMeasurements(ctx context.Context,
	image string, csp cloudprovider.Provider, attestationVariant variant.Variant,
	noVerify bool,
) (M, error) {
	version, err := versionsapi.NewVersionFromShortPath(image, versionsapi.VersionKindImage)
	if err != nil {
		return nil, fmt.Errorf("parsing image version: %w", err)
	}
	publicKey, err := keyselect.CosignPublicKeyForVersion(version)
	if err != nil {
		return nil, fmt.Errorf("getting public key: %w", err)
	}

	cosign, err := m.newCosignVerifier(publicKey)
	if err != nil {
		return nil, fmt.Errorf("creating cosign verifier: %w", err)
	}

	measurementsURL, signatureURL, err := versionsapi.MeasurementURL(version)
	if err != nil {
		return nil, err
	}
	var fetchedMeasurements M
	if noVerify {
		if err := fetchedMeasurements.FetchNoVerify(
			ctx,
			m.client,
			measurementsURL,
			version,
			csp,
			attestationVariant,
		); err != nil {
			return nil, fmt.Errorf("fetching measurements: %w", err)
		}
	} else {
		hash, err := fetchedMeasurements.FetchAndVerify(
			ctx,
			m.client,
			cosign,
			measurementsURL,
			signatureURL,
			version,
			csp,
			attestationVariant,
		)
		if err != nil {
			return nil, fmt.Errorf("fetching and verifying measurements: %w", err)
		}
		if err := sigstore.VerifyWithRekor(ctx, publicKey, m.rekor, hash); err != nil {
			return nil, &RekorError{err: err}
		}
	}
	return fetchedMeasurements, nil
}

type cosignVerifierConstructor func([]byte) (sigstore.Verifier, error)

type rekorVerifier interface {
	SearchByHash(context.Context, string) ([]string, error)
	VerifyEntry(context.Context, string, string) error
}
