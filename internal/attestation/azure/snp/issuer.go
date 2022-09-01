package snp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/oid"
	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm/tpm2"
)

const (
	lenHclHeader                   = 0x20
	lenSnpReport                   = 0x4a0
	lenSnpReportRuntimeDataPadding = 0x14
	tpmReportIdx                   = 0x01400001
)

// GetIdKeyDigest reads the idkeydigest from the snp report saved in the TPM's non-volatile memory.
func GetIdKeyDigest(open vtpm.TPMOpenFunc) ([]byte, error) {
	tpm, err := open()
	if err != nil {
		return nil, err
	}
	defer tpm.Close()

	reportRaw, err := tpm2.NVReadEx(tpm, tpmReportIdx, tpm2.HandleOwner, "", 0)
	if err != nil {
		return nil, fmt.Errorf("reading idx %x from TMP: %w", tpmReportIdx, err)
	}

	report, err := newSNPReportFromBytes(reportRaw[lenHclHeader:])
	if err != nil {
		return nil, fmt.Errorf("creating snp report: %w", err)
	}

	return report.IdKeyDigest[:], nil
}

// Issuer for Azure TPM attestation.
type Issuer struct {
	oid.AzureSNP
	*vtpm.Issuer
}

// NewIssuer initializes a new Azure Issuer.
func NewIssuer() *Issuer {
	imdsAPI := imdsClient{
		client: &http.Client{Transport: &http.Transport{Proxy: nil}},
	}

	return &Issuer{
		Issuer: vtpm.NewIssuer(
			vtpm.OpenVTPM,
			getAttestationKey,
			getInstanceInfo(&tpmReport{}, imdsAPI),
		),
	}
}

// getInstanceInfo loads and returns the SEV-SNP attestation report [1] and the
// AMD VCEK certificate chain.
// The attestation report is loaded from the TPM, the certificate chain is queried
// from the cloud metadata API.
// [1] https://github.com/AMDESE/sev-guest/blob/main/include/attestation.h
func getInstanceInfo(reportGetter tpmReportGetter, imdsAPI imdsApi) func(tpm io.ReadWriteCloser) ([]byte, error) {
	return func(tpm io.ReadWriteCloser) ([]byte, error) {
		hclReport, err := reportGetter.get(tpm)
		if err != nil {
			return nil, fmt.Errorf("reading report from TPM: %w", err)
		}
		if len(hclReport) < lenHclHeader+lenSnpReport+lenSnpReportRuntimeDataPadding {
			return nil, fmt.Errorf("report read from TPM is shorter then expected: %x", hclReport)
		}
		hclReport = hclReport[lenHclHeader:]

		runtimeData, _, _ := bytes.Cut(hclReport[lenSnpReport+lenSnpReportRuntimeDataPadding:], []byte{0})

		vcekResponse, err := imdsAPI.getVcek(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("getVcekFromIMDS: %w", err)
		}

		instanceInfo := azureInstanceInfo{
			Vcek:              []byte(vcekResponse.VcekCert),
			CertChain:         []byte(vcekResponse.CertificateChain),
			AttestationReport: hclReport[:0x4a0],
			RuntimeData:       runtimeData,
		}
		statement, err := json.Marshal(instanceInfo)
		if err != nil {
			return nil, fmt.Errorf("marshalling AzureInstanceInfo: %w", err)
		}

		return statement, nil
	}
}

func hclAkTemplate() tpm2.Public {
	akFlags := tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin | tpm2.FlagUserWithAuth | tpm2.FlagNoDA | tpm2.FlagRestricted | tpm2.FlagSign
	return tpm2.Public{
		Type:       tpm2.AlgRSA,
		NameAlg:    tpm2.AlgSHA256,
		Attributes: akFlags,
		RSAParameters: &tpm2.RSAParams{
			Sign: &tpm2.SigScheme{
				Alg:  tpm2.AlgRSASSA,
				Hash: tpm2.AlgSHA256,
			},
			KeyBits: 2048,
		},
	}
}

// getAttestationKey reads the attesation key put into the TPM during early boot.
func getAttestationKey(tpm io.ReadWriter) (*tpmclient.Key, error) {
	// A minor drawback of `NewCachedKey` is that it will transparently create/overwrite a key if it does not find one matching the template at the given index.
	// We actually wouldn't want to continue at this point if we realize that the key at the index is not present, due to
	// easier debuggability. If `NewCachedKey` creates a new key, attestation will fail at the validator.
	// The function in tpmclient that doesn't create a new key, ReadPublic, can't be used as we would have to create
	// a tpmclient.Key object manually, which we can't since there is no constructor exported.
	ak, err := tpmclient.NewCachedKey(tpm, tpm2.HandleOwner, hclAkTemplate(), 0x81000003)
	if err != nil {
		return nil, fmt.Errorf("reading HCL attestation key from TPM: %w", err)
	}

	return ak, nil
}

type tpmReport struct{}

func (s *tpmReport) get(tpm io.ReadWriteCloser) ([]byte, error) {
	return tpm2.NVReadEx(tpm, tpmReportIdx, tpm2.HandleOwner, "", 0)
}

type tpmReportGetter interface {
	get(tpm io.ReadWriteCloser) ([]byte, error)
}

type imdsApi interface {
	getVcek(ctx context.Context) (vcekResponse, error)
}
