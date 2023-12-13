/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAttestationSource(t *testing.T) {
	// Set the path to the Terraform binary for acceptance testing when running under Bazel.
	bazelPreCheck := func() { bazelSetTerraformBinaryPath(t) }

	testCases := map[string]resource.TestCase{
		"aws sev-snp succcess without explicit image_version": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion("v2.13.0"), // do this to test if a valid default version is set
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_attestation" "test" {
						csp = "aws"
						attestation_variant = "aws-sev-snp"
					}
					`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.variant", "aws-sev-snp"),

						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.bootloader_version", "3"),
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.microcode_version", "209"),
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.snp_version", "20"),
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.tee_version", "0"),
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.amd_root_key", "\"-----BEGIN CERTIFICATE-----\\nMIIGYzCCBBKgAwIBAgIDAQAAMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAIC\\nBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBMHsxFDAS\\nBgNVBAsMC0VuZ2luZWVyaW5nMQswCQYDVQQGEwJVUzEUMBIGA1UEBwwLU2FudGEg\\nQ2xhcmExCzAJBgNVBAgMAkNBMR8wHQYDVQQKDBZBZHZhbmNlZCBNaWNybyBEZXZp\\nY2VzMRIwEAYDVQQDDAlBUkstTWlsYW4wHhcNMjAxMDIyMTcyMzA1WhcNNDUxMDIy\\nMTcyMzA1WjB7MRQwEgYDVQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDAS\\nBgNVBAcMC1NhbnRhIENsYXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5j\\nZWQgTWljcm8gRGV2aWNlczESMBAGA1UEAwwJQVJLLU1pbGFuMIICIjANBgkqhkiG\\n9w0BAQEFAAOCAg8AMIICCgKCAgEA0Ld52RJOdeiJlqK2JdsVmD7FktuotWwX1fNg\\nW41XY9Xz1HEhSUmhLz9Cu9DHRlvgJSNxbeYYsnJfvyjx1MfU0V5tkKiU1EesNFta\\n1kTA0szNisdYc9isqk7mXT5+KfGRbfc4V/9zRIcE8jlHN61S1ju8X93+6dxDUrG2\\nSzxqJ4BhqyYmUDruPXJSX4vUc01P7j98MpqOS95rORdGHeI52Naz5m2B+O+vjsC0\\n60d37jY9LFeuOP4Meri8qgfi2S5kKqg/aF6aPtuAZQVR7u3KFYXP59XmJgtcog05\\ngmI0T/OitLhuzVvpZcLph0odh/1IPXqx3+MnjD97A7fXpqGd/y8KxX7jksTEzAOg\\nbKAeam3lm+3yKIcTYMlsRMXPcjNbIvmsBykD//xSniusuHBkgnlENEWx1UcbQQrs\\n+gVDkuVPhsnzIRNgYvM48Y+7LGiJYnrmE8xcrexekBxrva2V9TJQqnN3Q53kt5vi\\nQi3+gCfmkwC0F0tirIZbLkXPrPwzZ0M9eNxhIySb2npJfgnqz55I0u33wh4r0ZNQ\\neTGfw03MBUtyuzGesGkcw+loqMaq1qR4tjGbPYxCvpCq7+OgpCCoMNit2uLo9M18\\nfHz10lOMT8nWAUvRZFzteXCm+7PHdYPlmQwUw3LvenJ/ILXoQPHfbkH0CyPfhl1j\\nWhJFZasCAwEAAaN+MHwwDgYDVR0PAQH/BAQDAgEGMB0GA1UdDgQWBBSFrBrRQ/fI\\nrFXUxR1BSKvVeErUUzAPBgNVHRMBAf8EBTADAQH/MDoGA1UdHwQzMDEwL6AtoCuG\\nKWh0dHBzOi8va2RzaW50Zi5hbWQuY29tL3ZjZWsvdjEvTWlsYW4vY3JsMEYGCSqG\\nSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAICBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZI\\nAWUDBAICBQCiAwIBMKMDAgEBA4ICAQC6m0kDp6zv4Ojfgy+zleehsx6ol0ocgVel\\nETobpx+EuCsqVFRPK1jZ1sp/lyd9+0fQ0r66n7kagRk4Ca39g66WGTJMeJdqYriw\\nSTjjDCKVPSesWXYPVAyDhmP5n2v+BYipZWhpvqpaiO+EGK5IBP+578QeW/sSokrK\\ndHaLAxG2LhZxj9aF73fqC7OAJZ5aPonw4RE299FVarh1Tx2eT3wSgkDgutCTB1Yq\\nzT5DuwvAe+co2CIVIzMDamYuSFjPN0BCgojl7V+bTou7dMsqIu/TW/rPCX9/EUcp\\nKGKqPQ3P+N9r1hjEFY1plBg93t53OOo49GNI+V1zvXPLI6xIFVsh+mto2RtgEX/e\\npmMKTNN6psW88qg7c1hTWtN6MbRuQ0vm+O+/2tKBF2h8THb94OvvHHoFDpbCELlq\\nHnIYhxy0YKXGyaW1NjfULxrrmxVW4wcn5E8GddmvNa6yYm8scJagEi13mhGu4Jqh\\n3QU3sf8iUSUr09xQDwHtOQUVIqx4maBZPBtSMf+qUDtjXSSq8lfWcd8bLr9mdsUn\\nJZJ0+tuPMKmBnSH860llKk+VpVQsgqbzDIvOLvD6W1Umq25boxCYJ+TuBoa4s+HH\\nCViAvgT9kf/rBq1d+ivj6skkHxuzcxbk1xv6ZGxrteJxVH7KlX7YRdZ6eARKwLe4\\nAFZEAwoKCQ==\\n-----END CERTIFICATE-----\\n\""),

						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.measurements.0.expected", "7b068c0c3ac29afe264134536b9be26f1d4ccd575b88d3c3ceabf36ac99c0278"),
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.measurements.0.warn_only", "true"),
					),
				},
			},
		},
		"azure sev-snp success": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_attestation" "test" {
						csp = "azure"
						attestation_variant = "azure-sev-snp"
						image_version = "v2.13.0"
						maa_url = "https://www.example.com"
					}
					`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.variant", "azure-sev-snp"),

						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.bootloader_version", "3"),
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.microcode_version", "115"),
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.snp_version", "8"),
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.tee_version", "0"),

						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.azure_firmware_signer_config.accepted_key_digests.0", "0356215882a825279a85b300b0b742931d113bf7e32dde2e50ffde7ec743ca491ecdd7f336dc28a6e0b2bb57af7a44a3"),
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.azure_firmware_signer_config.enforcement_policy", "MAAFallback"),

						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.amd_root_key", "\"-----BEGIN CERTIFICATE-----\\nMIIGYzCCBBKgAwIBAgIDAQAAMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAIC\\nBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBMHsxFDAS\\nBgNVBAsMC0VuZ2luZWVyaW5nMQswCQYDVQQGEwJVUzEUMBIGA1UEBwwLU2FudGEg\\nQ2xhcmExCzAJBgNVBAgMAkNBMR8wHQYDVQQKDBZBZHZhbmNlZCBNaWNybyBEZXZp\\nY2VzMRIwEAYDVQQDDAlBUkstTWlsYW4wHhcNMjAxMDIyMTcyMzA1WhcNNDUxMDIy\\nMTcyMzA1WjB7MRQwEgYDVQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDAS\\nBgNVBAcMC1NhbnRhIENsYXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5j\\nZWQgTWljcm8gRGV2aWNlczESMBAGA1UEAwwJQVJLLU1pbGFuMIICIjANBgkqhkiG\\n9w0BAQEFAAOCAg8AMIICCgKCAgEA0Ld52RJOdeiJlqK2JdsVmD7FktuotWwX1fNg\\nW41XY9Xz1HEhSUmhLz9Cu9DHRlvgJSNxbeYYsnJfvyjx1MfU0V5tkKiU1EesNFta\\n1kTA0szNisdYc9isqk7mXT5+KfGRbfc4V/9zRIcE8jlHN61S1ju8X93+6dxDUrG2\\nSzxqJ4BhqyYmUDruPXJSX4vUc01P7j98MpqOS95rORdGHeI52Naz5m2B+O+vjsC0\\n60d37jY9LFeuOP4Meri8qgfi2S5kKqg/aF6aPtuAZQVR7u3KFYXP59XmJgtcog05\\ngmI0T/OitLhuzVvpZcLph0odh/1IPXqx3+MnjD97A7fXpqGd/y8KxX7jksTEzAOg\\nbKAeam3lm+3yKIcTYMlsRMXPcjNbIvmsBykD//xSniusuHBkgnlENEWx1UcbQQrs\\n+gVDkuVPhsnzIRNgYvM48Y+7LGiJYnrmE8xcrexekBxrva2V9TJQqnN3Q53kt5vi\\nQi3+gCfmkwC0F0tirIZbLkXPrPwzZ0M9eNxhIySb2npJfgnqz55I0u33wh4r0ZNQ\\neTGfw03MBUtyuzGesGkcw+loqMaq1qR4tjGbPYxCvpCq7+OgpCCoMNit2uLo9M18\\nfHz10lOMT8nWAUvRZFzteXCm+7PHdYPlmQwUw3LvenJ/ILXoQPHfbkH0CyPfhl1j\\nWhJFZasCAwEAAaN+MHwwDgYDVR0PAQH/BAQDAgEGMB0GA1UdDgQWBBSFrBrRQ/fI\\nrFXUxR1BSKvVeErUUzAPBgNVHRMBAf8EBTADAQH/MDoGA1UdHwQzMDEwL6AtoCuG\\nKWh0dHBzOi8va2RzaW50Zi5hbWQuY29tL3ZjZWsvdjEvTWlsYW4vY3JsMEYGCSqG\\nSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAICBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZI\\nAWUDBAICBQCiAwIBMKMDAgEBA4ICAQC6m0kDp6zv4Ojfgy+zleehsx6ol0ocgVel\\nETobpx+EuCsqVFRPK1jZ1sp/lyd9+0fQ0r66n7kagRk4Ca39g66WGTJMeJdqYriw\\nSTjjDCKVPSesWXYPVAyDhmP5n2v+BYipZWhpvqpaiO+EGK5IBP+578QeW/sSokrK\\ndHaLAxG2LhZxj9aF73fqC7OAJZ5aPonw4RE299FVarh1Tx2eT3wSgkDgutCTB1Yq\\nzT5DuwvAe+co2CIVIzMDamYuSFjPN0BCgojl7V+bTou7dMsqIu/TW/rPCX9/EUcp\\nKGKqPQ3P+N9r1hjEFY1plBg93t53OOo49GNI+V1zvXPLI6xIFVsh+mto2RtgEX/e\\npmMKTNN6psW88qg7c1hTWtN6MbRuQ0vm+O+/2tKBF2h8THb94OvvHHoFDpbCELlq\\nHnIYhxy0YKXGyaW1NjfULxrrmxVW4wcn5E8GddmvNa6yYm8scJagEi13mhGu4Jqh\\n3QU3sf8iUSUr09xQDwHtOQUVIqx4maBZPBtSMf+qUDtjXSSq8lfWcd8bLr9mdsUn\\nJZJ0+tuPMKmBnSH860llKk+VpVQsgqbzDIvOLvD6W1Umq25boxCYJ+TuBoa4s+HH\\nCViAvgT9kf/rBq1d+ivj6skkHxuzcxbk1xv6ZGxrteJxVH7KlX7YRdZ6eARKwLe4\\nAFZEAwoKCQ==\\n-----END CERTIFICATE-----\\n\""),

						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.measurements.1.expected", "3d458cfe55cc03ea1f443f1562beec8df51c75e14a9fcf9a7234a13f198e7969"),
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.measurements.1.warn_only", "true"),
					),
				},
			},
		},
		"gcp sev-snp succcess": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_attestation" "test" {
						csp = "gcp"
						attestation_variant = "gcp-sev-es"
						image_version = "v2.13.0"
					}
					`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.variant", "gcp-sev-es"),
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.bootloader_version", "0"), // since this is not supported on GCP, we expect 0

						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.measurements.1.expected", "745f2fb4235e4647aa0ad5ace781cd929eb68c28870e7dd5d1a1535854325e56"),
						resource.TestCheckResourceAttr("data.constellation_attestation.test", "attestation.measurements.1.warn_only", "true"),
					),
				},
			},
		},
		"azure fails without maa_url": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_attestation" "test" {
						csp = "azure"
						attestation_variant = "azure-sev-snp"
						image_version = "v2.13.0"
					}
					`,
					ExpectError: regexp.MustCompile(".*MAA URL must be set for Azure SEV-SNP.*"),
				},
			},
		},
		"aws sev-snp with maa_url fails": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_attestation" "test" {
						csp = "aws"
						attestation_variant = "aws-sev-snp"
						image_version = "v2.13.0"
						maa_url = "https://www.example.com"
					}
					`,
					ExpectError: regexp.MustCompile(".*MAA URL can only be set for Azure SEV-SNP.*"),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			resource.Test(t, tc)
		})
	}
}
