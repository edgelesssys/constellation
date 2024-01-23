/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package tdx

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/google/go-tpm/legacy/tpm2"
)

const (
	imdsURL                  = "http://169.254.169.254/acc/tdquote"
	indexHCLReport           = 0x1400001
	hclDataOffset            = 1216
	hclReportTypeOffset      = 8
	hclReportTypeOffsetStart = hclDataOffset + hclReportTypeOffset
	hclRequestDataSizeOffset = 16
	runtimeDataSizeOffset    = hclDataOffset + hclRequestDataSizeOffset
	hclRequestDataOffset     = 20
	runtimeDataOffset        = hclDataOffset + hclRequestDataOffset
	tdReportSize             = 1024
	hwReportStart            = 32
	hwReportEnd              = 1216
)

const (
	hclReportTypeInvalid uint32 = iota
	hclReportTypeReserved
	hclReportTypeSNP
	hclReportTypeTVM
	hclReportTypeTDX
)

// Issuer for Azure confidential VM attestation using TDX.
type Issuer struct {
	variant.AzureTDX
	*vtpm.Issuer

	quoteGetter quoteGetter
}

// NewIssuer initializes a new Azure Issuer.
func NewIssuer(log attestation.Logger) *Issuer {
	i := &Issuer{
		quoteGetter: imdsQuoteGetter{
			client: &http.Client{Transport: &http.Transport{Proxy: nil}},
		},
	}

	i.Issuer = vtpm.NewIssuer(
		vtpm.OpenVTPM,
		azure.GetAttestationKey,
		i.getInstanceInfo,
		log,
	)
	return i
}

func (i *Issuer) getInstanceInfo(ctx context.Context, tpm io.ReadWriteCloser, _ []byte) ([]byte, error) {
	// Read HCL report from TPM
	report, err := tpm2.NVReadEx(tpm, indexHCLReport, tpm2.HandleOwner, "", 0)
	if err != nil {
		return nil, err
	}

	// Parse the report from the TPM
	hwReport, runtimeData, err := parseHCLReport(report)
	if err != nil {
		return nil, fmt.Errorf("getting HCL report: %w", err)
	}

	// Get quote from IMDS API
	quote, err := i.quoteGetter.getQuote(ctx, hwReport)
	if err != nil {
		return nil, fmt.Errorf("getting quote: %w", err)
	}

	instanceInfo := instanceInfo{
		AttestationReport: quote,
		RuntimeData:       runtimeData,
	}
	instanceInfoJSON, err := json.Marshal(instanceInfo)
	if err != nil {
		return nil, fmt.Errorf("marshalling instance info: %w", err)
	}
	return instanceInfoJSON, nil
}

func parseHCLReport(report []byte) (hwReport, runtimeData []byte, err error) {
	// First, ensure the extracted report is actually for TDX
	if len(report) < hclReportTypeOffsetStart+4 {
		return nil, nil, fmt.Errorf("invalid HCL report: expected at least %d bytes to read HCL report type, got %d", hclReportTypeOffsetStart+4, len(report))
	}
	reportType := binary.LittleEndian.Uint32(report[hclReportTypeOffsetStart : hclReportTypeOffsetStart+4])
	if reportType != hclReportTypeTDX {
		return nil, nil, fmt.Errorf("invalid HCL report type: expected TDX (%d), got %d", hclReportTypeTDX, reportType)
	}

	// We need the td report (generally called HW report in Azure's samples) from the HCL report to send to the IMDS API
	if len(report) < hwReportStart+tdReportSize {
		return nil, nil, fmt.Errorf("invalid HCL report: expected at least %d bytes to read td report, got %d", hwReportStart+tdReportSize, len(report))
	}
	hwReport = report[hwReportStart : hwReportStart+tdReportSize]

	// We also need the runtime data to verify the attestation key later on the validator side
	if len(report) < runtimeDataSizeOffset+4 {
		return nil, nil, fmt.Errorf("invalid HCL report: expected at least %d bytes to read runtime data size, got %d", runtimeDataSizeOffset+4, len(report))
	}
	runtimeDataSize := int(binary.LittleEndian.Uint32(report[runtimeDataSizeOffset : runtimeDataSizeOffset+4]))
	if len(report) < runtimeDataOffset+int(runtimeDataSize) {
		return nil, nil, fmt.Errorf("invalid HCL report: expected at least %d bytes to read runtime data, got %d", runtimeDataOffset+runtimeDataSize, len(report))
	}
	runtimeData = report[runtimeDataOffset : runtimeDataOffset+runtimeDataSize]

	return hwReport, runtimeData, nil
}

// imdsQuoteGetter issues TDX quotes using Azure's IMDS API.
type imdsQuoteGetter struct {
	client *http.Client
}

func (i imdsQuoteGetter) getQuote(ctx context.Context, hwReport []byte) ([]byte, error) {
	encodedReportJSON, err := json.Marshal(quoteRequest{
		Report: base64.RawURLEncoding.EncodeToString(hwReport),
	})
	if err != nil {
		return nil, fmt.Errorf("marshalling encoded report: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, imdsURL, bytes.NewReader(encodedReportJSON))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := i.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	var quoteRes quoteResponse
	if err := json.NewDecoder(res.Body).Decode(&quoteRes); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return base64.RawURLEncoding.DecodeString(quoteRes.Quote)
}

type quoteRequest struct {
	Report string `json:"report"`
}

type quoteResponse struct {
	Quote string `json:"quote"`
}

type quoteGetter interface {
	getQuote(ctx context.Context, encodedHWReport []byte) ([]byte, error)
}
