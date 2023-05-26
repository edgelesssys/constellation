/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sigstore

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	versionsapi "github.com/edgelesssys/constellation/v2/internal/api/versions"
	"github.com/sigstore/rekor/pkg/client"
	genclient "github.com/sigstore/rekor/pkg/generated/client"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/client/index"
	"github.com/sigstore/rekor/pkg/generated/models"
	hashedrekord "github.com/sigstore/rekor/pkg/types/hashedrekord/v0.0.1"
	"github.com/sigstore/rekor/pkg/verify"
	"github.com/sigstore/sigstore/pkg/signature"
)

// VerifyWithRekor checks if the hash of a signature is present in Rekor.
func VerifyWithRekor(ctx context.Context, version versionsapi.Version, verifier rekorVerifier, hash string) error {
	publicKey, err := CosignPublicKeyForVersion(version)
	if err != nil {
		return fmt.Errorf("getting public key: %w", err)
	}

	uuids, err := verifier.SearchByHash(ctx, hash)
	if err != nil {
		return fmt.Errorf("searching Rekor for hash: %w", err)
	}

	if len(uuids) == 0 {
		return errors.New("no matching entries in Rekor")
	}

	// We expect the first entry in Rekor to be our original entry.
	// SHA256 should ensure there is no entry with the same hash.
	// Any subsequent hashes are treated as potential attacks and are ignored.
	// Attacks on Rekor will be monitored from other backend services.
	artifactUUID := uuids[0]

	return verifier.VerifyEntry(
		ctx, artifactUUID,
		base64.StdEncoding.EncodeToString(publicKey),
	)
}

// Rekor allows to interact with the transparency log at:
// https://rekor.sigstore.dev
// For more information see Rekor's Swagger definition:
// https://www.sigstore.dev/swagger/#/
type Rekor struct {
	client *genclient.Rekor
}

// NewRekor creates a new instance of Rekor to interact with the transparency
// log at: https://rekor.sigstore.dev
func NewRekor() (*Rekor, error) {
	client, err := client.GetRekorClient("https://rekor.sigstore.dev")
	if err != nil {
		return nil, err
	}

	return &Rekor{
		client: client,
	}, nil
}

// SearchByHash searches for the hash of an artifact in Rekor transparency log.
// A list of UUIDs will be returned, since multiple entries could be present for
// a single artifact in Rekor.
func (r *Rekor) SearchByHash(ctx context.Context, hash string) ([]string, error) {
	params := index.NewSearchIndexParamsWithContext(ctx)
	params.SetQuery(
		&models.SearchIndex{
			Hash: hash,
		},
	)

	index, err := r.client.Index.SearchIndex(params)
	if err != nil {
		return nil, fmt.Errorf("unable to search index: %w", err)
	}
	if !index.IsSuccess() {
		return nil, fmt.Errorf("search failed: %s", index.Error())
	}

	return index.GetPayload(), nil
}

// VerifyEntry performs log entry verification (see verifyLogEntry) and
// verifies that the provided publicKey was used to sign the entry.
// An error is returned if any verification fails.
func (r *Rekor) VerifyEntry(ctx context.Context, uuid, publicKey string) error {
	entry, err := r.getEntry(ctx, uuid)
	if err != nil {
		return err
	}

	err = r.verifyLogEntry(ctx, entry)
	if err != nil {
		return err
	}

	rekord, err := hashedRekordFromEntry(entry)
	if err != nil {
		return fmt.Errorf("extracting rekord from Rekor entry: %w", err)
	}

	if !isEntrySignedBy(rekord, publicKey) {
		return errors.New("rekord signed by unknown key")
	}

	return nil
}

// getEntry downloads entry for the provided UUID.
func (r *Rekor) getEntry(ctx context.Context, uuid string) (models.LogEntryAnon, error) {
	params := entries.NewGetLogEntryByUUIDParamsWithContext(ctx)
	params.SetEntryUUID(uuid)

	entry, err := r.client.Entries.GetLogEntryByUUID(params)
	if err != nil {
		return models.LogEntryAnon{}, fmt.Errorf("error getting entry: %w", err)
	}
	if !entry.IsSuccess() {
		return models.LogEntryAnon{}, fmt.Errorf("entries failure: %s", entry.Error())
	}

	if entires := len(entry.GetPayload()); entires != 1 {
		return models.LogEntryAnon{}, fmt.Errorf("excepted 1 entry, but rekor returned %d", entires)
	}

	for key := range entry.Payload {
		return entry.Payload[key], nil
	}

	return models.LogEntryAnon{}, fmt.Errorf("no entry returned")
}

// verifyLogEntry performs inclusion proof verification, SignedEntryTimestamp
// verification, and checkpoint verification of the provided entry in Rekor.
// A return value of nil indicates successful verification.
func (r *Rekor) verifyLogEntry(ctx context.Context, entry models.LogEntryAnon) error {
	keyResp, err := r.client.Pubkey.GetPublicKey(nil)
	if err != nil {
		return err
	}
	publicKey := keyResp.Payload

	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return errors.New("failed to decode key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	verifier, err := signature.LoadVerifier(pub, crypto.SHA256)
	if err != nil {
		return err
	}

	err = verify.VerifyLogEntry(ctx, &entry, verifier)
	if err != nil {
		return err
	}

	return nil
}

// hashedRekordFromEntry extracts the base64 encoded polymorphic Body field
// and unmarshals the contained JSON into the correct type.
func hashedRekordFromEntry(entry models.LogEntryAnon) (*hashedrekord.V001Entry, error) {
	var rekord models.Hashedrekord

	body, ok := entry.Body.(string)
	if !ok {
		return nil, errors.New("body is not a string")
	}
	decoded, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(bytes.NewReader(decoded)).Decode(&rekord)
	if err != nil {
		return nil, err
	}

	hashedRekord := &hashedrekord.V001Entry{}
	if err := hashedRekord.Unmarshal(&rekord); err != nil {
		return nil, errors.New("failed to unmarshal entry")
	}

	return hashedRekord, nil
}

// isEntrySignedBy checks whether rekord was signed with provided publicKey.
func isEntrySignedBy(rekord *hashedrekord.V001Entry, publicKey string) bool {
	if rekord == nil {
		return false
	}
	if rekord.HashedRekordObj.Signature == nil {
		return false
	}
	if rekord.HashedRekordObj.Signature.PublicKey == nil {
		return false
	}

	actualKey := rekord.HashedRekordObj.Signature.PublicKey.Content.String()
	return actualKey == publicKey
}

type rekorVerifier interface {
	SearchByHash(context.Context, string) ([]string, error)
	VerifyEntry(context.Context, string, string) error
}
