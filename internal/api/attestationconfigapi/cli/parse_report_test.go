/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var completeReport = `
Attestation Document:
	Quote:
		PCR 1 (Strict: false):
			Expected:	3d458cfe55cc03ea1f443f1562beec8df51c75e14a9fcf9a7234a13f198e7969
			Actual:		3d458cfe55cc03ea1f443f1562beec8df51c75e14a9fcf9a7234a13f198e7969
		PCR 2 (Strict: false):
			Expected:	3d458cfe55cc03ea1f443f1562beec8df51c75e14a9fcf9a7234a13f198e7969
			Actual:		3d458cfe55cc03ea1f443f1562beec8df51c75e14a9fcf9a7234a13f198e7969
		PCR 3 (Strict: false):
			Expected:	3d458cfe55cc03ea1f443f1562beec8df51c75e14a9fcf9a7234a13f198e7969
			Actual:		3d458cfe55cc03ea1f443f1562beec8df51c75e14a9fcf9a7234a13f198e7969
		PCR 4 (Strict: true):
			Expected:	1e0657268baf9006037ce952feaab4009ba50ffcb8ff38b2d8596071f2c5b28c
			Actual:		1e0657268baf9006037ce952feaab4009ba50ffcb8ff38b2d8596071f2c5b28c
		PCR 8 (Strict: true):
			Expected:	0000000000000000000000000000000000000000000000000000000000000000
			Actual:		0000000000000000000000000000000000000000000000000000000000000000
		PCR 9 (Strict: true):
			Expected:	96b8d52c5206be1720e088b28bf684fd54733a666e63aabe21d80a142f84d456
			Actual:		96b8d52c5206be1720e088b28bf684fd54733a666e63aabe21d80a142f84d456
		PCR 11 (Strict: true):
			Expected:	0000000000000000000000000000000000000000000000000000000000000000
			Actual:		0000000000000000000000000000000000000000000000000000000000000000
		PCR 12 (Strict: true):
			Expected:	52822e901c3e8be3a222979ba23092f171d148b87fa583ed953246a2083df457
			Actual:		52822e901c3e8be3a222979ba23092f171d148b87fa583ed953246a2083df457
		PCR 13 (Strict: true):
			Expected:	0000000000000000000000000000000000000000000000000000000000000000
			Actual:		0000000000000000000000000000000000000000000000000000000000000000
		PCR 14 (Strict: false):
			Expected:	d7c4cc7ff7933022f013e03bdee875b91720b5b86cf1753cad830f95e791926f
			Actual:		d7c4cc7ff7933022f013e03bdee875b91720b5b86cf1753cad830f95e791926f
		PCR 15 (Strict: true):
			Expected:	3501e426efdac775735f84bdaf0268d0b4ee286db11160dd38e4861fc28422f5
			Actual:		3501e426efdac775735f84bdaf0268d0b4ee286db11160dd38e4861fc28422f5
	Raw VCEK certificate:
		-----BEGIN CERTIFICATE-----
		MIIFTDCCAvugAwIBAgIBADBGBgkqhkiG9w0BAQowOaAPMA0GCWCGSAFlAwQCAgUA
		oRwwGgYJKoZIhvcNAQEIMA0GCWCGSAFlAwQCAgUAogMCATCjAwIBATB7MRQwEgYD
		VQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDASBgNVBAcMC1NhbnRhIENs
		YXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5jZWQgTWljcm8gRGV2aWNl
		czESMBAGA1UEAwwJU0VWLU1pbGFuMB4XDTIyMTEyOTA2NDU0NloXDTI5MTEyOTA2
		NDU0NlowejEUMBIGA1UECwwLRW5naW5lZXJpbmcxCzAJBgNVBAYTAlVTMRQwEgYD
		VQQHDAtTYW50YSBDbGFyYTELMAkGA1UECAwCQ0ExHzAdBgNVBAoMFkFkdmFuY2Vk
		IE1pY3JvIERldmljZXMxETAPBgNVBAMMCFNFVi1WQ0VLMHYwEAYHKoZIzj0CAQYF
		K4EEACIDYgAE1JskoXxFff+fwjPMmWWOxy5oGfmdxltZMU2VHZ1XS6VkNpS1fmW/
		IibUcKiig4HTf+ExbQFjzDVidnYbY1bMl1y/5GNAiI1oh/iBSwTRr1pTZgqd+YoC
		x9q/522iJF/Bo4IBFjCCARIwEAYJKwYBBAGceAEBBAMCAQAwFwYJKwYBBAGceAEC
		BAoWCE1pbGFuLUIwMBEGCisGAQQBnHgBAwEEAwIBAzARBgorBgEEAZx4AQMCBAMC
		AQAwEQYKKwYBBAGceAEDBAQDAgEAMBEGCisGAQQBnHgBAwUEAwIBADARBgorBgEE
		AZx4AQMGBAMCAQAwEQYKKwYBBAGceAEDBwQDAgEAMBEGCisGAQQBnHgBAwMEAwIB
		CDARBgorBgEEAZx4AQMIBAMCAXMwTQYJKwYBBAGceAEEBECReB86zeQgVrN+iTCw
		ZEvQZZ8TXHF7+HeAPyJREz05fjWLJziOfibHlpi+wTthY9deIHRLRI2sSKVocNga
		tpYZMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAICBQChHDAaBgkqhkiG9w0B
		AQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBA4ICAQAMxoEM1h7ZnGsobPq3j5Ze
		rzqQclJQUc+iuwnj7DfG4vttAcLiCkQJmBFwu50cv2v8x62yWrbURrSURcRH5zwj
		7woHN8XFcy5ArHvXt0HDFvuwwMYhnadKzNxHSXvSQmdTS7sr/S0aJGozK13YwHcq
		mcIMocb39ly+wdeDuI3TxzygNlPGqilMwmgMdF2Le1SjnM9FCxSqNtWDGSLYZBAS
		jmVNX/e1/j1Mzr1lQBJPWIJ8lDragBYPhw6Ptou40A9mXE67BCoQVyA9I7hFCSh6
		MCEaI5onwQT61GFMOqotU0iPojxTptl/e1IzgNjUwX/Os9Jm4lMwlgfWxYF862cM
		DYPhhFd9lyljpR8vgNxVESxlVZMscfvw2u/ZQtyJQVPiuUtZAVCCMDR02Vmh/SIF
		zKhw3azVg2m8YTR4aaWby8JmyWOMz0a+2TBJZuFXx4isSgH3gpijw/c6o5pfRyHG
		Nb/0RnxpDLtZ08UGnlRhnYwSCtPq0lg7LGbRohAW0r43Xg7qhYOcS8t4LxenE54v
		P/naIYb9IwM/qMP4ozO/IbtKb0GYOy2msMeFzuhPfy/LxjfLA+vMAtw/hO2LUxvN
		vpzlBD09rVHhBGvkE8AZQiWoB8Ye6CIyQj/tDD3r0BieK6AEWVcv2559Fa+jsVDr
		2IDvCVQUpxHPT7moggwNWA==
		-----END CERTIFICATE-----
	VCEK certificate (1):
		Serial Number: 0
		Subject: CN=SEV-VCEK,OU=Engineering,O=Advanced Micro Devices,L=Santa Clara,ST=CA,C=US
		Issuer: CN=SEV-Milan,OU=Engineering,O=Advanced Micro Devices,L=Santa Clara,ST=CA,C=US
		Not Before: 2022-11-29 06:45:46 +0000 UTC
		Not After: 2029-11-29 06:45:46 +0000 UTC
		Signature Algorithm: SHA384-RSAPSS
		Public Key Algorithm: ECDSA
		Struct version: 0
		Product name: Milan-B0
		Secure Processor bootloader SVN: 3
		Secure Processor operating system SVN: 0
		SVN 4 (reserved): 0
		SVN 5 (reserved): 0
		SVN 6 (reserved): 0
		SVN 7 (reserved): 0
		SEV-SNP firmware SVN: 8
		Microcode SVN: 115
		Hardware ID: 91781f3acde42056b37e8930b0644bd0659f135c717bf877803f2251133d397e358b27388e7e26c79698bec13b6163d75e20744b448dac48a56870d81ab69619
	Raw Certificate chain:
		-----BEGIN CERTIFICATE-----
		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
		-----END CERTIFICATE-----
		-----BEGIN CERTIFICATE-----
		MIIGYzCCBBKgAwIBAgIDAQAAMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAIC
		BQChHDAaBgkqhkiG9w0BAQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBMHsxFDAS
		BgNVBAsMC0VuZ2luZWVyaW5nMQswCQYDVQQGEwJVUzEUMBIGA1UEBwwLU2FudGEg
		Q2xhcmExCzAJBgNVBAgMAkNBMR8wHQYDVQQKDBZBZHZhbmNlZCBNaWNybyBEZXZp
		Y2VzMRIwEAYDVQQDDAlBUkstTWlsYW4wHhcNMjAxMDIyMTcyMzA1WhcNNDUxMDIy
		MTcyMzA1WjB7MRQwEgYDVQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDAS
		BgNVBAcMC1NhbnRhIENsYXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5j
		ZWQgTWljcm8gRGV2aWNlczESMBAGA1UEAwwJQVJLLU1pbGFuMIICIjANBgkqhkiG
		9w0BAQEFAAOCAg8AMIICCgKCAgEA0Ld52RJOdeiJlqK2JdsVmD7FktuotWwX1fNg
		W41XY9Xz1HEhSUmhLz9Cu9DHRlvgJSNxbeYYsnJfvyjx1MfU0V5tkKiU1EesNFta
		1kTA0szNisdYc9isqk7mXT5+KfGRbfc4V/9zRIcE8jlHN61S1ju8X93+6dxDUrG2
		SzxqJ4BhqyYmUDruPXJSX4vUc01P7j98MpqOS95rORdGHeI52Naz5m2B+O+vjsC0
		60d37jY9LFeuOP4Meri8qgfi2S5kKqg/aF6aPtuAZQVR7u3KFYXP59XmJgtcog05
		gmI0T/OitLhuzVvpZcLph0odh/1IPXqx3+MnjD97A7fXpqGd/y8KxX7jksTEzAOg
		bKAeam3lm+3yKIcTYMlsRMXPcjNbIvmsBykD//xSniusuHBkgnlENEWx1UcbQQrs
		+gVDkuVPhsnzIRNgYvM48Y+7LGiJYnrmE8xcrexekBxrva2V9TJQqnN3Q53kt5vi
		Qi3+gCfmkwC0F0tirIZbLkXPrPwzZ0M9eNxhIySb2npJfgnqz55I0u33wh4r0ZNQ
		eTGfw03MBUtyuzGesGkcw+loqMaq1qR4tjGbPYxCvpCq7+OgpCCoMNit2uLo9M18
		fHz10lOMT8nWAUvRZFzteXCm+7PHdYPlmQwUw3LvenJ/ILXoQPHfbkH0CyPfhl1j
		WhJFZasCAwEAAaN+MHwwDgYDVR0PAQH/BAQDAgEGMB0GA1UdDgQWBBSFrBrRQ/fI
		rFXUxR1BSKvVeErUUzAPBgNVHRMBAf8EBTADAQH/MDoGA1UdHwQzMDEwL6AtoCuG
		KWh0dHBzOi8va2RzaW50Zi5hbWQuY29tL3ZjZWsvdjEvTWlsYW4vY3JsMEYGCSqG
		SIb3DQEBCjA5oA8wDQYJYIZIAWUDBAICBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZI
		AWUDBAICBQCiAwIBMKMDAgEBA4ICAQC6m0kDp6zv4Ojfgy+zleehsx6ol0ocgVel
		ETobpx+EuCsqVFRPK1jZ1sp/lyd9+0fQ0r66n7kagRk4Ca39g66WGTJMeJdqYriw
		STjjDCKVPSesWXYPVAyDhmP5n2v+BYipZWhpvqpaiO+EGK5IBP+578QeW/sSokrK
		dHaLAxG2LhZxj9aF73fqC7OAJZ5aPonw4RE299FVarh1Tx2eT3wSgkDgutCTB1Yq
		zT5DuwvAe+co2CIVIzMDamYuSFjPN0BCgojl7V+bTou7dMsqIu/TW/rPCX9/EUcp
		KGKqPQ3P+N9r1hjEFY1plBg93t53OOo49GNI+V1zvXPLI6xIFVsh+mto2RtgEX/e
		pmMKTNN6psW88qg7c1hTWtN6MbRuQ0vm+O+/2tKBF2h8THb94OvvHHoFDpbCELlq
		HnIYhxy0YKXGyaW1NjfULxrrmxVW4wcn5E8GddmvNa6yYm8scJagEi13mhGu4Jqh
		3QU3sf8iUSUr09xQDwHtOQUVIqx4maBZPBtSMf+qUDtjXSSq8lfWcd8bLr9mdsUn
		JZJ0+tuPMKmBnSH860llKk+VpVQsgqbzDIvOLvD6W1Umq25boxCYJ+TuBoa4s+HH
		CViAvgT9kf/rBq1d+ivj6skkHxuzcxbk1xv6ZGxrteJxVH7KlX7YRdZ6eARKwLe4
		AFZEAwoKCQ==
		-----END CERTIFICATE-----
	Certificate chain (1):
		Serial Number: 65537
		Subject: CN=SEV-Milan,OU=Engineering,O=Advanced Micro Devices,L=Santa Clara,ST=CA,C=US
		Issuer: CN=ARK-Milan,OU=Engineering,O=Advanced Micro Devices,L=Santa Clara,ST=CA,C=US
		Not Before: 2020-10-22 18:24:20 +0000 UTC
		Not After: 2045-10-22 18:24:20 +0000 UTC
		Signature Algorithm: SHA384-RSAPSS
		Public Key Algorithm: RSA
	Certificate chain (2):
		Serial Number: 65536
		Subject: CN=ARK-Milan,OU=Engineering,O=Advanced Micro Devices,L=Santa Clara,ST=CA,C=US
		Issuer: CN=ARK-Milan,OU=Engineering,O=Advanced Micro Devices,L=Santa Clara,ST=CA,C=US
		Not Before: 2020-10-22 17:23:05 +0000 UTC
		Not After: 2045-10-22 17:23:05 +0000 UTC
		Signature Algorithm: SHA384-RSAPSS
		Public Key Algorithm: RSA
	SNP Report:
		Version: 2
		Guest SVN: 5
		Policy:
			ABI Minor: 31
			ABI Major: 0
			Symmetric Multithreading enabled: true
			Migration agent enabled: false
			Debugging enabled (host decryption of VM): false
			Single socket enabled: false
		Family ID: 01000000000000000000000000000000
		Image ID: 02000000000000000000000000000000
		VMPL: 0
		Signature Algorithm: 1
		Current TCB:
			Secure Processor bootloader SVN: 3
			Secure Processor operating system SVN: 0
			SVN 4 (reserved): 0
			SVN 5 (reserved): 0
			SVN 6 (reserved): 0
			SVN 7 (reserved): 0
			SEV-SNP firmware SVN: 8
			Microcode SVN: 210
		Platform Info:
			Symmetric Multithreading enabled (SMT): true
			Transparent secure memory encryption (TSME): false
		Author Key ID: 0
		Report Data: 2ba92bbf6dcdd960f3ad61c95da1211b38c64ee04632081816d1ef196345ecc60000000000000000000000000000000000000000000000000000000000000000
		Measurement: 6b657ba205dbf5cc63890cf73dee24c7338027e0ef5695be27289e27cc1a00e5081b92d63241049e6d55e795f43390ed
		Host Data: 0000000000000000000000000000000000000000000000000000000000000000
		ID Key Digest: 0356215882a825279a85b300b0b742931d113bf7e32dde2e50ffde7ec743ca491ecdd7f336dc28a6e0b2bb57af7a44a3
		Author Key Digest: 000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
		Report ID: a171633fa9fb8ab25f847321a742667114f4dee44693ad974e9a4eb97c339aa9
		Report ID MA: ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
		Reported TCB:
			Secure Processor bootloader SVN: 3
			Secure Processor operating system SVN: 0
			SVN 4 (reserved): 0
			SVN 5 (reserved): 0
			SVN 6 (reserved): 0
			SVN 7 (reserved): 0
			SEV-SNP firmware SVN: 8
			Microcode SVN: 115
		Chip ID: 91781f3acde42056b37e8930b0644bd0659f135c717bf877803f2251133d397e358b27388e7e26c79698bec13b6163d75e20744b448dac48a56870d81ab69619
		Committed TCB:
			Secure Processor bootloader SVN: 3
			Secure Processor operating system SVN: 0
			SVN 4 (reserved): 0
			SVN 5 (reserved): 0
			SVN 6 (reserved): 0
			SVN 7 (reserved): 0
			SEV-SNP firmware SVN: 8
			Microcode SVN: 115
		Current Build: 4
		Current Minor: 52
		Current Major: 1
		Committed Build: 4
		Committed Minor: 52
		Committed Major: 1
		Launch TCB:
			Secure Processor bootloader SVN: 3
			Secure Processor operating system SVN: 0
			SVN 4 (reserved): 0
			SVN 5 (reserved): 0
			SVN 6 (reserved): 0
			SVN 7 (reserved): 0
			SEV-SNP firmware SVN: 8
			Microcode SVN: 115
		Signature (DER):
			3066023100d6606c3fc55b649dc5700c63ed4298aedf2a201a47530fcd31f56c51dfa8c198d6a9b29666164130f36f2fe61f5b9796023100a3a16f9b8c996d60699a32651b458489ba99135484809fdd43098db43e3bb8519ccbd6643c09defb2ddb40ee00d6d9a0
	Microsoft Azure Attestation Token:
	***
		  "iss": "https://constell63936039.neu.attest.azure.net",
		  "exp": 1695500488,
		  "nbf": 1695471688,
		  "iat": 1695471688,
		  "jti": "fc0795364955b8eebb37df87df99fb5efa30e6af3a8c4cbb36c739c74bcb02d0",
		  "x-ms-attestation-type": "azurevm",
		  "x-ms-azurevm-attestation-protocol-ver": "2.0",
		  "x-ms-azurevm-attested-pcrs": [
		    0,
		    1,
		    2,
		    3,
		    4,
		    5,
		    6,
		    7
		  ],
		  "x-ms-azurevm-dbvalidated": true,
		  "x-ms-azurevm-dbxvalidated": true,
		  "x-ms-azurevm-debuggersdisabled": true,
		  "x-ms-azurevm-default-securebootkeysvalidated": true,
		  "x-ms-azurevm-osbuild": "Edgeless",
		  "x-ms-azurevm-osdistro": "Edgeless",
		  "x-ms-azurevm-ostype": "Edgeless",
		  "x-ms-azurevm-signingdisabled": true,
		  "x-ms-azurevm-vmid": "7B3535EE-6AB1-4EF3-B9B2-756596B298C1",
		  "x-ms-isolation-tee": ***
		    "x-ms-attestation-type": "sevsnpvm",
		    "x-ms-compliance-status": "azure-compliant-cvm",
		    "x-ms-runtime": ***
		      "keys": [
		        ***
		          "e": "AQAB",
		          "key_ops": [
		            "encrypt"
		          ],
		          "kid": "HCLAkPub",
		          "kty": "RSA",
		          "n": "tR8FpAAA0sSw4oY5iIIieWLGvn1xEZlyp7r4zPeP3c7QU6--7QWVOhbYd55VzL4XVE2NKqJpB-BmorP0iJ7QWll0IqYm780xQByrijunDR81VjUkpEcwDB_GZIne418GSCWG2K6tUBpBzo0kwvD1rKuz0Gcq0VRqNMMSzirKqebtJorOj-tmwOc5oQAud0ccxAQOC6iZA6GxDx_fzdjM2GNBF36Erg_-oixowY6Wwvc04lwc8XUAZ_SULNzKCCMBqRiazGKz597OQy2kWD6vIlRKGR90e5B0LOiRMbNDTX_G1lN3rYHFRZ-nmGb3hP06PaqIOQP0AKuK3blKo0uSqQ"
		        ***
		      ],
		      "vm-configuration": ***
		        "console-enabled": true,
		        "current-time": 1695470602,
		        "tpm-enabled": true,
		        "vmUniqueId": "7B3535EE-6AB1-4EF3-B9B2-756596B298C1"
		      ***
		    ***,
		    "x-ms-sevsnpvm-authorkeydigest": "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		    "x-ms-sevsnpvm-bootloader-svn": 3,
		    "x-ms-sevsnpvm-familyId": "01000000000000000000000000000000",
		    "x-ms-sevsnpvm-guestsvn": 5,
		    "x-ms-sevsnpvm-hostdata": "0000000000000000000000000000000000000000000000000000000000000000",
		    "x-ms-sevsnpvm-idkeydigest": "0356215882a825279a85b300b0b742931d113bf7e32dde2e50ffde7ec743ca491ecdd7f336dc28a6e0b2bb57af7a44a3",
		    "x-ms-sevsnpvm-imageId": "02000000000000000000000000000000",
		    "x-ms-sevsnpvm-launchmeasurement": "6b657ba205dbf5cc63890cf73dee24c7338027e0ef5695be27289e27cc1a00e5081b92d63241049e6d55e795f43390ed",
		    "x-ms-sevsnpvm-microcode-svn": 115,
		    "x-ms-sevsnpvm-reportdata": "2ba92bbf6dcdd960f3ad61c95da1211b38c64ee04632081816d1ef196345ecc60000000000000000000000000000000000000000000000000000000000000000",
		    "x-ms-sevsnpvm-reportid": "a171633fa9fb8ab25f847321a742667114f4dee44693ad974e9a4eb97c339aa9",
		    "x-ms-sevsnpvm-smt-allowed": true,
		    "x-ms-sevsnpvm-snpfw-svn": 8
		  ***,
		  "x-ms-policy-hash": "0BBrIt4wxrrxlM-HageeJq4lcdlVk9PEf7Xwl3T8Cfk",
		  "x-ms-runtime": ***
		    "client-payload": ***
		      "nonce": "Ofennd7waXrcRijaEazzdOKE+9ila5pwpt5pNvQNnBc="
		    ***,
		    "keys": [
		      ***
		        "e": "AQAB",
		        "key_ops": [
		          "encrypt"
		        ],
		        "kid": "TpmEphemeralEncryptionKey",
		        "kty": "RSA",
		        "n": "1t0xmAAAgzNTT_u-Gz_mTVPjUBgQWlhoqcBxtgnUw7WxiJhhze8rRp2cp5Y1MpOgVnZ7yyMRJn7KLyy4OOiPAntWDSbS0q7KY7fsWuPQxW7D_lfaPlIXZf0x5eEep7VABTUQU2jIeZmu8nIRdHyZnOUprtXalwNfjDmADi_o29wi7FmDPWL6xhRzAzw6Y5pSdYlS59GX9_LVqXZRQ25Lp92x0i3kENzNwx_nrL7HigdSNLkWuji-_ptbwFSHnskjYD0DKap_xkSLaWN3XFCW4Jq3cbM0z4XcAEYvpLyhcNQVu_Cz_g98i8NlsBDgT8ioB38_TL8DuAFyjdUElCGotQ"
		      ***
		    ]
		  ***,
		  "x-ms-ver": "1.0"
		***
`

func TestParse(t *testing.T) {
	reader := strings.NewReader(completeReport)
	version, err := ParseSNPReport(reader)
	require.NoError(t, err)
	assert.Equal(t, version.Bootloader, uint8(3))
	assert.Equal(t, version.TEE, uint8(0))
	assert.Equal(t, version.SNP, uint8(8))
	assert.Equal(t, version.Microcode, uint8(115))
}
