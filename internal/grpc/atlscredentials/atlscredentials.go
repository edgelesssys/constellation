/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package atlscredentials

import (
	"context"
	"errors"
	"net"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"google.golang.org/grpc/credentials"
)

type Credentials struct {
	issuer     atls.Issuer
	validators []atls.Validator
}

func New(issuer atls.Issuer, validators []atls.Validator) *Credentials {
	return &Credentials{
		issuer:     issuer,
		validators: validators,
	}
}

func (c *Credentials) ClientHandshake(ctx context.Context, authority string, rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	clientCfg, err := atls.CreateAttestationClientTLSConfig(c.issuer, c.validators)
	if err != nil {
		return nil, nil, err
	}

	return credentials.NewTLS(clientCfg).ClientHandshake(ctx, authority, rawConn)
}

func (c *Credentials) ServerHandshake(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	serverCfg, err := atls.CreateAttestationServerTLSConfig(c.issuer, c.validators)
	if err != nil {
		return nil, nil, err
	}

	return credentials.NewTLS(serverCfg).ServerHandshake(rawConn)
}

func (c *Credentials) Info() credentials.ProtocolInfo {
	return credentials.NewTLS(nil).Info()
}

func (c *Credentials) Clone() credentials.TransportCredentials {
	cloned := *c
	return &cloned
}

func (c *Credentials) OverrideServerName(s string) error {
	return errors.New("cannot override server name")
}
