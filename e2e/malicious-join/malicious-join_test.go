//go:build e2e

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// End-to-end test that issues various types of malicious join requests to a cluster.
package maliciousjoin

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/joinservice/joinproto"
	"github.com/stretchr/testify/require"
)

// Flags are defined globally as `go test` implicitly calls flag.Parse() before executing a testcase.
// Thus defining and parsing flags inside a testcase would result in a panic.
// See https://groups.google.com/g/golang-nuts/c/P6EdEdgvDuc/m/5-Dg6bPxmvQJ.
var (
	jsEndpoint = flag.String("js-endpoint", "", "Join service endpoint to use.")
	csp        = flag.String("csp", "", "Cloud service provider to use.")
	attVariant = flag.String(
		"variant",
		"",
		fmt.Sprintf("Attestation variant to use. Set to \"default\" to use the default attestation variant for the CSP, or one of: %s", variant.GetAvailableAttestationVariants()),
	)
)

func TestMaliciousJoin(t *testing.T) {
	fmt.Println(formatFlags())

	require := require.New(t)

	joiner, err := newMaliciousJoiner(logger.NewTest(t), *jsEndpoint)
	require.NoError(err)

	_, err = joiner.join(context.Background())
	require.Error(err)
}

func formatFlags() string {
	var sb strings.Builder
	sb.WriteString("Using Flags:\n")
	flag.VisitAll(func(f *flag.Flag) {
		sb.WriteString(fmt.Sprintf("\t%s: %s\n", f.Name, f.Value.String()))
	})
	return sb.String()
}

func newMaliciousJoiner(log *logger.Logger, endpoint string) (*maliciousJoiner, error) {
	var attVariantOid variant.Variant
	var err error
	if strings.EqualFold(*attVariant, "default") {
		attVariantOid = variant.GetDefaultAttestation(cloudprovider.FromString(*csp))
	} else {
		attVariantOid, err = variant.FromString(*attVariant)
		if err != nil {
			return nil, fmt.Errorf("parsing attestation variant: %w", err)
		}
	}

	issuer := newMaliciousIssuer(attVariantOid)

	return &maliciousJoiner{
		endpoint: endpoint,
		logger:   log,
		dialer:   dialer.New(issuer, nil, &net.Dialer{}),
	}, nil
}

type maliciousJoiner struct {
	endpoint string
	logger   *logger.Logger
	dialer   *dialer.Dialer
}

func (j *maliciousJoiner) join(ctx context.Context) (*joinproto.IssueJoinTicketResponse, error) {
	j.logger.Debugf("Dialing join service endpoint %s", j.endpoint)
	conn, err := j.dialer.Dial(ctx, j.endpoint)
	if err != nil {
		return nil, fmt.Errorf("dialing join service endpoint: %w", err)
	}
	defer conn.Close()
	j.logger.Debugf("Successfully dialed join service endpoint %s", j.endpoint)

	protoClient := joinproto.NewAPIClient(conn)

	j.logger.Debugf("Issuing join ticket")
	req := &joinproto.IssueJoinTicketRequest{
		DiskUuid:           "",
		CertificateRequest: []byte{},
		IsControlPlane:     false,
	}
	res, err := protoClient.IssueJoinTicket(ctx, req)
	j.logger.Debugf("Got join ticket response: %+v", res)
	if err != nil {
		return nil, fmt.Errorf("issuing join ticket: %w", err)
	}

	return res, nil
}

func newMaliciousIssuer(oid variant.Getter) *maliciousIssuer {
	return &maliciousIssuer{oid}
}

type maliciousIssuer struct {
	variant.Getter
}

func (i *maliciousIssuer) Issue(_ context.Context, userData, nonce []byte) ([]byte, error) {
	return json.Marshal(fakeAttestationDoc{UserData: userData, Nonce: nonce})
}

type fakeAttestationDoc struct {
	UserData []byte
	Nonce    []byte
}
