/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package trustedlaunch

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm/tpm2"
)

const (
	tpmAkIdx     = 0x81000003
	tpmAkCertIdx = 0x1C101D0
)

// Issuer for Azure trusted launch TPM attestation.
type Issuer struct {
	oid.AzureTrustedLaunch
	*vtpm.Issuer
	hClient httpClient
}

// NewIssuer initializes a new Azure Issuer.
func NewIssuer() *Issuer {
	i := &Issuer{
		hClient: &http.Client{},
	}
	i.Issuer = vtpm.NewIssuer(
		vtpm.OpenVTPM,
		getAttestationKey,
		i.getAttestationCert,
	)
	return i
}

// akSigner holds the attestation key certificate and the corresponding CA certificate.
type akSigner struct {
	AkCert []byte
	CA     []byte
}

// getAttestationCert returns the DER encoded certificate of the TPM's attestation key and it's CA.
func (i *Issuer) getAttestationCert(tpm io.ReadWriteCloser) ([]byte, error) {
	certDER, err := tpm2.NVReadEx(tpm, tpmAkCertIdx, tpm2.HandleOwner, "", 0)
	if err != nil {
		return nil, fmt.Errorf("reading attestation key certificate from TPM: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("parsing attestation key certificate: %w", err)
	}

	// try to load the CA certificate from the issuing certificate URL
	// any error is ignored and the next URL is tried
	// if no CA certificate can be loaded, an error is returned
	var caCert *x509.Certificate
	for _, caCertURL := range cert.IssuingCertificateURL {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, caCertURL, nil)
		if err != nil {
			continue
		}
		resp, err := i.hClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			continue
		}

		caCertDER, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		caCert, err = x509.ParseCertificate(caCertDER)
		if err != nil {
			continue
		}
		break
	}
	if caCert == nil {
		return nil, errors.New("failed to load CA certificate")
	}

	signerInfo := akSigner{
		AkCert: certDER,
		CA:     caCert.Raw,
	}

	return json.Marshal(signerInfo)
}

// getAttestationKey reads the Azure trusted launch attesation key.
func getAttestationKey(tpm io.ReadWriter) (*tpmclient.Key, error) {
	ak, err := tpmclient.LoadCachedKey(tpm, tpmAkIdx)
	if err != nil {
		return nil, fmt.Errorf("reading attestation key from TPM: %w", err)
	}

	return ak, nil
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}
