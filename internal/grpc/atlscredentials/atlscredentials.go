/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package atlscredentials handles creation of TLS credentials for attested TLS (ATLS).
package atlscredentials

import (
	"context"
	"errors"
	"net"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"google.golang.org/grpc/credentials"
)

// Credentials for attested TLS (ATLS).
type Credentials struct {
	issuer     atls.Issuer
	validators []atls.Validator
}

// New creates new ATLS Credentials.
func New(issuer atls.Issuer, validators []atls.Validator) *Credentials {
	return &Credentials{
		issuer:     issuer,
		validators: validators,
	}
}

// ClientHandshake performs the client handshake.
func (c *Credentials) ClientHandshake(ctx context.Context, authority string, rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	clientCfg, err := atls.CreateAttestationClientTLSConfig(c.issuer, c.validators)
	if err != nil {
		return nil, nil, err
	}

	return credentials.NewTLS(clientCfg).ClientHandshake(ctx, authority, rawConn)
}

// ServerHandshake performs the server handshake.
func (c *Credentials) ServerHandshake(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	serverCfg, err := atls.CreateAttestationServerTLSConfig(c.issuer, c.validators)
	if err != nil {
		return nil, nil, err
	}

	return credentials.NewTLS(serverCfg).ServerHandshake(rawConn)
}

// Info provides information about the protocol.
func (c *Credentials) Info() credentials.ProtocolInfo {
	return credentials.NewTLS(nil).Info()
}

// Clone the credentials object.
func (c *Credentials) Clone() credentials.TransportCredentials {
	cloned := *c
	return &cloned
}

// OverrideServerName is not supported and will fail.
func (c *Credentials) OverrideServerName(_ string) error {
	return errors.New("cannot override server name")
}
