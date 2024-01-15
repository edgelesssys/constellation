/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package verify

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/snp/testdata"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestParseCerts(t *testing.T) {
	validCertExpected := "\tRaw Some Cert:\n\t\t-----BEGIN CERTIFICATE-----\n\t\tMIIFTDCCAvugAwIBAgIBADBGBgkqhkiG9w0BAQowOaAPMA0GCWCGSAFlAwQCAgUA\n\t\toRwwGgYJKoZIhvcNAQEIMA0GCWCGSAFlAwQCAgUAogMCATCjAwIBATB7MRQwEgYD\n\t\tVQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDASBgNVBAcMC1NhbnRhIENs\n\t\tYXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5jZWQgTWljcm8gRGV2aWNl\n\t\tczESMBAGA1UEAwwJU0VWLU1pbGFuMB4XDTIzMDgzMDEyMTUyNFoXDTMwMDgzMDEy\n\t\tMTUyNFowejEUMBIGA1UECwwLRW5naW5lZXJpbmcxCzAJBgNVBAYTAlVTMRQwEgYD\n\t\tVQQHDAtTYW50YSBDbGFyYTELMAkGA1UECAwCQ0ExHzAdBgNVBAoMFkFkdmFuY2Vk\n\t\tIE1pY3JvIERldmljZXMxETAPBgNVBAMMCFNFVi1WQ0VLMHYwEAYHKoZIzj0CAQYF\n\t\tK4EEACIDYgAEhPX8Cl9uA7PxqNGzeqamJNYJLx/VFE/s3+8qOWtaztKNcn1PaAI4\n\t\tndE+yaVfMHsiA8CLTylumpWXcVBHPYV9kPEVrtozhvrrT5Oii9OpZPYHJ7/WPVmM\n\t\tJ3K8/Iz3AshTo4IBFjCCARIwEAYJKwYBBAGceAEBBAMCAQAwFwYJKwYBBAGceAEC\n\t\tBAoWCE1pbGFuLUIwMBEGCisGAQQBnHgBAwEEAwIBAjARBgorBgEEAZx4AQMCBAMC\n\t\tAQAwEQYKKwYBBAGceAEDBAQDAgEAMBEGCisGAQQBnHgBAwUEAwIBADARBgorBgEE\n\t\tAZx4AQMGBAMCAQAwEQYKKwYBBAGceAEDBwQDAgEAMBEGCisGAQQBnHgBAwMEAwIB\n\t\tBjARBgorBgEEAZx4AQMIBAMCAV0wTQYJKwYBBAGceAEEBECeRKrvAs/Kb926ymac\n\t\tbP0p4auNl+vJOYVxKKy7E7h0DfMUNtNOhuX4rgzf6zoOGF20beysF2zHfXYcIqG5\n\t\t3PJbMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAICBQChHDAaBgkqhkiG9w0B\n\t\tAQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBA4ICAQBoVGgDdFV9gWPHaEOBrHzd\n\t\tWVYyuuMBH340DDSXbCGlPR6rhgja0qALmkUPG50REQGvoPsikAskwqhzRG2XEDO2\n\t\tb6+fRPIq3DjEbz/8V89IiYiOZI/ycFACi3EEVECAWbzjXSfiOio1NfbniXP6tWzW\n\t\tD/8xpd/8N8166bHpgNgMl9pX4i0I9vaTl3qH+jBuSMZ5Q4heTHLB+v4V7q+H6SZo\n\t\t7htqpaI3keLEhQL/pCP72udMPAzU+/5W/x/t/LD6SbQcQQoHbWDU6kgTDuXabDxl\n\t\tA4JoEZfatr+/TO6jKQcGtqOLKT8JFGcigUlBi/TBVP+Xs8E4CWYGZZiTpYoLwNAu\n\t\tyuKOP9VVFViSCqPvzpNs2G+e0zXg2w3te7oMw/l0bD8iQCAS8rR0+r+8pZL4e010\n\t\tKLZ3yEfA0moXef66k5xyf4y37ZIP189wz6qJ+YXqOujDmeTomCU0SnZXlri6GhbF\n\t\t19rp2z5/lsZG+W27CRxvzTB3hk+ukZr35vCqNq4Rs+c7/hYcYzzyZ4ysATwdglNF\n\t\tWddfVw5Qunlu6Ngxr84ifz3HrnUx9bR5DzmFbztrb7IbkZhq7GjImwJULub1viyg\n\t\tYFa7X3p8b1WllienSEfvbadobbS9HeuLUrWyh0kZjQnz+0Q1UB1/zlzokeQmAYCf\n\t\t8H3kABPv6hqrFftRNbargQ==\n\t\t-----END CERTIFICATE-----\n\tSome Cert (1):\n\t\tSerial Number: 0\n\t\tSubject: CN=SEV-VCEK,OU=Engineering,O=Advanced Micro Devices,L=Santa Clara,ST=CA,C=US\n\t\tIssuer: CN=SEV-Milan,OU=Engineering,O=Advanced Micro Devices,L=Santa Clara,ST=CA,C=US\n\t\tNot Before: 2023-08-30 12:15:24 +0000 UTC\n\t\tNot After: 2030-08-30 12:15:24 +0000 UTC\n\t\tSignature Algorithm: SHA384-RSAPSS\n\t\tPublic Key Algorithm: ECDSA\n"

	testCases := map[string]struct {
		cert     []byte
		expected string
		wantErr  bool
	}{
		"one cert": {
			cert:     testdata.AzureThimVCEK,
			expected: validCertExpected,
		},
		"one cert with extra newlines": {
			cert:     []byte("\n\n" + string(testdata.AzureThimVCEK) + "\n\n"),
			expected: validCertExpected,
		},
		"invalid cert": {
			cert:    []byte("invalid"),
			wantErr: true,
		},
		"no cert": {
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			b := &strings.Builder{}

			certs, err := newCertificates("Some Cert", tc.cert, slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)))
			if err != nil {
				assert.True(tc.wantErr)
				return
			}

			err = formatCertificates(b, certs)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.expected, b.String())
			}
		})
	}
}
