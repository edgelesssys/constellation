/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestSSH(t *testing.T) {
	someSSHPubKey := "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBDA1yYg1PIJNjAGjyuv66r8AJtpfBDFLdp3u9lVwkgbVKv1AzcaeTF/NEw+nhNJOjuCZ61LTPj12LZ8Wy/oSm0A= motte@lolcatghost"
	someSSHPubKeyPath := "some-key.pub"
	someMasterSecret := `
	{
		"key": "MDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAK",
		"salt": "MDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAK"
	}
	`
	testCases := map[string]struct {
		fh           file.Handler
		pubKey       string
		masterSecret string
		wantErr      bool
	}{
		"everything exists": {
			fh:           file.NewHandler(afero.NewMemMapFs()),
			pubKey:       someSSHPubKey,
			masterSecret: someMasterSecret,
		},
		"no public key": {
			fh:           file.NewHandler(afero.NewMemMapFs()),
			masterSecret: someMasterSecret,
			wantErr:      true,
		},
		"no master secret": {
			fh:      file.NewHandler(afero.NewMemMapFs()),
			pubKey:  someSSHPubKey,
			wantErr: true,
		},
		"malformed public key": {
			fh:           file.NewHandler(afero.NewMemMapFs()),
			pubKey:       "asdf",
			masterSecret: someMasterSecret,
			wantErr:      true,
		},
		"malformed master secret": {
			fh:           file.NewHandler(afero.NewMemMapFs()),
			masterSecret: "asdf",
			pubKey:       someSSHPubKey,
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			if tc.pubKey != "" {
				require.NoError(tc.fh.Write(someSSHPubKeyPath, []byte(tc.pubKey)))
			}
			if tc.masterSecret != "" {
				require.NoError(tc.fh.Write(constants.MasterSecretFilename, []byte(tc.masterSecret)))
			}

			cmd := NewSSHCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(&bytes.Buffer{})

			err := writeCertificateForKey(cmd, someSSHPubKeyPath, tc.fh, logger.NewTest(t))
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				cert, err := tc.fh.Read("constellation_cert.pub")
				require.NoError(err)
				_, _, _, _, err = ssh.ParseAuthorizedKey(cert)
				require.NoError(err)
			}
		})
	}
}
