/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package trustedlaunch

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	certutil "github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm/tpm2"
)

// ameRoot is the AME root CA certificate used to sign Azure's AME Infra CA certificates.
// The certificate can be found at http://crl.microsoft.com/pkiinfra/certs/AMERoot_ameroot.crt.
var ameRoot = mustParseX509("-----BEGIN CERTIFICATE-----\nMIIFVjCCAz6gAwIBAgIQJdrLVcnGd4FAnlaUgt5N/jANBgkqhkiG9w0BAQsFADA8\nMRMwEQYKCZImiZPyLGQBGRYDR0JMMRMwEQYKCZImiZPyLGQBGRYDQU1FMRAwDgYD\nVQQDEwdhbWVyb290MB4XDTE2MDUyNDIyNTI1NFoXDTI2MDUyNDIyNTcwM1owPDET\nMBEGCgmSJomT8ixkARkWA0dCTDETMBEGCgmSJomT8ixkARkWA0FNRTEQMA4GA1UE\nAxMHYW1lcm9vdDCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBALv4uChY\noVuO+bxBOcn8v4FajoGkxo0YgVwEqEPDVPI6vzmnEqHVhQ1GMVeDyiRrgQT1vCk1\nHMMzo9LlWowPrzbXOwjOTFbXc36+UU41yNN2GeNa49RXbAkfbzKE/SYLfbqOD0dN\nZLwvOhgIb25oA1eAxW/DI/hvJLLKh2SscvkIyd3o2BUeFm7NtyYG/buCKJh8lOq8\n0iBwRoEoInb0vhorHaswSMmqY1g+AJndY/M7uGUqkhDGBhLu53bU9wbUPHsEI+wa\nq6WypCijZYT+C4BS5GJrEPZ2O92pztd+ULqhzNRoPj5RuElUww7+z5RnbCaupyBY\nOmmJMH30EiRSq8dK/irixXXwJraSywR5kyfmAkv6GYWlRlxFUiK3/co47JLA3TDK\nN0wfutbpqxdZQYyGfO2nZrr5JbKfSU0sMtOZDkK6hlafV++hfkVSvFfNHE5B5uN1\nMK6agl1dzi28HfJT9aO7cmjGxl1SJ5qoCvcwZNQ2SPHFdrslcwXEFOMDaEzVOA3V\n7j3+6lrT8sHXg0sErkcd8lrBImfzhLxM/Wh8CgOUNeUu3flUoxmFv3el+QWalSNy\n2SXs2NgWuYE5Iog7CHD/xCnoEnZwwjqLkrro4hYWE4Xj3VlA2Eq+VxqJOgdyFl3m\nckSZ08OcwLeprY4+2GEvCXNGNdXUmNNgk2PvAgMBAAGjVDBSMAsGA1UdDwQEAwIB\nhjASBgNVHRMBAf8ECDAGAQH/AgEBMB0GA1UdDgQWBBQpXlFeZK40ueusnA2njHUB\n0QkLKDAQBgkrBgEEAYI3FQEEAwIBADANBgkqhkiG9w0BAQsFAAOCAgEAcznFDnJx\nsXaazFY1DuIPvUaiWS7ELxAVXMGZ7ROjLrDq1FNYVewL4emDqyEIEMFncec8rqyk\nVBvLQA5YqMCxQWJpL0SlgRSknzLh9ZVcQw1TshC49/XV2N/CLOuyInEQwS//46so\nT20Cf8UGUiOK472LZlvM4KchyDR3FTNtmMg0B/LKVjevpX9sk5MiyjjLUj3jtPIP\n7jpsfZDd/BNsg/89kpsIF5O64I7iYFj3MHu9o4UJcEX0hRt7OzUxqa9THTssvzE5\nVkWo8Rtou2T5TobKV6Rr5Ob9wchLXqVtCyZF16voEKheBnalhGUvErI/6VtBwLb7\n13C0JkKLBNMen+HClNliicVIaubnpY2g+AqxOgKBHiZnzq2HhE1qqEUf4VfqahNU\niaXtbtyo54f2dCf9UL9uG9dllN3nxBE/Y/aWF6E1M8Bslj1aYAtfUQ/xlhEXCly6\nzohw697i3XFUt76RwvfW8quvqdH9Mx0PBpYo4wJJRwAecSJQNy6wIJhAuDgOemXJ\nYViBi/bDnhPcFEVQxsypQSw91BUw7Mxh+W59H5MC25SAIw9fLMT9LRqSYpPyasNp\n4nACjR+bv/6cI+ICOrGmD2mrk2c4dNnYpDx96FfX/Y158RV0wotqIglACk6m1qyo\nyTra6P0Kvo6xz4KaVm8F7VDzUP+heAAhPAs=\n-----END CERTIFICATE-----\n")

// Validator for Azure trusted launch VM attestation.
type Validator struct {
	variant.AzureTrustedLaunch
	*vtpm.Validator
	roots *x509.CertPool
}

// NewValidator initializes a new Azure validator with the provided PCR values.
func NewValidator(cfg *config.AzureTrustedLaunch, log attestation.Logger) *Validator {
	rootPool := x509.NewCertPool()
	rootPool.AddCert(ameRoot)
	v := &Validator{roots: rootPool}
	v.Validator = vtpm.NewValidator(
		cfg.Measurements,
		v.verifyAttestationKey,
		validateVM,
		log,
	)
	return v
}

// verifyAttestationKey establishes trust in an attestation key.
// It does so by verifying the certificate chain of the attestation key certificate.
func (v *Validator) verifyAttestationKey(_ context.Context, attDoc vtpm.AttestationDocument, _ []byte) (crypto.PublicKey, error) {
	pubArea, err := tpm2.DecodePublic(attDoc.Attestation.AkPub)
	if err != nil {
		return nil, fmt.Errorf("decoding attestation key public area: %w", err)
	}

	var akSigner akSigner
	if err := json.Unmarshal(attDoc.InstanceInfo, &akSigner); err != nil {
		return nil, fmt.Errorf("unmarshaling attestation key signer info: %w", err)
	}

	akCert, err := x509.ParseCertificate(akSigner.AkCert)
	if err != nil {
		return nil, fmt.Errorf("parsing attestation key certificate: %w", err)
	}
	akCertCA, err := x509.ParseCertificate(akSigner.CA)
	if err != nil {
		return nil, fmt.Errorf("parsing attestation key CA certificate: %w", err)
	}

	intermediates := x509.NewCertPool()
	intermediates.AddCert(akCertCA)

	if _, err := akCert.Verify(x509.VerifyOptions{
		Roots:         v.roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}); err != nil {
		return nil, fmt.Errorf("verifying attestation key certificate: %w", err)
	}

	pubKey, err := pubArea.Key()
	if err != nil {
		return nil, fmt.Errorf("getting public key: %w", err)
	}

	pubKeyRSA, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("attestation key is not an RSA key")
	}

	if !pubKeyRSA.Equal(akCert.PublicKey) {
		return nil, errors.New("certificate public key does not match attestation key")
	}

	return pubKeyRSA, nil
}

// validateVM returns nil.
func validateVM(vtpm.AttestationDocument, *attest.MachineState) error {
	return nil
}

func mustParseX509(pem string) *x509.Certificate {
	cert, err := certutil.PemToX509Cert([]byte(pem))
	if err != nil {
		panic(err)
	}
	return cert
}
