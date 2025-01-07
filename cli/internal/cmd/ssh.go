/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
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

type secret struct {
	Key  []byte `json:"key,omitempty"`
	Salt []byte `json:"salt,omitempty"`
}

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
	cmd.MarkFlagRequired("key")
	return cmd
}

func runSSH(cmd *cobra.Command, _ []string) error {
	fh := file.NewHandler(afero.NewOsFs())
	debugLogger, err := newDebugFileLogger(cmd, fh)
	if err != nil {
		return err
	}

	var mastersecret secret
	if err = fh.ReadJSON(fmt.Sprintf("%s.json", constants.ConstellationMasterSecretStoreName), &mastersecret); err != nil {
		return err
	}

	mastersecret_uri := uri.MasterSecret{Key: mastersecret.Key, Salt: mastersecret.Salt}
	kms, err := setup.KMS(cmd.Context(), uri.NoStoreURI, mastersecret_uri.EncodeToURI())
	if err != nil {
		return err
	}
	key, err := kms.GetDEK(cmd.Context(), crypto.DEKPrefix, 256)
	if err != nil {
		return err
	}

	_, priv, err := ed25519.GenerateKey(bytes.NewReader(key))
	if err != nil {
		return err
	}

	ca, err := ssh.NewSignerFromSigner(priv)
	if err != nil {
		return err
	}

	debugLogger.Debug("CA KEY generated", "key", string(ssh.MarshalAuthorizedKey(ca.PublicKey())))

	key_path, err := cmd.Flags().GetString("key")
	if err != nil {
		return err
	}

	key_buf, err := fh.Read(key_path)
	if err != nil {
		return err
	}

	pub, _, _, _, err := ssh.ParseAuthorizedKey(key_buf)
	if err != nil {
		return err
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
		return err
	}

	debugLogger.Debug("Signed certificate", "certificate", string(ssh.MarshalAuthorizedKey(&certificate)))
	fh.Write(fmt.Sprintf("%s/ca_cert.pub", constants.TerraformWorkingDir), ssh.MarshalAuthorizedKey(&certificate), file.OptOverwrite, file.OptMkdirAll)
	fmt.Printf("You can now connect to a node using 'ssh -F %s/ssh_config -i <your private key> <node ip>'.\nYou can obtain the private node IP via the web UI of your CSP.\n", constants.TerraformWorkingDir)

	return nil
}
