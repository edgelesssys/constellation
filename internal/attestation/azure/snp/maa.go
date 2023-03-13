/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/edgelesssys/go-azguestattestation/maa"
	"github.com/google/go-tpm-tools/client"
	ptpm "github.com/google/go-tpm-tools/proto/tpm"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

type maaClient struct {
	client *http.Client
}

func newMAAClient() *maaClient {
	return &maaClient{
		client: &http.Client{Transport: &http.Transport{Proxy: nil}},
	}
}

func (m *maaClient) createToken(
	ctx context.Context, tpm io.ReadWriter, maaURL string,
	data, hclReport, vcekCert, vcekChain []byte,
) (string, error) {
	attestation, err := tpmAttest(tpm, data)
	if err != nil {
		return "", fmt.Errorf("creating TPM attestation: %w", err)
	}

	encKey, encKeyCert, encKeyCertSig, err := tpmGetEncryptionKey(tpm, attestation.Quotes)
	if err != nil {
		return "", fmt.Errorf("getting encryption key: %w", err)
	}

	params := maa.Parameters{
		SNPReport:         cutSNPReport(hclReport),
		RuntimeData:       cutRuntimeData(hclReport),
		VcekCert:          vcekCert,
		VcekChain:         vcekChain,
		Attestation:       attestation,
		EncKey:            encKey,
		EncKeyCertInfo:    encKeyCert,
		EncKeyCertInfoSig: encKeyCertSig,
	}

	tokenEnc, err := maa.GetEncryptedToken(ctx, params, data, maaURL, m.client)
	if err != nil {
		return "", fmt.Errorf("getting encrypted token: %w", err)
	}

	token, err := maa.DecryptToken(tokenEnc, tpm)
	if err != nil {
		return "", fmt.Errorf("decrypting token: %w", err)
	}

	return token, nil
}

func (m *maaClient) validateToken(ctx context.Context, maaURL, token string, extraData []byte) error {
	keySet, err := maa.GetKeySet(ctx, maaURL, m.client)
	if err != nil {
		return fmt.Errorf("getting key set from MAA: %w", err)
	}
	claims, err := maa.ValidateToken(token, keySet)
	if err != nil {
		return fmt.Errorf("validating token: %w", err)
	}
	return m.validateClaims(claims, extraData)
}

func (m *maaClient) validateClaims(claims map[string]interface{}, extraData []byte) error {
	runtime, ok := claims["x-ms-runtime"].(map[string]interface{})
	if !ok {
		return errors.New("invalid claims: missing x-ms-runtime")
	}
	payload, ok := runtime["client-payload"].(map[string]interface{})
	if !ok {
		return errors.New("invalid claims: missing client-payload")
	}
	nonce, ok := payload["nonce"].(string)
	if !ok {
		return errors.New("invalid claims: missing nonce")
	}

	if nonce != base64.StdEncoding.EncodeToString(extraData) {
		return errors.New("invalid claims: nonce does not match extra data")
	}
	return nil
}

func tpmAttest(tpm io.ReadWriter, nonce []byte) (*maa.Attestation, error) {
	cert, err := tpm2.NVReadEx(tpm, tpmAkCertIdx, tpm2.HandleOwner, "", 0)
	if err != nil {
		return nil, err
	}
	key, err := client.LoadCachedKey(tpm, tpmAkIdx)
	if err != nil {
		return nil, err
	}
	defer key.Close()
	attestation, err := key.Attest(client.AttestOpts{Nonce: nonce})
	if err != nil {
		return nil, err
	}
	attestation.AkCert = cert
	return attestation, nil
}

func tpmGetEncryptionKey(tpm io.ReadWriter, quotes []*maa.Quote) ([]byte, []byte, []byte, error) {
	quoteIdx, err := getSHA256QuoteIndex(quotes)
	if err != nil {
		return nil, nil, nil, err
	}
	pcrDigest, sel, err := getSHA256PCRDigest(quotes[quoteIdx].Pcrs.Pcrs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getting PCR digest: %w", err)
	}

	template, err := getEncryptionKeyTemplate(tpm, pcrDigest, sel)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getting key template: %w", err)
	}

	handle, pubKey, _, _, _, _, err := tpm2.CreatePrimaryEx(tpm, tpm2.HandleNull, sel, "", "", template)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating key: %w", err)
	}
	defer flushContext(tpm, handle)

	certifyInfo, signature, err := tpm2.Certify(tpm, "", "", handle, tpmAkIdx, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("certifying key: %w", err)
	}

	pubKey, err = tpmMarshal(pubKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("marshaling key: %w", err)
	}

	// signature is TPMT_SIGNATURE, MAA wants TPM2B_PUBLIC_KEY_RSA.buffer, which starts at offset 6
	return pubKey, certifyInfo, signature[6:], nil
}

func tpmMarshal(data []byte) ([]byte, error) {
	tpmBytes := tpmutil.U16Bytes(data)
	buf := bytes.Buffer{}
	if err := tpmBytes.TPMMarshal(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func getSHA256QuoteIndex(quotes []*maa.Quote) (int, error) {
	for idx, quote := range quotes {
		if quote.GetPcrs().GetHash() == ptpm.HashAlgo_SHA256 {
			return idx, nil
		}
	}
	return 0, errors.New("attestation did not include SHA256 hashed PCRs")
}

func getEncryptionKeyTemplate(tpm io.ReadWriter, pcrDigest []byte, sel tpm2.PCRSelection) (tpm2.Public, error) {
	session, _, err := tpm2.StartAuthSession(tpm, tpm2.HandleNull, tpm2.HandleNull, make([]byte, 32), nil, tpm2.SessionTrial, tpm2.AlgNull, tpm2.AlgSHA256)
	if err != nil {
		return tpm2.Public{}, fmt.Errorf("starting session: %w", err)
	}
	defer flushContext(tpm, session)
	if err := tpm2.PolicyPCR(tpm, session, pcrDigest, sel); err != nil {
		return tpm2.Public{}, fmt.Errorf("setting PCR policy: %w", err)
	}
	policyDigest, err := tpm2.PolicyGetDigest(tpm, session)
	if err != nil {
		return tpm2.Public{}, fmt.Errorf("getting policy digest: %w", err)
	}
	return tpm2.Public{
		Type:       tpm2.AlgRSA,
		NameAlg:    tpm2.AlgSHA256,
		Attributes: tpm2.FlagDecrypt | tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin | tpm2.FlagNoDA,
		AuthPolicy: policyDigest,
		RSAParameters: &tpm2.RSAParams{
			KeyBits: 2048,
		},
	}, nil
}

func getSHA256PCRDigest(pcrs map[uint32][]byte) ([]byte, tpm2.PCRSelection, error) {
	sel := tpm2.PCRSelection{Hash: tpm2.AlgSHA256}
	for k := range pcrs {
		sel.PCRs = append(sel.PCRs, int(k))
	}
	sort.Ints(sel.PCRs)
	hasher := sha256.New()
	for _, k := range sel.PCRs {
		if _, err := hasher.Write(pcrs[uint32(k)]); err != nil {
			return nil, tpm2.PCRSelection{}, err
		}
	}
	return hasher.Sum(nil), sel, nil
}

func flushContext(tpm io.ReadWriter, handle tpmutil.Handle) {
	_ = tpm2.FlushContext(tpm, handle)
}
