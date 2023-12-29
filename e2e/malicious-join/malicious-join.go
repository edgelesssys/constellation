/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// End-to-end test that issues various types of malicious join requests to a cluster.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/joinservice/joinproto"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	jsEndpoint := flag.String("js-endpoint", "", "Join service endpoint to use.")
	csp := flag.String("csp", "", "Cloud service provider to use.")
	attVariant := flag.String(
		"variant",
		"",
		fmt.Sprintf("Attestation variant to use. Set to \"default\" to use the default attestation variant for the CSP,"+
			"or one of: %s", variant.GetAvailableAttestationVariants()),
	)
	flag.Parse()
	log.With(
		slog.String("js-endpoint", *jsEndpoint),
		slog.String("csp", *csp),
		slog.String("variant", *attVariant),
	).Info("Running tests with flags")

	testCases := map[string]struct {
		fn      func(attVariant, csp, jsEndpoint string, log *slog.Logger) error
		wantErr bool
	}{
		"JoinFromUnattestedNode": {
			fn:      JoinFromUnattestedNode,
			wantErr: true,
		},
	}

	allPassed := true
	testOutput := &testOutput{
		TestCases: make(map[string]testCaseOutput),
	}
	for name, tc := range testCases {
		log.With(slog.String("testcase", name)).Info("Running testcase")

		err := tc.fn(*attVariant, *csp, *jsEndpoint, log)

		switch {
		case err == nil && tc.wantErr:
			log.With(slog.Any("error", err), slog.String("testcase", name)).Error("Test case failed: Expected error but got none")
			testOutput.TestCases[name] = testCaseOutput{
				Passed:  false,
				Message: "Expected error but got none",
			}
			allPassed = false
		case !tc.wantErr && err != nil:
			log.With(slog.Any("error", err), slog.String("testcase", name)).Error("Test case failed: Got unexpected error")
			testOutput.TestCases[name] = testCaseOutput{
				Passed:  false,
				Message: fmt.Sprintf("Got unexpected error: %s", err),
			}
			allPassed = false
		case tc.wantErr && err != nil:
			log.With(slog.String("testcase", name)).Info("Test case succeeded")
			testOutput.TestCases[name] = testCaseOutput{
				Passed:  true,
				Message: fmt.Sprintf("Got expected error: %s", err),
			}
		case !tc.wantErr && err == nil:
			log.With(slog.String("testcase", name)).Info("Test case succeeded")
			testOutput.TestCases[name] = testCaseOutput{
				Passed:  true,
				Message: "No error, as expected",
			}
		default:
			log.With(slog.String("testcase", name)).Error("invalid result")
      os.Exit(1)
		}
	}

	testOutput.AllPassed = allPassed
	log.With(slog.Any("result", testOutput)).Info("Test completed")
}

type testOutput struct {
	AllPassed bool                      `json:"allPassed"`
	TestCases map[string]testCaseOutput `json:"testCases"`
}

type testCaseOutput struct {
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

// JoinFromUnattestedNode simulates a join request from a Node that uses a stub issuer
// and thus cannot be attested correctly.
func JoinFromUnattestedNode(attVariant, csp, jsEndpoint string, log *slog.Logger) error {
	joiner, err := newMaliciousJoiner(attVariant, csp, jsEndpoint, log)
	if err != nil {
		return fmt.Errorf("creating malicious joiner: %w", err)
	}

	_, err = joiner.join(context.Background())
	if err != nil {
		return fmt.Errorf("joining cluster: %w", err)
	}
	return nil
}

// newMaliciousJoiner creates a new malicious joiner, i.e. a simulated node that issues
// an invalid join request.
func newMaliciousJoiner(attVariant, csp, endpoint string, log *slog.Logger) (*maliciousJoiner, error) {
	var attVariantOid variant.Variant
	var err error
	if strings.EqualFold(attVariant, "default") {
		attVariantOid = variant.GetDefaultAttestation(cloudprovider.FromString(csp))
	} else {
		attVariantOid, err = variant.FromString(attVariant)
		if err != nil {
			return nil, fmt.Errorf("parsing attestation variant: %w", err)
		}
	}

	issuer := newFakeIssuer(attVariantOid)

	return &maliciousJoiner{
		endpoint: endpoint,
		logger:   log,
		dialer:   dialer.New(issuer, nil, &net.Dialer{}),
	}, nil
}

// maliciousJoiner simulates a malicious node joining a cluster.
type maliciousJoiner struct {
	endpoint string
	logger   *slog.Logger
	dialer   *dialer.Dialer
}

// join issues a join request to the join service endpoint.
func (j *maliciousJoiner) join(ctx context.Context) (*joinproto.IssueJoinTicketResponse, error) {
	j.logger.Debug("Dialing join service endpoint %s", j.endpoint, "")
	conn, err := j.dialer.Dial(ctx, j.endpoint)
	if err != nil {
		return nil, fmt.Errorf("dialing join service endpoint: %w", err)
	}
	defer conn.Close()
	j.logger.Debug("Successfully dialed join service endpoint %s", j.endpoint, "")

	protoClient := joinproto.NewAPIClient(conn)

	j.logger.Debug("Issuing join ticket")
	req := &joinproto.IssueJoinTicketRequest{
		DiskUuid:           "",
		CertificateRequest: []byte{},
		IsControlPlane:     false,
	}
	res, err := protoClient.IssueJoinTicket(ctx, req)
	j.logger.Debug("Got join ticket response: %s", fmt.Sprintf("%+v", res), "")
	if err != nil {
		return nil, fmt.Errorf("issuing join ticket: %w", err)
	}

	return res, nil
}

// newFakeIssuer creates a new fake issuer for a given attestation variant.
func newFakeIssuer(oid variant.Getter) *fakeIssuer {
	return &fakeIssuer{oid}
}

// fakeIssuer simulates an issuer that issues a fake / invalid attestation document.
type fakeIssuer struct {
	variant.Getter
}

// Issue issues a fake attestation document.
func (i *fakeIssuer) Issue(_ context.Context, userData, nonce []byte) ([]byte, error) {
	return json.Marshal(fakeAttestationDoc{UserData: userData, Nonce: nonce})
}

// fakeAttestationDoc is a fake attestation document.
type fakeAttestationDoc struct {
	UserData []byte
	Nonce    []byte
}
