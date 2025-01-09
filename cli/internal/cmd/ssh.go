/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kms/setup"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"golang.org/x/crypto/ssh"
)

var permissions = ssh.Permissions{
	Extensions: map[string]string{
		"permit-port-forwarding": "yes",
		"permit-pty":             "yes",
	},
}

// NewSSHCmd returns a new cobra.Command for the ssh command.
func NewSSHCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "Prepare your cluster for emergency ssh access",
		Long:  "Prepare your cluster for emergency ssh access and sign a given key pair for authorization.",
		Args:  cobra.ExactArgs(0),
		RunE:  runSSH,
	}
	cmd.Flags().String("key", "", "The path to an existing ssh public key.")
	must(cmd.MarkFlagRequired("key"))
	return cmd
}

func runSSH(cmd *cobra.Command, _ []string) error {
	fh := file.NewHandler(afero.NewOsFs())
	debugLogger, err := newDebugFileLogger(cmd, fh)
	if err != nil {
		return err
	}

	_, err = fh.Stat(constants.TerraformWorkingDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory %q does not exist", constants.TerraformWorkingDir)
	}
	if err != nil {
		return err
	}

	// NOTE(miampf): Since other KMS aren't fully implemented yet, this commands assumes that the cKMS is used and derives the key accordingly.
	var mastersecret uri.MasterSecret
	if err = fh.ReadJSON(fmt.Sprintf("%s.json", constants.ConstellationMasterSecretStoreName), &mastersecret); err != nil {
		return fmt.Errorf("reading master secret: %s", err)
	}

	mastersecretURI := uri.MasterSecret{Key: mastersecret.Key, Salt: mastersecret.Salt}
	kms, err := setup.KMS(cmd.Context(), uri.NoStoreURI, mastersecretURI.EncodeToURI())
	if err != nil {
		return fmt.Errorf("setting up KMS: %s", err)
	}
	key, err := kms.GetDEK(cmd.Context(), crypto.DEKPrefix+constants.SSHCAKeySuffix, 256)
	if err != nil {
		return fmt.Errorf("retrieving key from KMS: %s", err)
	}

	ca, err := crypto.GenerateEmergencySSHCAKey(key)
	if err != nil {
		return fmt.Errorf("generating ssh emergency CA key: %s", err)
	}

	debugLogger.Debug("SSH CA KEY generated", "public-key", string(ssh.MarshalAuthorizedKey(ca.PublicKey())))

	keyPath, err := cmd.Flags().GetString("key")
	if err != nil {
		return fmt.Errorf("retrieving path to public key from flags: %s", err)
	}

	keyBuffer, err := fh.Read(keyPath)
	if err != nil {
		return fmt.Errorf("reading public key %q: %s", keyPath, err)
	}

	pub, _, _, _, err := ssh.ParseAuthorizedKey(keyBuffer)
	if err != nil {
		return fmt.Errorf("parsing public key %q: %s", keyPath, err)
	}

	certificate := ssh.Certificate{
		Key:             pub,
		CertType:        ssh.UserCert,
		ValidAfter:      uint64(time.Now().Unix()),
		ValidBefore:     uint64(time.Now().Add(24 * time.Hour).Unix()),
		ValidPrincipals: []string{"root"},
		Permissions:     permissions,
	}
	if err := certificate.SignCert(rand.Reader, ca); err != nil {
		return fmt.Errorf("signing certificate: %s", err)
	}

	debugLogger.Debug("Signed certificate", "certificate", string(ssh.MarshalAuthorizedKey(&certificate)))
	if err := fh.Write(fmt.Sprintf("%s/ca_cert.pub", constants.TerraformWorkingDir), ssh.MarshalAuthorizedKey(&certificate), file.OptOverwrite); err != nil {
		return fmt.Errorf("writing certificate: %s", err)
	}
	fmt.Printf("You can now connect to a node using 'ssh -F %s/ssh_config -i <your private key> <node ip>'.\nYou can obtain the private node IP via the web UI of your CSP.\n", constants.TerraformWorkingDir)

	return nil
}
