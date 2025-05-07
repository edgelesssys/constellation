/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
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

// NewSSHCmd returns a new cobra.Command for the ssh command.
func NewSSHCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "Generate a certificate for emergency SSH access",
		Long:  "Generate a certificate for emergency SSH access to your SSH-enabled constellation cluster.",
		Args:  cobra.ExactArgs(0),
		RunE:  runSSH,
	}
	cmd.Flags().String("key", "", "the path to an existing SSH public key")
	must(cmd.MarkFlagRequired("key"))
	return cmd
}

func runSSH(cmd *cobra.Command, _ []string) error {
	fh := file.NewHandler(afero.NewOsFs())
	debugLogger, err := newDebugFileLogger(cmd, fh)
	if err != nil {
		return err
	}

	keyPath, err := cmd.Flags().GetString("key")
	if err != nil {
		return fmt.Errorf("retrieving path to public key from flags: %s", err)
	}

	return writeCertificateForKey(cmd, keyPath, fh, debugLogger)
}

func writeCertificateForKey(cmd *cobra.Command, keyPath string, fh file.Handler, debugLogger debugLog) error {
	// NOTE(miampf): Since other KMS aren't fully implemented yet, this commands assumes that the cKMS is used and derives the key accordingly.
	var mastersecret uri.MasterSecret
	if err := fh.ReadJSON(constants.MasterSecretFilename, &mastersecret); err != nil {
		return fmt.Errorf("reading master secret (does %q exist?): %w", constants.MasterSecretFilename, err)
	}

	mastersecretURI := uri.MasterSecret{Key: mastersecret.Key, Salt: mastersecret.Salt}
	kms, err := setup.KMS(cmd.Context(), uri.NoStoreURI, mastersecretURI.EncodeToURI())
	if err != nil {
		return fmt.Errorf("setting up KMS: %s", err)
	}
	sshCAKeySeed, err := kms.GetDEK(cmd.Context(), crypto.DEKPrefix+constants.SSHCAKeySuffix, ed25519.SeedSize)
	if err != nil {
		return fmt.Errorf("retrieving key from KMS: %s", err)
	}

	ca, err := crypto.GenerateEmergencySSHCAKey(sshCAKeySeed)
	if err != nil {
		return fmt.Errorf("generating SSH emergency CA key: %s", err)
	}

	marshalledKey := string(ssh.MarshalAuthorizedKey(ca.PublicKey()))
	debugLogger.Debug("SSH CA KEY generated", "public-key", marshalledKey)
	knownHostsContent := fmt.Sprintf("@cert-authority * %s", marshalledKey)
	if err := fh.Write("./known_hosts", []byte(knownHostsContent), file.OptMkdirAll); err != nil {
		return fmt.Errorf("writing known hosts file: %w", err)
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
		Permissions: ssh.Permissions{
			Extensions: map[string]string{
				"permit-port-forwarding": "",
				"permit-pty":             "",
			},
		},
	}
	if err := certificate.SignCert(rand.Reader, ca); err != nil {
		return fmt.Errorf("signing certificate: %s", err)
	}

	debugLogger.Debug("Signed certificate", "certificate", string(ssh.MarshalAuthorizedKey(&certificate)))
	if err := fh.Write("constellation_cert.pub", ssh.MarshalAuthorizedKey(&certificate), file.OptOverwrite); err != nil {
		return fmt.Errorf("writing certificate: %s", err)
	}
	cmd.Printf("You can now connect to a node using the \"constellation_cert.pub\" certificate.\nLook at the documentation for a how-to guide:\n\n\thttps://docs.edgeless.systems/constellation/workflows/troubleshooting#emergency-ssh-access\n")

	return nil
}
