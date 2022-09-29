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
	"log"

	"github.com/sigstore/rekor/pkg/client"
	genclient "github.com/sigstore/rekor/pkg/generated/client"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/client/index"
	"github.com/sigstore/rekor/pkg/generated/models"
	hashedrekord "github.com/sigstore/rekor/pkg/types/hashedrekord/v0.0.1"
	"github.com/sigstore/rekor/pkg/verify"
	"github.com/sigstore/sigstore/pkg/signature"
)

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

// GetEntry downloads entries for the provided UUID.
func (r *Rekor) GetEntry(ctx context.Context, uuid string) (models.LogEntryAnon, error) {
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

func (r *Rekor) VerifyEntry(ctx context.Context, entry models.LogEntryAnon) (bool, error) {
	keyResp, err := r.client.Pubkey.GetPublicKey(nil)
	if err != nil {
		return false, err
	}
	publicKey := keyResp.Payload

	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return false, errors.New("failed to decode key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, err
	}

	verifier, err := signature.LoadVerifier(pub, crypto.SHA256)
	if err != nil {
		return false, err
	}

	err = verify.VerifyLogEntry(ctx, &entry, verifier)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Rekor) GetAndVerifyEntry(ctx context.Context, uuid string) (models.LogEntryAnon, bool, error) {
	entry, err := r.GetEntry(ctx, uuid)
	if err != nil {
		return models.LogEntryAnon{}, false, err
	}

	valid, err := r.VerifyEntry(ctx, entry)
	if err != nil {
		return models.LogEntryAnon{}, valid, err
	}

	return entry, valid, nil
}

func (r *Rekor) HashedRekordFromEntry(entry models.LogEntryAnon) (*hashedrekord.V001Entry, error) {
	var rekord models.Hashedrekord

	decoded, err := base64.StdEncoding.DecodeString(entry.Body.(string))
	if err != nil {
		return nil, err
	}

	log.Printf("%s\n", decoded)

	err = json.NewDecoder(bytes.NewReader(decoded)).Decode(&rekord)
	if err != nil {
		return nil, err
	}

	hashedRekord := &hashedrekord.V001Entry{}
	if err := hashedRekord.Unmarshal(&rekord); err != nil {
		return nil, errors.New("lkasjd")
	}

	return hashedRekord, nil
}
