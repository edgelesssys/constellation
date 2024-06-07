/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package maa

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/attestation/attestation"
	azpolicy "github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// NewAzurePolicyPatcher returns a new AzurePolicyPatcher.
func NewAzurePolicyPatcher() AzurePolicyPatcher {
	return AzurePolicyPatcher{}
}

// AzurePolicyPatcher patches attestation policies on Azure.
type AzurePolicyPatcher struct{}

// Patch updates the attestation policy to the base64-encoded attestation policy JWT for the given attestation URL.
// https://learn.microsoft.com/en-us/azure/attestation/author-sign-policy#next-steps
func (p AzurePolicyPatcher) Patch(ctx context.Context, attestationURL string) error {
	// hacky way to update the MAA attestation policy. This should be changed as soon as either the Terraform provider supports it
	// or the Go SDK gets updated to a recent API version.
	// https://github.com/hashicorp/terraform-provider-azurerm/issues/20804
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("retrieving default Azure credentials: %w", err)
	}
	token, err := cred.GetToken(ctx, azpolicy.TokenRequestOptions{
		Scopes: []string{"https://attest.azure.net/.default"},
	})
	if err != nil {
		return fmt.Errorf("retrieving token from default Azure credentials: %w", err)
	}

	client := attestation.NewPolicyClient()

	// azureGuest is the id for the "Azure VM" attestation type. Other types are documented here:
	// https://learn.microsoft.com/en-us/rest/api/attestation/policy/set
	req, err := client.SetPreparer(ctx, attestationURL, "azureGuest", p.encodeAttestationPolicy())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Token))
	if err != nil {
		return fmt.Errorf("preparing request: %w", err)
	}

	resp, err := client.Send(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("updating attestation policy: unexpected status code: %s: %s", resp.Status, string(body))
	}

	return nil
}

// encodeAttestationPolicy encodes the base64-encoded attestation policy in the JWS format specified here:
// https://learn.microsoft.com/en-us/azure/attestation/author-sign-policy#creating-the-policy-file-in-json-web-signature-format
func (p AzurePolicyPatcher) encodeAttestationPolicy() string {
	const policy = `
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
	encodedPolicy := base64.RawURLEncoding.EncodeToString([]byte(policy))
	const header = `{"alg":"none"}`
	payload := fmt.Sprintf(`{"AttestationPolicy":"%s"}`, encodedPolicy)

	encodedHeader := base64.RawURLEncoding.EncodeToString([]byte(header))
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))

	return fmt.Sprintf("%s.%s.", encodedHeader, encodedPayload)
}
