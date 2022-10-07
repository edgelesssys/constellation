package cmd

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/constants"
)

type RekorVerifier interface {
	SearchByHash(context.Context, string) ([]string, error)
	VerifyEntry(context.Context, string, string) error
}

func verifyWithRekor(ctx context.Context, verifier RekorVerifier, hash string) error {
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

	return verifier.VerifyEntry(ctx, artifactUUID, constants.CosignPublicKeyBase64())
}
