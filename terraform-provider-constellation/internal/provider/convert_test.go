/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAttestationConfig(t *testing.T) {
	testAttestation := attestation{
		BootloaderVersion: 1,
		TEEVersion:        2,
		SNPVersion:        3,
		MicrocodeVersion:  4,
		AMDRootKey:        "\"-----BEGIN CERTIFICATE-----\\nMIIGYzCCBBKgAwIBAgIDAQAAMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAIC\\nBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBMHsxFDAS\\nBgNVBAsMC0VuZ2luZWVyaW5nMQswCQYDVQQGEwJVUzEUMBIGA1UEBwwLU2FudGEg\\nQ2xhcmExCzAJBgNVBAgMAkNBMR8wHQYDVQQKDBZBZHZhbmNlZCBNaWNybyBEZXZp\\nY2VzMRIwEAYDVQQDDAlBUkstTWlsYW4wHhcNMjAxMDIyMTcyMzA1WhcNNDUxMDIy\\nMTcyMzA1WjB7MRQwEgYDVQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDAS\\nBgNVBAcMC1NhbnRhIENsYXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5j\\nZWQgTWljcm8gRGV2aWNlczESMBAGA1UEAwwJQVJLLU1pbGFuMIICIjANBgkqhkiG\\n9w0BAQEFAAOCAg8AMIICCgKCAgEA0Ld52RJOdeiJlqK2JdsVmD7FktuotWwX1fNg\\nW41XY9Xz1HEhSUmhLz9Cu9DHRlvgJSNxbeYYsnJfvyjx1MfU0V5tkKiU1EesNFta\\n1kTA0szNisdYc9isqk7mXT5+KfGRbfc4V/9zRIcE8jlHN61S1ju8X93+6dxDUrG2\\nSzxqJ4BhqyYmUDruPXJSX4vUc01P7j98MpqOS95rORdGHeI52Naz5m2B+O+vjsC0\\n60d37jY9LFeuOP4Meri8qgfi2S5kKqg/aF6aPtuAZQVR7u3KFYXP59XmJgtcog05\\ngmI0T/OitLhuzVvpZcLph0odh/1IPXqx3+MnjD97A7fXpqGd/y8KxX7jksTEzAOg\\nbKAeam3lm+3yKIcTYMlsRMXPcjNbIvmsBykD//xSniusuHBkgnlENEWx1UcbQQrs\\n+gVDkuVPhsnzIRNgYvM48Y+7LGiJYnrmE8xcrexekBxrva2V9TJQqnN3Q53kt5vi\\nQi3+gCfmkwC0F0tirIZbLkXPrPwzZ0M9eNxhIySb2npJfgnqz55I0u33wh4r0ZNQ\\neTGfw03MBUtyuzGesGkcw+loqMaq1qR4tjGbPYxCvpCq7+OgpCCoMNit2uLo9M18\\nfHz10lOMT8nWAUvRZFzteXCm+7PHdYPlmQwUw3LvenJ/ILXoQPHfbkH0CyPfhl1j\\nWhJFZasCAwEAAaN+MHwwDgYDVR0PAQH/BAQDAgEGMB0GA1UdDgQWBBSFrBrRQ/fI\\nrFXUxR1BSKvVeErUUzAPBgNVHRMBAf8EBTADAQH/MDoGA1UdHwQzMDEwL6AtoCuG\\nKWh0dHBzOi8va2RzaW50Zi5hbWQuY29tL3ZjZWsvdjEvTWlsYW4vY3JsMEYGCSqG\\nSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAICBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZI\\nAWUDBAICBQCiAwIBMKMDAgEBA4ICAQC6m0kDp6zv4Ojfgy+zleehsx6ol0ocgVel\\nETobpx+EuCsqVFRPK1jZ1sp/lyd9+0fQ0r66n7kagRk4Ca39g66WGTJMeJdqYriw\\nSTjjDCKVPSesWXYPVAyDhmP5n2v+BYipZWhpvqpaiO+EGK5IBP+578QeW/sSokrK\\ndHaLAxG2LhZxj9aF73fqC7OAJZ5aPonw4RE299FVarh1Tx2eT3wSgkDgutCTB1Yq\\nzT5DuwvAe+co2CIVIzMDamYuSFjPN0BCgojl7V+bTou7dMsqIu/TW/rPCX9/EUcp\\nKGKqPQ3P+N9r1hjEFY1plBg93t53OOo49GNI+V1zvXPLI6xIFVsh+mto2RtgEX/e\\npmMKTNN6psW88qg7c1hTWtN6MbRuQ0vm+O+/2tKBF2h8THb94OvvHHoFDpbCELlq\\nHnIYhxy0YKXGyaW1NjfULxrrmxVW4wcn5E8GddmvNa6yYm8scJagEi13mhGu4Jqh\\n3QU3sf8iUSUr09xQDwHtOQUVIqx4maBZPBtSMf+qUDtjXSSq8lfWcd8bLr9mdsUn\\nJZJ0+tuPMKmBnSH860llKk+VpVQsgqbzDIvOLvD6W1Umq25boxCYJ+TuBoa4s+HH\\nCViAvgT9kf/rBq1d+ivj6skkHxuzcxbk1xv6ZGxrteJxVH7KlX7YRdZ6eARKwLe4\\nAFZEAwoKCQ==\\n-----END CERTIFICATE-----\\n\"",
		AzureSNPFirmwareSignerConfig: azureSnpFirmwareSignerConfig{
			AcceptedKeyDigests: []string{"0356215882a825279a85b300b0b742931d113bf7e32dde2e50ffde7ec743ca491ecdd7f336dc28a6e0b2bb57af7a44a3"},
			EnforcementPolicy:  "equal",
			MAAURL:             "https://example.com",
		},
		Measurements: map[string]measurement{
			"1": {Expected: "48656c6c6f", WarnOnly: false}, // "Hello" in hex
			"2": {Expected: "776f726c64", WarnOnly: true},  // "world" in hex
		},
	}
	t.Run("Azure SEV-SNP success", func(t *testing.T) {
		attestationVariant := variant.AzureSEVSNP{}

		cfg, err := convertFromTfAttestationCfg(testAttestation, attestationVariant)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		azureCfg, ok := cfg.(*config.AzureSEVSNP)
		require.True(t, ok)

		require.Equal(t, "ARK-Milan", azureCfg.AMDRootKey.Issuer.CommonName)

		require.Equal(t, uint8(1), azureCfg.BootloaderVersion.Value)
		require.Equal(t, uint8(2), azureCfg.TEEVersion.Value)
		require.Equal(t, uint8(3), azureCfg.SNPVersion.Value)
		require.Equal(t, uint8(4), azureCfg.MicrocodeVersion.Value)

		require.Equal(t, []byte("Hello"), azureCfg.Measurements[1].Expected)
		require.Equal(t, measurements.Enforce, azureCfg.Measurements[1].ValidationOpt)

		require.Equal(t, []byte("world"), azureCfg.Measurements[2].Expected)
		require.Equal(t, measurements.WarnOnly, azureCfg.Measurements[2].ValidationOpt)

		require.Equal(t, "https://example.com", azureCfg.FirmwareSignerConfig.MAAURL)
		require.Equal(t, idkeydigest.Equal, azureCfg.FirmwareSignerConfig.EnforcementPolicy)

		assert.Len(t, azureCfg.FirmwareSignerConfig.AcceptedKeyDigests, 1)
	})

	// Test error scenarios
	t.Run("invalid_measurement_index", func(t *testing.T) {
		testAttestation.Measurements = map[string]measurement{"invalid": {Expected: "data"}}
		attestationVariant := variant.AzureSEVSNP{}

		_, err := convertFromTfAttestationCfg(testAttestation, attestationVariant)
		assert.Error(t, err)
	})

	t.Run("invalid_amd_root_key", func(t *testing.T) {
		invalidAttestation := testAttestation
		invalidAttestation.AMDRootKey = "not valid json"

		attestationVariant := variant.AzureSEVSNP{}

		_, err := convertFromTfAttestationCfg(invalidAttestation, attestationVariant)
		assert.Error(t, err)
	})
}
