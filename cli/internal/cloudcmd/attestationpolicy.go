/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package cloudcmd

import (
	"encoding/base64"
	"fmt"
)

// maaAttestationPolicy is the default attestation policy for Azure VMs on Constellation.
const maaAttestationPolicy = `
                version= 1.0;
                authorizationrules
                {
                    [type=="x-ms-azurevm-default-securebootkeysvalidated", value==false] => deny();
                    [type=="x-ms-azurevm-debuggersdisabled", value==false] => deny();
                    // The line below was edited by the Constellation CLI. Do not edit manually.
                    //[type=="secureboot", value==false] => deny();
                    [type=="x-ms-azurevm-signingdisabled", value==false] => deny();
                    [type=="x-ms-azurevm-dbvalidated", value==false] => deny();
                    [type=="x-ms-azurevm-dbxvalidated", value==false] => deny();
                    => permit();
                };
                issuancerules
                {
                };`

// NewAzureMaaAttestationPolicy returns a new AzureAttestationPolicy to use with MAA.
func NewAzureMaaAttestationPolicy() AzureAttestationPolicy {
	return AzureAttestationPolicy{
		policy: maaAttestationPolicy,
	}
}

// AzureAttestationPolicy patches attestation policies on Azure.
type AzureAttestationPolicy struct {
	policy string
}

// Encode encodes the base64-encoded attestation policy in the JWS format specified here:
// https://learn.microsoft.com/en-us/azure/attestation/author-sign-policy#creating-the-policy-file-in-json-web-signature-format
func (p AzureAttestationPolicy) Encode() string {
	encodedPolicy := base64.RawURLEncoding.EncodeToString([]byte(p.policy))
	const header = `{"alg":"none"}`
	payload := fmt.Sprintf(`{"AttestationPolicy":"%s"}`, encodedPolicy)

	encodedHeader := base64.RawURLEncoding.EncodeToString([]byte(header))
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))

	return fmt.Sprintf("%s.%s.", encodedHeader, encodedPayload)
}
