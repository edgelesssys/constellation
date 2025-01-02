/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"golang.org/x/crypto/hkdf"
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
		Long:  "Prepare your cluster for emergency ssh access and derive the necessary key.",
		Args:  cobra.ExactArgs(0),
		RunE:  runSSH,
	}
	cmd.Flags().String("key", "", "The path to an existing ssh private key.")
	cmd.MarkFlagRequired("key")
	return cmd
}

func runSSH(cmd *cobra.Command, _ []string) error {
	fh := file.NewHandler(afero.NewOsFs())
	var mastersecret secret
	err := fh.ReadJSON(fmt.Sprintf("%s.json", constants.ConstellationMasterSecretStoreName), &mastersecret)
	if err != nil {
		return err
	}

	hkdf := hkdf.New(sha256.New, mastersecret.Key, mastersecret.Salt, []byte("ssh-ca"))
	_, priv, err := ed25519.GenerateKey(hkdf)
	if err != nil {
		return err
	}

	ca, err := ssh.NewSignerFromSigner(priv)
	if err != nil {
		return err
	}

	// TODO(miampf): Remove or replace with better logger
	log.Printf("CA KEY: %s", string(ssh.MarshalAuthorizedKey(ca.PublicKey())))

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

	log.Printf("Signed certificate: %s", string(ssh.MarshalAuthorizedKey(&certificate)))

	return nil
}
