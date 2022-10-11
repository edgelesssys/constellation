/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/constants"
)

type rekorVerifier interface {
	SearchByHash(context.Context, string) ([]string, error)
	VerifyEntry(context.Context, string, string) error
}

func verifyWithRekor(ctx context.Context, verifier rekorVerifier, hash string) error {
	uuids, err := verifier.SearchByHash(ctx, hash)
	if err != nil {
		return fmt.Errorf("searching Rekor for hash: %w", err)
	}

	if len(uuids) == 0 {
		return fmt.Errorf("no matching entries in Rekor")
	}

	// We expect the first entry in Rekor to be our original entry.
	// SHA256 should ensure there is no entry with the same hash.
	// Any subsequent hashes are treated as potential attacks and are ignored.
	// Attacks on Rekor will be monitored from other backend services.
	artifactUUID := uuids[0]

	return verifier.VerifyEntry(
		ctx, artifactUUID,
		base64.StdEncoding.EncodeToString([]byte(constants.CosignPublicKey)),
	)
}
