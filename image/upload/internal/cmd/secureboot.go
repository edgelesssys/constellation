/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/internal/osimage/secureboot"
	"github.com/spf13/afero"
)

func loadSecureBootKeys(basePath string) (secureboot.Database, secureboot.UEFIVarStore, error) {
	platformKeyCert := filepath.Join(basePath, "PK.cer")
	keyExchangeKeyCerts := []string{
		filepath.Join(basePath, "KEK.cer"),
		filepath.Join(basePath, "MicCorKEKCA2011_2011-06-24.crt"),
	}
	signatureDBCerts := []string{
		filepath.Join(basePath, "db.cer"),
		filepath.Join(basePath, "MicWinProPCA2011_2011-10-19.crt"),
		filepath.Join(basePath, "MicCorUEFCA2011_2011-06-27.crt"),
	}
	sbDatabase, err := secureboot.DatabaseFromFiles(afero.NewOsFs(), platformKeyCert, keyExchangeKeyCerts, signatureDBCerts)
	if err != nil {
		return secureboot.Database{},
			secureboot.UEFIVarStore{},
			fmt.Errorf("preparing secure boot database: %w", err)
	}
	platformKeyESL := filepath.Join(basePath, "PK.esl")
	keyExchangeKeyESL := filepath.Join(basePath, "KEK.esl")
	signatureDBESL := filepath.Join(basePath, "db.esl")
	uefiVarStore, err := secureboot.VarStoreFromFiles(afero.NewOsFs(), platformKeyESL, keyExchangeKeyESL, signatureDBESL, "")
	if err != nil {
		return secureboot.Database{},
			secureboot.UEFIVarStore{},
			fmt.Errorf("preparing secure boot variable store: %w", err)
	}
	return sbDatabase, uefiVarStore, nil
}
