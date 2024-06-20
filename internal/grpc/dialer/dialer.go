/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package dialer provides a grpc dialer that can be used to create grpc client connections with different levels of ATLS encryption / verification.
package dialer

import (
	"context"
	"net"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dialer can open grpc client connections with different levels of ATLS encryption / verification.
type Dialer struct {
	issuer    atls.Issuer
	validator atls.Validator
	netDialer NetDialer
}

// New creates a new Dialer.
func New(issuer atls.Issuer, validator atls.Validator, netDialer NetDialer) *Dialer {
	return &Dialer{
		issuer:    issuer,
		validator: validator,
		netDialer: netDialer,
	}
}

// Dial creates a new grpc client connection to the given target using the atls validator.
func (d *Dialer) Dial(target string) (*grpc.ClientConn, error) {
	var validators []atls.Validator
	if d.validator != nil {
		validators = append(validators, d.validator)
	}
	credentials := atlscredentials.New(d.issuer, validators)

	return grpc.NewClient(target,
		d.grpcWithDialer(),
		grpc.WithTransportCredentials(credentials),
	)
}

// DialInsecure creates a new grpc client connection to the given target without using encryption or verification.
// Only use this method when using another kind of encryption / verification (VPN, etc).
func (d *Dialer) DialInsecure(target string) (*grpc.ClientConn, error) {
	return grpc.NewClient(target,
		d.grpcWithDialer(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

// DialNoVerify creates a new grpc client connection to the given target without verifying the server's attestation.
func (d *Dialer) DialNoVerify(target string) (*grpc.ClientConn, error) {
	credentials := atlscredentials.New(nil, nil)

	return grpc.NewClient(target,
		d.grpcWithDialer(),
		grpc.WithTransportCredentials(credentials),
	)
}

func (d *Dialer) grpcWithDialer() grpc.DialOption {
	return grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return d.netDialer.DialContext(ctx, "tcp", addr)
	})
}

// NetDialer implements the net Dialer interface.
type NetDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}
