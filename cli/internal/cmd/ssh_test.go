package cmd

import (
	"context"
	"fmt"
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

	newFsWithDirectory := func() file.Handler {
		require := require.New(t)
		fh := file.NewHandler(afero.NewMemMapFs())
		require.NoError(fh.MkdirAll(constants.TerraformWorkingDir))
		return fh
	}
	newFsNoDirectory := func() file.Handler {
		fh := file.NewHandler(afero.NewMemMapFs())
		return fh
	}

	testCases := map[string]struct {
		fh           file.Handler
		pubKey       string
		masterSecret string
		wantErr      bool
	}{
		"everything exists": {
			fh:           newFsWithDirectory(),
			pubKey:       someSSHPubKey,
			masterSecret: someMasterSecret,
		},
		"no public key": {
			fh:           newFsWithDirectory(),
			masterSecret: someMasterSecret,
			wantErr:      true,
		},
		"no master secret": {
			fh:      newFsWithDirectory(),
			pubKey:  someSSHPubKey,
			wantErr: true,
		},
		"malformed public key": {
			fh:           newFsWithDirectory(),
			pubKey:       "asdf",
			masterSecret: someMasterSecret,
			wantErr:      true,
		},
		"malformed master secret": {
			fh:           newFsWithDirectory(),
			masterSecret: "asdf",
			pubKey:       someSSHPubKey,
			wantErr:      true,
		},
		"directory does not exist": {
			fh:           newFsNoDirectory(),
			pubKey:       someSSHPubKey,
			masterSecret: someMasterSecret,
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

			err := generateKey(context.Background(), someSSHPubKeyPath, tc.fh, logger.NewTest(t))
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				cert, err := tc.fh.Read(fmt.Sprintf("%s/ca_cert.pub", constants.TerraformWorkingDir))
				require.NoError(err)
				_, _, _, _, err = ssh.ParseAuthorizedKey(cert)
				require.NoError(err)
			}
		})
	}
}
