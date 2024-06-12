/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
package tdx implements attestation for TDX on Azure.

Quotes are generated using an Azure provided vTPM and the IMDS API.
They are verified using the go-tdx-guest library.

More specifically:
- The vTPM is used to collected a TPM attestation and a Hardware Compatibility Layer (HCL) report.
- The HCL report is sent to the IMDS API to generate a TDX quote.
- The quote is verified using the go-tdx-guest library.
- The quote's report data can be used to verify the TPM's attestation key.
- The attestation key can be used to verify the TPM attestation.
*/
package tdx

// InstanceInfo wraps the TDX report with additional Azure specific runtime data.
type InstanceInfo struct {
	AttestationReport []byte
	RuntimeData       []byte
}
